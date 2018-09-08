package logrus

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
	"net/http"
	"google.golang.org/api/logging/v2"
	"golang.org/x/net/context"
)

// FluentdFormatter is similar to logrus.JSONFormatter but with log level that are recongnized
// by kubernetes fluentd.
type FluentdFormatter struct {
	TimestampFormat string
}

// Format the log entry. Implements logrus.Formatter.
func (f *FluentdFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Make the slice 4 longer due to field clashes and traceid
	data := make(logrus.Fields, len(entry.Data)+4)
	var httpReq *logging.HttpRequest
	var err error
	for k, v := range entry.Data {
		switch x := v.(type) {
		case string:
			data[k] = x
		case *http.Request:
			httpReq = &logging.HttpRequest{
				Referer:		x.Referer(),
				RemoteIp:		x.RemoteAddr,
				RequestMethod:	x.Method,
				RequestUrl:		x.URL.String(),
				UserAgent:		x.UserAgent(),
			}
			data[k] = httpReq
			// Extract the traceId from the request
			span := trace.FromContext(x.Context())
			data["trace"] = span.SpanContext().TraceID.String()
		case context.Context:
			span := trace.FromContext(x)
			data["trace"] = span.SpanContext().TraceID.String()
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/Sirupsen/logrus/issues/137
			data[k] = x.Error()
		default:
			data[k] = fmt.Sprintf("%v", v)
		}
	}
	prefixFieldClashes(data)

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.RFC3339Nano
	}

	data["time"] = entry.Time.Format(timestampFormat)
	data["message"] = entry.Message
	data["severity"] = entry.Level.String()

	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("fluentd_formatter: Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}

func prefixFieldClashes(data logrus.Fields) {
	if t, ok := data["time"]; ok {
		data["fields.time"] = t
	}

	if m, ok := data["msg"]; ok {
		data["fields.msg"] = m
	}

	if l, ok := data["level"]; ok {
		data["fields.level"] = l
	}
}

