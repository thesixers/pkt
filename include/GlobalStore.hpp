#pragma once

#include <string>
#include <vector>
#include <map>
#include <nlohmann/json.hpp>

namespace pkg {

using json = nlohmann::json;

/**
 * Package metadata structure
 */
struct Package {
    std::string name;
    std::string version;
    std::string language;
};

/**
 * Manages the global package store (~/.pkg_global_store/)
 * Organizes packages by language with native folder structures.
 * Maintains .deps files for each language.
 */
class GlobalStore {
public:
    GlobalStore();
    
    // Store operations
    bool packageExists(const std::string& language, const std::string& name, const std::string& version) const;
    std::string getPackagePath(const std::string& language, const std::string& name, const std::string& version) const;
    bool addPackage(const std::string& language, const std::string& name, const std::string& version);
    bool removePackage(const std::string& language, const std::string& name, const std::string& version);
    
    // .deps file management
    std::map<std::string, std::string> getGlobalDeps(const std::string& language) const;
    bool addToGlobalDeps(const std::string& language, const std::string& name, const std::string& version);
    bool removeFromGlobalDeps(const std::string& language, const std::string& name, const std::string& version);
    
    // Language configuration
    std::string getDepFolderForLanguage(const std::string& language) const;
    std::vector<std::string> getSupportedLanguages() const;
    
    // Utility
    std::string getStorePath() const;
    std::string getLanguagePath(const std::string& language) const;
    
private:
    std::string storePath;
    std::map<std::string, std::string> languageDepFolders;
    
    void ensureStoreExists();
    void initializeLanguageFolders();
    std::string getDepsFilePath(const std::string& language) const;
    json loadDepsFile(const std::string& language) const;
    void saveDepsFile(const std::string& language, const json& deps);
};

} // namespace pkg
