#pragma once

#include <string>
#include <map>

namespace pkg {

/**
 * HTTP client for querying package registries
 * Supports npm, PyPI, RubyGems, Maven Central, and Go proxy
 */
class RegistryClient {
public:
    RegistryClient();
    
    // Version resolution
    std::string getLatestVersion(const std::string& language, const std::string& packageName);
    bool packageExists(const std::string& language, const std::string& packageName);
    
    // Package download
    bool downloadPackage(const std::string& language, const std::string& packageName, 
                        const std::string& version, const std::string& destPath);
    
private:
    std::map<std::string, std::string> registryUrls;
    
    void initializeRegistryUrls();
    std::string httpGet(const std::string& url);
    bool httpDownload(const std::string& url, const std::string& destPath);
    
    // Language-specific implementations
    std::string getLatestVersionNpm(const std::string& packageName);
    std::string getLatestVersionPyPI(const std::string& packageName);
    std::string getLatestVersionRubyGems(const std::string& packageName);
    std::string getLatestVersionMaven(const std::string& packageName);
    std::string getLatestVersionGo(const std::string& packageName);
    
    bool downloadNpmPackage(const std::string& packageName, const std::string& version, const std::string& destPath);
    bool downloadPyPIPackage(const std::string& packageName, const std::string& version, const std::string& destPath);
    bool downloadRubyGem(const std::string& packageName, const std::string& version, const std::string& destPath);
    bool downloadMavenPackage(const std::string& packageName, const std::string& version, const std::string& destPath);
    bool downloadGoModule(const std::string& packageName, const std::string& version, const std::string& destPath);
};

} // namespace pkg
