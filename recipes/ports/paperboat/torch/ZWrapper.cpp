#include "ZWrapper.h"
#include <cstring>
#include <filesystem>
#include <fstream>
#include <stdexcept>

#include "spdlog/spdlog.h"
#include <Companion.h>
#include <miniz/zip_file.hpp>

namespace fs = std::filesystem;

#define ZIP(p) (static_cast<mz_zip_archive*>(p))

ZWrapper::ZWrapper(const std::string& path) {
    this->mPath = path;
    this->mZip = new mz_zip_archive;
    std::memset(this->mZip, 0, sizeof(mz_zip_archive));
}

ZWrapper::~ZWrapper() {
    if (this->mOpen) {
        mz_zip_writer_end(ZIP(this->mZip));
    }
    delete ZIP(this->mZip);
}

int32_t ZWrapper::CreateArchive() {
    SPDLOG_INFO("Loaded ZIP (O2R) archive: {}", mPath.c_str());
    return 0;
}

// Opened on first use, matching the old writer which only touched the file in Close()
void ZWrapper::Open() {
    if (this->mOpen) {
        return;
    }
    if (!mz_zip_writer_init_file(ZIP(this->mZip), this->mPath.c_str(), 0)) {
        throw std::runtime_error("Failed to create O2R archive: " + this->mPath);
    }
    this->mOpen = true;
}

bool ZWrapper::AddFile(const std::string& path, std::vector<char> data) {
    std::lock_guard<std::mutex> lock(this->mMutex);
    char* fileData = data.data();
    size_t fileSize = data.size();

    if (Companion::Instance != nullptr && Companion::Instance->IsDebug()) {
        SPDLOG_INFO("Creating debug file: debug/{}", path);
        std::string dpath = "debug/" + path;
        if (!fs::exists(fs::path(dpath).parent_path())) {
            fs::create_directories(fs::path(dpath).parent_path());
        }
        std::ofstream stream(dpath, std::ios::binary);
        stream.write(fileData, fileSize);
        stream.close();
    }

    Open();
    if (!mz_zip_writer_add_mem(ZIP(this->mZip), path.c_str(), fileData, fileSize, MZ_BEST_COMPRESSION)) {
        // miniz's own error says nothing useful, so log which entry blew up and how big.
        SPDLOG_ERROR("[ZWrapper] Failed to add '{}' ({} bytes)", path, fileSize);
        throw std::runtime_error("Failed to add " + path + " to O2R archive");
    }
    return true;
}

int32_t ZWrapper::Close(void) {
    std::lock_guard<std::mutex> lock(this->mMutex);
    Open();
    bool ok = mz_zip_writer_finalize_archive(ZIP(this->mZip));
    mz_zip_writer_end(ZIP(this->mZip));
    this->mOpen = false;
    if (!ok) {
        SPDLOG_ERROR("[ZWrapper] Failed to finalize {}", this->mPath);
        return -1;
    }
    return 0;
}
