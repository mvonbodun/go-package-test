package logrus

import (
	"github.com/sirupsen/logrus"
	"cloud.google.com/go/errorreporting"
	"golang.org/x/net/context"
	"log"
	"errors"
	"cloud.google.com/go/logging"
	"github.com/facebookgo/stack"
	logging2 "google.golang.org/genproto/googleapis/logging/v2"
	"net/http"
)

type StackdriverHook struct {
	// levels are the levels that logrus will log
	levels []logrus.Level

	// projectID is the projectID
	projectID string

	// loggingClient is the stackdriver logging reporting client.
	loggingClient *logging.Client

	// Standard logger
	logger *logging.Logger

	// logName is the name of the log written to.
	logName string

	// errorClient is the stackdriver error reporting client.
	errorClient *errorreporting.Client

	// errorReportingServiceName defines the value of the field <service>,
	// required for a valid error reporting payload. If this value is set,
	// messages where level/severity is higher than or equal to "error" will
	// be sent to Stackdriver error reporting.
	// See more at:
	// https://cloud.google.com/error-reporting/docs/formatting-error-messages
	errorReportingServiceName string

}

// New creates a StackdriverHook using the ServiceName for using with
// logrus to write to Google Stackdriver.
func New(projectID string, logName string, errServiceName string) (*StackdriverHook, error) {
	// Set the levels this hook applies for.  Only use for error and higher.
	sh := &StackdriverHook{
		levels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
			logrus.InfoLevel,
			logrus.DebugLevel,
		},
	}
	// Sets the logName
	sh.logName = logName
	// Sets the service name
	sh.errorReportingServiceName = errServiceName
	// Sets the project id
	sh.projectID = projectID
	// Create a context
	ctx := context.Background()

	// Create the client to the stackdriver logging service.
	loggingClient, err := logging.NewClient(ctx, sh.projectID)
	if err != nil {
		log.Fatalf("Failed to create stackdriver logging client: %v", err)
	}
	if err := loggingClient.Ping(ctx); err != nil {
		log.Printf("Error pinging logging service: %v", err)
	}
	// Create the logger
	sh.logger = loggingClient.Logger(sh.logName)

	// Create the client to the stackdriver error reporting service.
	errorClient, err := errorreporting.NewClient(ctx, sh.projectID, errorreporting.Config{
		ServiceName: sh.errorReportingServiceName,
		OnError: func(err error) {
			log.Printf("Could not log error: %v", err)
		},
	})
	if err != nil {
		log.Fatalf("Failed to create stackdriver error reporting client: %v ", err)
	}
	sh.errorClient = errorClient
	return sh, err
}

// Determines if an error should be logged
func isError(entry *logrus.Entry) bool {
	if entry != nil {
		switch entry.Level {
		case logrus.ErrorLevel:
			return true
		case logrus.FatalLevel:
			return true
		case logrus.PanicLevel:
			return true
		}
	}
	return false
}

// Levels returns the logrus levels that this hook is applied to.
// This can be set using the Levels Option.
func (sh *StackdriverHook) Levels() []logrus.Level {
	return sh.levels
}

// Fire writes the message to teh Stackdriver entry service.
func (sh *StackdriverHook) Fire(entry *logrus.Entry) error {
	if isError(entry) {
		log.Print("Inside stackdriver hook isError")

		// extract the stack trace from entry.Data
		st, _ := entry.Data["stackTrace"].([]byte)
		// extract the request data
		httpRequest, _ := entry.Data["httprequest"].(*http.Request)
		sh.errorClient.Report(errorreporting.Entry{
			Error: errors.New(entry.Message),
			Req: httpRequest,
			Stack: st,
		})
		log.Print("Finished sending error")
	} else {
		sh.sendLogMessage(entry)
	}
	return nil
}

func (sh *StackdriverHook) sendLogMessage(entry *logrus.Entry) {
	log.Print("Inside stackdriver_hook sendLogMessage")
	sh.logger.Log(logging.Entry{
		Severity: sh.translateLogrusLevel(entry.Level),
		Payload: entry.Message,
		SourceLocation: sh.extractCallerFields(entry),
	})
	log.Print("after stackdriver_hook sendLogMessage call")
	//for k, v := range entry.Data {
	//	log.Printf("key: %v, value: %v, type: %v", k, v, reflect.TypeOf(v))
	//}

}

// Translates the logrus level to the stackdriver logging level
func (sh *StackdriverHook) translateLogrusLevel(level logrus.Level) logging.Severity {
	var sdLevel logging.Severity
	switch level {
	case logrus.DebugLevel:
		sdLevel = logging.Debug
	case logrus.InfoLevel:
		sdLevel = logging.Info
	case logrus.WarnLevel:
		sdLevel = logging.Warning
	case logrus.ErrorLevel:
		sdLevel = logging.Error
	case logrus.FatalLevel:
		sdLevel = logging.Critical
	case logrus.PanicLevel:
		sdLevel = logging.Emergency
	}
	return sdLevel
}


// Extracts the caller field to populate the source code location
func (sh *StackdriverHook) extractCallerFields(entry *logrus.Entry) *logging2.LogEntrySourceLocation {
	var sl logging2.LogEntrySourceLocation
	// Extract the "caller" field from Data
	for _, v := range entry.Data {
		switch x := v.(type) {
		case stack.Frame:
			sl.File = x.File
			sl.Line = int64(x.Line)
			sl.Function = x.Name
		}
	}
	return &sl
}

// Close closes the logging and error reporting clients
func (sh *StackdriverHook) Close() {
	err := sh.loggingClient.Close()
	if err != nil {
		log.Fatalf("Failed to close stackdriver logging client: %v", err)
	}
	err = sh.errorClient.Close()
	if err != nil {
		log.Fatalf("Failed to close stackdriver error client: %v", err)
	}
}


