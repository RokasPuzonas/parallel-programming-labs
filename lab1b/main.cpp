#include <thread>
#include <iostream>
#include <fstream>
#include <string>
#include <vector>
#include <omp.h>

#include "json.hpp"

using namespace std;
using json = nlohmann::json;

const int threadCount = 15;

struct DataEntry {
    std::string name;
    float sugar;
    int criteria;
};

void sortedAdd(vector<DataEntry> *entries, DataEntry element) {
    entries->push_back(element);
    for (int i = entries->size()-1; i >= 1; i--) {
        if ((*entries)[i].sugar > (*entries)[i-1].sugar) {
            break;
        }
        swap((*entries)[i], (*entries)[i-1]);
    }
}

bool processEntry(DataEntry *entry, vector<DataEntry> *entries, omp_lock_t *outputLock) {
    this_thread::sleep_for(chrono::milliseconds(100 + entry->criteria));
    if (entry->sugar > entry->criteria) {
        omp_set_lock(outputLock);
        cout << "Output: " << entry->name << endl;
        sortedAdd(entries, *entry);
        omp_unset_lock(outputLock);
        return true;
    }

    return false;
}

tuple<int, int> getDataRange(int N, int threadCount, int tid) {
    int start = ((N/threadCount)+1) * tid;
    int end   = ((N/threadCount)+1) * (tid+1);
    if (tid >= N%threadCount) {
        start -= tid - N%threadCount;
        end   -= (tid+1) - N%threadCount;
    }

    return make_tuple(start, end);
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
    omp_lock_t outputListLock;
    omp_init_lock(&outputListLock);

    int N = entries.size();

    // Test if 'getDataRange' works correcly
    // for (int tid = 0; tid < threadCount; tid++) {
    //     auto data_range = get_data_range(N, threadCount, tid);
    //     int start = get<0>(data_range);
    //     int end = get<1>(data_range);
    //     cout << start << " -> " << end << endl;
    // }

    int sugarSum;
    float criteriaSum;
    #pragma omp parallel \
        num_threads(threadCount) \
        reduction(+:sugarSum, criteriaSum) \
        shared(cout, N, threadCount, outputList, outputListLock, entries)
    {
        int tid = omp_get_thread_num();
        auto data_range = getDataRange(N, threadCount, tid);
        int start = get<0>(data_range);
        int end = get<1>(data_range);

        sugarSum = 0;
        criteriaSum = 0;
        for (int i = start; i < end; i++) {
            DataEntry *entry = &entries[i];
            cout << "Started to filter: " << entry->name << endl;
            if (processEntry(entry, &outputList, &outputListLock)) {
                sugarSum += entry->sugar;
                criteriaSum += entry->criteria;
            }
            cout << "Finished to filter: " << entry->name << endl;
        }
    }

    omp_destroy_lock(&outputListLock);

    std::ofstream outputFile(outputPath);
    outputFile << "{" << endl;
    outputFile << "\"data\": [" << endl;
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
    outputFile << "]," << endl;
    outputFile << "\"sugar_sum\": " << sugarSum << "," << endl;
    outputFile << "\"criteria_sum\": " << criteriaSum << endl;
    outputFile << "}" << endl;

    cout << "Finished" << endl;
    return 0;
}
