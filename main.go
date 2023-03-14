package main

import (
	// "context"
	"crypto/tls"
	"net/http"
	"strings"

	// "net/http"

	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	opensearch "github.com/opensearch-project/opensearch-go/v2"
	opensearchapi "github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

const IndexName = "sipfront-go-test"

// Initialization of the logrus logger
func SetUp(l *logrus.Logger) {
	// Log as JSON instead of the default ASCII formatter.
	l.Formatter = new(logrus.JSONFormatter)
	l.Formatter = new(logrus.TextFormatter)
	l.Formatter.(*logrus.TextFormatter).DisableColors = true

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	l.Out = os.Stdout

	// Only log the warning severity or above.
	l.Level = logrus.TraceLevel
}

func main() {
	// Set Up Logger ------------------------------------------------------------------------------
	var log *logrus.Logger = logrus.New()
	SetUp(log)

	log.WithFields(logrus.Fields{
		"dummy-field-1": "fizz",
		"dummy-field-2": "buzz",
		"dummy-field-3": "fizzbuzz",
	}).Info("this-is-a-test")

	// Initialize the client with SSL/TLS enabled. ------------------------------------------------
	var clientConfiguration opensearch.Config = opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
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

	// Make a new index ---------------------------------------------------------------------------
	settings := strings.NewReader(`{
		'settings': {
			'index': {
				'number_of_shards': 1,
				'number_of_replicas': 0
				}
			}
		}`)

	result := opensearchapi.IndicesCreateRequest{
		Index: IndexName,
		Body:  settings,
	}

}
