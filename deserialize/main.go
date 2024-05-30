package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
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
	logFilePath := "deserialize/out2.log"
	outputFilePath := "heap.pprof"

	// 1. Open log file
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
		// 2. Read log file line by line
		line, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() != "EOF" {
				fmt.Printf("Error reading log file: %v\n", err)
			}
			break
		}

		// 3. Parse log entry from JSON string
		var logEntry LogEntry
		err = json.Unmarshal([]byte(line), &logEntry)
		if err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
			panic(err)
		}

		// 4. Mark the beginning of heap profile
		if !start {
			if strings.HasPrefix(logEntry.Msg, "H4") {
				// start of heap profile
				start = true
			} else {
				continue
			}
		}

		// 5. Extract heap profile data (start, end).
		// Mostly observed that heap profile base64 encoded data starts with "H4".
		if logEntry.Msg != "" && strings.HasPrefix(logEntry.Caller, "dbutils/mem.go") {
			fmt.Println("heap profile chunk found")
			data, err := base64.RawStdEncoding.DecodeString(logEntry.Msg)
			if err != nil {
				fmt.Printf("Error decoding base64: %v\n", err)
				panic(err)
			}
			logBytes = append(logBytes, data...)
		}

		// 6. We print MLimit only in the last block of heap profile
		if logEntry.Mlimit != "" {
			// end of heap profile
			break
		}
	}

	// 7. Write heap profile to file
	err = os.WriteFile(outputFilePath, logBytes, 0644)
	if err != nil {
		fmt.Printf("Error writing heap.pprof file: %v\n", err)
		return
	}

	fmt.Printf("heap.pprof file has been reconstructed and saved to %s\n", outputFilePath)
}
