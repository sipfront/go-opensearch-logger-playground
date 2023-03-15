package main

import (
	// "context"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	opensearchapi "github.com/opensearch-project/opensearch-go/opensearchapi"
	opensearch "github.com/opensearch-project/opensearch-go/v2"
)

const IndexName = "sipfront-go-test"

// Custom type which will later implement the Write method for logging directly to
// Opensearch, without the help of using logstash.
type OpenSearchWriter struct {
	Client *opensearch.Client
}

// Writer interface to log directly to opensearch. Based on [SO-Post]
//
// [SO-Post] https://bit.ly/3Tj0fqe
func (ow *OpenSearchWriter) Write(p []byte) (n int, err error) {

	document := strings.NewReader(string(p))
	req := opensearchapi.IndexRequest{
		Index:      IndexName,
		DocumentID: "1",
		Body:       document,
	}

	_, err = req.Do(context.Background(), ow.Client)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

// Initialization of the logrus logger
func SetUp(l *logrus.Logger) {
	// Log as JSON instead of the default ASCII formatter.
	l.Formatter = new(logrus.JSONFormatter)
	l.Formatter = new(logrus.TextFormatter)
	l.Formatter.(*logrus.TextFormatter).DisableColors = true

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	//
	// TODO -> change this line to somthing like http.Post(...)
	// l.Out = os.Stdout

	// Only log the warning severity or above.
	l.Level = logrus.InfoLevel
}

// The main entry point, who would have guessed, duh'?!
func main() {
	// Set Up Logger ----------------------------------------------------------------------------
	var l *logrus.Logger = logrus.New()
	SetUp(l)
	l.WithFields(logrus.Fields{
		"dummy-field-1": "fizz",
		"dummy-field-2": "buzz",
		"dummy-field-3": "fizzbuzz",
	})

	// Initialize the client with SSL/TLS enabled. ----------------------------------------------
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

	// Set up writer
	var session *OpenSearchWriter = &OpenSearchWriter{Client: client}
	l.SetOutput(session)
	l.Info("this-is-a-test")

	// Make a new index -------------------------------------------------------------------------
	// settings := strings.NewReader(`{
	// 	'settings': {
	// 		'index': {
	// 			'number_of_shards': 1,
	// 			'number_of_replicas': 0
	// 			}
	// 		}
	// 	}`)
	// result := opensearchapi.IndicesCreateRequest{
	// 	Index: IndexName,
	// 	Body:  settings,
	// }
}
