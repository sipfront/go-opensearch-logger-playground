package main

import (
	// "context"
	// "crypto/tls"
	// "strings"
	// "net/http"

	"fmt"
	"os"

	// opensearchapi "github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	opensearch "github.com/opensearch-project/opensearch-go/v2"
	log "github.com/sirupsen/logrus"
)

const IndexName = "sipfront-go-test"

// Initialization of the logrus logger -> set-up o
func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}

func main() {
	log.WithFields(log.Fields{
		"dummy-field-1": "fizz",
		"dummy-field-2": "buzz",
		"dummy-field-3": "fizzbuzz",
	}).Info("this-is-a-test")

	// Initialize the client with SSL/TLS enabled.
	var clientConfiguration opensearch.Config = opensearch.Config{
		Addresses: []string{
			"https://vpc-sipfront-os-iepreu6yviwjk5rnzetncw7dfm.eu-central-1.es.amazonaws.com"},
	}

	client, err := opensearch.NewClient(clientConfiguration)
	if err != nil {
		fmt.Println("cannot initialize", err)
		os.Exit(1)
	}

	// Print OpenSearch version information on console.
	fmt.Println(client.Info())
}
