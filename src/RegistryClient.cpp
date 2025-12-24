#include "RegistryClient.hpp"
#include "Utils.hpp"
#include <curl/curl.h>
#include <nlohmann/json.hpp>

using json = nlohmann::json;

namespace pkg {

// Callback for libcurl to write data
static size_t WriteCallback(void* contents, size_t size, size_t nmemb, std::string* userp) {
    userp->append((char*)contents, size * nmemb);
    return size * nmemb;
}

RegistryClient::RegistryClient() {
    initializeRegistryUrls();
    curl_global_init(CURL_GLOBAL_DEFAULT);
}

void RegistryClient::initializeRegistryUrls() {
    registryUrls["node"] = "https://registry.npmjs.org";
    registryUrls["python"] = "https://pypi.org/pypi";
    registryUrls["ruby"] = "https://rubygems.org/api/v1";
    registryUrls["java"] = "https://search.maven.org/solrsearch/select";
    registryUrls["go"] = "https://proxy.golang.org";
}

std::string RegistryClient::httpGet(const std::string& url) {
    CURL* curl = curl_easy_init();
    std::string response;
    
    if (curl) {
        curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, WriteCallback);
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, &response);
        curl_easy_setopt(curl, CURLOPT_FOLLOWLOCATION, 1L);
        curl_easy_setopt(curl, CURLOPT_TIMEOUT, 30L);
        
        CURLcode res = curl_easy_perform(curl);
        
        if (res != CURLE_OK) {
            Utils::logError("HTTP request failed: " + std::string(curl_easy_strerror(res)));
        }
        
        curl_easy_cleanup(curl);
    }
    
    return response;
}

bool RegistryClient::httpDownload(const std::string& url, const std::string& destPath) {
    CURL* curl = curl_easy_init();
    bool success = false;
    
    if (curl) {
        FILE* fp = fopen(destPath.c_str(), "wb");
        if (!fp) {
            curl_easy_cleanup(curl);
            return false;
        }
        
        curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, fp);
        curl_easy_setopt(curl, CURLOPT_FOLLOWLOCATION, 1L);
        curl_easy_setopt(curl, CURLOPT_TIMEOUT, 120L);
        
        CURLcode res = curl_easy_perform(curl);
        
        if (res == CURLE_OK) {
            success = true;
        } else {
            Utils::logError("Download failed: " + std::string(curl_easy_strerror(res)));
        }
        
        fclose(fp);
        curl_easy_cleanup(curl);
    }
    
    return success;
}

std::string RegistryClient::getLatestVersion(const std::string& language, const std::string& packageName) {
    if (language == "node") {
        return getLatestVersionNpm(packageName);
    } else if (language == "python") {
        return getLatestVersionPyPI(packageName);
    } else if (language == "ruby") {
        return getLatestVersionRubyGems(packageName);
    } else if (language == "java") {
        return getLatestVersionMaven(packageName);
    } else if (language == "go") {
        return getLatestVersionGo(packageName);
    }
    
    Utils::logError("Unsupported language: " + language);
    return "";
}

bool RegistryClient::packageExists(const std::string& language, const std::string& packageName) {
    return !getLatestVersion(language, packageName).empty();
}

bool RegistryClient::downloadPackage(const std::string& language, const std::string& packageName,
                                    const std::string& version, const std::string& destPath) {
    if (language == "node") {
        return downloadNpmPackage(packageName, version, destPath);
    } else if (language == "python") {
        return downloadPyPIPackage(packageName, version, destPath);
    } else if (language == "ruby") {
        return downloadRubyGem(packageName, version, destPath);
    } else if (language == "java") {
        return downloadMavenPackage(packageName, version, destPath);
    } else if (language == "go") {
        return downloadGoModule(packageName, version, destPath);
    }
    
    Utils::logError("Unsupported language: " + language);
    return false;
}

// NPM implementation
std::string RegistryClient::getLatestVersionNpm(const std::string& packageName) {
    std::string url = registryUrls["node"] + "/" + packageName;
    std::string response = httpGet(url);
    
    if (response.empty()) return "";
    
    try {
        json data = json::parse(response);
        if (data.contains("dist-tags") && data["dist-tags"].contains("latest")) {
            return data["dist-tags"]["latest"].get<std::string>();
        }
    } catch (const json::exception& e) {
        Utils::logError("Failed to parse npm response: " + std::string(e.what()));
    }
    
    return "";
}

