package main

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	// "github.com/sirupsen/logrus"
)

var (
	logger_client *logrus.Logger = nil
	logger        *logrus.Entry  = nil
	environment   string         = "dev"
)

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
		fields := logrus.Fields{
			"function_name": "test_function",
		}

		logger = logger_client.WithFields(fields)
		logger_client.SetOutput(io.Discard) // Send all logs to nowhere by default
		logger_client.SetLevel(logrus.InfoLevel)

		logger_client.AddHook(
			&FormatterHook{
				Writer: os.Stdout,
				LogLevels: []logrus.Level{
					logrus.PanicLevel,
					logrus.FatalLevel,
					logrus.ErrorLevel,
					logrus.WarnLevel,
					logrus.InfoLevel,
				},
				Formatter: &logrus.TextFormatter{
					DisableTimestamp: true,
				},
			})

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

	logger.Info("This-is-a-test")
	defer func() {
		if r := recover(); r != nil {
			logger.Infof("recovered in main: +v%\n", r)
			OpensearchWriter.Convert()
			panic(r)
		}
	}()

	panic("Test panic")
}
