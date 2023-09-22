package main

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