bool RegistryClient::downloadNpmPackage(const std::string& packageName, const std::string& version,
                                       const std::string& destPath) {
    // For now, we'll create a placeholder directory
    // In a full implementation, we would download and extract the tarball
    Utils::logInfo("Downloading npm package: " + packageName + "@" + version);
    Utils::createDirRecursive(destPath);
    
    // Create a simple package.json
    json pkgJson;
    pkgJson["name"] = packageName;
    pkgJson["version"] = version;
    
    std::string pkgJsonPath = Utils::joinPath(destPath, "package.json");
    Utils::writeFile(pkgJsonPath, pkgJson.dump(2));
    
    return true;
}

// PyPI implementation
std::string RegistryClient::getLatestVersionPyPI(const std::string& packageName) {
    std::string url = registryUrls["python"] + "/" + packageName + "/json";
    std::string response = httpGet(url);
    
    if (response.empty()) return "";
    
    try {
        json data = json::parse(response);
        if (data.contains("info") && data["info"].contains("version")) {
            return data["info"]["version"].get<std::string>();
        }
    } catch (const json::exception& e) {
        Utils::logError("Failed to parse PyPI response: " + std::string(e.what()));
    }
    
    return "";
}

bool RegistryClient::downloadPyPIPackage(const std::string& packageName, const std::string& version,
                                        const std::string& destPath) {
    Utils::logInfo("Downloading PyPI package: " + packageName + "@" + version);
    Utils::createDirRecursive(destPath);
    
    // Create a simple metadata file
    std::string metadataPath = Utils::joinPath(destPath, "PKG-INFO");
    std::string metadata = "Name: " + packageName + "\nVersion: " + version + "\n";
    Utils::writeFile(metadataPath, metadata);
    
    return true;
}

// RubyGems implementation
std::string RegistryClient::getLatestVersionRubyGems(const std::string& packageName) {
    std::string url = registryUrls["ruby"] + "/gems/" + packageName + ".json";
    std::string response = httpGet(url);
    
    if (response.empty()) return "";
    
    try {
        json data = json::parse(response);
        if (data.contains("version")) {
            return data["version"].get<std::string>();
        }
    } catch (const json::exception& e) {
        Utils::logError("Failed to parse RubyGems response: " + std::string(e.what()));
    }
    
    return "";
}

bool RegistryClient::downloadRubyGem(const std::string& packageName, const std::string& version,
                                    const std::string& destPath) {
    Utils::logInfo("Downloading Ruby gem: " + packageName + "@" + version);
    Utils::createDirRecursive(destPath);
    
    // Create a simple gemspec placeholder
    std::string gemspecPath = Utils::joinPath(destPath, packageName + ".gemspec");
    std::string gemspec = "Gem::Specification.new do |s|\n  s.name = '" + packageName + 
                         "'\n  s.version = '" + version + "'\nend\n";
    Utils::writeFile(gemspecPath, gemspec);
    
    return true;
}

// Maven implementation
std::string RegistryClient::getLatestVersionMaven(const std::string& packageName) {
    // Maven packages use group:artifact format
    // For simplicity, we'll return a placeholder
    Utils::logWarning("Maven version lookup not fully implemented");
    return "1.0.0";
}

bool RegistryClient::downloadMavenPackage(const std::string& packageName, const std::string& version,
                                         const std::string& destPath) {
    Utils::logInfo("Downloading Maven package: " + packageName + "@" + version);
    Utils::createDirRecursive(destPath);
    
    // Create a simple pom.xml placeholder
    std::string pomPath = Utils::joinPath(destPath, "pom.xml");
    std::string pom = "<project>\n  <artifactId>" + packageName + 
                     "</artifactId>\n  <version>" + version + "</version>\n</project>\n";
    Utils::writeFile(pomPath, pom);
    
    return true;
}

// Go implementation
std::string RegistryClient::getLatestVersionGo(const std::string& packageName) {
    // Go modules use semantic versioning with v prefix
    Utils::logWarning("Go version lookup not fully implemented");
    return "v1.0.0";
}

bool RegistryClient::downloadGoModule(const std::string& packageName, const std::string& version,
                                     const std::string& destPath) {
    Utils::logInfo("Downloading Go module: " + packageName + "@" + version);
    Utils::createDirRecursive(destPath);
    
    // Create a simple go.mod placeholder
    std::string goModPath = Utils::joinPath(destPath, "go.mod");
    std::string goMod = "module " + packageName + "\n\ngo 1.21\n";
    Utils::writeFile(goModPath, goMod);
    
    return true;
}

} // namespace pkg
