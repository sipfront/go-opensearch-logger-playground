package main

import (
	// "context"
	// "crypto/tls"
	// "strings"
	// "net/http"

	"fmt"
	"os"

	opensearch "github.com/opensearch-project/opensearch-go/v2"
	// opensearchapi "github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

const IndexName = "sipfront-go-test"

func main() {

	// Initialize the client with SSL/TLS enabled.
	var clientConfiguration opensearch.Config = opensearch.Config{
		Addresses: []string{
			"https://vpc-sipfront-os-iepreu6yviwjk5rnzetncw7dfm.eu-central-1.es.amazonaws.com"},
	}

	client, err := opensearch.NewClient(clientConfiguration)
	if err != nil {
		fmt.Println("cannot initialize", err)
		os.Exit(1)
	}

	// Print OpenSearch version information on console.
	fmt.Println(client.Info())
}
