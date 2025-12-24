#include "DependencyManager.hpp"
#include "Utils.hpp"
#include <iostream>
#include <fstream>

#include <nlohmann/json.hpp>

namespace pkg {

using json = nlohmann::json;

DependencyManager::DependencyManager(GlobalRegistry& registry, GlobalStore& store, 
                                     RegistryClient& client)
    : registry(registry), store(store), client(client) {}

DependencyManager::PackageSpec DependencyManager::parsePackageSpec(const std::string& spec) {
    PackageSpec result;
    
    size_t atPos = spec.find('@');
    if (atPos != std::string::npos) {
        result.name = spec.substr(0, atPos);
        result.version = spec.substr(atPos + 1);
        result.hasVersion = true;
    } else {
        result.name = spec;
        result.version = "";
        result.hasVersion = false;
    }
    
    return result;
}

std::map<std::string, std::string> DependencyManager::loadProjectDeps() const {
    std::string depsPath = Utils::joinPath(Utils::getCurrentDir(), ".pkg.deps");
    
    if (!Utils::fileExists(depsPath)) {
        return {};
    }
    
    std::string content = Utils::readFile(depsPath);
    if (content.empty()) {
        return {};
    }
    
    try {
        json data = json::parse(content);
        std::map<std::string, std::string> deps;
        
        for (auto it = data.begin(); it != data.end(); ++it) {
            deps[it.key()] = it.value().get<std::string>();
        }
        
        return deps;
    } catch (const json::exception& e) {
        Utils::logError("Failed to parse .pkg.deps: " + std::string(e.what()));
        return {};
    }
}

bool DependencyManager::saveProjectDeps(const std::map<std::string, std::string>& deps) {
    std::string depsPath = Utils::joinPath(Utils::getCurrentDir(), ".pkg.deps");
    
    json data = json::object();
    for (const auto& pair : deps) {
        data[pair.first] = pair.second;
    }
    
    return Utils::writeFile(depsPath, data.dump(2));
}

std::string DependencyManager::resolveVersion(const std::string& language, 
                                              const std::string& packageName,
                                              const std::string& requestedVersion) {
    if (!requestedVersion.empty()) {
        return requestedVersion;
    }
    
    // Query registry for latest version
    Utils::logInfo("Fetching latest version of " + packageName + "...");
    std::string latestVersion = client.getLatestVersion(language, packageName);
    
    if (latestVersion.empty()) {
        Utils::logError("Failed to fetch version for " + packageName);
        return "";
    }
    
    return latestVersion;
}

bool DependencyManager::addDependency(const std::string& packageSpec) {
    // Check if in project folder
    std::string pkgInfoPath = Utils::joinPath(Utils::getCurrentDir(), ".pkg.info");
    if (!Utils::fileExists(pkgInfoPath)) {
        Utils::logError("Not in a PKG project directory");
        return false;
    }
    
    // Load project info
    std::string content = Utils::readFile(pkgInfoPath);
    json projectInfo = json::parse(content);
    std::string language = projectInfo["language"];
    std::string depFolder = projectInfo["default_dep_folder"];
    
    // Parse package specification
    PackageSpec spec = parsePackageSpec(packageSpec);
    
    return addDependencyRecursive(language, spec.name, spec.version, depFolder, true);
}

bool DependencyManager::addDependencyRecursive(const std::string& language, 
                                              const std::string& packageName,
                                              const std::string& version,
                                              const std::string& depFolder,
                                              bool isDirect) {
    // Load current dependencies
    auto deps = loadProjectDeps();
    
    // Resolve version
    std::string resolvedVersion = resolveVersion(language, packageName, version);
    if (resolvedVersion.empty()) {
        return false;
    }
    
    // Check if already installed (if direct, we might want to update, but for now just skip)
    if (isDirect && deps.find(packageName) != deps.end() && deps[packageName] == resolvedVersion) {
        Utils::logWarning("Package '" + packageName + "' is already installed at version " + resolvedVersion);
        return true;
    }
    
    // Check if package exists in global store
    if (!store.packageExists(language, packageName, resolvedVersion)) {
        Utils::logInfo("Downloading " + packageName + "@" + resolvedVersion + "...");
        
        // Add to global store
        if (!store.addPackage(language, packageName, resolvedVersion)) {
            Utils::logError("Failed to add package to global store");
            return false;
        }
        
        // Download package
        std::string pkgPath = store.getPackagePath(language, packageName, resolvedVersion);
        if (!client.downloadPackage(language, packageName, resolvedVersion, pkgPath)) {
            Utils::logError("Failed to download package");
            return false;
        }
    }
    
    // Create symlink
    if (!createSymlink(packageName, resolvedVersion, language, depFolder)) {
        Utils::logError("Failed to create symlink for " + packageName);
        return false;
    }
    
    // Update project dependencies if it's a direct dependency
    if (isDirect) {
        deps[packageName] = resolvedVersion;
        if (!saveProjectDeps(deps)) {
            Utils::logError("Failed to update .pkg.deps");
            return false;
        }
        Utils::logSuccess("Added " + packageName + "@" + resolvedVersion);
    } else {
        Utils::logInfo("Installed sub-dependency: " + packageName + "@" + resolvedVersion);
    }
    
    // Recursive resolution: fetch sub-dependencies from the actual package.json in the store
    std::string packagePath = store.getPackagePath(language, packageName, resolvedVersion);
    std::string pkgJsonPath = Utils::joinPath(packagePath, "package.json");
    
    if (language == "node" && Utils::fileExists(pkgJsonPath)) {
        try {
            json pkgJson = json::parse(Utils::readFile(pkgJsonPath));
            if (pkgJson.contains("dependencies")) {
                const auto& subDeps = pkgJson["dependencies"];
                
                // Create node_modules folder inside the package's store directory
                std::string storeNodeModules = Utils::joinPath(packagePath, "node_modules");
                Utils::createDirRecursive(storeNodeModules);
                
                for (auto it = subDeps.begin(); it != subDeps.end(); ++it) {
                    std::string subName = it.key();
                    std::string subVersionSpec = it.value().get<std::string>();
                    
                    // Install sub-dependency recursively
                    // Note: depFolder for sub-dependencies is the node_modules folder inside the package's store dir
                    if (!addDependencyRecursive(language, subName, subVersionSpec, storeNodeModules, false)) {
                        Utils::logWarning("Failed to install sub-dependency: " + subName);
                    }
                }
            }
        } catch (const std::exception& e) {
            Utils::logWarning("Failed to parse package.json for " + packageName + ": " + e.what());
        }
    }
    
    return true;
}

bool DependencyManager::removeDependency(const std::string& packageName) {
    // Check if in project folder
    std::string pkgInfoPath = Utils::joinPath(Utils::getCurrentDir(), ".pkg.info");
    if (!Utils::fileExists(pkgInfoPath)) {
        Utils::logError("Not in a PKG project directory");
        return false;
    }
    
    // Load project info
    std::string content = Utils::readFile(pkgInfoPath);
    json projectInfo = json::parse(content);
    std::string depFolder = projectInfo["default_dep_folder"];
    
    // Load current dependencies
    auto deps = loadProjectDeps();
    
    // Check if package exists
    if (deps.find(packageName) == deps.end()) {
        Utils::logError("Package '" + packageName + "' is not installed");
        return false;
    }
    
    std::string version = deps[packageName];
    
    // Remove symlink
    if (!removeSymlink(packageName, depFolder)) {
        Utils::logWarning("Failed to remove symlink (may not exist)");
    }
    
    // Update project dependencies
    deps.erase(packageName);
    if (!saveProjectDeps(deps)) {
        Utils::logError("Failed to update .pkg.deps");
        return false;
    }
    
    Utils::logSuccess("Removed " + packageName + "@" + version);
    return true;
}

bool DependencyManager::updateDependency(const std::string& packageSpec) {
    // Parse package specification
    PackageSpec spec = parsePackageSpec(packageSpec);
    
    // Remove old version
    if (!removeDependency(spec.name)) {
        return false;
    }
    
    // Add new version
    return addDependency(packageSpec);
}

void DependencyManager::listProjectDeps() {
    auto deps = loadProjectDeps();
    
    if (deps.empty()) {
        Utils::logInfo("No dependencies installed");
        return;
    }
    
    std::cout << "\n📦 Project Dependencies:\n\n";
    for (const auto& pair : deps) {
        std::cout << "  • " << pair.first << "@" << pair.second << "\n";
    }
    std::cout << "\n";
}

void DependencyManager::listGlobalDeps(const std::string& language, bool allLanguages) {
    if (allLanguages) {
        auto languages = store.getSupportedLanguages();
        
        std::cout << "\n🌍 Global Dependencies (All Languages):\n\n";
        for (const auto& lang : languages) {
            auto deps = store.getGlobalDeps(lang);
            
            if (!deps.empty()) {
                std::cout << "  " << lang << ":\n";
                for (const auto& pair : deps) {
                    std::cout << "    • " << pair.first << "@" << pair.second << "\n";
                }
                std::cout << "\n";
            }
        }
    } else {
        auto deps = store.getGlobalDeps(language);
        
        if (deps.empty()) {
            Utils::logInfo("No global dependencies for " + language);
            return;
        }
        
        std::cout << "\n🌍 Global Dependencies (" << language << "):\n\n";
        for (const auto& pair : deps) {
            std::cout << "  • " << pair.first << "@" << pair.second << "\n";
        }
        std::cout << "\n";
    }
}

bool DependencyManager::createSymlink(const std::string& packageName, const std::string& version,
                                     const std::string& language, const std::string& depFolder) {
    std::string packagePath = store.getPackagePath(language, packageName, version);
    
    // CRITICAL: Link the package directory, not the entry point file.
    // This ensures Node.js can find package.json and its own node_modules.
    std::string target = packagePath;
    
    // depFolder can be an absolute path (for sub-deps in store) or relative (for project deps)
    std::string link;
    if (Utils::isAbsolutePath(depFolder)) {
        link = Utils::joinPath(depFolder, packageName);
    } else {
        link = Utils::joinPath({Utils::getCurrentDir(), depFolder, packageName});
    }
    
    // Remove existing symlink if present
    if (Utils::isSymlink(link)) {
        Utils::removeSymlink(link);
    }
    
    return Utils::createSymlink(target, link);
}

bool DependencyManager::removeSymlink(const std::string& packageName, const std::string& depFolder) {
    std::string link = Utils::joinPath({Utils::getCurrentDir(), depFolder, packageName});
    
    if (Utils::isSymlink(link)) {
        return Utils::removeSymlink(link);
    }
    
    return true;
}

} // namespace pkg
