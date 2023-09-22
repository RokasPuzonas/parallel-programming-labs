package main

import "sync"

type LimitedBuffer struct {
	items       []DataEntry
	readIndex   int
	writeIndex  int
	currentSize int
	mutex       *sync.Mutex
	updateCond  *sync.Cond

	estimatedCount int
	reservedCount int
}

func makeLimitedBuffer(bufferSize int) LimitedBuffer {
	var container = make([]DataEntry, bufferSize)

	var itemsMutex = sync.Mutex{}
	var cond = sync.NewCond(&itemsMutex)
	return LimitedBuffer{
		items: container,
		mutex: &itemsMutex,
		updateCond: cond,
	}
}

func (buffer *LimitedBuffer) Insert(item DataEntry) {
	buffer.mutex.Lock()
	for buffer.currentSize == len(buffer.items) {
		buffer.updateCond.Wait()
	}
	buffer.items[buffer.writeIndex] = item
	buffer.writeIndex = (buffer.writeIndex + 1) % len(buffer.items)
	buffer.currentSize++
	buffer.updateCond.Broadcast()
	buffer.mutex.Unlock()
}

func (buffer *LimitedBuffer) Remove() DataEntry {
	buffer.mutex.Lock()
	for buffer.currentSize == 0 {
		buffer.updateCond.Wait()
	}
	var item = buffer.items[buffer.readIndex]
	buffer.readIndex = (buffer.readIndex + 1) % len(buffer.items)
	buffer.currentSize--
	buffer.updateCond.Broadcast()
	buffer.mutex.Unlock()
	return item
}

func (buffer *LimitedBuffer) ReserveEntry() bool {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()

	if buffer.estimatedCount == buffer.reservedCount {
		return false
	} else {
		buffer.reservedCount++
		return true
	}
}
