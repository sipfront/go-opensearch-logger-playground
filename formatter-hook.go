package main

import (
	"io"

	"github.com/sirupsen/logrus"
)

// Copied from https://github.com/sirupsen/logrus/issues/894#issuecomment-1284051207
//
// FormatterHook is a hook that writes logs of specified LogLevels with a formatter to specified Writer
type FormatterHook struct {
	Writer    io.Writer
	LogLevels []logrus.Level
	Formatter logrus.Formatter
}

// Copied from https://github.com/sirupsen/logrus/issues/894#issuecomment-1284051207
//
// Fire will be called when some logging function is called with current hook
// It will format log entry and write it to appropriate writer
func (hook *FormatterHook) Fire(entry *logrus.Entry) error {
	line, err := hook.Formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write(line)
	return err
}

// Copied from https://github.com/sirupsen/logrus/issues/894#issuecomment-1284051207
//
// Levels define on which log levels this hook would trigger
func (hook *FormatterHook) Levels() []logrus.Level {
	return hook.LogLevels
}
