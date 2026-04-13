package main

import (
	"bytes"
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

//go:embed dashboard.html
var dashboardHTML string

var httpDotHTML = template.Must(template.New("dashboard").Parse(dashboardHTML))

type CheckType int

const (
	TypeCritical CheckType = iota
	TypeDegraded
	TypeAlert
	TypeInfo
)

type CheckStatus int

const (
	StatusUnknown CheckStatus = iota
	StatusUp
	StatusDown
)

type HealthCheck struct {
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

type HealthResponse struct {
	Status    string        `json:"status"`
	Timestamp string        `json:"timestamp"`
	Uptime    string        `json:"uptime"`
	Checks    []HealthCheck `json:"checks"`
}

var (
	startTime  = time.Now()
	lastStatus string
	cfg        config
)

type config struct {
	DBURL            string
	RedisURL         string
	APIURL           string
	SMTPHost         string
	SMTPPort         string
	TelegramBotToken string
	BackupPath       string
	DiskThreshold    int
}

func loadConfig() config {
	return config{
		DBURL:            getEnv("DATABASE_URL", "postgres://localhost:5432/db"),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379"),
		APIURL:           getEnv("API_URL", "http://localhost:8080"),
		SMTPHost:         getEnv("SMTP_HOST", "localhost"),
		SMTPPort:         getEnv("SMTP_PORT", "587"),
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		BackupPath:       getEnv("BACKUP_PATH", "/backups"),
		DiskThreshold:    getEnvInt("DISK_THRESHOLD", 20),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}
	return defaultValue
}

func runChecks() []HealthCheck {
	checks := []HealthCheck{
		checkRedis(),
		checkPostgres(),
		checkSMTP(),
		checkTelegram(),
		checkDisk(),
		checkMemory(),
		checkBackup(),
		checkSSL(),
		checkFrontend(),
		checkAPI(),
	}
	return checks
}

func calculateStatus(checks []HealthCheck) string {
	hasCriticalDown := false
	hasNonCriticalDown := false

	for _, c := range checks {
		if c.Status == "down" {
			switch c.Type {
			case "critical":
				hasCriticalDown = true
			case "degraded", "alert":
				hasNonCriticalDown = true
			}
		}
	}

	if hasCriticalDown {
		return "unhealthy"
	}
	if hasNonCriticalDown {
		return "degraded"
	}
	return "healthy"
}

func checkRedis() HealthCheck {
	addr := strings.TrimPrefix(cfg.RedisURL, "redis://")
	if !strings.Contains(addr, ":") {
		addr += ":6379"
	}

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return HealthCheck{Name: "Redis", Type: "critical", Status: "down", Message: err.Error()}
	}
	defer func() { _ = conn.Close() }()

	if _, err := conn.Write([]byte("*1\r\n$4\r\nPING\r\n")); err != nil {
		return HealthCheck{Name: "Redis", Type: "critical", Status: "down", Message: err.Error()}
	}
	buf := make([]byte, 32)
	n, _ := conn.Read(buf)
	if n == 0 || !bytes.Contains(buf[:n], []byte("PONG")) {
		return HealthCheck{Name: "Redis", Type: "critical", Status: "down", Message: "No PONG response"}
	}

	if _, err := conn.Write([]byte("*2\r\n$6\r\nDBSIZE\r\n")); err != nil {
		return HealthCheck{Name: "Redis", Type: "critical", Status: "down", Message: err.Error()}
	}
	n, _ = conn.Read(buf)
	_ = n

	return HealthCheck{Name: "Redis", Type: "critical", Status: "up", Message: "PING + DBSIZE OK"}
}

func checkPostgres() HealthCheck {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		return HealthCheck{Name: "PostgreSQL", Type: "critical", Status: "down", Message: err.Error()}
	}
	defer func() { _ = db.Close() }()

	if err := db.PingContext(ctx); err != nil {
		return HealthCheck{Name: "PostgreSQL", Type: "critical", Status: "down", Message: err.Error()}
	}

	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&count)
	msg := fmt.Sprintf("%d tables", count)
	if err != nil {
		msg = "connected"
	}

	return HealthCheck{Name: "PostgreSQL", Type: "critical", Status: "up", Message: msg}
}

func checkSMTP() HealthCheck {
	addr := net.JoinHostPort(cfg.SMTPHost, cfg.SMTPPort)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return HealthCheck{Name: "SMTP", Type: "degraded", Status: "down", Message: err.Error()}
	}
	defer func() { _ = conn.Close() }()
	return HealthCheck{Name: "SMTP", Type: "degraded", Status: "up", Message: "TCP dial OK"}
}

