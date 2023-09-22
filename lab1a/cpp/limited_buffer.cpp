#include <vector>
#include <mutex>
#include <condition_variable>

#include "common.h"

class LimitedBuffer {
	std::vector<DataEntry> items;
	int capacity;
	std::mutex items_mutex;
	std::condition_variable items_cv;

	int reserved_count;
public:
	int estimated_count;

	LimitedBuffer(int capacity) {
		this->capacity = capacity;
		estimated_count = 0;
		reserved_count = 0;
	}

	void insert(DataEntry entry);
	DataEntry remove();

	bool reserve_entry();
};

void LimitedBuffer::insert(DataEntry entry) {
	std::unique_lock<std::mutex> guard(items_mutex);
	items_cv.wait(guard, [&](){ return items.size() < this->capacity; });

	items.push_back(entry);
	items_cv.notify_all();
}

DataEntry LimitedBuffer::remove() {
	std::unique_lock<std::mutex> guard(items_mutex);
	items_cv.wait(guard, [&](){ return items.size() > 0; });

	DataEntry entry = items[items.size()-1];
	items.pop_back();
	items_cv.notify_all();
	return entry;
}

bool LimitedBuffer::reserve_entry() {
	std::unique_lock<std::mutex> guard(items_mutex);

	if (reserved_count == estimated_count) {
		return false;
	} else {
		reserved_count++;
		return true;
	}
}
