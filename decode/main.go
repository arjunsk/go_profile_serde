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
	logFilePath := "decode/out.log"
	outputFilePath := "heap.pprof"

	file, err := os.Open(logFilePath)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer file.Close()

	var base64Heap []byte
	reader := bufio.NewReader(file)

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

		if logEntry.Msg != "" && logEntry.Caller == "dbutils/mem.go:124" {
			fmt.Println("heap profile chunk found")
			data, err := base64.RawStdEncoding.DecodeString(logEntry.Msg)
			if err != nil {
				fmt.Printf("Error decoding base64: %v\n", err)
				panic(err)
			}
			base64Heap = append(base64Heap, data...)
		}
	}

	err = os.WriteFile(outputFilePath, base64Heap, 0644)
	if err != nil {
		fmt.Printf("Error writing heap.pprof file: %v\n", err)
		return
	}

	fmt.Printf("heap.pprof file has been reconstructed and saved to %s\n", outputFilePath)
}
