#include "Utils.hpp"
#include <iostream>
#include <fstream>
#include <sstream>
#include <algorithm>
#include <ctime>
#include <random>
#include <chrono>
#include <sys/stat.h>

#ifdef _WIN32
    #include <windows.h>
    #include <direct.h>
    #include <io.h>
    #define mkdir(path, mode) _mkdir(path)
    #define getcwd _getcwd
#else
    #include <unistd.h>
    #include <sys/types.h>
    #include <dirent.h>
#endif

namespace pkg {

// ANSI color codes
const std::string RESET = "\033[0m";
const std::string RED = "\033[31m";
const std::string GREEN = "\033[32m";
const std::string YELLOW = "\033[33m";
const std::string BLUE = "\033[34m";
const std::string CYAN = "\033[36m";

// Filesystem operations
bool Utils::fileExists(const std::string& path) {
    struct stat buffer;
    return (stat(path.c_str(), &buffer) == 0 && S_ISREG(buffer.st_mode));
}

bool Utils::dirExists(const std::string& path) {
    struct stat buffer;
    return (stat(path.c_str(), &buffer) == 0 && S_ISDIR(buffer.st_mode));
}

bool Utils::createDir(const std::string& path) {
    return mkdir(path.c_str(), 0755) == 0 || dirExists(path);
}

bool Utils::createDirRecursive(const std::string& path) {
    if (dirExists(path)) return true;
    
    size_t pos = 0;
    std::string currentPath;
    
    while ((pos = path.find('/', pos)) != std::string::npos) {
        currentPath = path.substr(0, pos++);
        if (!currentPath.empty() && !dirExists(currentPath)) {
            if (!createDir(currentPath)) return false;
        }
    }
    
    return createDir(path);
}

bool Utils::removeFile(const std::string& path) {
    return remove(path.c_str()) == 0;
}

bool Utils::removeDir(const std::string& path) {
#ifdef _WIN32
    return RemoveDirectoryA(path.c_str()) != 0;
#else
    return rmdir(path.c_str()) == 0;
#endif
}

bool Utils::copyFile(const std::string& src, const std::string& dest) {
    std::ifstream srcFile(src, std::ios::binary);
    std::ofstream destFile(dest, std::ios::binary);
    
    if (!srcFile || !destFile) return false;
    
    destFile << srcFile.rdbuf();
    return true;
}

std::string Utils::getCurrentDir() {
    char buffer[1024];
    if (getcwd(buffer, sizeof(buffer)) != nullptr) {
        return std::string(buffer);
    }
    return "";
}

std::string Utils::getHomeDir() {
#ifdef _WIN32
    const char* home = getenv("USERPROFILE");
#else
    const char* home = getenv("HOME");
#endif
    return home ? std::string(home) : "";
}

std::string Utils::joinPath(const std::string& a, const std::string& b) {
    if (a.empty()) return b;
    if (b.empty()) return a;
    
    char sep = '/';
#ifdef _WIN32
    sep = '\\';
#endif
    
    if (a.back() == sep) {
        return a + b;
    }
    return a + sep + b;
}

std::string Utils::joinPath(const std::vector<std::string>& parts) {
    if (parts.empty()) return "";
    
    std::string result = parts[0];
    for (size_t i = 1; i < parts.size(); ++i) {
        result = joinPath(result, parts[i]);
    }
    return result;
}

std::string Utils::getFileName(const std::string& path) {
    size_t pos = path.find_last_of("/\\");
    if (pos == std::string::npos) return path;
    return path.substr(pos + 1);
}

std::string Utils::getBaseName(const std::string& path) {
    std::string filename = getFileName(path);
    size_t pos = filename.find_last_of('.');
    if (pos == std::string::npos) return filename;
    return filename.substr(0, pos);
}

// Symlink operations
bool Utils::createSymlink(const std::string& target, const std::string& link) {
#ifdef _WIN32
    // On Windows, use junction points for directories or symbolic links for files
    DWORD attrs = GetFileAttributesA(target.c_str());
    if (attrs != INVALID_FILE_ATTRIBUTES && (attrs & FILE_ATTRIBUTE_DIRECTORY)) {
        // Create directory junction
        return CreateSymbolicLinkA(link.c_str(), target.c_str(), 
                                  SYMBOLIC_LINK_FLAG_DIRECTORY) != 0;
    } else {
        return CreateSymbolicLinkA(link.c_str(), target.c_str(), 0) != 0;
    }
#else
    return symlink(target.c_str(), link.c_str()) == 0;
#endif
}

bool Utils::isSymlink(const std::string& path) {
#ifdef _WIN32
    DWORD attrs = GetFileAttributesA(path.c_str());
    return (attrs != INVALID_FILE_ATTRIBUTES && 
            (attrs & FILE_ATTRIBUTE_REPARSE_POINT));
#else
    struct stat buffer;
    return (lstat(path.c_str(), &buffer) == 0 && S_ISLNK(buffer.st_mode));
#endif
}

bool Utils::isAbsolutePath(const std::string& path) {
    if (path.empty()) return false;
#ifdef _WIN32
    return path.size() >= 2 && isalpha(path[0]) && path[1] == ':';
#else
    return path[0] == '/';
#endif
}

bool Utils::removeSymlink(const std::string& path) {
    if (!isSymlink(path)) return false;
    return removeFile(path);
}

// JSON file operations
std::string Utils::readFile(const std::string& path) {
    std::ifstream file(path);
    if (!file) return "";
    
    std::stringstream buffer;
    buffer << file.rdbuf();
    return buffer.str();
}

bool Utils::writeFile(const std::string& path, const std::string& content) {
    std::ofstream file(path);
    if (!file) return false;
    
    file << content;
    return true;
}

// String utilities
std::string Utils::trim(const std::string& str) {
    size_t start = str.find_first_not_of(" \t\n\r");
    if (start == std::string::npos) return "";
    
    size_t end = str.find_last_not_of(" \t\n\r");
    return str.substr(start, end - start + 1);
}

std::vector<std::string> Utils::split(const std::string& str, char delimiter) {
    std::vector<std::string> tokens;
    std::stringstream ss(str);
    std::string token;
    
    while (std::getline(ss, token, delimiter)) {
        tokens.push_back(token);
    }
    
    return tokens;
}

bool Utils::startsWith(const std::string& str, const std::string& prefix) {
    return str.size() >= prefix.size() && 
           str.compare(0, prefix.size(), prefix) == 0;
}

bool Utils::endsWith(const std::string& str, const std::string& suffix) {
    return str.size() >= suffix.size() && 
           str.compare(str.size() - suffix.size(), suffix.size(), suffix) == 0;
}

std::string Utils::toLower(const std::string& str) {
    std::string result = str;
    std::transform(result.begin(), result.end(), result.begin(), ::tolower);
    return result;
}

int Utils::fuzzyMatch(const std::string& query, const std::string& target) {
    std::string lowerQuery = toLower(query);
    std::string lowerTarget = toLower(target);
    
    // Exact match
    if (lowerQuery == lowerTarget) return 100;
    
    // Starts with
    if (startsWith(lowerTarget, lowerQuery)) return 90;
    
    // Contains
    if (lowerTarget.find(lowerQuery) != std::string::npos) return 70;
    
    // Fuzzy character matching
    size_t queryIdx = 0;
    size_t targetIdx = 0;
    int matches = 0;
    
    while (queryIdx < lowerQuery.size() && targetIdx < lowerTarget.size()) {
        if (lowerQuery[queryIdx] == lowerTarget[targetIdx]) {
            matches++;
            queryIdx++;
        }
        targetIdx++;
    }
    
    if (queryIdx == lowerQuery.size()) {
        return static_cast<int>((matches * 50) / lowerQuery.size());
    }
    
    return 0;
}

// ID generation
std::string Utils::generateId(const std::string& prefix) {
    static std::random_device rd;
    static std::mt19937 gen(rd());
    static std::uniform_int_distribution<> dis(0, 15);
    
    const char* hexChars = "0123456789abcdef";
    std::string id = prefix;
    
    for (int i = 0; i < 8; ++i) {
        id += hexChars[dis(gen)];
    }
    
    return id;
}

// Time utilities
std::string Utils::getCurrentTimestamp() {
    time_t now = time(nullptr);
    char buffer[100];
    strftime(buffer, sizeof(buffer), "%Y-%m-%dT%H:%M:%SZ", gmtime(&now));
    return std::string(buffer);
}

// Error handling
void Utils::logError(const std::string& message) {
    std::cerr << RED << "✗ Error: " << message << RESET << std::endl;
}

void Utils::logInfo(const std::string& message) {
    std::cout << CYAN << "ℹ " << message << RESET << std::endl;
}

void Utils::logSuccess(const std::string& message) {
    std::cout << GREEN << "✓ " << message << RESET << std::endl;
}

void Utils::logWarning(const std::string& message) {
    std::cout << YELLOW << "⚠ Warning: " << message << RESET << std::endl;
}

} // namespace pkg
