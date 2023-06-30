package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"os"
	"time"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/sirupsen/logrus"
)

//-------------------------------------------------------------------------------------------------
// LogMessage describes a simple log message, which is then encoded into a json
type LogMessage struct {
	Timestamp time.Time `json:"@timestamp"`
	Message   string    `json:"message"`
	Function  string    `json:"function_name"`
	Level     string    `json:"level"`
}

//-------------------------------------------------------------------------------------------------
// Custom type that will later implement the Write method/interface for logging 
// directly to Opensearch, without the help of using logstash.
type OpenSearchWriter struct {
	Client *opensearch.Client
}

//-------------------------------------------------------------------------------------------------
// Write function/method for writting directly to opensearch\
// For mor information, see:
//
// - https://github.com/elastic/ecs-logging-go-logrus/blob/main/formatter.go or
// - https://github.com/sirupsen/logrus/issues/719
func (ow *OpenSearchWriter) Write(p []byte) (n int, err error) {
	// pre-processig step for parsing the byte slice
	// Trims {left brace and }right brace
	trimmedString := strings.Trim(string(p), "{")

	// Solves the issue of chopped up messages. From https://www.json.org/json-en.html
	// An object is an unordered set of name/value pairs. 
	// An object begins with {left brace and ends with }right brace. 
	// Each name is followed by :colon and the name/value pairs are separated by ,comma.
	// 
	// Because the byte slice provided can contain a string formatted as a json,
	// splitting after each ,comma would lead to a chopped up message. Therefore we split
	// the string into three parts.
	//
	// But: The entry 'time' of the byte slice p is now included and actually not needed.
	// To drop it, we need to handle it in some way. For that see line 71-73
	splittedString := strings.SplitAfterN(trimmedString, ",", 3)

	// 'function_name' does not contain any ':', otherwise we have the same
	// issue as with message
	function := strings.SplitAfter(splittedString[0], ":")[1]
	logLevel := strings.SplitAfter(splittedString[1], ":")[1]

	// Solves the issues with chopped up messages -> splits the string
	// in 'message' and the 'rest'
	r := strings.NewReplacer("\\n", "", "\\r", "", "\\t", "", "\"", "", "\\", "")
	message := strings.SplitAfterN(splittedString[2], ":", 2)[1]
	messageCleaned := r.Replace(message)

	// The reason for using len(...)-2 is, to trim the newline char and the 
	// last "double quote character
	logMessage := LogMessage{
		Timestamp: time.Now().UTC(),
		// We're looking for the last ,colon and slice the string from index 1 to
		// the position where it is encountered
		Message:   messageCleaned[1:strings.LastIndex(messageCleaned, ",")],
		Function:  function[2 : len(function)-2],
		Level:     logLevel[2 : len(logLevel)-2],
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

//-------------------------------------------------------------------------------------------------
func main() {
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
	l.SetOutput(&OpenSearchWriter{Client: client})
	l.SetLevel(logrus.InfoLevel)
	l.SetFormatter(&OpensearchFormatter{PrettyPrint: false})

	e := l.WithFields(
		logrus.Fields{"function_name": "main"},
	)

	e = e.WithFields(
		logrus.Fields{"extra_field_1": "extra_value_1"},
	)

	e = e.WithFields(
		logrus.Fields{"extra_field_2": "extra_value_2"},
	)

	e.Info("this-is-a-test")
}
