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

	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/sirupsen/logrus"
)

// For test purposes
// var IndexName = "sipfront-gotest-v3-" + time.Now().UTC().Format("2006-01-02")

// Custom type which will later implement the Write method for logging directly to
// Opensearch, without the help of using logstash.
type OpenSearchWriter struct {
	Client *opensearch.Client
}

// LogMessage describes a simple log message, which is then encoded into a json
type LogMessage struct {
	Timestamp time.Time `json:"@timestamp"`
	Message   string    `json:"message"`
	Function  string    `json:"function_name"`
	Level     string    `json:"level"`
}

// Write function/method for writting directly to opensearch
func (ow *OpenSearchWriter) Write(p []byte) (n int, err error) {
	// pre-processig step for parsing the byte slice
	//
	trimmedString := strings.Trim(string(p), "{}")
	splittedString := strings.SplitAfter(trimmedString, ",")

	// Trims the last entry 'time' of the byte slice p. Make sure that
	// 'function_name' does not contain any ':', otherwise we have the same 
	// issue with message
	function := strings.SplitAfter(splittedString[0], ":")[1]
	logLevel := strings.SplitAfter(splittedString[1], ":")[1]

	// olves the issues with chopped up messages -> splits the string
	// in 'message' and the 'rest'
	message := strings.SplitAfterN(splittedString[2], ":", 2)[1]

	// ------------------------------------------------------------------------
	// reason for len(...)-2 >> to trim the newline char and the last "
	logMessage := LogMessage{
		Timestamp: time.Now().UTC(),
		Message:   message[1 : len(message)-2],
		Function:  function[1 : len(function)-2],
		Level:     logLevel[1 : len(logLevel)-2],
	}

	logJson, err := json.Marshal(logMessage)
	if err != nil {
		return 0, err
	}

	req := opensearchapi.IndexRequest{
		Index: "sipfront-gotest-" + time.Now().UTC().Format("2006.01.02"),
		Body:  strings.NewReader(string(logJson)),
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
	// Initialize the client with SSL/TLS enabled.
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

	var l *logrus.Logger = logrus.New()
	e := l.WithField("function_name", "main")

	l.SetOutput(&OpenSearchWriter{Client: client})
	l.SetLevel(logrus.InfoLevel)
	l.SetFormatter(&OpensearchFormatter{})

	e.Info("stompSource: state= " + "test" + " destination: " + "middle-earth" + " dataText: " + "frodo")
}
