package main

import (
	"github.com/sirupsen/logrus"
)

// // Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Formatter is a logrus.Formatter, formatting log entries as ECS-compliant JSON.
type OpensearchFormatter struct {
	// DisableHTMLEscape allows disabling html escaping in output
	DisableHTMLEscape bool

	// DataKey allows users to put all the log entry parameters into a
	// nested dictionary at a given key.
	//
	// DataKey is ignored for well-defined fields, such as "error",
	// which will instead be stored under the appropriate ECS fields.
	DataKey string

	// PrettyPrint will indent all json logs
	PrettyPrint bool
}

type errorObject struct {
	Message string `json:"message,omitempty"`
}

// Format formats e as ECS-compliant JSON.
func (f *OpensearchFormatter) Format(e *logrus.Entry) ([]byte, error) {
	datahint := len(e.Data)
	if f.DataKey != "" {
		datahint = 2
	}
	data := make(logrus.Fields, datahint)

	if len(e.Data) > 0 {
		extraData := data
		if f.DataKey != "" {
			extraData = make(logrus.Fields, len(e.Data))
		}
		for k, v := range e.Data {
			switch k {
			case logrus.ErrorKey:
				err, ok := v.(error)
				if ok {
					data["error"] = errorObject{
						Message: err.Error(),
					}
					break
				}
				fallthrough // error has unexpected type
			default:
				extraData[k] = v
			}
		}
		if f.DataKey != "" && len(extraData) > 0 {
			data[f.DataKey] = extraData
		}
	}

	ecopy := *e
	ecopy.Data = data
	e = &ecopy

	// https://stackoverflow.com/questions/63396766/is-it-possible-to-swap-msg-for-message-with-logrus-logging
	jf := logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.999999999Z07:00",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyMsg:  "message",
			logrus.FieldKeyTime: "@timestamp",
		},
		DisableHTMLEscape: f.DisableHTMLEscape,
		PrettyPrint:       f.PrettyPrint,
	}
	return jf.Format(e)
}
