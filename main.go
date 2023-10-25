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
	// "github.com/sirupsen/logrus"
)

//-------------------------------------------------------------------------------------------------
type SF_LogLevel int64
type SF_LogMessage struct {
	AwsRequestId 	string    		`json:"aws_request_id,omitempty"`
	Function     	string    		`json:"function_name"`
	LogLevel 		SF_LogLevel		`json:"level"`
	Message  		string			`json:"message"`
	Timestamp    	time.Time		`json:"@timestamp"`
}

const (
	SF_LogLevel_Info SF_LogLevel = iota
	SF_LogLevel_Error
	SF_LogLevel_Warn
)

var logc chan SF_LogMessage = make(chan SF_LogMessage, 10)


//-------------------------------------------------------------------------------------------------
func log_info(aws_request_id, message string) {
	logc <- SF_LogMessage{
		AwsRequestId: 	aws_request_id,
		Function: 		"test-function",
		LogLevel: 		SF_LogLevel_Info,
		Message: 		message,
		Timestamp: 		time.Now().UTC(),
	}
}

//-------------------------------------------------------------------------------------------------
func log_error(aws_request_id, message string) {
	logc <- SF_LogMessage{
		AwsRequestId: 	aws_request_id,
		Function: 		"test-function",
		LogLevel: 		SF_LogLevel_Error,
		Message: 		message,
		Timestamp: 		time.Now().UTC(),
	}
}

//-------------------------------------------------------------------------------------------------
func log_warn(aws_request_id, message string) {
	logc <- SF_LogMessage{
		AwsRequestId: 	aws_request_id,
		Function: 		"test-function",
		LogLevel: 		SF_LogLevel_Warn,
		Message: 		message,
		Timestamp: 		time.Now().UTC(),
	}
}


//-------------------------------------------------------------------------------------------------
// LogMessage describes a simple log message, which is then encoded into a json
type LogMessage struct {
	Function     	string    	`json:"function_name"`
	Level        	string    	`json:"level"`
	Message      	string    	`json:"message"`
	Timestamp    	time.Time	`json:"@timestamp"`
	AwsRequestId 	string    	`json:"aws_request_id,omitempty"`
	CustomerId   	string	   	`json:"customer_id,omitempty"`
	SessionId    	string    	`json:"session_id,omitempty"`
	UserId    		string    	`json:"user_id,omitempty"`
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
// Write function/method for writting directly to opensearch
func (ow *OpenSearchWriter) Write(p []byte) (n int, err error) {
	log := LogMessage{}
	if err := json.Unmarshal(p, &log); err != nil {
		fmt.Printf("[ERROR]: %s\n", err)
		return 0, err
    }

	log.Timestamp = log.Timestamp.UTC()
	logJson, err := json.Marshal(log)
	if err != nil {
		fmt.Printf("[ERROR]: %s\n", err)
		return 0, err
	}

	req := opensearchapi.IndexRequest{
		Index: "sipfront-gotest-" + time.Now().Format("2006.01.02"),
		Body:  strings.NewReader(string(logJson)),
	}

	// https://stackoverflow.com/questions/16280176/go-panic-runtime-error-invalid-memory-address-or-nil-pointer-dereference
	insertResponse, err := req.Do(context.Background(), ow.Client)
	if err != nil {
		fmt.Printf("[ERROR]: %s\nResponseBody: %s\n", err, insertResponse)
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
		//	"https://vpc-sipfront-os-iepreu6yviwjk5rnzetncw7dfm.eu-central-1.es.amazonaws.com", // dev
			"https://vpc-sipfront-os-fgqfem5p72z6mzvlm542j43uvy.eu-central-1.es.amazonaws.com",	// prod
		},
	}
	client, err := opensearch.NewClient(clientConfiguration)
	if err != nil {
		fmt.Println("cannot initialize", err)
		os.Exit(1)
	}

	// var l *logrus.Logger = logrus.New()
	// l.SetOutput(&OpenSearchWriter{Client: client})
	// l.SetFormatter(&OpensearchFormatter{PrettyPrint: false})
	// l.SetLevel(logrus.InfoLevel)
	// e := l.WithFields(
	// 	logrus.Fields{"function_name": "main"},
	// )
	// e = e.WithFields(
	//  	logrus.Fields{"aws_request_id": "1"},
	// )

	// https://github.com/Sirupsen/logrus/issues/338
	log_info("101", "this-is-a-test")
	log_error("201", "this-is-another-test")

	test := ``

	close(logc)
	s := " "
	for i := range logc {
		log, err := json.Marshal(i)
		if err != nil {
			fmt.Printf("[ERROR]: %s\n", err)
		}

		index := "sipfront-playground-" + time.Now().Format("2006.01.02")
		// test += fmt.Sprintf(`{"index" : { "_index" : %s, "_id" : "1" }}`, index)+"\n"
		test += fmt.Sprintf(`{"index" : { "_index" : %s }}`, index)+"\n"
		s = string(log)+"\n"
		test += s
	}

	print(test)

	res, err := client.Bulk(strings.NewReader(test))
	fmt.Println(res, err)
}
