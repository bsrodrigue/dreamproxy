package logger

import (
	"bufio"
	"dreamproxy/format"
	"encoding/json"
	"fmt"
	_log "log"
	"os"
	"time"

	"github.com/google/uuid"
)

const (
	ACCESS_LOG_FILE = "/var/log/dreamserver/access.log"
	ERROR_LOG_FILE  = "/var/log/dreamserver/error.log"
)

// ================================ LogLevel =======================================

type LogLevel string

const (
	INFO  LogLevel = "INFO"
	DEBUG LogLevel = "DEBUG"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

func (lvl *LogLevel) ToStr() string {
	return string(*lvl)
}

// ================================ LogEvent =======================================

type LogEvent string

const (
	BAD_REQUEST       LogEvent = "BAD_REQUEST"
	REQUEST           LogEvent = "REQUEST"
	REQ_READING_ERROR LogEvent = "REQ_READING_ERROR"
	REQ_PARSE_ERROR   LogEvent = "REQ_PARSE_ERROR"
)

func (event *LogEvent) ToStr() string {
	return string(*event)
}

// ================================ Service =======================================

type Service string

const (
	DREAM_SERVER Service = "DREAM_SERVER"
	HTTP_PARSER  Service = "HTTP_PARSER"
)

func (service *Service) ToStr() string {
	return string(*service)
}

// ================================ RequestLogEntry =======================================

type RequestLogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Service   string `json:"service"`
	Event     string `json:"event"`
	Message   string `json:"message"`

	Request struct {
		ID        string `json:"id"`
		Method    string `json:"method"`
		Host      string `json:"host"`
		Path      string `json:"path"`
		Query     string `json:"query,omitempty"`
		ClientIP  string `json:"client_ip"`
		UserAgent string `json:"user_agent,omitempty"`
	} `json:"request"`

	Response struct {
		StatusCode int   `json:"status_code"`
		BytesSent  int64 `json:"bytes_sent"`
		LatencyMS  int64 `json:"latency_ms"`
	} `json:"response"`
}

func NewRequestLogEntry(service Service, level LogLevel, event LogEvent, message string) RequestLogEntry {
	req_log := RequestLogEntry{
		Timestamp: time.Now().UTC().Format(format.DateTimeFormat),
		Service:   service.ToStr(),
		Level:     level.ToStr(),
		Event:     event.ToStr(),
		Message:   message,
	}

	req_log.Request.ID = uuid.New().String()

	return req_log
}

func (rl RequestLogEntry) ToText() string {
	return fmt.Sprintf(
		"[%s][%s][%s] %s -> \"%s %s%s\" %d %dB %dms: %s\n",
		rl.Timestamp,
		rl.Service,
		rl.Level,
		rl.Request.ClientIP,
		rl.Request.Method,
		rl.Request.Host,
		rl.Request.Path,
		rl.Response.StatusCode,
		rl.Response.BytesSent,
		rl.Response.LatencyMS,
		rl.Message,
	)
}

func (rl RequestLogEntry) ToJSON() string {
	data, _ := json.Marshal(rl)
	return string(data)
}

// ================================ LogEntry =======================================

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Service   string `json:"service"`
	Event     string `json:"event"`
	Message   string `json:"message"`
}

func NewLogEntry(service Service, level LogLevel, event LogEvent, message string) LogEntry {
	req_log := LogEntry{
		Timestamp: time.Now().UTC().Format(format.DateTimeFormat),
		Service:   service.ToStr(),
		Level:     level.ToStr(),
		Event:     event.ToStr(),
		Message:   message,
	}

	return req_log
}

func (rl LogEntry) ToText() string {
	return fmt.Sprintf(
		"[%s][%s][%s]: %s\n",
		rl.Timestamp,
		rl.Service,
		rl.Level,
		rl.Message,
	)
}

func (rl LogEntry) ToJSON() string {
	data, _ := json.Marshal(rl)
	return string(data)
}

func (rl LogEntry) Log() {
	var logpath string = ACCESS_LOG_FILE

	if rl.Level == string(ERROR) {
		logpath = ERROR_LOG_FILE
	}

	log_file, err := os.OpenFile(logpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		_log.Println(err)
		return
	}

	buf_writer := bufio.NewWriter(log_file)
	written_bytes, err := buf_writer.WriteString(rl.ToText())

	if err != nil {
		_log.Println(err, " ", written_bytes)
		return
	}

	buf_writer.Flush()
}

func (rl LogEntry) LogTo(logpath string) {
	log_file, err := os.OpenFile(logpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		_log.Println(err)
		return
	}

	buf_writer := bufio.NewWriter(log_file)
	written_bytes, err := buf_writer.WriteString(rl.ToText())

	if err != nil {
		_log.Println(err, " ", written_bytes)
		return
	}

	buf_writer.Flush()
}

// ================================ ResponseLog =======================================
