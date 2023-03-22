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

var replacer = strings.NewReplacer("\n", "", " ", "")

// ------------------------------------------------------------------------------------------------
// Custom type which will later implement the Write method for logging directly to
// Opensearch, without the help of using logstash.
type OpenSearchWriter struct {
	Client *opensearch.Client
}

// Writer interface to log directly to opensearch. Based on [SO-Post]
// [SO-Post] https://bit.ly/3Tj0fqe
func (ow *OpenSearchWriter) Write(p []byte) (n int, err error) {
	// https://stackoverflow.com/questions/45792309/bulk-api-malformed-action-metadata-line-3-expected-start-object-but-found
	mapping := `{"mappings":{"properties":{"@timestamp":{"type":"date","format":"yyyy-MM-dd HH:mm:ss||yyyy-MM-dd||epoch_millis"},"function_name":{"type":"text"}}}}`

	fmt.Printf("%s\n%s\n", mapping, replacer.Replace(string(p)))
	var document *strings.Reader = strings.NewReader(fmt.Sprintf("%s\n%s\n", mapping, replacer.Replace(string(p))))
	req := opensearchapi.BulkRequest{
		Index: IndexName,
		Body:  document,
	}
	insertResponse, err := req.Do(context.Background(), ow.Client)
	if err != nil {
		return 0, err
	}
	defer insertResponse.Body.Close()
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

	// Set Up Logger ----------------------------------------------------------------------------
	var l *logrus.Logger = &logrus.Logger{
		Out:   &OpenSearchWriter{Client: client},
		Level: logrus.InfoLevel,
		Formatter: &OpensearchFormatter{
			DisableHTMLEscape: true,
			PrettyPrint:       true,
		},
	}
	l.Info("Plonk")
}
