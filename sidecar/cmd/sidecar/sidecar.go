package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

/*
this is the sidecar container that will run alongside the main app container.
it will constantly poll the config server for changes to the .json or any config file and
update the file in the shared volume that both containers have access to. then
it sends a signal to the main app container to reload the config file, this can be done
using a SIGHUP or SIGKILL signal.
*/

type ConfigResponse struct {
	Hash    string `json:"hash"`
	Content string `json:"content"`
}

// this is to store the last hash of the config file
// so we can compare it with the new hash and only update the file if it has changed
var lastHash string

func main() {
	serverURL := flag.String("server-url", "http://localhost:8080/config", "URL of the config server")
	configFile := flag.String("config-file", "/shared/config.json", "Path to write config file")
	pidFile := flag.String("pid-file", "/shared/app.pid", "Path to app PID file")
	pollInterval := flag.Duration("poll-interval", 5*time.Second, "Poll interval in seconds")
	flag.Parse()

	for {
		poll(*serverURL, *configFile, *pidFile)
		time.Sleep(*pollInterval)
	}
}

func poll(url string, configFile string, pidFile string) {
	res, err := http.Get(url)
	if err != nil {
		fmt.Println("Error polling config server:", err)
		return
	}
	defer res.Body.Close()

	var cfgResp ConfigResponse
	err = json.NewDecoder(res.Body).Decode(&cfgResp)
	if err != nil {
		fmt.Println("Error decoding config response:", err)
		return
	}

	if cfgResp.Hash == lastHash {
		log.Println("Config has not changed, skipping update")
		return
	}

	log.Printf("Config has changed, updating config file with hash: %s", cfgResp.Hash)
	if err := os.WriteFile(configFile, []byte(cfgResp.Content), 0644); err != nil {
		log.Printf("Error writing config file: %v", err)
		return
	}

	if err := signalApp(pidFile); err != nil {
		log.Printf("Error sending SIGHUP to app: %v", err)
		return
	}

	log.Println("Sent SIGHUP to app")
	lastHash = cfgResp.Hash
}

func signalApp(pidFile string) error {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return err
	}

	return syscall.Kill(pid, syscall.SIGHUP)
}
