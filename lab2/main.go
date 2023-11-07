package main;

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

const processThreadCount = 3

type DataEntry struct {
	Name string `json:"name"`
	Sugar float32 `json:"sugar"`
	Criteria int `json:"criteria"`
}

type SortedList struct {
	items []DataEntry
	count int
}

func makeSortedList(capacity int) SortedList {
	var items = make([]DataEntry, capacity)
	return SortedList{
		items: items,
	}
}

func (self *SortedList) Append(item DataEntry) {
	if len(self.items) == self.count {
		panic("SortedList is full")
	}

	self.items[self.count] = item
	self.count++
	for i := self.count-1; i >= 1; i-- {
		if self.items[i].Sugar > self.items[i-1].Sugar { break }

		self.items[i], self.items[i-1] = self.items[i-1], self.items[i] // Swap elements [i] and [i-1]
	}
}

func (self *SortedList) GetSlice() []DataEntry {
	return self.items[:self.count]
}

func processEntry(entry DataEntry, output chan DataEntry) {
	time.Sleep(time.Millisecond * 100 + time.Millisecond * time.Duration(entry.Criteria))
	if entry.Sugar > float32(entry.Criteria) {
		fmt.Println("Output:", entry)
		output <- entry;
	}
}

func processEntries(input chan DataEntry, output chan DataEntry, waitGroup *sync.WaitGroup) {
	for entry := range input {
		fmt.Println("Started to filter:", entry)
		processEntry(entry, output)
		fmt.Println("Finished to filter:", entry)
	}
	waitGroup.Done();
}

func collectResults(channel chan DataEntry, output *SortedList, waitGroup *sync.WaitGroup) {
	for entry := range channel {
		output.Append(entry)
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
	waitGroup := sync.WaitGroup{}
	outputWaitGroup := sync.WaitGroup{}
	inputChannel := make(chan DataEntry);
	outputChannel := make(chan DataEntry);

	for i := 0; i < processThreadCount; i++ {
		go processEntries(inputChannel, outputChannel, &waitGroup)
		waitGroup.Add(1)
	}
	outputWaitGroup.Add(1)
	go collectResults(outputChannel, &outputList, &outputWaitGroup)

	for _, entry := range entries {
		inputChannel <- entry
	}
	close(inputChannel)
	waitGroup.Wait()
	close(outputChannel)
	outputWaitGroup.Wait()

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

	fmt.Printf("Finished, %d entries\n", outputList.count)
}
