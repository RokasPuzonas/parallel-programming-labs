use crate::common::DataEntry;

pub struct LimitedBuffer {
	items: Vec<DataEntry>,
	capacity: usize,

	estimated_size: i32,
	consumed_count: u32,
	processed_count: u32
}

impl LimitedBuffer {
	pub fn new(capacity: usize) -> LimitedBuffer {
		return LimitedBuffer{
			items: Vec::with_capacity(capacity),
			capacity,
			estimated_size: 0,
			consumed_count: 0,
			processed_count: 0
		}
	}

	pub fn try_insert(&mut self, entry: DataEntry) -> bool {
		if self.items.len() == self.capacity {
			return false
		}
		self.items.push(entry);
		return true
	}

	pub fn remove(&mut self) -> Option<DataEntry> {
		if self.items.is_empty() {
			None
		} else {
			Some(self.items.remove(0))
		}
	}

	pub fn is_done(&self) -> bool {
		return (self.estimated_size as u32) == self.processed_count
	}

	pub fn consume_estimated(&mut self) -> bool {
		if (self.estimated_size as u32) == self.consumed_count {
			return false;
		} else {
			self.consumed_count += 1;
			return true;
		}
	}

	pub fn update_estimated(&mut self, change: i32) {
		self.estimated_size += change
	}

	pub fn mark_processed(&mut self) {
		self.processed_count += 1
	}
}
