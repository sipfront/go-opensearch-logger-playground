package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"

	opensearchapi "github.com/opensearch-project/opensearch-go/opensearchapi"
	opensearch "github.com/opensearch-project/opensearch-go/v2"
	"github.com/sirupsen/logrus"
)

// ------------------------------------------------------------------------------------------------
// For test purposes
const IndexName = "sipfront-gotest-v2-2023.03.22"

// ------------------------------------------------------------------------------------------------
// Custom type which will later implement the Write method for logging directly to
// Opensearch, without the help of using logstash.
type OpenSearchWriter struct {
	Client *opensearch.Client
}

// Writer interface to log directly to opensearch. Based on [SO-Post]
// [SO-Post] https://bit.ly/3Tj0fqe
func (ow *OpenSearchWriter) Write(p []byte) (n int, err error) {
	var document *strings.Reader = strings.NewReader(string(p))
	req := opensearchapi.IndexRequest{
		Index: IndexName,
		Body:  document,
	}

	insertResponse, err := req.Do(context.Background(), ow.Client)
	if err != nil {
		return 0, err
	}
	fmt.Println(insertResponse)

	return len(p), nil
}

// TODO Write a custom formater, such that the log is ESC compliant, see:
// - https://github.com/elastic/ecs-logging-go-logrus/blob/main/formatter.go or
// - https://github.com/sirupsen/logrus/issues/719
func main() {
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

	// IGNORE: Noy needed for creating an index document!
	// Define index mapping.
	mapping := strings.NewReader(`{
		"mappings":{
			"properties":{
				"@timestamp":{
					"type":"date",
					"format":"yyyy-MM-dd HH:mm:ss||yyyy-MM-dd||epoch_millis"},
				"function_name":{"type":"text"}
			}
		}
	}`)
	// Create an index with non-default settings.
	res := opensearchapi.IndicesCreateRequest{
		Index: IndexName,
		Body:  mapping,
	}
	insertResponse, err := res.Do(context.Background(), client)
	if err != nil {
		fmt.Println("cannot create", err)
		os.Exit(1)
	}
	fmt.Println(insertResponse)

	// Set Up Logger ----------------------------------------------------------------------------
	var l *logrus.Logger = &logrus.Logger{
		Out:   &OpenSearchWriter{Client: client},
		Level: logrus.InfoLevel,
		Formatter: &OpensearchFormatter{
			// DataKey:           "labels",
			DisableHTMLEscape: true,
			PrettyPrint:       true,
		},
	}
	l.Info("Plonk")
}
