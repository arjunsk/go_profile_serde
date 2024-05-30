package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Level  string `json:"level"`
	Time   string `json:"time"`
	Caller string `json:"caller"`
	Msg    string `json:"msg"`
	Mlimit string `json:"mlimit,omitempty"`
}

func main() {
	logFilePath := "deserialize/out.log"
	outputFilePath := "heap.pprof"

	file, err := os.Open(logFilePath)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer file.Close()

	var logBytes []byte
	reader := bufio.NewReader(file)
	start := false
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() != "EOF" {
				fmt.Printf("Error reading log file: %v\n", err)
			}
			break
		}

		var logEntry LogEntry
		err = json.Unmarshal([]byte(line), &logEntry)
		if err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
			panic(err)
		}

		if !start {
			if logEntry.Msg[:2] == "H4" {
				// start of heap profile
				start = true
			} else {
				continue
			}
		}

		if logEntry.Msg != "" && logEntry.Caller == "dbutils/mem.go:124" {
			fmt.Println("heap profile chunk found")
			data, err := base64.RawStdEncoding.DecodeString(logEntry.Msg)
			if err != nil {
				fmt.Printf("Error decoding base64: %v\n", err)
				panic(err)
			}
			logBytes = append(logBytes, data...)
		}
		if logEntry.Mlimit != "" {
			// end of heap profile
			break
		}
	}

	err = os.WriteFile(outputFilePath, logBytes, 0644)
	if err != nil {
		fmt.Printf("Error writing heap.pprof file: %v\n", err)
		return
	}

	fmt.Printf("heap.pprof file has been reconstructed and saved to %s\n", outputFilePath)
}
