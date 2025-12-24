#pragma once

#include <string>
#include <vector>
#include <map>
#include <nlohmann/json.hpp>

namespace pkg {

using json = nlohmann::json;

/**
 * Project metadata structure
 */
struct Project {
    std::string id;
    std::string name;
    std::string path;
    std::string language;
    std::string defaultDepFolder;
    std::string createdAt;
    std::string defaultEditor;
    std::vector<std::string> tags;
    std::string description;
};

/**
 * Manages the global project registry (~/.pkg_registry.json)
 * Provides CRUD operations for projects and search functionality.
 */
class GlobalRegistry {
public:
    GlobalRegistry();
    
    // Project operations
    bool addProject(const Project& project);
    bool removeProject(const std::string& projectId);
    bool updateProject(const Project& project);
    Project getProject(const std::string& nameOrId) const;
    std::vector<Project> getAllProjects() const;
    bool projectExists(const std::string& nameOrId) const;
    
    // Search operations
    std::vector<Project> searchProjects(const std::string& query) const;
    std::vector<Project> getProjectsByLanguage(const std::string& language) const;
    
    // Utility
    std::string generateProjectId() const;
    void save();
    void load();
    
private:
    std::string registryPath;
    json registryData;
    
    void ensureRegistryExists();
    Project jsonToProject(const json& j) const;
    json projectToJson(const Project& project) const;
};

} // namespace pkg
