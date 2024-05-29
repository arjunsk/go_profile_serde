package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"os"
	"runtime/debug"
	"runtime/pprof"
)

func main() {
	heapp := pprof.Lookup("heap")
	buf := &bytes.Buffer{}
	heapp.WriteTo(buf, 0)
	_ = debug.SetMemoryLimit(-1)
	logBytes := buf.Bytes()

	// 1. Copy the byte slice
	logBytesCopy := make([]byte, len(logBytes))
	copy(logBytesCopy, logBytes)

	// 2.a Encode the entire byte slice to base64 using StdEncoding
	encode1 := base64.RawStdEncoding.EncodeToString(logBytes)

	// 2.b Encode in chunks of 50 bytes using RawStdEncoding
	var encode2 []string
	chunkSize := 50
	for len(logBytes) > chunkSize {
		chunk := base64.RawStdEncoding.EncodeToString(logBytes[:chunkSize])
		logutil.Info(chunk)
		encode2 = append(encode2, chunk)
		logBytes = logBytes[chunkSize:]
	}
	logutil.Info(base64.RawStdEncoding.EncodeToString(logBytes))
	encode2 = append(encode2, base64.RawStdEncoding.EncodeToString(logBytes))

	// 3.a Decode the base64-encoded string
	decode1, _ := base64.RawStdEncoding.DecodeString(encode1)

	// 3.b Decode the base64-encoded strings
	var decode2 []byte
	for _, s := range encode2 {
		data, _ := base64.RawStdEncoding.DecodeString(s)
		decode2 = append(decode2, data...)
	}

	// 4. Print the lengths and equality check
	fmt.Printf("decode1 == decode2 %v\n", bytes.Equal(decode1, decode2))
	fmt.Printf("logBytesCopy==decode2 %v\n", bytes.Equal(logBytesCopy, decode2))

	// 5. Write OS file
	err := os.WriteFile("heap.pprof", decode1, 0644)
	if err != nil {
		panic(err)
	}
	// go tool pprof -http=:8080 heap.pprof
}
