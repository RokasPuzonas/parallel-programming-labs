package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

const bufferSize = 10
const threadCount = 3

type DataEntry struct {
	Name string `json:"name"`
	Sugar float32 `json:"sugar"`
	Criteria int `json:"criteria"`
}

func processEntry(entry DataEntry, outputEntries *SortedList, outputMutex *sync.Mutex) {
	time.Sleep(time.Millisecond * 100 + time.Millisecond * time.Duration(entry.Criteria))
	if entry.Sugar > float32(entry.Criteria) {
		outputMutex.Lock()
		fmt.Println("Output:", entry)
		outputEntries.Append(entry)
		outputMutex.Unlock()
	}
}

func processEntries(monitor *LimitedBuffer, outputEntries *SortedList, outputMutex *sync.Mutex, waitGroup *sync.WaitGroup) {
	for monitor.ReserveEntry() {
		entry := monitor.Remove()
		fmt.Println("Started to filter:", entry)
		processEntry(entry, outputEntries, outputMutex)
		fmt.Println("Finished to filter:", entry)
	}
	waitGroup.Done()
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

	outputList := makeSortedList(len(entries))
	outputMutex := sync.Mutex{}
	waitGroup := sync.WaitGroup{}

	monitor := makeLimitedBuffer(bufferSize)
	monitor.estimatedCount = len(entries)

	for i := 0; i < threadCount; i++ {
		go processEntries(&monitor, &outputList, &outputMutex, &waitGroup)
		waitGroup.Add(1)
	}

	for _, entry := range entries {
		monitor.Insert(entry);
	}
	waitGroup.Wait()

	outputFile, err := os.OpenFile(outputFilename, os.O_TRUNC | os.O_CREATE | os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	outputListBytes, err := json.MarshalIndent(outputList.GetSlice(), "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	outputFile.Write(outputListBytes)

	fmt.Println("Finished")
}
