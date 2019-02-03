package logrus

import (
	"github.com/sirupsen/logrus"
	"cloud.google.com/go/errorreporting"
	"golang.org/x/net/context"
	"log"
	"errors"
	"net/http"
	"fmt"
)

type StackdriverHook struct {
	// levels are the levels that logrus will log
	levels []logrus.Level

	// projectID is the projectID
	projectID string

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
func New(projectID string, errServiceName string) (*StackdriverHook, error) {
	// Set the levels this hook applies for.  Only use for error and higher.
	sh := &StackdriverHook{
		levels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
		},
	}
	// Sets the service name
	sh.errorReportingServiceName = errServiceName
	// Sets the project id
	sh.projectID = projectID
	// Create a context
	ctx := context.Background()

	// Create the client to the stackdriver error reporting service.
	errorClient, err := errorreporting.NewClient(ctx, sh.projectID, errorreporting.Config{
		ServiceName: sh.errorReportingServiceName,
		OnError: func(err error) {
			log.Printf("logrus stackdriver_hook: Could not log error: %v", err)
		},
	})
	if err != nil {
		log.Fatalf("logrus stackdriver_hook: Failed to create stackdriver error reporting client: %v ", err)
	} else {
		sh.errorClient = errorClient
	}
	sh.errorClient = errorClient
	return sh, err
}

// Levels returns the logrus levels that this hook is applied to.
// This can be set using the Levels Option.
func (sh *StackdriverHook) Levels() []logrus.Level {
	return sh.levels
}

// Fire writes the message to teh Stackdriver entry service.
func (sh *StackdriverHook) Fire(entry *logrus.Entry) error {
	// extract the stack trace from entry.Data
	st, _ := entry.Data["stackTrace"].([]byte)
	log.Print("stackdriver_hook: just finished getting stackTrace")
	// extract the request data
	httpRequest, _ := entry.Data["httpRequest"].(*http.Request)
	sh.errorClient.Report(errorreporting.Entry{
		Error: errors.New(entry.Message),
		Req: httpRequest,
		Stack: st,
	})
	// Remove the binary stacktrace from the logrus.Entry so it is not
	// written out to by the fluentd formatter
	delete(entry.Data, "stackTrace")
	return nil
}

// Close closes the logging and error reporting clients
func (sh *StackdriverHook) Close() error {
	err := sh.errorClient.Close()
	if err != nil {
		return fmt.Errorf("failed to close stackdriver error client: %v", err)
	}
	return nil
}


