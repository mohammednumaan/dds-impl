package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var proxyServerAddr string
var pollInterval time.Duration

func main() {
	flag.StringVar(&proxyServerAddr, "proxy", "http://localhost:8092/", "address of the proxy server")
	flag.DurationVar(&pollInterval, "interval", time.Second, "interval between requests")
	flag.Parse()

	targetResource := "some-resource"
	prodCount := 0
	betaCount := 0
	errorCount := 0

	for {
		resp, err := http.Get(proxyServerAddr + targetResource)
		if err != nil {
			errorCount++
			fmt.Printf("request failed: %v | prod=%d beta=%d errors=%d\n", err, prodCount, betaCount, errorCount)
			time.Sleep(pollInterval)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorCount++
			fmt.Printf("response read failed: %v | prod=%d beta=%d errors=%d\n", err, prodCount, betaCount, errorCount)
			time.Sleep(pollInterval)
			continue
		}

		responseBody := string(body)
		switch {
		case strings.Contains(responseBody, "prod server"):
			prodCount++
		case strings.Contains(responseBody, "beta server"):
			betaCount++
		default:
			errorCount++
		}

		total := prodCount + betaCount
		betaPercent := 0.0
		if total > 0 {
			betaPercent = float64(betaCount) / float64(total) * 100
		}

		fmt.Printf("received: '%s' | prod=%d beta=%d beta_percent=%.2f errors=%d\n",
			responseBody,
			prodCount,
			betaCount,
			betaPercent,
			errorCount,
		)
		time.Sleep(pollInterval)
	}
}
