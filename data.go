package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Result struct {
	Size            int
	WriteTime       time.Duration
	ReadTime        time.Duration
	LatencyTime     time.Duration
	WriteThroughput float64
	ReadThroughput  float64
}

type megaByte int

const (
	megaByteFloat   float64  = 1024 * 1024
	megaByteInt     megaByte = 1024 * 1024
	tenMegabyte     megaByte = 10 * 1024 * 1024
	hundredMegabyte megaByte = 100 * 1024 * 1024
	gigaByte        megaByte = 1000 * 1024 * 1024
)

func runAES(sizes []int, filename string) []Result {
	numCpus := runtime.NumCPU()
	result := make([]Result, len(sizes))
	key := make([]byte, 32) // 256bit key
	rand.Read(key)

	for i, size := range sizes {
		data := make([]byte, size)
		rand.Read(data)

		latencyTime, err := latency(data, key, filename)
		if err != nil {
			fmt.Printf("Error measuring latency: %v\n", err)
			continue
		}

		coreChunks := spiltCores(data, numCpus)

		var wg sync.WaitGroup
		writeResultsChannel := make(chan writeResults, numCpus)

		writeStart := time.Now()
		for w := 0; w < numCpus; w++ {
			wg.Add(1)
			go writeConcurrent(coreChunks[w], key, &wg, writeResultsChannel)
		}
		wg.Wait()
		totalWriteTime := time.Since(writeStart)

		for w := 0; w < numCpus; w++ {
			writeResult := <-writeResultsChannel
			if writeResult.Error != nil {
				fmt.Printf("Write error: %v\n", writeResult.Error)
			}
		}

		readFiles := make([]string, numCpus)
		for w := 0; w < numCpus; w++ {
			readFiles[w] = fmt.Sprintf("%s_read_%d", filename, w)
			_, _, err := write(coreChunks[w], key, readFiles[w])
			if err != nil {
				fmt.Printf("Error creating read test file %d: %v\n", w, err)
				continue
			}
			defer os.Remove(readFiles[w])
		}

		readResultsChannel := make(chan readResults, numCpus)
		readStart := time.Now()

		for w := 0; w < numCpus; w++ {
			wg.Add(1)
			go readConcurrent(readFiles[w], key, &wg, readResultsChannel)
		}
		wg.Wait()
		totalReadTime := time.Since(readStart)

		for w := 0; w < numCpus; w++ {
			readResult := <-readResultsChannel
			if readResult.Error != nil {
				fmt.Printf("Read error: %v\n", readResult.Error)
			}
		}

		writeThroughput := float64(size) / totalWriteTime.Seconds() / megaByteFloat
		readThroughput := float64(size) / totalReadTime.Seconds() / megaByteFloat

		if totalWriteTime == 0 {
			writeThroughput = 0
		}
		if totalReadTime == 0 {
			readThroughput = 0
		}

		result[i] = Result{
			Size:            size,
			WriteTime:       totalWriteTime,
			ReadTime:        totalReadTime,
			LatencyTime:     latencyTime,
			WriteThroughput: writeThroughput,
			ReadThroughput:  readThroughput,
		}
	}
	return result
}

func testAES() {
	fmt.Println("Welcome to Gopeed! A basic encryption file transfer benchmark using AES.")
	file := "encryption_test.txt"
	defer os.Remove(file)
	sizes := []int{
		int(megaByteInt),
		int(tenMegabyte),
		int(hundredMegabyte),
		int(gigaByte),
	}
	result := runAES(sizes, file)

	header := fmt.Sprintf("%-12s | %-12s | %-11s | %-12s", "Size (MB)", "Write (MB/s)", "Read (MB/s)", "Latency (ms)")
	separator := strings.Repeat("-", len(header))

	fmt.Println(separator)
	fmt.Println(header)
	fmt.Println(separator)

	for _, r := range result {
		sizeMB := r.Size / int(megaByteInt)
		var sizeStr string
		switch {
		case sizeMB < 1.0:
			sizeStr = fmt.Sprintf("%d", sizeMB)
		case sizeMB < 10.0:
			sizeStr = fmt.Sprintf("%d", sizeMB)
		default:
			sizeStr = fmt.Sprintf("%d", sizeMB)
		}

		fmt.Printf("%-12s | %-12.2f | %-11.2f | %-12.2f\n",
			sizeStr,
			r.WriteThroughput,
			r.ReadThroughput,
			float64(r.LatencyTime.Microseconds())/1000.0)
	}
	fmt.Println(separator)
}
