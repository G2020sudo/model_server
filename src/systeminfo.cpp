//*****************************************************************************
// Copyright 2023 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************
#include "systeminfo.hpp"

#include <fstream>
#include <sstream>
#include <string>
#include <thread>

#include "logging.hpp"
#include "status.hpp"
#include "systeminfo_impl.hpp"

namespace ovms {
const char* CPUSET_FILENAME = "/sys/fs/cgroup/cpuset/cpuset.cpus";
// Ubuntu 22.04 LTS changed CPUSET path
const char* CPUSET_FILENAME22 = "/sys/fs/cgroup/cpuset.cpus.effective";
    
uint16_t getCoreCount() {
    std::ifstream fs;
    uint16_t coreCount = 1;    
    std::string fileName(CPUSET_FILENAME);
    FILE *file = fopen(CPUSET_FILENAME, "r");
    
    if (!file) {
      fileName = CPUSET_FILENAME22;
      file = fopen(CPUSET_FILENAME22, "r");
    }
    if (file)
      fclose(file); 
    
    // below throws a segfault on ubuntu 22.04 LTS when using CPUSEt_FILENAME
    auto status = getCPUSetFile(fs, fileName.c_str());
    if (status.ok()) {
        std::string cpusets;
        fs >> cpusets;
        status = getCoreCountImpl(cpusets, coreCount);
    }
    if (status.ok()) {
        return coreCount;
    } else {
        SPDLOG_ERROR("Failed to read system core count from cpuset file. Falling back to std::thread::hardware_concurrency");
        auto hwConcurrency = std::thread::hardware_concurrency();
        if (hwConcurrency == 0) {
            SPDLOG_ERROR("Failed to read core count number of system. Fallback to treating system as 1 core only");
            return 1;
        }
        return hwConcurrency;
    }
}
}  // namespace ovms
