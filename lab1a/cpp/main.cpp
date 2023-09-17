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

const int inputBufferSize = 10;
const int inputMonitorCount = 3;

const int outputBufferSize = 10;
const int outputMonitorCount = 3;

#include "common.h"
#include "limited_buffer.cpp"

bool filterEntry(DataEntry *entry) {
	this_thread::sleep_for(chrono::milliseconds(100 + entry->criteria));
	return entry->sugar > entry->criteria;
}

void outputEntry(vector<DataEntry> *entries, DataEntry *entry) {
	this_thread::sleep_for(chrono::milliseconds(200));
	cout << "Output: " << entry->name << endl;
	entries->push_back(*entry);
	sort(entries->begin(), entries->end(), [](DataEntry &a, DataEntry &b) {
			return a.sugar < b.sugar;
	});
}

void filterEntries(LimitedBuffer *inputBuffer, LimitedBuffer *outputBuffer) {
	while (!inputBuffer->is_done()) {
		if (inputBuffer->consume_estimated()) {
			DataEntry entry = inputBuffer->remove();
			cout << "Started to filter: " << entry.name << endl;
			bool is_filtered = filterEntry(&entry);
			cout << "Finished to filter: " << entry.name << endl;

			if (is_filtered) {
				outputBuffer->update_estimated(+1);
				outputBuffer->insert(entry);
			}

			inputBuffer->mark_processed();
		}
	}
}

void outputEntries(LimitedBuffer *inputBuffer, LimitedBuffer *outputBuffer, vector<DataEntry> *outputList, mutex *outputMutex) {
	while (!inputBuffer->is_done() || !outputBuffer->is_done()) {
		if (outputBuffer->consume_estimated()) {
			DataEntry entry = outputBuffer->remove();
			cout << "Started to output: " << entry.name << endl;
			outputEntry(outputList, &entry);
			cout << "Finished to output: " << entry.name << endl;

			outputBuffer->mark_processed();
		}
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

	LimitedBuffer inputBuffer(inputBufferSize);
	LimitedBuffer outputBuffer(outputBufferSize);
	inputBuffer.update_estimated(entries.size());

	vector<thread> threads;
	for (int i = 0; i < inputMonitorCount; i++) {
		threads.push_back(thread(filterEntries, &inputBuffer, &outputBuffer));
	}
	for (int i = 0; i < outputMonitorCount; i++) {
		threads.push_back(thread(outputEntries, &inputBuffer, &outputBuffer, &outputList, &outputListMutex));
	}

	for (int i = 0; i < entries.size(); i++) {
		inputBuffer.insert(entries[i]);
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

void execute(const string &name) {
	cout << name << ": one" << endl;
	cout << name << ": two" << endl;
	cout << name << ": three" << endl;
}
