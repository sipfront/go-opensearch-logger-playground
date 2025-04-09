package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// LogMessage describes a simple log message, which is then encoded into a json
type LogMessage struct {
	Function     string    `json:"function_name"`
	Level        string    `json:"level"`
	Message      string    `json:"message"`
	Timestamp    time.Time `json:"@timestamp"`
	AwsRequestId string    `json:"aws_request_id,omitempty"`
	CustomerId   string    `json:"cid,omitempty"`
	SessionId    string    `json:"sid,omitempty"`
	UserId       string    `json:"uid,omitempty"`
}

// Custom type that will later implement the Write method/interface for collecting all
// log messages which are sent to SQS queue for later processing
type OpenSearchWriterProxy struct {
	// Channel to collect all log messages
	LogMessagesChannel chan []byte

	// The collected log messages are appended to a slice
	LogMessagesSlice []string
}

// Construct a New OpenSearchWriterProxy with a default buffered channel
// https://stackoverflow.com/questions/37135193/how-to-set-default-values-in-go-structs
func NewOpenSearchWriterProxy(size int) *OpenSearchWriterProxy {
	return &OpenSearchWriterProxy{
		LogMessagesChannel: make(chan []byte, size),
		LogMessagesSlice:   make([]string, 0, size),
	}
}

// Write function/method for writting directly to opensearch\
// For mor information, see:
//
// - https://github.com/elastic/ecs-logging-go-logrus/blob/main/formatter.go or
// - https://github.com/sirupsen/logrus/issues/719
// Write function/method for writting directly to opensearch
func (ow *OpenSearchWriterProxy) Write(p []byte) (n int, err error) {
	log := LogMessage{}
	if err := json.Unmarshal(p, &log); err != nil {
		fmt.Printf("[ERROR]: %s\n", err)
		return 0, err
	}
	log.Timestamp = log.Timestamp.UTC()

	// Type Hints:
	// 		logJson: []byte
	// 		err: error
	logJson, err := json.Marshal(log)
	if err != nil {
		fmt.Printf("[ERROR]: %s\n", err)
		return 0, err
	}

	// Collect all logJson into the channel and then generate
	// the payload for SQS.
	//
	ow.LogMessagesChannel <- logJson

	return len(p), nil
}

// Added Close method for the writer
func (ow *OpenSearchWriterProxy) Close() {
	close(ow.LogMessagesChannel)
}

// 'Convert' the content of the channel into a slice
func (ow *OpenSearchWriterProxy) Convert() {

	var (
		err         error
		totalLength int
	)

	ow.Close()
	for message := range ow.LogMessagesChannel {
		messageLength := len(message)

		if messageLength+totalLength > 50000 {
			err = ow.SendToSqs("sipfront-log-sqs-" + environment)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
			}

			// Clear Slice
			ow.LogMessagesSlice = nil
			totalLength = 0
		}

		totalLength += messageLength
		ow.LogMessagesSlice = append(ow.LogMessagesSlice, string(message))
	}

	err = ow.SendToSqs("sipfront-log-sqs-" + environment)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
}

// Send the slice to
func (ow *OpenSearchWriterProxy) SendToSqs(QueueName string) error {
	// SqsSession := session.Must(session.NewSessionWithOptions(
	// 	session.Options{SharedConfigState: session.SharedConfigEnable},
	// ),
	// )

	//var (
	//	LogSqsClientErr error
	// 	LogSqsClient    *sqs.SQS = sqs.New(SqsSession)
	// )

	MessageBodyByte, _ := json.Marshal(ow.LogMessagesSlice)
	MessageBodyString := string(MessageBodyByte)
	// _, LogSqsClientErr = LogSqsClient.SendMessage(&sqs.SendMessageInput{
	// 	MessageBody: &MessageBodyString,
	// 	QueueUrl:    &QueueName,
	// },
	// )
	fmt.Println(MessageBodyString)

	// if LogSqsClientErr != nil {
	// 	return LogSqsClientErr
	// }

	return nil
}
