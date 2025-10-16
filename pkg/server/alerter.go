package server

import (
	"fmt"
	"net/smtp"
	"sync"
	"time"

	"github.com/jyxjjj/Monitor/pkg/models"
)

// Alerter handles alert checking and notifications
type Alerter struct {
	db           *Database
	config       *models.Config
	alertStates  map[int]map[string]time.Time // rule_id -> agent_id -> first_trigger_time
	mu           sync.RWMutex
}

// NewAlerter creates a new alerter
func NewAlerter(db *Database, config *models.Config) *Alerter {
	return &Alerter{
		db:          db,
		config:      config,
		alertStates: make(map[int]map[string]time.Time),
	}
}

// CheckMetrics checks metrics against alert rules
func (a *Alerter) CheckMetrics(metrics *models.Metrics) error {
	rules, err := a.db.GetAlertRules()
	if err != nil {
		return err
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// Only check rules for this agent or global rules
		if rule.AgentID != "" && rule.AgentID != metrics.AgentID {
			continue
		}

		value := a.getMetricValue(metrics, rule.MetricType)
		if a.checkThreshold(value, rule.Threshold, rule.Operator) {
			a.handleAlertTrigger(rule, metrics, value)
		} else {
			a.handleAlertClear(rule, metrics.AgentID)
		}
	}

	return nil
}

// getMetricValue extracts the metric value based on type
func (a *Alerter) getMetricValue(metrics *models.Metrics, metricType string) float64 {
	switch metricType {
	case "cpu":
		return metrics.CPUPercent
	case "memory":
		if metrics.MemoryTotal > 0 {
			return float64(metrics.MemoryUsed) / float64(metrics.MemoryTotal) * 100
		}
	case "disk":
		if metrics.DiskTotal > 0 {
			return float64(metrics.DiskUsed) / float64(metrics.DiskTotal) * 100
		}
	case "load":
		return metrics.LoadAvg1
	}
	return 0
}

// checkThreshold checks if value meets threshold condition
func (a *Alerter) checkThreshold(value, threshold float64, operator string) bool {
	switch operator {
	case "gt":
		return value > threshold
	case "lt":
		return value < threshold
	case "gte":
		return value >= threshold
	case "lte":
		return value <= threshold
	}
	return false
}

// handleAlertTrigger handles alert trigger logic
func (a *Alerter) handleAlertTrigger(rule *models.AlertRule, metrics *models.Metrics, value float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.alertStates[rule.ID] == nil {
		a.alertStates[rule.ID] = make(map[string]time.Time)
	}

	firstTrigger, exists := a.alertStates[rule.ID][metrics.AgentID]
	if !exists {
		a.alertStates[rule.ID][metrics.AgentID] = time.Now()
		return
	}

	// Check if alert has been triggered for the required duration
	if time.Since(firstTrigger) >= time.Duration(rule.Duration)*time.Second {
		alert := &models.Alert{
			RuleID:    rule.ID,
			AgentID:   metrics.AgentID,
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("%s: %.2f%s %s %.2f%s", rule.MetricType, value, a.getUnit(rule.MetricType), rule.Operator, rule.Threshold, a.getUnit(rule.MetricType)),
			Value:     value,
			Resolved:  false,
		}

		if err := a.db.SaveAlert(alert); err == nil {
			a.sendEmailNotification(alert, rule)
		}

		// Reset state after alert is sent
		delete(a.alertStates[rule.ID], metrics.AgentID)
	}
}

// handleAlertClear clears alert state when condition is no longer met
func (a *Alerter) handleAlertClear(rule *models.AlertRule, agentID string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.alertStates[rule.ID] != nil {
		delete(a.alertStates[rule.ID], agentID)
	}
}

// getUnit returns the unit for a metric type
func (a *Alerter) getUnit(metricType string) string {
	switch metricType {
	case "cpu", "memory", "disk":
		return "%"
	case "load":
		return ""
	}
	return ""
}

// sendEmailNotification sends email notification for an alert
func (a *Alerter) sendEmailNotification(alert *models.Alert, rule *models.AlertRule) error {
	if a.config.SMTPHost == "" || a.config.AlertEmail == "" {
		return nil // Email not configured
	}

	subject := fmt.Sprintf("Alert: %s", rule.Description)
	body := fmt.Sprintf("Alert triggered at %s\n\nAgent: %s\nRule: %s\nMessage: %s\n",
		alert.Timestamp.Format(time.RFC3339),
		alert.AgentID,
		rule.Description,
		alert.Message,
	)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		a.config.EmailFrom, a.config.AlertEmail, subject, body)

	auth := smtp.PlainAuth("", a.config.SMTPUser, a.config.SMTPPassword, a.config.SMTPHost)
	addr := fmt.Sprintf("%s:%d", a.config.SMTPHost, a.config.SMTPPort)

	return smtp.SendMail(addr, auth, a.config.EmailFrom, []string{a.config.AlertEmail}, []byte(msg))
}
