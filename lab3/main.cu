#include "json.hpp"
#include "cuda_runtime.h"
#include "device_launch_parameters.h"

#include <stdio.h>
#include <iostream>
#include <fstream>
#include <vector>

using namespace std;
using json = nlohmann::json;


#define checkCudaErrors(val) check_cuda((val), #val, __FILE__, __LINE__)

void check_cuda(cudaError_t result, const char* func, const char* file, int line) {
    if (result) {
        std::cout << "CUDA error: " << cudaGetErrorString(result) << " (error code " << static_cast<unsigned int>(result) << ")";
        std::cout << " at " << file << ":" << line << " '" << func << "' \n";
        cudaDeviceReset();
        exit(-1);
    }
}

struct DataEntry {
    char name[64];
    float sugar;
    int criteria;
};

struct Result {
    char text[64];
};

__device__ static size_t get_string_size(char *text) {
    size_t size = 0;
    while (text[size] != 0) {
        size++;
    }
    return size;
}

__global__ void processEntriesKernel(DataEntry *entries, Result *results, size_t entry_count)
{
    int idx = blockDim.x * blockIdx.x + threadIdx.x;
    if (idx >= entry_count) return;

    auto entry = &entries[idx];
    if (entry->sugar < entry->criteria) return;
    auto result = &results[idx];
 
    for (int i = 0; i < get_string_size(entry->name)/2; i++) {
        result->text[i] = entry->name[2*i];
        if ('a' <= result->text[i] && result->text[i] <= 'z') {
            result->text[i] -= 32; // Convert to uppercase
        }
    }
}

int main(int argc, char** argv)
{
    int block_count = 8;
    int block_size  = 32;
    const char* input_path = "IF-1-1_PuzonasR_L3_dat_1.json";
    const char* output_path = "output.txt";

    std::ifstream f(input_path);
    json data = json::parse(f);

    vector<DataEntry> entries;
    for (auto &it : data) {
        auto entry = DataEntry{ 0 };
        strcpy(entry.name, it["name"].get<std::string>().c_str());
        entry.sugar = it["sugar"];
        entry.criteria = it["criteria"];
        entries.push_back(entry);
    }

    int entry_count = entries.size();
    cout << "Input data count: " << entry_count << endl;

    if (entry_count > block_count * block_size) {
        cout << "WARNING! Not enough blocks/threads, the total number threads is " << block_count * block_size << ", but you need " << entry_count << endl;
    }

    DataEntry* device_entries = NULL;
    Result* device_results = NULL;
    checkCudaErrors(cudaMalloc((void**)&device_entries, entry_count * sizeof(DataEntry)));
    checkCudaErrors(cudaMalloc((void**)&device_results, entry_count * sizeof(Result)));

    checkCudaErrors(cudaMemcpy(device_entries, &entries[0], entry_count * sizeof(DataEntry), cudaMemcpyHostToDevice));
    checkCudaErrors(cudaMemset(device_results, 0, entry_count * sizeof(Result)));
    checkCudaErrors(cudaDeviceSynchronize());

    processEntriesKernel<<<block_count, block_size>>>(device_entries, device_results, entry_count);
    checkCudaErrors(cudaDeviceSynchronize());

    Result* results = (Result*)malloc(entry_count * sizeof(Result));
    checkCudaErrors(cudaMemcpy(results, device_results, entry_count * sizeof(Result), cudaMemcpyDeviceToHost));

    int result_count = 0;
    std::ofstream output_file(output_path);
    for (int i = 0; i < entry_count; i++) {
        if (results[i].text[0] == 0) continue;

        output_file << results[i].text << endl;
        result_count++;
    }

    cout << "Result data count: " << result_count << endl;

    free(results);
    cudaFree(device_entries);
    cudaFree(device_results);
    return 0;
}
