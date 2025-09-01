package main

func spiltCores(data []byte, numParts int) [][]byte {
	partSize := len(data) / numParts
	chunks := make([][]byte, 0, numParts)

	for i := 0; i < numParts; i++ {
		start := i * partSize
		end := start + partSize

		if i == numParts-1 {
			end = len(data)
		}
		chunks = append(chunks, data[start:end])
	}
	return chunks
}
