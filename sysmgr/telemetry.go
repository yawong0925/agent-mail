// sysmgr/telemetry.go
package sysmgr

import (
	"encoding/json"
	"net/http"
)

// handleDashboardStats compiles system-wide metrics.
func (s *Server) handleDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats := make(map[string]interface{})

	var totalEmails int
	s.db.Emails.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM emails").Scan(&totalEmails)
	stats["total_emails"] = totalEmails

	var diskBytes int64
	s.db.Emails.QueryRowContext(r.Context(), "SELECT COALESCE(SUM(file_size), 0) FROM email_attachments").Scan(&diskBytes)
	stats["disk_usage_mb"] = float64(diskBytes) / 1024 / 1024

	// Eventually, we will query the bandwidth history tables here for the 24h / 1w / 1m graphs.
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}