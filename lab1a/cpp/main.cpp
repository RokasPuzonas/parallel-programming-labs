#include <algorithm>
#include <chrono>
#include <thread>
#include <iostream>
#include <fstream>
#include <string>
#include <vector>

#include "json.hpp"

using namespace std;
using json = nlohmann::json;

const int bufferSize = 10;
const int threadCount = 1;

#include "common.h"
#include "limited_buffer.cpp"

void processEntry(DataEntry *entry, vector<DataEntry> *entries, mutex *outputMutex) {
	this_thread::sleep_for(chrono::milliseconds(100 + entry->criteria));
	if (entry->sugar > entry->criteria) {
		outputMutex->lock();
		cout << "Output: " << entry->name << endl;
		// TODO: make this a sorted add
		entries->push_back(*entry);
		sort(entries->begin(), entries->end(), [](DataEntry &a, DataEntry &b) {
				return a.sugar < b.sugar;
		});
		outputMutex->unlock();
	}
}

void processEntries(LimitedBuffer *monitor, vector<DataEntry> *outputList, mutex *outputMutex) {
	while (monitor->reserve_entry()) {
		DataEntry entry = monitor->remove();
		cout << "Started to filter: " << entry.name << endl;
		processEntry(&entry, outputList, outputMutex);
		cout << "Finished to filter: " << entry.name << endl;
	}
}

int main(int argc, char **argv) {
	if (argc != 3) {
		cout << "Usage: " << argv[0] << " <data-file> <output-file>" << endl;
		return -1;
	}

	char *inputPath = argv[1];
	char *outputPath = argv[2];

	std::ifstream f(inputPath);
	json data = json::parse(f);

	vector<DataEntry> entries;
	for (auto it : data) {
		entries.push_back({
				.name = it["name"],
				.sugar = it["sugar"],
				.criteria = it["criteria"],
		});
	}

	vector<DataEntry> outputList;
	mutex outputListMutex;

	LimitedBuffer monitor(bufferSize);
	monitor.estimated_count = entries.size();

	vector<thread> threads;
	for (int i = 0; i < threadCount; i++) {
		threads.push_back(thread(processEntries, &monitor, &outputList, &outputListMutex));
	}

	for (int i = 0; i < entries.size(); i++) {
		monitor.insert(entries[i]);
	}
	for (int i = 0; i < threads.size(); i++) {
		threads[i].join();
	}

	std::ofstream outputFile(outputPath);
	outputFile << "[" << endl;
	for (int i = 0; i < outputList.size(); i++) {
		DataEntry *entry = &outputList[i];
		json json_entry;
		json_entry["name"] = entry->name;
		json_entry["sugar"] = entry->sugar;
		json_entry["criteria"] = entry->criteria;

		std::string json_str = json_entry.dump();
		outputFile << "\t" << json_str;
		if (i < outputList.size()-1) {
			outputFile << ",";
		}
		outputFile << endl;
	}
	outputFile << "]" << endl;

	cout << "Finished" << endl;
	return 0;
}
