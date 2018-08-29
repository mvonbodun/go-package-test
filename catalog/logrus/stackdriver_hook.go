package logrus

import (
	"github.com/sirupsen/logrus"
	"cloud.google.com/go/errorreporting"
	"golang.org/x/net/context"
	"log"
	"errors"
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
func New(errServiceName string) (*StackdriverHook, error) {
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
	sh.projectID = "demogeauxcommerce"
	// Create a context
	ctx := context.Background()

	// Create the client to the error reporting service
	errorClient, err := errorreporting.NewClient(ctx, sh.projectID, errorreporting.Config{
		ServiceName: sh.errorReportingServiceName,
		OnError: func(err error) {
			logrus.Printf("Could not log error: %v", err)
		},
	})
	if err != nil {
		log.Fatal(err)
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

// Fire writes teh message to teh Stackdriver entry service.
func (sh *StackdriverHook) Fire(entry *logrus.Entry) error {
	if isError(entry) {
		sh.errorClient.Report(errorreporting.Entry{
			Error: errors.New(entry.Message),
		})
	}
	return nil
}


