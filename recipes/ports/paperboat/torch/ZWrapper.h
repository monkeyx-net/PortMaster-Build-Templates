#pragma once

#include <vector>
#include <string>
#include "BinaryWrapper.h"

// PortMaster: streams entries straight into the O2R file with miniz's file writer instead of
// building the whole archive in RAM (miniz_cpp::zip_file), which runs 1GB devices out of memory.
class ZWrapper : public BinaryWrapper {
public:
    explicit ZWrapper(const std::string& path);
    ~ZWrapper() override;

    int32_t CreateArchive(void) override;
    bool AddFile(const std::string& path, std::vector<char> data) override;
    int32_t Close(void) override;

private:
    void Open();

    void* mZip; // mz_zip_archive, opaque so miniz stays out of this header
    bool mOpen = false;
};
