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
type SF_LogLevel int64
type SF_LogMessage struct {
	AwsRequestId 	string    		`json:"aws_request_id,omitempty"`
	Message  		string			`json:"message"`
	LogLevel 		SF_LogLevel		`json:"level"`
	Timestamp    	time.Time		`json:"@timestamp"`
}

const (
	SF_LogLevel_Info SF_LogLevel = iota
	SF_LogLevel_Error
)

var logc chan SF_LogMessage


//-------------------------------------------------------------------------------------------------
func log_info(aws_request_id, message string) {
	logc <- SF_LogMessage{
		AwsRequestId: 	aws_request_id,
		LogLevel: 		SF_LogLevel_Info,
		Message: 		message,
		Timestamp: 		time.Now.UTC(),
	}
}

//-------------------------------------------------------------------------------------------------
func log_error(message string) {
	logc <- SF_LogMessage{
		AwsRequestId: 	aws_request_id,
		LogLevel: 		SF_LogLevel_Error,
		Message: 		message,
		Timestamp: 		time.Now.UTC(),
	}
}

//-------------------------------------------------------------------------------------------------
func logSink(
	ctx context.Context,
	function_name string,
	aws_request_id string,
	in <-chan SF_LogMessage) (<-chan error, error) {

	errc := make(chan error, 1)
	go func() {
		defer close(errc)

		// create logger!
		var go_logger_client *logrus.Logger = nil
		var go_logger *logrus.Entry = nil

		if go_logger_client == nil {
			go_logger_client = logrus.New()
			fields := logrus.Fields{
				"function_name":  function_name,
				"aws_request_id": aws_request_id,
			}
			go_logger = go_logger_client.WithFields(fields)

			// Set Up OpenSearch Client and initialize with SSL/TLS enabled.
			var clientConfiguration opensearch.Config = opensearch.Config{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
				Addresses: []string{
					es_endpoint},
			}
			client, err := opensearch.NewClient(clientConfiguration)
			if err != nil {
				println("init logging failed")
				panic("init logging failed")
			} else {
				// Set Up logger with non-default params
				// logger_client.SetOutput(&OpenSearchWriter{Client: client})
				logger_client.SetFormatter(&OpensearchFormatter{})
				logger_client.SetLevel(logrus.InfoLevel)
			}
		}

		for {
			select {
			case s := <-in:
				switch level := s.LogLevel; level {
				case SF_LogLevel_Info:
					//log info
					//log_info("logSink: " + s.Message)
					go_logger.Info("logSink: " + s.Message)
				case SF_LogLevel_Error:
					//log error
					//logger.Info("logSink: " + s.Message)
					go_logger.Error("logSink: " + s.Message)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return errc, nil
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
		logrus.Fields{"aws_request_id": "1"},
	)

	e.Info("this-is-a-test")
}
