#include <vector>
#include <mutex>
#include <condition_variable>

#include "common.h"

class LimitedBuffer {
	std::vector<DataEntry> items;
	int capacity;
	std::mutex items_mutex;
	std::condition_variable items_cv;

public:
	int estimated_size;
	int consumed_count;
	int processed_count;
	std::mutex processed_mutex;

	LimitedBuffer(int capacity) {
		this->capacity = capacity;
		estimated_size = 0;
		consumed_count = 0;
		processed_count = 0;
	}

	void insert(DataEntry entry);
	DataEntry remove();

	void update_estimated(int change);
	bool consume_estimated();
	void mark_processed();
	bool is_done();
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

void LimitedBuffer::update_estimated(int change) {
	std::unique_lock<std::mutex> guard(processed_mutex);
	estimated_size += change;
}

bool LimitedBuffer::consume_estimated() {
	std::unique_lock<std::mutex> guard(processed_mutex);

	if (consumed_count == estimated_size) {
		return false;
	} else {
		consumed_count++;
		return true;
	}
}

void LimitedBuffer::mark_processed() {
	std::unique_lock<std::mutex> guard(processed_mutex);
	processed_count++;
}

bool LimitedBuffer::is_done() {
	return estimated_size == processed_count;
}
