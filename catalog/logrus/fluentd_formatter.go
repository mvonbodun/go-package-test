package logrus

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/facebookgo/stack"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
	"golang.org/x/net/context"
	"google.golang.org/api/logging/v2"
	"net/http"
)

// FluentdFormatter is similar to logrus.JSONFormatter but with log level that are recongnized
// by kubernetes fluentd.
type FluentdFormatter struct {
	TimestampFormat string
	TracePrefix     string
}

const (
	// Define constants for the Google json fields read by fluentd
	// See: https://cloud.google.com/logging/docs/agent/configuration#special_fields_in_structured_payloads
	httpRequestField = "httpRequest"
	traceField       = "logging.googleapis.com/trace"
	spanIdField      = "logging.googleapis.com/spanId"
	sourceLocField   = "logging.googleapis.com/sourceLocation"
)

// Format the log entry. Implements logrus.Formatter.
func (f *FluentdFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Setup the trace prefix.  Will set the prefix to "default" if not running in GCP.
	if f.TracePrefix == "" {
		p, err := metadata.ProjectID()
		if err != nil {
			p = "default"
			log.Printf("metadata: error retrieving ProjectID: %v", err)
		}
		f.TracePrefix = "projects/" + p + "/traces/"
	}
	// Make the slice 5 longer due to field clashes, traceid, source location
	data := make(logrus.Fields, len(entry.Data)+5)
	var httpReq *logging.HttpRequest
	var sourceLoc *logging.LogEntrySourceLocation
	var err error
	for k, v := range entry.Data {
		switch x := v.(type) {
		case string:
			data[k] = x
		case *http.Request:
			httpReq = &logging.HttpRequest{
				Referer:       x.Referer(),
				RemoteIp:      x.RemoteAddr,
				RequestMethod: x.Method,
				RequestUrl:    x.URL.String(),
				UserAgent:     x.UserAgent(),
			}
			data[httpRequestField] = httpReq
			// Extract the traceId from the request
			span := trace.FromContext(x.Context())
			data[traceField] = f.TracePrefix + span.SpanContext().TraceID.String()
			data[spanIdField] = span.SpanContext().SpanID.String()
		case context.Context:
			span := trace.FromContext(x)
			data[traceField] = f.TracePrefix + span.SpanContext().TraceID.String()
			data[spanIdField] = span.SpanContext().SpanID.String()
		case stack.Frame:
			sourceLoc = &logging.LogEntrySourceLocation{
				File:     x.File,
				Line:     int64(x.Line),
				Function: x.Name,
			}
			data[sourceLocField] = sourceLoc
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
