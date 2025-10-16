package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
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

	// Update agent status based on reporting history
	for _, agent := range agents {
		// Default fallback interval if we don't have enough history
		fallbackInterval := 2 * time.Minute

		// Get metrics from the last hour to estimate reporting interval
		since := time.Now().Add(-1 * time.Hour)
		metrics, err := s.db.GetMetricsHistory(agent.ID, since)
		if err != nil || len(metrics) < 2 {
			// Not enough data to estimate - fallback to previous simple rule
			if time.Since(time.Time(agent.LastSeen)) > fallbackInterval {
				agent.Status = "offline"
			} else {
				agent.Status = "online"
			}
			continue
		}

		// metrics are ordered DESC (newest first)
		var totalInterval time.Duration
		var count int64
		for i := 0; i < len(metrics)-1; i++ {
			// interval between consecutive reports
			delta := time.Time(metrics[i].Timestamp).Sub(time.Time(metrics[i+1].Timestamp))
			if delta > 0 {
				totalInterval += delta
				count++
			}
		}

		var avgInterval time.Duration
		if count > 0 {
			avgInterval = time.Duration(int64(totalInterval) / count)
			// if avgInterval is zero or unreasonable, fallback
			if avgInterval <= 0 {
				avgInterval = fallbackInterval
			}
		} else {
			avgInterval = fallbackInterval
		}

		// Use the latest known report time as the base
		latest := metrics[0].Timestamp

		// Predict next three expected report times (only third needed)
		next3 := time.Time(latest).Add(3 * avgInterval)

		// If now has passed the third expected report time, mark offline
		if time.Now().After(next3) {
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

	// Get time range from query params. Default to last 5 minutes if not provided.
	now := time.Now()
	var since time.Time
	if sinceParam := r.URL.Query().Get("since"); sinceParam != "" {
		// Try to parse with millisecond precision first, then without milliseconds.
		// All times expected in local "2006-01-02 15:04:05" (optional .000)
		if t, err := time.Parse("2006-01-02 15:04:05.000", sinceParam); err == nil {
			since = t
		} else if t, err := time.Parse("2006-01-02 15:04:05", sinceParam); err == nil {
			since = t
		} else {
			http.Error(w, "Invalid since parameter", http.StatusBadRequest)
			return
		}
	} else {
		since = now.Add(-5 * time.Minute)
	}

	// Fetch raw metrics from DB (limited by DB implementation)
	rawMetrics, err := s.db.GetMetricsHistory(agentID, since)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Determine aggregation bucket count based on requested duration
	duration := now.Sub(since)
	var buckets int
	switch {
	case duration <= 5*time.Minute:
		buckets = 60
	case duration <= 1*time.Hour:
		buckets = 120
	case duration <= 6*time.Hour:
		buckets = 240
	case duration <= 24*time.Hour:
		buckets = 480
	default:
		buckets = 500
	}

	// Note: previous behavior allowed clients to request a specific number of points
	// via ?points=. To make time windows more accurate we ignore client-requested
	// points and compute aggregation buckets strictly based on the requested
	// time duration. This keeps server-side time-based aggregation deterministic.

	// If rawMetrics is empty, return empty array
	if len(rawMetrics) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]*models.Metrics{})
		return
	}

	// rawMetrics returned by DB are ordered DESC (newest first) - reverse to ascending
	for i, j := 0, len(rawMetrics)-1; i < j; i, j = i+1, j-1 {
		rawMetrics[i], rawMetrics[j] = rawMetrics[j], rawMetrics[i]
	}

	// If raw count <= buckets, return rawMetrics as-is
	if len(rawMetrics) <= buckets {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rawMetrics)
		return
	}

	// Use LTTB to downsample while preserving shape (use CPUPercent as primary axis)
	sampled := lttbDownsample(rawMetrics, buckets)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sampled)
}

// lttbDownsampleWithAccessor implements Largest-Triangle-Three-Buckets downsampling.
// It selects 'threshold' points from the input while preserving shape using
// timestamps as the X axis and a provided value accessor for the Y axis.
// valueFunc allows callers to choose which metric field to use (e.g., CPUPercent,
// Memory usage percent, etc.).
func lttbDownsampleWithAccessor(data []*models.Metrics, threshold int, valueFunc func(*models.Metrics) float64) []*models.Metrics {
	n := len(data)
	if threshold >= n || threshold == 0 {
		return data
	}

	sampled := make([]*models.Metrics, 0, threshold)
	// Always include first point
	sampled = append(sampled, data[0])

	bucketSize := float64(n-2) / float64(threshold-2)
	a := 0 // index of previously selected

	for i := 0; i < threshold-2; i++ {
		// bucket range
		start := int(math.Floor(float64(i)*bucketSize)) + 1
		end := int(math.Floor(float64(i+1)*bucketSize)) + 1
		if end >= n-1 {
			end = n - 1
		}

		// calculate avg of next bucket
		avgX := 0.0
		avgY := 0.0
		avgCount := 0
		nextStart := end
		nextEnd := int(math.Floor(float64(i+2)*bucketSize)) + 1
		if nextEnd >= n {
			nextEnd = n
		}
		for j := nextStart; j < nextEnd; j++ {
			avgX += float64(time.Time(data[j].Timestamp).UnixNano())
			avgY += valueFunc(data[j])
			avgCount++
		}
		if avgCount > 0 {
			avgX /= float64(avgCount)
			avgY /= float64(avgCount)
		}

		// Find point in this bucket that maximizes triangle area
		maxArea := -1.0
		maxIdx := start
		ax := float64(time.Time(data[a].Timestamp).UnixNano())
		ay := valueFunc(data[a])
		for j := start; j < end; j++ {
			bx := float64(time.Time(data[j].Timestamp).UnixNano())
			by := valueFunc(data[j])
			// area of triangle A-B-C where C is avg
			area := math.Abs((ax-bx)*(avgY-ay)-(ax-avgX)*(by-ay)) / 2.0
			if area > maxArea {
				maxArea = area
				maxIdx = j
			}
		}

		sampled = append(sampled, data[maxIdx])
		a = maxIdx
	}

	// Always include last point
	sampled = append(sampled, data[n-1])

	return sampled
}

// Backward-compatible wrapper that preserves existing behavior using CPUPercent
func lttbDownsample(data []*models.Metrics, threshold int) []*models.Metrics {
	return lttbDownsampleWithAccessor(data, threshold, func(m *models.Metrics) float64 { return m.CPUPercent })
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
	// Serve static assets and support SPA fallback to index.html

	// Prevent accidental interception of API routes
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}

	fs := http.FileServer(http.Dir("./frontend/build"))

	// Try to open the requested file. If it doesn't exist, serve index.html
	// so client-side routing (SPA) works.
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	// Strip leading slash for opening with http.Dir
	tryPath := strings.TrimPrefix(path, "/")
	if tryPath == "" {
		tryPath = "index.html"
	}

	// Use http.Dir Open to check existence
	if f, err := http.Dir("./frontend/build").Open(tryPath); err == nil {
		// file exists, close and serve via file server
		f.Close()
		// Serve the file directly
		fs.ServeHTTP(w, r)
		return
	}

	// file not found - fallback to index.html for SPA
	indexFile, err := http.Dir("./frontend/build").Open("index.html")
	if err != nil {
		// If index.html missing, return a minimal HTML message
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "<html><body><h1>Monitor Server</h1><p>Frontend not built. Run 'npm run build' in the frontend directory.</p></body></html>")
		return
	}
	defer indexFile.Close()

	// Serve index.html
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeContent(w, r, "index.html", time.Now(), indexFile)
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
