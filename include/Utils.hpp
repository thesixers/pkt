#pragma once

#include <string>
#include <vector>
#include <map>

namespace pkg {

/**
 * Cross-platform utility functions for filesystem operations,
 * JSON handling, string manipulation, and symlink management.
 */
class Utils {
public:
    // Filesystem operations
    static bool fileExists(const std::string& path);
    static bool dirExists(const std::string& path);
    static bool createDir(const std::string& path);
    static bool createDirRecursive(const std::string& path);
    static bool removeFile(const std::string& path);
    static bool removeDir(const std::string& path);
    static bool copyFile(const std::string& src, const std::string& dest);
    static std::string getCurrentDir();
    static std::string getHomeDir();
    static std::string joinPath(const std::string& a, const std::string& b);
    static std::string joinPath(const std::vector<std::string>& parts);
    static std::string getFileName(const std::string& path);
    static std::string getBaseName(const std::string& path);
    
    // Symlink operations
    static bool createSymlink(const std::string& target, const std::string& link);
    static bool isSymlink(const std::string& path);
    static bool isAbsolutePath(const std::string& path);
    static bool removeSymlink(const std::string& path);
    
    // JSON file operations
    static std::string readFile(const std::string& path);
    static bool writeFile(const std::string& path, const std::string& content);
    
    // String utilities
    static std::string trim(const std::string& str);
    static std::vector<std::string> split(const std::string& str, char delimiter);
    static bool startsWith(const std::string& str, const std::string& prefix);
    static bool endsWith(const std::string& str, const std::string& suffix);
    static std::string toLower(const std::string& str);
    static int fuzzyMatch(const std::string& query, const std::string& target);
    
    // ID generation
    static std::string generateId(const std::string& prefix);
    
    // Time utilities
    static std::string getCurrentTimestamp();
    
    // Error handling
    static void logError(const std::string& message);
    static void logInfo(const std::string& message);
    static void logSuccess(const std::string& message);
    static void logWarning(const std::string& message);
};

} // namespace pkg
