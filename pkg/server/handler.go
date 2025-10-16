package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jyxjjj/Monitor/pkg/compress"
	"github.com/jyxjjj/Monitor/pkg/models"
)

var jwtSecret = []byte("change-this-secret-in-production")

// Server represents the monitoring server
type Server struct {
	db      *Database
	config  *models.Config
	alerter *Alerter
}

// NewServer creates a new server instance
func NewServer(config *models.Config) (*Server, error) {
	// Support backward compatibility with DBPath
	if config.Database.Driver == "" && config.DBPath != "" {
		config.Database.Driver = "sqlite3"
		config.Database.Database = config.DBPath
	}
	
	// Set defaults if not configured
	if config.Database.Driver == "" {
		config.Database.Driver = "sqlite3"
		config.Database.Database = "./monitor.db"
	}
	
	db, err := NewDatabase(config.Database)
	if err != nil {
		return nil, err
	}

	// Check if database is installed
	installed, err := db.CheckInstalled()
	if err != nil {
		return nil, err
	}
	config.Installed = installed

	alerter := NewAlerter(db, config)

	return &Server{
		db:      db,
		config:  config,
		alerter: alerter,
	}, nil
}

// Start starts the server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Installation endpoints (no auth required)
	mux.HandleFunc("/api/install/check", s.handleInstallCheck)
	mux.HandleFunc("/api/install/setup", s.handleInstallSetup)

	// API endpoints
	mux.HandleFunc("/api/login", s.handleLogin)
	mux.HandleFunc("/api/agents", s.withAuth(s.handleAgents))
	mux.HandleFunc("/api/metrics/", s.withAuth(s.handleMetrics))
	mux.HandleFunc("/api/metrics/report", s.handleMetricsReport)
	mux.HandleFunc("/api/alerts", s.withAuth(s.handleAlerts))
	mux.HandleFunc("/api/alert-rules", s.withAuth(s.handleAlertRules))
	mux.HandleFunc("/api/config", s.withAuth(s.handleConfig))

	// Static files
	mux.HandleFunc("/", s.handleStatic)

	fmt.Printf("Server starting on %s\n", s.config.ServerAddr)

	if s.config.TLSCertFile != "" && s.config.TLSKeyFile != "" {
		return http.ListenAndServeTLS(s.config.ServerAddr, s.config.TLSCertFile, s.config.TLSKeyFile, mux)
	}

	return http.ListenAndServe(s.config.ServerAddr, mux)
}

// handleLogin handles admin login
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Hash the password
	hash := sha256.Sum256([]byte(req.Password))
	hashStr := hex.EncodeToString(hash[:])

	// Compare with configured password (also hashed)
	configHash := sha256.Sum256([]byte(s.config.AdminPassword))
	configHashStr := hex.EncodeToString(configHash[:])

	if hashStr != configHashStr {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin": true,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// withAuth wraps handlers with authentication
func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

// handleAgents handles agent listing
func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agents, err := s.db.GetAgents()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update agent status based on last seen
	for _, agent := range agents {
		if time.Since(agent.LastSeen) > 2*time.Minute {
			agent.Status = "offline"
		} else {
			agent.Status = "online"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

// handleMetrics handles metrics retrieval
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract agent ID from URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	agentID := parts[3]

	// Get time range from query params
	since := time.Now().Add(-24 * time.Hour)
	if sinceParam := r.URL.Query().Get("since"); sinceParam != "" {
		if t, err := time.Parse(time.RFC3339, sinceParam); err == nil {
			since = t
		}
	}

	metrics, err := s.db.GetMetricsHistory(agentID, since)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// handleMetricsReport handles metrics reporting from agents
func (s *Server) handleMetricsReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Check if body is compressed
	if r.Header.Get("Content-Encoding") == "br" {
		body, err = compress.DecompressBrotli(body)
		if err != nil {
			http.Error(w, "Failed to decompress", http.StatusBadRequest)
			return
		}
	}

	var metrics models.Metrics
	if err := json.Unmarshal(body, &metrics); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Save metrics
	if err := s.db.SaveMetrics(&metrics); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update agent info
	agent := &models.Agent{
		ID:       metrics.AgentID,
		Name:     metrics.AgentID,
		Host:     r.RemoteAddr,
		LastSeen: time.Now(),
		Status:   "online",
		Platform: runtime.GOOS,
		Version:  "1.0.0",
	}
	s.db.UpdateAgent(agent)

	// Check alerts
	s.alerter.CheckMetrics(&metrics)

	w.WriteHeader(http.StatusOK)
}

// handleAlerts handles alert listing
func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	alerts, err := s.db.GetAlerts(100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

// handleAlertRules handles alert rules CRUD
func (s *Server) handleAlertRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rules, err := s.db.GetAlertRules()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rules)

	case http.MethodPost:
		var rule models.AlertRule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if err := s.db.SaveAlertRule(&rule); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rule)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleConfig handles configuration retrieval
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Return config without sensitive data
	config := map[string]interface{}{
		"server_addr": s.config.ServerAddr,
		"smtp_host":   s.config.SMTPHost,
		"smtp_port":   s.config.SMTPPort,
		"email_from":  s.config.EmailFrom,
		"alert_email": s.config.AlertEmail,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// handleStatic serves static files
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	// Serve frontend files from ./frontend/build directory
	// For now, serve a simple HTML page
	html := `<!DOCTYPE html>
<html>
<head>
	<title>Monitor</title>
	<meta charset="utf-8">
</head>
<body>
	<h1>Monitor Server</h1>
	<p>API is running. Frontend React app should be built and placed in ./frontend/build directory.</p>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// handleInstallCheck checks if the database is installed
func (s *Server) handleInstallCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"installed": s.config.Installed,
	})
}

// handleInstallSetup handles database installation
func (s *Server) handleInstallSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if already installed
	if s.config.Installed {
		http.Error(w, "Database already installed", http.StatusBadRequest)
		return
	}

	// Initialize schema
	if err := s.db.InitSchema(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to initialize schema: %v", err), http.StatusInternalServerError)
		return
	}

	s.config.Installed = true

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Database installed successfully",
	})
}

// Close closes the server resources
func (s *Server) Close() error {
	return s.db.Close()
}
