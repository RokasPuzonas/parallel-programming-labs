use std::{env::args, process::exit, thread, time::Duration, sync::{Arc, Mutex}};

use anyhow::Result;

mod limited_buffer;
mod common;

use common::DataEntry;
use limited_buffer::LimitedBuffer;

// This was a mistake

const INPUT_BUFFER_SIZE: usize = 2;
const INPUT_MONITOR_COUNT: i32 = 2;

const OUTPUT_BUFFER_SIZE: usize = 10;
const OUTPUT_MONITOR_COUNT: i32 = 1;

fn filter_entry(entry: &DataEntry) -> bool {
	thread::sleep(Duration::from_millis(100 + entry.criteria as u64));
	return entry.sugar > entry.criteria as f32
}

fn output_entry(output_list: &mut Vec<DataEntry>, entry: DataEntry) {
	thread::sleep(Duration::from_millis(200));
	output_list.push(entry);
	output_list.sort_by(|a, b| a.sugar.total_cmp(&b.sugar))
}

fn filter_entries(input: Arc<Mutex<LimitedBuffer>>, output: Arc<Mutex<LimitedBuffer>>) {
	while !input.lock().unwrap().is_done() {
		if input.lock().unwrap().consume_estimated() {
			let entry = input.lock().unwrap().remove();
			if entry.is_none() { continue; }
			let entry = entry.unwrap();

			println!("Started to filter: {}", entry.name);
			let is_filtered = filter_entry(&entry);
			println!("Finished to filter: {}", entry.name);

			if is_filtered {
				// output.lock().unwrap().update_estimated(1);
				// while !output.lock().unwrap().try_insert(entry) {}
			}

			input.lock().unwrap().mark_processed()
		}
	}
}

fn ouput_entries(input: Arc<Mutex<LimitedBuffer>>, output: Arc<Mutex<LimitedBuffer>>, output_list: Arc<Mutex<Vec<DataEntry>>>) {
	while !input.lock().unwrap().is_done() {
		let entry = input.lock().unwrap().remove();
		if entry.is_none() { continue; }
		let entry = entry.unwrap();
		output.lock().unwrap().consume_estimated();

		let entry_name = entry.name.clone();
		println!("Started to output: {}", entry_name);
		{
			let mut output_list = output_list.lock().unwrap();
			output_entry(&mut output_list, entry);
		}
		println!("Finished to output: {}", entry_name);

		output.lock().unwrap().mark_processed()
	}
}

fn main() -> Result<()> {
	let args = args().collect::<Vec<_>>();
	if args.len() != 3 {
		eprintln!("Usage: {} <data-file> <output-file>", args[0]);
		exit(-1)
	}

	let entries = vec![
		DataEntry{ name: "foo1".into(), sugar: 100.0, criteria: 100 },
		DataEntry{ name: "foo2".into(), sugar: 100.0, criteria: 100 },
		DataEntry{ name: "foo3".into(), sugar: 100.0, criteria: 100 },
		DataEntry{ name: "foo4".into(), sugar: 100.0, criteria: 100 },
		DataEntry{ name: "foo5".into(), sugar: 100.0, criteria: 100 }
	];

	let mut input_threads = vec![];
	// let mut output_threads = vec![];

	// let output_list = Arc::new(Mutex::new(Vec::new()));
	let input_buffer = Arc::new(Mutex::new(LimitedBuffer::new(INPUT_BUFFER_SIZE)));
	let output_buffer = Arc::new(Mutex::new(LimitedBuffer::new(OUTPUT_BUFFER_SIZE)));
	input_buffer.lock().unwrap().update_estimated(entries.len() as i32);

	for _ in 0..INPUT_MONITOR_COUNT {
		let input_buffer = input_buffer.clone();
		let output_buffer = output_buffer.clone();
		input_threads.push(thread::spawn(|| filter_entries(input_buffer, output_buffer)));
	}
	// for _ in 0..OUTPUT_MONITOR_COUNT {
	// 	let input_buffer = input_buffer.clone();
	// 	let output_buffer = output_buffer.clone();
	// 	let output_list = output_list.clone();
	// 	output_threads.push(thread::spawn(|| ouput_entries(input_buffer, output_buffer, output_list)));
	// }

	for entry in entries {
		input_buffer.lock().unwrap().try_insert(entry);
	}

	for handle in input_threads {
		handle.join().expect("Failed to join input thread");
	}
	// for handle in output_threads {
	// 	handle.join().expect("Failed to join output thread");
	// }

	println!("Finished");
	Ok(())
}
