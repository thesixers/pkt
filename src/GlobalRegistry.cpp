#include "GlobalRegistry.hpp"
#include "Utils.hpp"
#include <fstream>

namespace pkg {

GlobalRegistry::GlobalRegistry() {
    registryPath = Utils::joinPath(Utils::getHomeDir(), ".pkg_registry.json");
    ensureRegistryExists();
    load();
}

void GlobalRegistry::ensureRegistryExists() {
    if (!Utils::fileExists(registryPath)) {
        json emptyRegistry = json::object();
        emptyRegistry["projects"] = json::array();
        emptyRegistry["version"] = "1.0.0";
        Utils::writeFile(registryPath, emptyRegistry.dump(2));
    }
}

void GlobalRegistry::load() {
    std::string content = Utils::readFile(registryPath);
    if (!content.empty()) {
        try {
            registryData = json::parse(content);
        } catch (const json::exception& e) {
            Utils::logError("Failed to parse registry: " + std::string(e.what()));
            registryData = json::object();
            registryData["projects"] = json::array();
        }
    }
}

void GlobalRegistry::save() {
    Utils::writeFile(registryPath, registryData.dump(2));
}

bool GlobalRegistry::addProject(const Project& project) {
    if (projectExists(project.name)) {
        Utils::logError("Project with name '" + project.name + "' already exists");
        return false;
    }
    
    registryData["projects"].push_back(projectToJson(project));
    save();
    return true;
}

bool GlobalRegistry::removeProject(const std::string& projectId) {
    auto& projects = registryData["projects"];
    
    for (size_t i = 0; i < projects.size(); ++i) {
        if (projects[i]["id"] == projectId) {
            projects.erase(i);
            save();
            return true;
        }
    }
    
    return false;
}

bool GlobalRegistry::updateProject(const Project& project) {
    auto& projects = registryData["projects"];
    
    for (auto& proj : projects) {
        if (proj["id"] == project.id) {
            proj = projectToJson(project);
            save();
            return true;
        }
    }
    
    return false;
}

Project GlobalRegistry::getProject(const std::string& nameOrId) const {
    auto& projects = registryData["projects"];
    
    // First try exact ID match
    for (const auto& proj : projects) {
        if (proj["id"] == nameOrId) {
            return jsonToProject(proj);
        }
    }
    
    // Then try exact name match
    for (const auto& proj : projects) {
        if (proj["name"] == nameOrId) {
            return jsonToProject(proj);
        }
    }
    
    // Return empty project if not found
    return Project();
}

std::vector<Project> GlobalRegistry::getAllProjects() const {
    std::vector<Project> projects;
    
    for (const auto& proj : registryData["projects"]) {
        projects.push_back(jsonToProject(proj));
    }
    
    return projects;
}

bool GlobalRegistry::projectExists(const std::string& nameOrId) const {
    auto& projects = registryData["projects"];
    
    for (const auto& proj : projects) {
        if (proj["id"] == nameOrId || proj["name"] == nameOrId) {
            return true;
        }
    }
    
    return false;
}

std::vector<Project> GlobalRegistry::searchProjects(const std::string& query) const {
    std::vector<std::pair<int, Project>> scoredProjects;
    
    for (const auto& proj : registryData["projects"]) {
        Project project = jsonToProject(proj);
        
        // Calculate fuzzy match score
        int nameScore = Utils::fuzzyMatch(query, project.name);
        int descScore = Utils::fuzzyMatch(query, project.description);
        int maxScore = std::max(nameScore, descScore);
        
        // Check tags
        for (const auto& tag : project.tags) {
            int tagScore = Utils::fuzzyMatch(query, tag);
            maxScore = std::max(maxScore, tagScore);
        }
        
        if (maxScore > 0) {
            scoredProjects.push_back({maxScore, project});
        }
    }
    
    // Sort by score (descending)
    std::sort(scoredProjects.begin(), scoredProjects.end(),
              [](const auto& a, const auto& b) { return a.first > b.first; });
    
    // Extract projects
    std::vector<Project> results;
    for (const auto& pair : scoredProjects) {
        results.push_back(pair.second);
    }
    
    return results;
}

std::vector<Project> GlobalRegistry::getProjectsByLanguage(const std::string& language) const {
    std::vector<Project> projects;
    
    for (const auto& proj : registryData["projects"]) {
        Project project = jsonToProject(proj);
        if (project.language == language) {
            projects.push_back(project);
        }
    }
    
    return projects;
}

std::string GlobalRegistry::generateProjectId() const {
    return Utils::generateId("proj-");
}

Project GlobalRegistry::jsonToProject(const json& j) const {
    Project project;
    
    project.id = j.value("id", "");
    project.name = j.value("name", "");
    project.path = j.value("path", "");
    project.language = j.value("language", "");
    project.defaultDepFolder = j.value("default_dep_folder", "");
    project.createdAt = j.value("created_at", "");
    project.defaultEditor = j.value("default_editor", "");
    project.description = j.value("description", "");
    
    if (j.contains("tags") && j["tags"].is_array()) {
        for (const auto& tag : j["tags"]) {
            project.tags.push_back(tag.get<std::string>());
        }
    }
    
    return project;
}

json GlobalRegistry::projectToJson(const Project& project) const {
    json j;
    
    j["id"] = project.id;
    j["name"] = project.name;
    j["path"] = project.path;
    j["language"] = project.language;
    j["default_dep_folder"] = project.defaultDepFolder;
    j["created_at"] = project.createdAt;
    
    if (!project.defaultEditor.empty()) {
        j["default_editor"] = project.defaultEditor;
    }
    
    if (!project.description.empty()) {
        j["description"] = project.description;
    }
    
    if (!project.tags.empty()) {
        j["tags"] = project.tags;
    }
    
    return j;
}

} // namespace pkg
