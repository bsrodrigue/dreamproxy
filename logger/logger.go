package logger

import (
	"dreamproxy/format"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	ACCESS_LOG_FILE = "/var/log/dreamserver/access.log"
	ERROR_LOG_FILE  = "/var/log/dreamserver/error.log"
)

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

type Service string

const (
	DREAM_SERVER Service = "DREAM_SERVER"
	HTTP_PARSER  Service = "HTTP_PARSER"
)

func (service *Service) ToStr() string {
	return string(*service)
}

type RequestLog struct {
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

	Trace struct {
		CorrelationID     string `json:"correlation_id,omitempty"`
		UpstreamIP        string `json:"upstream_ip,omitempty"`
		UpstreamLatencyMS int64  `json:"upstream_latency_ms,omitempty"`
	} `json:"trace"`
}

func NewRequestLog(service Service, level LogLevel, event LogEvent, message string) RequestLog {
	req_log := RequestLog{
		Timestamp: time.Now().UTC().Format(format.DateTimeFormat),
		Level:     level.ToStr(),
		Service:   service.ToStr(),
		Event:     event.ToStr(),
		Message:   message,
	}

	req_log.Request.ID = uuid.New().String()

	return req_log
}

func (rl RequestLog) ToText() string {
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

func (rl RequestLog) ToJSON() string {
	data, _ := json.Marshal(rl)
	return string(data)
}
