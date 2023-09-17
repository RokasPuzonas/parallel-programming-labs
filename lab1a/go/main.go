package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"time"
)

const inputBufferSize = 10
const inputMonitorCount = 3

const outputBufferSize = 10
const outputMonitorCount = 1

type DataEntry struct {
	Name string `json:"name"`
	Sugar float32 `json:"sugar"`
	Criteria int `json:"criteria"`
}

func filterEntry(entry DataEntry) bool {
	time.Sleep(time.Millisecond * 100 + time.Millisecond * time.Duration(entry.Criteria))
	return entry.Sugar > float32(entry.Criteria)
}

func outputEntry(outputEntries *[]DataEntry, entry DataEntry) {
	time.Sleep(time.Millisecond * 200)
	fmt.Println("Output:", entry)
	*outputEntries = append(*outputEntries, entry)
	sort.Slice(*outputEntries, func(i, j int) bool {
		return (*outputEntries)[i].Sugar < (*outputEntries)[j].Sugar
	})
}

func filterEntries(inputBuffer *LimitedBuffer, outputBuffer *LimitedBuffer) {
	for inputBuffer.ConsumeEstimatedEntry() {
		entry := inputBuffer.Remove()
		fmt.Println("Started to filter:", entry)
		isFiltered := filterEntry(entry)
		fmt.Println("Finished to filter:", entry)
		if (isFiltered) {
			outputBuffer.UpdateEstimated(+1)
			outputBuffer.Insert(entry)
		}

		inputBuffer.MarkAsProcessed()
	}
}

func outputEntries(outputBuffer *LimitedBuffer, inputBuffer *LimitedBuffer, outputList *[]DataEntry, outputMutex *sync.Mutex) {
	for !inputBuffer.IsDone() {
		for outputBuffer.ConsumeEstimatedEntry() {
			entry := outputBuffer.Remove()
			outputMutex.Lock()
			fmt.Println("Started to output:", entry)
			outputEntry(outputList, entry)
			fmt.Println("Finished to output:", entry)
			outputMutex.Unlock()

			outputBuffer.MarkAsProcessed()
		}
	}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage:", os.Args[0], "<data-file> <output-file>")
		os.Exit(-1)
	}

	dataFilename := os.Args[1]
	outputFilename := os.Args[2]
	fileData, err := os.ReadFile(dataFilename)
	if err != nil {
		log.Fatal(err)
	}

	entries := []DataEntry{}
	err = json.Unmarshal(fileData, &entries)
	if err != nil {
		log.Fatal(err)
	}

	outputList := []DataEntry{}

	inputBuffer := makeLimitedBuffer(inputBufferSize)
	outputBuffer := makeLimitedBuffer(outputBufferSize)
	outputMutex := sync.Mutex{}
	inputBuffer.UpdateEstimated(len(entries))

	for i := 0; i < inputMonitorCount; i++ {
		go filterEntries(&inputBuffer, &outputBuffer)
	}
	for i := 0; i < outputMonitorCount; i++ {
		go outputEntries(&outputBuffer, &inputBuffer, &outputList, &outputMutex)
	}

	for _, entry := range entries {
		inputBuffer.Insert(entry);
	}
	inputBuffer.WaitUntilDone()
	outputBuffer.WaitUntilDone()

	outputFile, err := os.OpenFile(outputFilename, os.O_TRUNC | os.O_CREATE | os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	outputListBytes, err := json.MarshalIndent(outputList, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	outputFile.Write(outputListBytes)

	fmt.Println("Finished")
}
