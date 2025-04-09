package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	logger_client *logrus.Logger = nil
	logger        *logrus.Entry  = nil
	environment   string         = "dev"
)

// finalizeMsg takes a string message (`msg_string`) and sends it to the `recordc` channel.
// If a panic occurs, it recovers, logs the panic, and re-panics to signal a serious error.
func finalizeMsg(msg_string string, recordc chan<- string, ow *OpenSearchWriterProxy) {
	fmt.Println("in finalizeMsg")
	defer func() {
		fmt.Println("defered func call in finalizeMsg")
		if r := recover(); r != nil {
			defer func() {
				fmt.Println("recovered ow")
			}()
			ow.Convert()
		}
	}()

	// Send the message to the `recordc` channel
	recordc <- msg_string
}

// conditionProcessor continuously processes messages from the `in` channel and
// finalizes them using the `finalizeMsg` function. It stops processing when the context is done.
func conditionProcessor(in <-chan int, recordc chan<- string, ctx context.Context, ow *OpenSearchWriterProxy) {
	// Launch a goroutine to process messages from the `in` channel
	go func() {
		for {
			select {
			case s := <-in:
				// Process the input by repeating "A" `s` times and sending it to the record channel
				finalizeMsg(strings.Repeat("A", s), recordc, ow)
			case x := <-ctx.Done():
				// Handle context cancellation or timeout
				fmt.Printf("%+v\n", x)
				return
			}
		}
	}()
}

// psqlSink continuously processes messages from the `in` channel and
// finalizes them using the `finalizeMsg` function. It stops processing when the context is done.
func psqlSink(in <-chan int, recordc chan<- string, ctx context.Context, ow *OpenSearchWriterProxy) {
	// Launch a goroutine to process messages from the `in` channel
	go func() {
		for {
			select {
			case s := <-in:
				// Process the input by repeating "A" `s` times and sending it to the record channel
				finalizeMsg(strings.Repeat("B", s), recordc, ow)
			case x := <-ctx.Done():
				// Handle context cancellation or timeout
				fmt.Printf("%+v\n", x)
				return
			}
		}
	}()
}

// -------------------------------------------------------------------------------------------------
func main() {
	logger_client = nil

	// IMPORTANT: IN CASE SOMETHING GOES WRONG DURING DEPLOYMENT, COMMENT OUT
	// 1. writerCap
	// 2. OpensearchWriter
	// 3. logger_client.SetOutput(OpensearchWriter) AND
	// 4. OpensearchWriter.Convert()
	//

	writerCap := 10000
	OpensearchWriter := NewOpenSearchWriterProxy(writerCap)
	if logger_client == nil {
		logger_client = logrus.New()
		fields := logrus.Fields{"function_name": "test_function"}
		logger = logger_client.WithFields(fields)
		logger_client.SetOutput(io.Discard)
		logger_client.SetLevel(logrus.InfoLevel)

		// logger_client.AddHook(
		// 	&FormatterHook{
		// 		Writer: os.Stdout,
		// 		LogLevels: []logrus.Level{
		// 			logrus.PanicLevel,
		// 			logrus.FatalLevel,
		// 			logrus.ErrorLevel,
		// 			logrus.WarnLevel,
		// 			logrus.InfoLevel,
		// 		},
		// 		Formatter: &logrus.TextFormatter{
		// 			DisableTimestamp: true,
		// 		},
		// 	})

		logger_client.AddHook(
			&FormatterHook{
				Writer: OpensearchWriter,
				LogLevels: []logrus.Level{
					logrus.PanicLevel,
					logrus.FatalLevel,
					logrus.ErrorLevel,
					logrus.WarnLevel,
					logrus.InfoLevel,
				},
				Formatter: &OpensearchFormatter{},
			})
	}

	// Create a context with a timeout of 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel() // Ensure the context is canceled when the main function exits

	// Channel to send processed strings
	recordc := make(chan string)
	// defer close(recordc)
	// Note: the `recordc` channel is not explicitly closed in the original code.
	// Uncommenting `defer close(recordc)` would cause an issue since `conditionProcessor`
	// and the main loop access this channel concurrently.

	// Channel for synchronization with the main goroutine
	wg := make(chan int)
	defer close(wg)

	// Channel for sending integer inputs to the conditionProcessor
	in := make(chan int)
	defer close(in)

	defer func() {
		fmt.Println(<-wg)
		if r := recover(); r != nil {
			fmt.Println("Hello")
			OpensearchWriter.Convert()
		}
	}()

	// Start the conditionProcessor to process input values
	for i := 0; i < 3; i++ {
		conditionProcessor(in, recordc, ctx, OpensearchWriter)
	}

	// Start the psqlSink to process input values
	for i := 0; i < 3; i++ {
		psqlSink(in, recordc, ctx, OpensearchWriter)
	}

	// Launch a goroutine to send increasing integers to the `in` channel
	go func() {
		defer func() { wg <- 42 }() // Notify the main goroutine when processing is done
		c := 1
		for {
			select {
			case in <- c:
				// Increment and send integers to the `in` channel
			case <-ctx.Done():
				// Stop sending integers if the context is done
				return
			}
			c++
		}
	}()

	// Close the `recordc` channel (this will terminate the loop reading from `recordc`)
	// Read and print 7 messages from the `recordc` channel or stop when it's closed
	// defer close(recordc)
	for i := 0; i < 7; i++ {
		if _, ok := <-recordc; !ok {
			// Exit the loop if the `recordc` channel is closed
			fmt.Println("stopped")
			break
		} else {
			// Print the next message from the `recordc` channel
			fmt.Println(<-recordc)
		}

		if i == 2 {
			close(recordc)
		}
	}

	// Synchronize with the goroutine launched on line 68 using the `wg` channel
}
