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

	// Solves the issue of chopped up messages -> json contains
	splittedString := strings.SplitAfterN(trimmedString, ",", 3)

	// Trims the last entry 'time' of the byte slice p. Make sure that
	// 'function_name' does not contain any ':', otherwise we have the same
	// issue with message
	function := strings.SplitAfter(splittedString[0], ":")[1]
	logLevel := strings.SplitAfter(splittedString[1], ":")[1]

	// Solves the issues with chopped up messages -> splits the string
	// in 'message' and the 'rest'
	message := strings.SplitAfterN(splittedString[2], ":", 2)[1]

	r := strings.NewReplacer("\\n", "", "\\r", "", "\\t", "", "\"", "", "\\", "")
	test := r.Replace(message)

	// ------------------------------------------------------------------------
	// reason for len(...)-2 >> to trim the newline char and the last "
	logMessage := LogMessage{
		Timestamp: time.Now().UTC(),
		// We want the last '}', therefore len(message)-1
		Message:  test[1 : len(test)-1],
		Function: function[1 : len(function)-2],
		Level:    logLevel[1 : len(logLevel)-2],
	}

	logJson, err := json.Marshal(logMessage)
	if err != nil {
		return 0, err
	}

	// ------------------------------------------------------------------------
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
	l.SetFormatter(&OpensearchFormatter{PrettyPrint: true})

	test := `{
		"timestamp": 1679903463.1590829,
		"currentstatistics": {
			"sessionsown": 1,
			"sessionsforeign": 0,
			"sessionstotal": 1,
			"transcodedmedia": 0,
			"packetrate_user": 2,
			"byterate_user": 24,
			"errorrate_user": 1,
			"packetrate_kernel": 0,
			"byterate_kernel": 0,
			"errorrate_kernel": 0,
			"packetrate": 2,
			"byterate": 24,
			"errorrate": 1,
			"media_userspace": 1,
			"media_kernel": 0,
			"media_mixed": 0
		},
		"totalstatistics": {
			"uptime": "9",
			"managedsessions": 0,
			"rejectedsessions": 0,
			"timeoutsessions": 0,
			"silenttimeoutsessions": 0,
			"finaltimeoutsessions": 0,
			"offertimeoutsessions": 0,
			"regularterminatedsessions": 0,
			"forcedterminatedsessions": 0,
			"relayedpackets_user": 8,
			"relayedpacketerrors_user": 4,
			"relayedbytes_user": 96,
			"relayedpackets_kernel": 0,
			"relayedpacketerrors_kernel": 0,
			"relayedbytes_kernel": 0,
			"relayedpackets": 8,
			"relayedpacketerrors": 4,
			"relayedbytes": 96,
			"zerowaystreams": 0,
			"onewaystreams": 0,
			"avgcallduration": "0.000000",
			"totalcallsduration": "0.000000",
			"totalcallsduration2": "0.000000",
			"totalcallsduration_stddev": "0.000000"
		},
		"mos": {
			"mos_total": 0,
			"mos2_total": 0,
			"mos_samples_total": 0,
			"mos_average": 0,
			"mos_stddev": 0
		},
		"voip_metrics": {
			"jitter_total": 0,
			"jitter2_total": 0,
			"jitter_samples_total": 0,
			"jitter_average": 0,
			"jitter_stddev": 0,
			"rtt_e2e_total": 0,
			"rtt_e2e2_total": 0,
			"rtt_e2e_samples_total": 0,
			"rtt_e2e_average": 0,
			"rtt_e2e_stddev": 0,
			"rtt_dsct_total": 0,
			"rtt_dsct2_total": 0,
			"rtt_dsct_samples_total": 0,
			"rtt_dsct_average": 0,
			"rtt_dsct_stddev": 0,
			"packetloss_total": 0,
			"packetloss2_total": 0,
			"packetloss_samples_total": 0,
			"packetloss_average": 0,
			"packetloss_stddev": 0,
			"jitter_measured_total": 0,
			"jitter_measured2_total": 0,
			"jitter_measured_samples_total": 0,
			"jitter_measured_average": 0,
			"jitter_measured_stddev": 0,
			"packets_lost": 0,
			"rtp_duplicates": 0,
			"rtp_skips": 0,
			"rtp_seq_resets": 0,
			"rtp_reordered": 0
		},
		"controlstatistics": {
			"proxies": [
				{
					"proxy": "127.0.0.1",
					"pingcount": 1,
					"pingduration": 9.9999999999999995e-7,
					"offercount": 2,
					"offerduration": 0.00037800000000000003,
					"answercount": 0,
					"answerduration": 0,
					"deletecount": 1,
					"deleteduration": 0.000057000000000000003,
					"querycount": 0,
					"queryduration": 0,
					"listcount": 0,
					"listduration": 0,
					"startreccount": 0,
					"startrecduration": 0,
					"stopreccount": 0,
					"stoprecduration": 0,
					"pausereccount": 0,
					"pauserecduration": 0,
					"startfwdcount": 0,
					"startfwdduration": 0,
					"stopfwdcount": 0,
					"stopfwdduration": 0,
					"blkdtmfcount": 0,
					"blkdtmfduration": 0,
					"unblkdtmfcount": 0,
					"unblkdtmfduration": 0,
					"blkmediacount": 0,
					"blkmediaduration": 0,
					"unblkmediacount": 0,
					"unblkmediaduration": 0,
					"playmediacount": 0,
					"playmediaduration": 0,
					"stopmediacount": 0,
					"stopmediaduration": 0,
					"playdtmfcount": 0,
					"playdtmfduration": 0,
					"statscount": 0,
					"statsduration": 0,
					"slnmediacount": 0,
					"slnmediaduration": 0,
					"unslnmediacount": 0,
					"unslnmediaduration": 0,
					"pubcount": 0,
					"pubduration": 0,
					"subreqcount": 0,
					"subreqduration": 0,
					"subanscount": 0,
					"subansduration": 0,
					"unsubcount": 0,
					"unsubduration": 0,
					"errorcount": 0
				}
			],
			"totalpingcount": 1,
			"totaloffercount": 2,
			"totalanswercount": 0,
			"totaldeletecount": 1,
			"totalquerycount": 0,
			"totallistcount": 0,
			"totalstartreccount": 0,
			"totalstopreccount": 0,
			"totalpausereccount": 0,
			"totalstartfwdcount": 0,
			"totalstopfwdcount": 0,
			"totalblkdtmfcount": 0,
			"totalunblkdtmfcount": 0,
			"totalblkmediacount": 0,
			"totalunblkmediacount": 0,
			"totalplaymediacount": 0,
			"totalstopmediacount": 0,
			"totalplaydtmfcount": 0,
			"totalstatscount": 0,
			"totalslnmediacount": 0,
			"totalunslnmediacount": 0,
			"totalpubcount": 0,
			"totalsubreqcount": 0,
			"totalsubanscount": 0,
			"totalunsubcount": 0
		},
		"interfaces": [
			{
				"name": "internal",
				"address": "10.0.243.239",
				"ports": {
					"min": 10000,
					"max": 39999,
					"used": 4,
					"used_pct": 0.013333333333333334,
					"free": 29996,
					"totals": 30000,
					"last": 10022
				},
				"packets_lost": 0,
				"duplicates": 0,
				"interval": {
					"packets_lost": 0,
					"duplicates": 0
				},
				"rate": {
					"packets_lost": 0,
					"duplicates": 0
				},
				"voip_metrics": {
					"mos_total": 0,
					"mos2_total": 0,
					"mos_samples_total": 0,
					"mos_average": 0,
					"mos_stddev": 0,
					"jitter_total": 0,
					"jitter2_total": 0,
					"jitter_samples_total": 0,
					"jitter_average": 0,
					"jitter_stddev": 0,
					"rtt_e2e_total": 0,
					"rtt_e2e2_total": 0,
					"rtt_e2e_samples_total": 0,
					"rtt_e2e_average": 0,
					"rtt_e2e_stddev": 0,
					"rtt_dsct_total": 0,
					"rtt_dsct2_total": 0,
					"rtt_dsct_samples_total": 0,
					"rtt_dsct_average": 0,
					"rtt_dsct_stddev": 0,
					"packetloss_total": 0,
					"packetloss2_total": 0,
					"packetloss_samples_total": 0,
					"packetloss_average": 0,
					"packetloss_stddev": 0,
					"jitter_measured_total": 0,
					"jitter_measured2_total": 0,
					"jitter_measured_samples_total": 0,
					"jitter_measured_average": 0,
					"jitter_measured_stddev": 0
				},
				"voip_metrics_interval": {
					"mos": 0,
					"mos_stddev": 0,
					"jitter": 0,
					"jitter_stddev": 0,
					"rtt_e2e": 0,
					"rtt_e2e_stddev": 0,
					"rtt_dsct": 0,
					"rtt_dsct_stddev": 0,
					"packetloss": 0,
					"packetloss_stddev": 0,
					"jitter_measured": 0,
					"jitter_measured_stddev": 0
				},
				"ingress": {
					"packets": 0,
					"bytes": 0,
					"errors": 0
				},
				"egress": {
					"packets": 8,
					"bytes": 96,
					"errors": 0
				},
				"ingress_interval": {
					"packets": 0,
					"bytes": 0,
					"errors": 0
				},
				"egress_interval": {
					"packets": 2,
					"bytes": 24,
					"errors": 0
				},
				"ingress_rate": {
					"packets": 0,
					"bytes": 0,
					"errors": 0
				},
				"egress_rate": {
					"packets": 2,
					"bytes": 24,
					"errors": 0
				}
			},
			{
				"name": "external",
				"address": "10.0.243.239",
				"ports": {
					"min": 10000,
					"max": 39999,
					"used": 4,
					"used_pct": 0.013333333333333334,
					"free": 29996,
					"totals": 30000,
					"last": 10022
				},
				"packets_lost": 0,
				"duplicates": 0,
				"interval": {
					"packets_lost": 0,
					"duplicates": 0
				},
				"rate": {
					"packets_lost": 0,
					"duplicates": 0
				},
				"voip_metrics": {
					"mos_total": 0,
					"mos2_total": 0,
					"mos_samples_total": 0,
					"mos_average": 0,
					"mos_stddev": 0,
					"jitter_total": 0,
					"jitter2_total": 0,
					"jitter_samples_total": 0,
					"jitter_average": 0,
					"jitter_stddev": 0,
					"rtt_e2e_total": 0,
					"rtt_e2e2_total": 0,
					"rtt_e2e_samples_total": 0,
					"rtt_e2e_average": 0,
					"rtt_e2e_stddev": 0,
					"rtt_dsct_total": 0,
					"rtt_dsct2_total": 0,
					"rtt_dsct_samples_total": 0,
					"rtt_dsct_average": 0,
					"rtt_dsct_stddev": 0,
					"packetloss_total": 0,
					"packetloss2_total": 0,
					"packetloss_samples_total": 0,
					"packetloss_average": 0,
					"packetloss_stddev": 0,
					"jitter_measured_total": 0,
					"jitter_measured2_total": 0,
					"jitter_measured_samples_total": 0,
					"jitter_measured_average": 0,
					"jitter_measured_stddev": 0
				},
				"voip_metrics_interval": {
					"mos": 0,
					"mos_stddev": 0,
					"jitter": 0,
					"jitter_stddev": 0,
					"rtt_e2e": 0,
					"rtt_e2e_stddev": 0,
					"rtt_dsct": 0,
					"rtt_dsct_stddev": 0,
					"packetloss": 0,
					"packetloss_stddev": 0,
					"jitter_measured": 0,
					"jitter_measured_stddev": 0
				},
				"ingress": {
					"packets": 8,
					"bytes": 96,
					"errors": 4
				},
				"egress": {
					"packets": 0,
					"bytes": 0,
					"errors": 0
				},
				"ingress_interval": {
					"packets": 2,
					"bytes": 24,
					"errors": 1
				},
				"egress_interval": {
					"packets": 0,
					"bytes": 0,
					"errors": 0
				},
				"ingress_rate": {
					"packets": 2,
					"bytes": 24,
					"errors": 1
				},
				"egress_rate": {
					"packets": 0,
					"bytes": 0,
					"errors": 0
				}
			}
		],
		"transcoders": [],
		"sf_mqtt_topic": "dt/agent/1fc17a6b-e932-40fe-91c9-05b8438e1471/d7452484-c88c-11ed-8aad-9cbba9e2c686/18554f0e-cc74-11ed-a694-174ae68c282a/caller/rtp"
	}`

	e.Info("stompSource: state= " + "test" + " destination: " + "middle-earth" + " dataText: " + "frodo" + " addInfo: " + test)
}
