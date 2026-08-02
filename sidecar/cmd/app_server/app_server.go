package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	path := flag.String("config-file", "/shared/config.json", "Path to config file")
	pidFile := flag.String("pid-file", "/shared/app.pid", "Path to PID file of the main app")
	startupWait := flag.Duration("startup-wait", 30*time.Second, "How long to wait for the initial config file")
	flag.Parse()

	if err := writePID(*pidFile); err != nil {
		log.Fatalf("Failed to write PID file: %v", err)
	}

	if err := waitForConfig(*path, *startupWait); err != nil {
		log.Fatalf("Config file was not ready: %v", err)
	}

	loadConfig(*path)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)

	for {
		<-sigChan
		log.Println("Received SIGHUP signal, reloading config...")
		loadConfig(*path)
	}
}

func writePID(path string) error {
	pid := os.Getpid()
	return os.WriteFile(path, []byte(fmt.Sprintf("%d\n", pid)), 0644)
}

func waitForConfig(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		if _, err := os.Stat(path); err == nil {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for %s", path)
		}

		log.Printf("Waiting for config file %s", path)
		time.Sleep(1 * time.Second)
	}
}

func loadConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Error reading config file: %v", err)
		return
	}

	log.Printf("Loaded config: %s", string(data))
}
