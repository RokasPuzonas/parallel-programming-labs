package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
)

type DataEntry struct {
	Name string `json:"name"`
	Sugar float32 `json:"sugar"`
	Criteria int `json:"criteria"`
}

var PossibleNames = []string{
	"Cake",
	"Candy",
	"Apple",
	"Fruit juice",
	"Chocolate",
	"Tea",
	"Smoothie",
	"Ice cream",
};

func generateEntry(percent float64) DataEntry {
	entry := DataEntry{}
	nameIndex := rand.Uint32() % uint32(len(PossibleNames))
	entry.Name = fmt.Sprintf("%s %d", PossibleNames[nameIndex], rand.Uint32() % 100)
	entry.Sugar = rand.Float32() * 100 + 100
	if rand.Float64() < percent {
		entry.Criteria = int(entry.Sugar - rand.Float32() * 50 - 10)
	} else {
		entry.Criteria = int(entry.Sugar + rand.Float32() * 50 + 10)
	}
	return entry
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage:", os.Args[0], "<data-file> <count> <filter-percent>")
		os.Exit(-1)
	}

	ouputFilename := os.Args[1]
	count, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	filterPercent, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		log.Fatal(err)
	}

	outputFile, err := os.OpenFile(ouputFilename, os.O_TRUNC | os.O_CREATE | os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	outputFile.WriteString("[\n")
	for i := 0; i < count; i++ {
		entry := generateEntry(filterPercent)
		entry_bytes, err := json.Marshal(entry)
		if err != nil {
			log.Println(err)
			continue
		}
		outputFile.WriteString("\t");
		outputFile.Write(entry_bytes)
		if (i < count-1) {
			outputFile.WriteString(",");
		}
		outputFile.WriteString("\n");
	}
	outputFile.WriteString("]")
}
