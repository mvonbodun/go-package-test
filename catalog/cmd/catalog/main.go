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
	"cloud.google.com/go/profiler"
	"golang.org/x/oauth2/google"
	"golang.org/x/net/context"
)

const (
	production = "PRODUCTION"
)

func main() {
	// Get the environment the program is executing in.
	// PRODUCTION, PERFORMANCE represent environment with reduced tracing and logging.
	environment := envString("ENVIRONMENT", "DEVELOPMENT")
	useStackdriver := envString("USE_STACKDRIVER", "FALSE")

	// Initialize logrus logging hooks
	var err error
	// Initialize logrus standard logger.  This globally
	// Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})
	// Log for fluentd formatter for Kubernetes or Google Cloud
	log.SetFormatter(&logrus.FluentdFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

	// Add the logrus stack hook to generate the source location when the log is written.
	log.AddHook(logrus_stack.StandardHook())
	log.Info("Finished initializing logrus.")

	// If useStackdriver is set to "TRUE", enable the various stackdriver components
	if useStackdriver == "TRUE" {
		ctx := context.Background()
		// Get the Application Default Credentials
		creds, err := google.FindDefaultCredentials(ctx, defaultAuthScopes()...)
		if err != nil {
			log.Fatalf("stackdriver - failed to get default credentials: %v", err)
		}
		if creds.ProjectID == "" {
			log.Fatal("stackdriver: no project found with application default credentials")
		}

		// Profiler initialization, best done as early as possible.
		if err := profiler.Start(profiler.Config{
			Service:        "catalogservice",
			ServiceVersion: "1.0.0",
			// ProjectID must be set if not running on GCP.
			ProjectID: creds.ProjectID,
		}); err != nil {
			log.Warningf("Error initializing profiler: %v", err)
		}

		// Setup Stackdriver trace exporter
		exporter, err := stackdriver.NewExporter(stackdriver.Options{
			ProjectID: creds.ProjectID,
		})
		if err != nil {
			log.Warningf("Unable to create stackdriver exporter: %v", err)
		}
		trace.RegisterExporter(exporter)
		view.RegisterExporter(exporter)
		// Only set trace to Always sample if non-production environment,
		// otherwise open census samples on a much less frequent basis
		if environment != production {
			trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
		}
		view.SetReportingPeriod(1 * time.Second)

		// Add the Stackdriver Error reporting logrus hook
		sdHook, err := logrus.New(creds.ProjectID, "catalog-err")
		if err != nil {
			log.Errorf("unable to create hook for stackdriver error reporting: %v", err)
		}
		log.AddHook(sdHook)
	}

	// Connect to the database
	client := mysql.NewClient()
	log.Info("Created new MySql client")
	err = client.Open()
	if err != nil {
		log.Fatalf("Failed to open MySql client: %v", err)
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
}

// DefaultAuthScopes reports the default set of authentication scopes to use with this application.
func defaultAuthScopes() []string {
	return []string{
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/trace.append",
	}
}

// envString retrieves an environment variable from the os, or uses the fallack if not set.
func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}
