package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	opensearchapi "github.com/opensearch-project/opensearch-go/opensearchapi"
	opensearch "github.com/opensearch-project/opensearch-go/v2"
	"github.com/sirupsen/logrus"
)

// ------------------------------------------------------------------------------------------------
// For test purposes
const IndexName = "sipfront-gotest-v5-2023.03.22"

// var replacer = strings.NewReplacer("\n", "", " ", "")

// ------------------------------------------------------------------------------------------------
// Custom type which will later implement the Write method for logging directly to
// Opensearch, without the help of using logstash.
type OpenSearchWriter struct {
	Client *opensearch.Client
}

type LogMessage struct {
	Timestamp time.Time `json:"@timestamp"`
	Message   string    `json:"message"`
	Level     string    `json:"level"`
}

func (ow *OpenSearchWriter) Write(p []byte) (n int, err error) {
	now := time.Now().UTC()
	logMessage := LogMessage{
		Timestamp: now,
		Message:   string(p),
	}

	logJson, err := json.Marshal(logMessage)
	if err != nil {
		return 0, err
	}

	var document *strings.Reader = strings.NewReader(string(logJson))
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
	l.Info("Blub")
}
