package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"runtime/debug"
	"runtime/pprof"
)

func main() {

	//0. Do some heap allocation
	arr := make([][]int, 100)
	for i := 0; i < 100; i++ {
		arr[i] = make([]int, 100)
	}

	heapp := pprof.Lookup("heap")
	buf := &bytes.Buffer{}
	_ = heapp.WriteTo(buf, 0)
	_ = debug.SetMemoryLimit(-1)
	logBytes := buf.Bytes()

	// 1. Copy the byte slice
	logBytesCopy := make([]byte, len(logBytes))
	copy(logBytesCopy, logBytes)

	// 2.a Encode the entire byte slice to base64 using StdEncoding
	encode1 := base64.RawStdEncoding.EncodeToString(logBytes)

	// 2.b Encode in chunks of 50 bytes using RawStdEncoding
	chunkSize := 50
	encode2 := base64Chunk(logBytes, chunkSize)

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

// base64Chunk encodes the byte slice in chunks of chunkSize bytes
// This function is extracted from
// https://github.com/matrixorigin/matrixone/blob/dfa158c1073a3db6eccd5e7c1fd6b4541a744a73/pkg/vm/engine/tae/db/dbutils/mem.go#L111-L131
func base64Chunk(logBytes []byte, chunkSize int) []string {
	var encoded []string
	for len(logBytes) > chunkSize {
		encoded = append(encoded, base64.RawStdEncoding.EncodeToString(logBytes[:chunkSize]))
		logBytes = logBytes[chunkSize:]
	}
	encoded = append(encoded, base64.RawStdEncoding.EncodeToString(logBytes))
	return encoded
}
