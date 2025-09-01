package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"runtime"
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
	megaByteF       float64  = 1024 * 1024
	regular         megaByte = 1024 * 1024
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

		numWorkers := min(numCpus, 8)
		coreChunks := spiltCores(data, numWorkers)

		var wg sync.WaitGroup
		writeResultsChannel := make(chan writeResults, numWorkers)

		writeStart := time.Now()
		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go writeConcurrent(coreChunks[w], key, &wg, writeResultsChannel)
		}
		wg.Wait()
		totalWriteTime := time.Since(writeStart)

		for w := 0; w < numWorkers; w++ {
			writeResult := <-writeResultsChannel
			if writeResult.Error != nil {
				fmt.Printf("Write error: %v\n", writeResult.Error)
			}
		}

		readFiles := make([]string, numWorkers)
		for w := 0; w < numWorkers; w++ {
			readFiles[w] = fmt.Sprintf("%s_read_%d", filename, w)
			_, _, err := write(coreChunks[w], key, readFiles[w])
			if err != nil {
				fmt.Printf("Error creating read test file %d: %v\n", w, err)
				continue
			}
			defer os.Remove(readFiles[w])
		}

		readResultsChannel := make(chan readResults, numWorkers)
		readStart := time.Now()

		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go readConcurrent(readFiles[w], key, &wg, readResultsChannel)
		}
		wg.Wait()
		totalReadTime := time.Since(readStart)

		for w := 0; w < numWorkers; w++ {
			readResult := <-readResultsChannel
			if readResult.Error != nil {
				fmt.Printf("Read error: %v\n", readResult.Error)
			}
		}

		writeThroughput := float64(size) / totalWriteTime.Seconds() / megaByteF
		readThroughput := float64(size) / totalReadTime.Seconds() / megaByteF

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
		int(regular),
		int(tenMegabyte),
		int(hundredMegabyte),
		int(gigaByte),
	}
	result := runAES(sizes, file)

	fmt.Println("Size (bytes) | Write (MB/s) | Read (MB/s) | Latency (ms)")
	for _, r := range result {
		fmt.Printf("%-13d | %-12.2f | %-11.2f | %-12.2f\n",
			r.Size,
			r.WriteThroughput,
			r.ReadThroughput,
			float64(r.LatencyTime.Microseconds())/1000.0)
	}
}
