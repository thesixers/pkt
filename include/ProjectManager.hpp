#pragma once

#include "GlobalRegistry.hpp"
#include "GlobalStore.hpp"
#include <string>

namespace pkg {

/**
 * Manages project lifecycle operations
 * Handles creation, initialization, deletion, and configuration of projects
 */
class ProjectManager {
public:
    ProjectManager(GlobalRegistry& registry, GlobalStore& store);
    
    // Project lifecycle
    bool createProject(const std::string& language, const std::string& name = "");
    bool initProject(const std::string& language);
    bool deleteProject(const std::string& nameOrId);
    
    // Project operations
    bool openProject(const std::string& nameOrId);
    bool runFile(const std::string& filename);
    bool setEditor(const std::string& editor);
    bool unsetEditor();
    
    // Project queries
    void listProjects();
    void searchProjects(const std::string& query);
    
    // Utility
    bool isProjectFolder() const;
    Project getCurrentProject() const;
    
private:
    GlobalRegistry& registry;
    GlobalStore& store;
    
    bool createPkgInfo(const Project& project);
    bool createPkgDeps();
    Project loadPkgInfo(const std::string& path) const;
    bool savePkgInfo(const Project& project);
    std::string getCurrentFolderName() const;
};

} // namespace pkg
