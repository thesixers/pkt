#pragma once

#include "GlobalRegistry.hpp"
#include "GlobalStore.hpp"
#include "RegistryClient.hpp"
#include <string>
#include <map>

namespace pkg {

/**
 * Manages project and global dependencies
 * Handles installation, removal, updates, and symlink management
 */
class DependencyManager {
public:
    DependencyManager(GlobalRegistry& registry, GlobalStore& store, RegistryClient& client);
    
    // Dependency operations
    bool addDependency(const std::string& packageSpec);
    bool removeDependency(const std::string& packageName);
    bool updateDependency(const std::string& packageSpec);
    
    // Dependency queries
    void listProjectDeps();
    void listGlobalDeps(const std::string& language, bool allLanguages = false);
    
private:
    GlobalRegistry& registry;
    GlobalStore& store;
    RegistryClient& client;
    
    // Helper methods
    struct PackageSpec {
        std::string name;
        std::string version;
        bool hasVersion;
    };
    
    PackageSpec parsePackageSpec(const std::string& spec);
    std::map<std::string, std::string> loadProjectDeps() const;
    bool saveProjectDeps(const std::map<std::string, std::string>& deps);
    bool createSymlink(const std::string& packageName, const std::string& version, 
                      const std::string& language, const std::string& depFolder);
    bool removeSymlink(const std::string& packageName, const std::string& depFolder);
    std::string resolveVersion(const std::string& language, const std::string& packageName, 
                              const std::string& requestedVersion);
    bool addDependencyRecursive(const std::string& language, 
                               const std::string& packageName,
                               const std::string& version,
                               const std::string& depFolder,
                               bool isDirect);
};

} // namespace pkg
