package main

import (
	"github.com/mvonbodun/go-package-test/catalog/http"
	"github.com/mvonbodun/go-package-test/catalog/mysql"
	log "github.com/sirupsen/logrus"
	"os"
	"github.com/Gurpartap/logrus-stack"
	"github.com/mvonbodun/go-package-test/catalog/logrus"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/trace"
	"go.opencensus.io/stats/view"
	"time"
)

func init() {
	// Initialize logrus logging hooks
	var err error
	// Initialize logrus standard logger.  This globally
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)

	// Add the stack hook
	log.AddHook(logrus_stack.StandardHook())

	var sdHook *logrus.StackdriverHook
	// Add the Stackdriver Error reporting hook
	sdHook, err = logrus.New("catalog-app")
	if err != nil {
		log.Error("unable to create hook for stackdriver error reporting.")
	}
	log.AddHook(sdHook)
	// Finished initializing logrus.
	log.Info("Finished initializing logrus.")

	// Setup Stackdriver trace exporter
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: "demogeauxcommerce",
	})
	if err != nil {
		log.Warningf("Unable to create stackdriver exporter: %v", err)
	}
	trace.RegisterExporter(exporter)
	view.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	view.SetReportingPeriod(1 * time.Second)

}

func main() {

	// Connect to the database
	client := mysql.NewClient()
	log.Info("Created new MySql client")
	err := client.Open()
	if err != nil {
		panic(err)
	}
	// Close the Database connection when the program exits
	defer client.Close()

	// Create the http Handler
	h := http.NewHandler()
	h.ProductService = client.ProductService()
	h.Handler = h
	//h.ErrorClient = errorClient

	// Register the handlers and Start the web server
	h.ListenAndServe()

	// TODO: Need to close the error reporting client when program exits
	// See the logrus hook extension
}