func checkTelegram() HealthCheck {
	if cfg.TelegramBotToken == "" {
		return HealthCheck{Name: "Telegram Bot", Type: "degraded", Status: "down", Message: "No token configured"}
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", cfg.TelegramBotToken)
	resp, err := http.Get(url)
	if err != nil {
		return HealthCheck{Name: "Telegram Bot", Type: "degraded", Status: "down", Message: err.Error()}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return HealthCheck{Name: "Telegram Bot", Type: "degraded", Status: "down", Message: fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}

	return HealthCheck{Name: "Telegram Bot", Type: "degraded", Status: "up", Message: "Bot info retrieved"}
}

func checkDisk() HealthCheck {
	var statfs syscall.Statfs_t
	err := syscall.Statfs("/", &statfs)
	if err != nil {
		return HealthCheck{Name: "Disk", Type: "alert", Status: "unknown", Message: err.Error()}
	}

	available := int64(statfs.Bavail)
	total := int64(statfs.Blocks)
	usedPercent := 100 - int((float64(available)/float64(total))*100)
	threshold := cfg.DiskThreshold

	status := "up"
	if usedPercent > (100 - threshold) {
		status = "down"
	}

	return HealthCheck{
		Name:    "Disk",
		Type:    "alert",
		Status:  status,
		Message: fmt.Sprintf("%d%% used", usedPercent),
		Details: map[string]int64{"used_percent": int64(usedPercent), "total_gb": total / 1_000_000_000, "available_gb": available / 1_000_000_000},
	}
}

func checkMemory() HealthCheck {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	used := m.Alloc / 1024 / 1024
	total := m.Sys / 1024 / 1024

	return HealthCheck{
		Name:    "Memory",
		Type:    "info",
		Status:  "up",
		Message: fmt.Sprintf("%d MB / %d MB", used, total),
		Details: map[string]uint64{"alloc_mb": m.Alloc / 1024 / 1024, "sys_mb": m.Sys / 1024 / 1024, "total_alloc_mb": m.TotalAlloc / 1024 / 1024},
	}
}

func checkBackup() HealthCheck {
	path := cfg.BackupPath
	if path == "" {
		return HealthCheck{Name: "Backup", Type: "alert", Status: "down", Message: "No backup path configured"}
	}

	cmd := exec.Command("sh", "-c", fmt.Sprintf("ls -t %s/*.dump.gz 2>/dev/null | head -1", path))
	out, err := cmd.Output()
	if err != nil {
		return HealthCheck{Name: "Backup", Type: "alert", Status: "down", Message: "No backup found"}
	}

	filePath := strings.TrimSpace(string(out))
	if filePath == "" {
		return HealthCheck{Name: "Backup", Type: "alert", Status: "down", Message: "No .dump.gz files"}
	}

	cmd = exec.Command("sh", "-c", fmt.Sprintf("stat -c %%Y %s", filePath))
	out, err = cmd.Output()
	if err != nil {
		return HealthCheck{Name: "Backup", Type: "alert", Status: "down", Message: "Cannot read file time"}
	}

	unixTime, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return HealthCheck{Name: "Backup", Type: "alert", Status: "down", Message: "Cannot parse file time"}
	}

	age := time.Since(time.Unix(unixTime, 0)).Hours()
	status := "up"
	if age > 25 {
		status = "down"
	}

	return HealthCheck{Name: "Backup", Type: "alert", Status: status, Message: fmt.Sprintf("%.1fh old", age), Details: map[string]interface{}{"age_hours": age, "file": filePath}}
}

func checkSSL() HealthCheck {
	conn, err := net.DialTimeout("tcp", ":443", 5*time.Second)
	if err != nil {
		return HealthCheck{Name: "SSL", Type: "degraded", Status: "down", Message: err.Error()}
	}
	defer func() { _ = conn.Close() }()
	return HealthCheck{Name: "SSL", Type: "degraded", Status: "up", Message: "TCP dial :443 OK"}
}

func checkFrontend() HealthCheck {
	paths := []string{"../frontend/dist/index.html", "frontend/dist/index.html", "/app/frontend/dist/index.html"}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return HealthCheck{Name: "Frontend", Type: "degraded", Status: "up", Message: "dist/index.html exists"}
		}
	}
	return HealthCheck{Name: "Frontend", Type: "degraded", Status: "down", Message: "dist/index.html not found"}
}

func checkAPI() HealthCheck {
	resp, err := http.Get(cfg.APIURL + "/health")
	if err != nil {
		return HealthCheck{Name: "API", Type: "critical", Status: "down", Message: err.Error()}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 500 {
		return HealthCheck{Name: "API", Type: "critical", Status: "down", Message: fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}

	return HealthCheck{Name: "API", Type: "critical", Status: "up", Message: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func sendTelegramAlert(newStatus string) {
	if cfg.TelegramBotToken == "" {
		return
	}

	msg := fmt.Sprintf("🔴 Health Monitor Alert: Status changed to *%s*", newStatus)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?text=%s&chat_id=%s", cfg.TelegramBotToken, msg, os.Getenv("TELEGRAM_CHAT_ID"))
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	checks := runChecks()
	status := calculateStatus(checks)

	if r.URL.Query().Get("format") == "json" {
		w.Header().Set("Content-Type", "application/json")
		if status == "unhealthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_ = json.NewEncoder(w).Encode(HealthResponse{
			Status:    status,
			Timestamp: time.Now().Format(time.RFC3339),
			Uptime:    time.Since(startTime).String(),
			Checks:    checks,
		})
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if status == "unhealthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	data := struct {
		Status    string
		Timestamp string
		Uptime    string
		Checks    []HealthCheck
	}{
		Status:    status,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Uptime:    time.Since(startTime).String(),
		Checks:    checks,
	}

	_ = httpDotHTML.Execute(w, data)
}

func main() {
	cfg = loadConfig()

	go func() {
		for {
			time.Sleep(60 * time.Second)
			checks := runChecks()
			newStatus := calculateStatus(checks)
			if lastStatus != "" && lastStatus != newStatus && (newStatus == "degraded" || newStatus == "unhealthy") {
				sendTelegramAlert(newStatus)
			}
			lastStatus = newStatus
		}
	}()

	http.HandleFunc("/", healthHandler)
	fmt.Println("Health monitor started on :8081")
	_ = http.ListenAndServe(":8081", nil)
}
