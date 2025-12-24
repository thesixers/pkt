#include "GlobalStore.hpp"
#include "Utils.hpp"
#include <fstream>

namespace pkg {

GlobalStore::GlobalStore() {
    storePath = Utils::joinPath(Utils::getHomeDir(), ".pkg_global_store");
    
    // Initialize language-to-folder mappings
    languageDepFolders = {
        {"node", "node_modules"},
        {"python", "site-packages"},
        {"ruby", "gems"},
        {"java", "maven"},
        {"go", "pkg"}
    };
    
    ensureStoreExists();
    initializeLanguageFolders();
}

void GlobalStore::ensureStoreExists() {
    if (!Utils::dirExists(storePath)) {
        Utils::createDirRecursive(storePath);
    }
}

void GlobalStore::initializeLanguageFolders() {
    for (const auto& pair : languageDepFolders) {
        std::string langPath = getLanguagePath(pair.first);
        std::string depFolderPath = Utils::joinPath(langPath, pair.second);
        
        Utils::createDirRecursive(langPath);
        Utils::createDirRecursive(depFolderPath);
        
        // Create .deps file if it doesn't exist
        std::string depsFile = getDepsFilePath(pair.first);
        if (!Utils::fileExists(depsFile)) {
            json emptyDeps = json::object();
            Utils::writeFile(depsFile, emptyDeps.dump(2));
        }
    }
}

bool GlobalStore::packageExists(const std::string& language, const std::string& name, 
                                const std::string& version) const {
    std::string pkgPath = getPackagePath(language, name, version);
    return Utils::dirExists(pkgPath);
}

std::string GlobalStore::getPackagePath(const std::string& language, const std::string& name, 
                                        const std::string& version) const {
    std::string depFolder = getDepFolderForLanguage(language);
    std::string langPath = getLanguagePath(language);
    
    return Utils::joinPath({langPath, depFolder, name, version});
}

bool GlobalStore::addPackage(const std::string& language, const std::string& name, 
                            const std::string& version) {
    if (packageExists(language, name, version)) {
        return true; // Already exists
    }
    
    std::string pkgPath = getPackagePath(language, name, version);
    if (!Utils::createDirRecursive(pkgPath)) {
        Utils::logError("Failed to create package directory: " + pkgPath);
        return false;
    }
    
    return addToGlobalDeps(language, name, version);
}

bool GlobalStore::removePackage(const std::string& language, const std::string& name, 
                               const std::string& version) {
    std::string pkgPath = getPackagePath(language, name, version);
    
    if (Utils::dirExists(pkgPath)) {
        Utils::removeDir(pkgPath);
    }
    
    return removeFromGlobalDeps(language, name, version);
}

std::map<std::string, std::string> GlobalStore::getGlobalDeps(const std::string& language) const {
    json deps = loadDepsFile(language);
    std::map<std::string, std::string> result;
    
    for (auto it = deps.begin(); it != deps.end(); ++it) {
        result[it.key()] = it.value().get<std::string>();
    }
    
    return result;
}

bool GlobalStore::addToGlobalDeps(const std::string& language, const std::string& name, 
                                  const std::string& version) {
    json deps = loadDepsFile(language);
    
    // Add or update the package version
    deps[name] = version;
    
    saveDepsFile(language, deps);
    return true;
}

bool GlobalStore::removeFromGlobalDeps(const std::string& language, const std::string& name, 
                                       const std::string& /*version*/) {
    json deps = loadDepsFile(language);
    
    if (deps.contains(name)) {
        deps.erase(name);
        saveDepsFile(language, deps);
        return true;
    }
    
    return false;
}

std::string GlobalStore::getDepFolderForLanguage(const std::string& language) const {
    auto it = languageDepFolders.find(language);
    if (it != languageDepFolders.end()) {
        return it->second;
    }
    return "packages"; // Default fallback
}

std::vector<std::string> GlobalStore::getSupportedLanguages() const {
    std::vector<std::string> languages;
    for (const auto& pair : languageDepFolders) {
        languages.push_back(pair.first);
    }
    return languages;
}

std::string GlobalStore::getStorePath() const {
    return storePath;
}

std::string GlobalStore::getLanguagePath(const std::string& language) const {
    return Utils::joinPath(storePath, language);
}

std::string GlobalStore::getDepsFilePath(const std::string& language) const {
    return Utils::joinPath(getLanguagePath(language), ".deps");
}

json GlobalStore::loadDepsFile(const std::string& language) const {
    std::string depsPath = getDepsFilePath(language);
    
    if (!Utils::fileExists(depsPath)) {
        return json::object();
    }
    
    std::string content = Utils::readFile(depsPath);
    if (content.empty()) {
        return json::object();
    }
    
    try {
        return json::parse(content);
    } catch (const json::exception& e) {
        Utils::logError("Failed to parse .deps file: " + std::string(e.what()));
        return json::object();
    }
}

void GlobalStore::saveDepsFile(const std::string& language, const json& deps) {
    std::string depsPath = getDepsFilePath(language);
    Utils::writeFile(depsPath, deps.dump(2));
}

} // namespace pkg
