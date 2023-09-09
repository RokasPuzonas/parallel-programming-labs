package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

const inputQueueSize = 10
const inputMonitorCount = 3

const outputQueueSize = 10
const outputMonitorCount = 3

type DataEntry struct {
	Name string `json:"name"`
	Sugar float32 `json:"sugar"`
	Criteria float32 `json:"criteria"`
}

func filterEntry(entry DataEntry) bool {
	time.Sleep(time.Millisecond * 100 + time.Millisecond * time.Duration(entry.Criteria))
	return entry.Sugar > entry.Criteria
}

func outputEntry(file *os.File, entry DataEntry) {
	time.Sleep(time.Millisecond * 200)
	fmt.Println("Output:", entry)
	file.WriteString(entry.Name)
	file.WriteString("\n")
}

func filterEntries(inputGroup, outputGroup *sync.WaitGroup, input <-chan DataEntry, output chan<- DataEntry) {
	for entry := range input {
		fmt.Println("Started to filter:", entry)
		isFiltered := filterEntry(entry)
		fmt.Println("Finished to filter:", entry)
		if (isFiltered) {
			outputGroup.Add(1)
			output <- entry
		}
		inputGroup.Done()
	}
}

func ouputEntries(file *os.File, group *sync.WaitGroup, channel <-chan DataEntry) {
	for entry := range channel {
		outputEntry(file, entry)
		group.Done()
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

	outputFile, err := os.OpenFile(outputFilename, os.O_TRUNC | os.O_CREATE | os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	inputChannel := make(chan DataEntry, inputQueueSize)
	outputChannel := make(chan DataEntry, outputQueueSize)

	var inputsGroups = sync.WaitGroup{}
	var ouputsGroups = sync.WaitGroup{}
	inputsGroups.Add(len(entries))

	for i := 0; i < inputMonitorCount; i++ {
		go filterEntries(&inputsGroups, &ouputsGroups, inputChannel, outputChannel)
	}
	for i := 0; i < outputMonitorCount; i++ {
		go ouputEntries(outputFile, &ouputsGroups, outputChannel)
	}

	for _, entry := range entries {
		inputChannel <- entry
	}
	close(inputChannel)
	inputsGroups.Wait()
	ouputsGroups.Wait()
	close(outputChannel)

	fmt.Println("Finished")
}
