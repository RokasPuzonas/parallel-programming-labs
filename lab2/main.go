package main;

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

var Terminator = DataEntry{ Name: "Terminator", Sugar: -1, Criteria: -1 };

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

func dataThread(fromMain, toWorker chan DataEntry, fromWorkerRequest chan bool, arraySize int, workerCount int) {
	var gotTerminator = false
	var storage []DataEntry = make([]DataEntry, arraySize)
	var used = 0

	var appendtoStorage = func(entry DataEntry) {
		if (entry == Terminator) {
			gotTerminator = true
			return
		}

		storage[used] = entry
		used += 1
	}

	var removeFromStorage = func() DataEntry {
		var entry = storage[used-1]
		used -= 1
		return entry
	}

	for {
		if (used == arraySize) {
			<-fromWorkerRequest
			toWorker <- removeFromStorage()
		} else if (used == 0) {
			appendtoStorage(<-fromMain)
		} else {
			select {
			case entry := <-fromMain:
				appendtoStorage(entry)
			case <-fromWorkerRequest:
				toWorker <- removeFromStorage()
			}
		}

		if (gotTerminator) {
			for i := 0; i < used; i++ {
				<-fromWorkerRequest
				toWorker <- storage[i]
			}
			for i := 0; i < workerCount; i++ {
				<-fromWorkerRequest
				toWorker <- Terminator
			}
			break
		}
	}
}

func workerThread(fromData, toResult chan DataEntry, toDataRequest chan bool) {
	for {
		toDataRequest <-true
		var entry = <-fromData
		if entry == Terminator {
			toResult <-Terminator
			break
		}
		fmt.Println("Started to filter:", entry)
		processEntry(entry, toResult)
		fmt.Println("Finished to filter:", entry)
	}
}

func resultThread(fromWorker chan DataEntry, toMain chan SortedList, workerCount int, maxResults int) {
	result := makeSortedList(maxResults)
	finishedWorkers := 0

	for {
		if finishedWorkers == workerCount {
			break
		}

		var entry = <-fromWorker
		if entry == Terminator {
			finishedWorkers += 1
			continue
		}

		result.Append(entry)
	}

	toMain <-result
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

	mainToData          := make(chan DataEntry)
	dataToWorker        := make(chan DataEntry)
	workerToDataRequest := make(chan bool)
	workerToResult      := make(chan DataEntry)
	resultToMain        := make(chan SortedList)

	var workerCount = 5

	for i := 0; i < workerCount; i++ {
		go workerThread(dataToWorker, workerToResult, workerToDataRequest)
	}
	go dataThread(mainToData, dataToWorker, workerToDataRequest, 3, workerCount)
	go resultThread(workerToResult, resultToMain, workerCount, len(entries))

	for _, entry := range entries {
		mainToData <-entry
	}
	mainToData <-Terminator
	var result = <-resultToMain

	outputFile, err := os.OpenFile(outputFilename, os.O_TRUNC | os.O_CREATE | os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	outputListBytes, err := json.MarshalIndent(result.GetSlice(), "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	outputFile.Write(outputListBytes)

	fmt.Printf("Initial amount %d\n", len(entries))
	fmt.Printf("Finished, %d entries\n", result.count)
}
