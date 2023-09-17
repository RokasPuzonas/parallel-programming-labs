package main

import "sync"

type LimitedBuffer struct {
	items       []DataEntry
	readIndex   int
	writeIndex  int
	currentSize int
	mutex       *sync.Mutex
	updateCond  *sync.Cond

	estimatedSize int
	estimatedSizeMutex *sync.Mutex

	consumedCount int
	consumedMutex *sync.Mutex

	processedCount int
	processedMutex *sync.Mutex
}

func makeLimitedBuffer(bufferSize int) LimitedBuffer {
	var container = make([]DataEntry, bufferSize)
	var estimatedSizeMutex = sync.Mutex{}
	var processedMutex = sync.Mutex{}
	var consumedMutex = sync.Mutex{}

	var itemsMutex = sync.Mutex{}
	var cond = sync.NewCond(&itemsMutex)
	return LimitedBuffer{
		items: container,
		mutex: &itemsMutex,
		updateCond: cond,
		estimatedSizeMutex: &estimatedSizeMutex,
		processedMutex: &processedMutex,
		consumedMutex: &consumedMutex,
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

func (buffer *LimitedBuffer) IsDone() bool {
	return buffer.estimatedSize == buffer.processedCount
}

func (buffer *LimitedBuffer) WaitUntilDone() {
	for !buffer.IsDone() {}
}

func (buffer *LimitedBuffer) ConsumeEstimatedEntry() bool {
	buffer.estimatedSizeMutex.Lock()
	defer buffer.estimatedSizeMutex.Unlock()

	if (buffer.estimatedSize == buffer.consumedCount) {
		return false
	} else {
		buffer.consumedCount += 1
		return true
	}
}

func (buffer *LimitedBuffer) IsEstimatedEmpty() bool {
	buffer.estimatedSizeMutex.Lock()
	defer buffer.estimatedSizeMutex.Unlock()

	return buffer.estimatedSize == 0
}

func (buffer *LimitedBuffer) UpdateEstimated(change int) {
	buffer.estimatedSizeMutex.Lock()
	defer buffer.estimatedSizeMutex.Unlock()

	buffer.estimatedSize += change
}

func (buffer *LimitedBuffer) MarkAsProcessed() {
	buffer.processedMutex.Lock()
	defer buffer.processedMutex.Unlock()

	buffer.processedCount += 1
}
