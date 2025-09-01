package main

import (
	"encoding/binary"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ncw/directio"
)

type writeResults struct {
	Duration  time.Duration
	bytesTime float64
	Error     error
}

type readResults struct {
	Duration  time.Duration
	bytesTime float64
	data      []byte
	Error     error
}

var fileCounter int64

/* Previous implementations relied on the page cache which slowed down performance tremendously. DirectIO bypasses the page cache, allowing for more efficient I/O operations.
I also fixed an issues with data alignment and padding. */

func alignDataForDirectIO(data []byte) []byte {
	blockSize := directio.BlockSize
	header := make([]byte, 8) // 64bit systems

	dataLength := uint64(len(data))
	binary.LittleEndian.PutUint64(header, dataLength)

	dataWithHeader := append(header, data...)
	dataWithHeaderLen := len(dataWithHeader)

	requiredSize := ((dataWithHeaderLen + blockSize - 1) / blockSize) * blockSize
	alignedData := directio.AlignedBlock(requiredSize)

	copy(alignedData, dataWithHeader)

	for i := dataWithHeaderLen; i < requiredSize; i++ {
		alignedData[i] = 0
	}

	return alignedData
}

func extractDataFromAligned(alignedData []byte) []byte {
	originalLength := binary.LittleEndian.Uint64(alignedData[:8])

	if len(alignedData) < int(8+originalLength) {
		return alignedData[8:]
	}

	return alignedData[8 : 8+originalLength]
}

func write(data []byte, key []byte, file string) (time.Duration, float64, error) {
	start := time.Now()

	encryptedData, err := encrypt(data, key)
	if err != nil {
		return 0, 0, err
	}

	alignedData := alignDataForDirectIO(encryptedData)

	oFile, err := directio.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return 0, 0, err
	}
	defer oFile.Close()

	_, err = oFile.Write(alignedData)
	if err != nil {
		return 0, 0, err
	}

	err = oFile.Sync()
	if err != nil {
		return 0, 0, err
	}

	duration := time.Since(start)
	bytesTime := float64(len(encryptedData))

	return duration, bytesTime, nil
}

func writeConcurrent(data []byte, key []byte, wg *sync.WaitGroup, resultChannel chan<- writeResults) {
	defer wg.Done()

	start := time.Now()

	encryptedData, err := encrypt(data, key)
	if err != nil {
		resultChannel <- writeResults{Error: err}
		return
	}

	counter := atomic.AddInt64(&fileCounter, 1)
	writtenFiles := "tmp" + strconv.FormatInt(counter, 10)

	alignedData := alignDataForDirectIO(encryptedData)

	f, err := directio.OpenFile(writtenFiles, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		resultChannel <- writeResults{Error: err}
		return
	}

	defer func() {
		f.Close()
		os.Remove(writtenFiles)
	}()

	_, err = f.Write(alignedData)
	if err != nil {
		resultChannel <- writeResults{Error: err}
		return
	}

	err = f.Sync()
	if err != nil {
		resultChannel <- writeResults{Error: err}
		return
	}

	duration := time.Since(start)
	bytesTime := float64(len(encryptedData))

	resultChannel <- writeResults{
		Duration:  duration,
		bytesTime: bytesTime,
		Error:     nil,
	}
}

func read(file string, key []byte) (time.Duration, float64, []byte, error) {
	start := time.Now()

	oFile, err := directio.OpenFile(file, os.O_RDONLY, 0)
	if err != nil {
		return 0, 0, nil, err
	}
	defer oFile.Close()

	stat, err := oFile.Stat()
	if err != nil {
		return 0, 0, nil, err
	}

	fileSize := int(stat.Size())
	alignedBuffer := directio.AlignedBlock(fileSize)

	_, err = oFile.Read(alignedBuffer)
	if err != nil {
		return 0, 0, nil, err
	}

	encryptedData := extractDataFromAligned(alignedBuffer)
	if encryptedData == nil {
		return 0, 0, nil, err
	}

	decryptedData, err := decrypt(encryptedData, key)
	if err != nil {
		return 0, 0, nil, err
	}

	duration := time.Since(start)
	bytesTime := float64(len(encryptedData))

	return duration, bytesTime, decryptedData, nil
}

func readConcurrent(file string, key []byte, wg *sync.WaitGroup, resultChannel chan<- readResults) {
	defer wg.Done()

	start := time.Now()

	oFile, err := directio.OpenFile(file, os.O_RDONLY, 0)
	if err != nil {
		resultChannel <- readResults{Error: err}
		return
	}
	defer oFile.Close()

	stat, err := oFile.Stat()
	if err != nil {
		resultChannel <- readResults{Error: err}
		return
	}

	fileSize := int(stat.Size())
	alignedBuffer := directio.AlignedBlock(fileSize)

	_, err = oFile.Read(alignedBuffer)
	if err != nil {
		resultChannel <- readResults{Error: err}
		return
	}

	encryptedData := extractDataFromAligned(alignedBuffer)
	if encryptedData == nil {
		resultChannel <- readResults{Error: err}
		return
	}

	decryptedData, err := decrypt(encryptedData, key)
	if err != nil {
		resultChannel <- readResults{Error: err}
		return
	}

	duration := time.Since(start)
	bytesTime := float64(len(encryptedData))

	resultChannel <- readResults{
		Duration:  duration,
		bytesTime: bytesTime,
		data:      decryptedData,
		Error:     nil,
	}
}

func latency(data []byte, key []byte, file string) (time.Duration, error) {
	start := time.Now()

	_, _, err := write(data, key, file)
	if err != nil {
		return 0, err
	}

	_, _, _, err = read(file, key)
	if err != nil {
		return 0, err
	}

	return time.Since(start), nil
}
