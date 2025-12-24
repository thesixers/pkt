#include "ProjectManager.hpp"
#include "Utils.hpp"
#include <iostream>
#include <fstream>

namespace pkg {

ProjectManager::ProjectManager(GlobalRegistry& registry, GlobalStore& store)
    : registry(registry), store(store) {}

bool ProjectManager::createProject(const std::string& language) {
    std::string currentDir = Utils::getCurrentDir();
    std::string folderName = getCurrentFolderName();
    
    // Check if already a project
    if (isProjectFolder()) {
        Utils::logError("Current directory is already a PKG project");
        return false;
    }
    
    // Validate language
    auto supportedLangs = store.getSupportedLanguages();
    if (std::find(supportedLangs.begin(), supportedLangs.end(), language) == supportedLangs.end()) {
        Utils::logError("Unsupported language: " + language);
        Utils::logInfo("Supported languages: node, python, ruby, java, go");
        return false;
    }
    
    // Create project structure
    Project project;
    project.id = registry.generateProjectId();
    project.name = folderName;
    project.path = currentDir;
    project.language = language;
    project.defaultDepFolder = store.getDepFolderForLanguage(language);
    project.createdAt = Utils::getCurrentTimestamp();
    
    // Create .pkg.info
    if (!createPkgInfo(project)) {
        Utils::logError("Failed to create .pkg.info");
        return false;
    }
    
    // Create .pkg.deps
    if (!createPkgDeps()) {
        Utils::logError("Failed to create .pkg.deps");
        return false;
    }
    
    // Create dependency folder
    std::string depFolder = Utils::joinPath(currentDir, project.defaultDepFolder);
    Utils::createDir(depFolder);
    
    // Register project
    if (!registry.addProject(project)) {
        Utils::logError("Failed to register project");
        return false;
    }
    
    Utils::logSuccess("Created project '" + project.name + "' (" + language + ")");
    return true;
}

bool ProjectManager::initProject(const std::string& language) {
    std::string currentDir = Utils::getCurrentDir();
    std::string folderName = getCurrentFolderName();
    
    // Check if already registered
    if (registry.projectExists(folderName)) {
        Utils::logError("Project '" + folderName + "' is already registered");
        return false;
    }
    
    // If .pkg.info exists, load it
    std::string pkgInfoPath = Utils::joinPath(currentDir, ".pkg.info");
    if (Utils::fileExists(pkgInfoPath)) {
        Project project = loadPkgInfo(currentDir);
        if (!project.id.empty()) {
            registry.addProject(project);
            Utils::logSuccess("Registered existing project '" + project.name + "'");
            return true;
        }
    }
    
    // Otherwise, create new project
    return createProject(language);
}

bool ProjectManager::deleteProject(const std::string& nameOrId) {
    Project project = registry.getProject(nameOrId);
    
    if (project.id.empty()) {
        Utils::logError("Project not found: " + nameOrId);
        return false;
    }
    
    // Confirm deletion
    std::cout << "Are you sure you want to delete project '" << project.name 
              << "' at " << project.path << "? (y/N): ";
    std::string response;
    std::getline(std::cin, response);
    
    if (response != "y" && response != "Y") {
        Utils::logInfo("Deletion cancelled");
        return false;
    }
    
    // Remove from registry
    if (!registry.removeProject(project.id)) {
        Utils::logError("Failed to remove project from registry");
        return false;
    }
    
    // Optionally remove project files
    std::cout << "Delete project files? (y/N): ";
    std::getline(std::cin, response);
    
    if (response == "y" || response == "Y") {
        Utils::removeFile(Utils::joinPath(project.path, ".pkg.info"));
        Utils::removeFile(Utils::joinPath(project.path, ".pkg.deps"));
        Utils::logSuccess("Deleted project files");
    }
    
    Utils::logSuccess("Removed project '" + project.name + "' from registry");
    return true;
}

bool ProjectManager::openProject(const std::string& nameOrId) {
    Project project = registry.getProject(nameOrId);
    
    if (project.id.empty()) {
        // Try fuzzy search
        auto results = registry.searchProjects(nameOrId);
        if (results.empty()) {
            Utils::logError("Project not found: " + nameOrId);
            return false;
        }
        
        if (results.size() > 1) {
            std::cout << "Multiple projects found:\n";
            for (size_t i = 0; i < results.size(); ++i) {
                std::cout << "  " << (i + 1) << ". " << results[i].name 
                         << " (" << results[i].language << ")\n";
            }
            std::cout << "Select project (1-" << results.size() << "): ";
            
            int choice;
            std::cin >> choice;
            std::cin.ignore();
            
            if (choice < 1 || choice > static_cast<int>(results.size())) {
                Utils::logError("Invalid selection");
                return false;
            }
            
            project = results[choice - 1];
        } else {
            project = results[0];
        }
    }
    
    // Check if editor is set
    if (project.defaultEditor.empty()) {
        Utils::logError("No default editor set. Use 'pkg editor set <command>' to set one");
        return false;
    }
    
    // Open project
    std::string command = project.defaultEditor + " " + project.path;
    Utils::logInfo("Opening project: " + project.name);
    
    return system(command.c_str()) == 0;
}

bool ProjectManager::setEditor(const std::string& editor) {
    if (!isProjectFolder()) {
        Utils::logError("Not in a PKG project directory");
        return false;
    }
    
    Project project = getCurrentProject();
    project.defaultEditor = editor;
    
    if (!savePkgInfo(project)) {
        Utils::logError("Failed to save editor setting");
        return false;
    }
    
    registry.updateProject(project);
    Utils::logSuccess("Set default editor to: " + editor);
    return true;
}

bool ProjectManager::unsetEditor() {
    if (!isProjectFolder()) {
        Utils::logError("Not in a PKG project directory");
        return false;
    }
    
    Project project = getCurrentProject();
    project.defaultEditor = "";
    
    if (!savePkgInfo(project)) {
        Utils::logError("Failed to unset editor");
        return false;
    }
    
    registry.updateProject(project);
    Utils::logSuccess("Unset default editor");
    return true;
}

void ProjectManager::listProjects() {
    auto projects = registry.getAllProjects();
    
    if (projects.empty()) {
        Utils::logInfo("No projects registered");
        return;
    }
    
    std::cout << "\n📦 Registered Projects:\n\n";
    for (const auto& proj : projects) {
        std::cout << "  • " << proj.name << " (" << proj.language << ")\n";
        std::cout << "    Path: " << proj.path << "\n";
        std::cout << "    ID: " << proj.id << "\n";
        if (!proj.defaultEditor.empty()) {
            std::cout << "    Editor: " << proj.defaultEditor << "\n";
        }
        std::cout << "\n";
    }
}

void ProjectManager::searchProjects(const std::string& query) {
    auto results = registry.searchProjects(query);
    
    if (results.empty()) {
        Utils::logInfo("No projects found matching: " + query);
        return;
    }
    
    std::cout << "\n🔍 Search Results for '" << query << "':\n\n";
    for (const auto& proj : results) {
        std::cout << "  • " << proj.name << " (" << proj.language << ")\n";
        std::cout << "    Path: " << proj.path << "\n";
        std::cout << "\n";
    }
}

bool ProjectManager::isProjectFolder() const {
    std::string pkgInfoPath = Utils::joinPath(Utils::getCurrentDir(), ".pkg.info");
    return Utils::fileExists(pkgInfoPath);
}

Project ProjectManager::getCurrentProject() const {
    return loadPkgInfo(Utils::getCurrentDir());
}

bool ProjectManager::createPkgInfo(const Project& project) {
    return savePkgInfo(project);
}

bool ProjectManager::createPkgDeps() {
    std::string depsPath = Utils::joinPath(Utils::getCurrentDir(), ".pkg.deps");
    json emptyDeps = json::object();
    return Utils::writeFile(depsPath, emptyDeps.dump(2));
}

Project ProjectManager::loadPkgInfo(const std::string& path) const {
    std::string pkgInfoPath = Utils::joinPath(path, ".pkg.info");
    
    if (!Utils::fileExists(pkgInfoPath)) {
        return Project();
    }
    
    std::string content = Utils::readFile(pkgInfoPath);
    if (content.empty()) {
        return Project();
    }
    
    try {
        json data = json::parse(content);
        
        Project project;
        project.id = data.value("id", "");
        project.name = data.value("name", "");
        project.path = data.value("path", "");
        project.language = data.value("language", "");
        project.defaultDepFolder = data.value("default_dep_folder", "");
        project.createdAt = data.value("created_at", "");
        project.defaultEditor = data.value("default_editor", "");
        project.description = data.value("description", "");
        
        if (data.contains("tags") && data["tags"].is_array()) {
            for (const auto& tag : data["tags"]) {
                project.tags.push_back(tag.get<std::string>());
            }
        }
        
        return project;
    } catch (const json::exception& e) {
        Utils::logError("Failed to parse .pkg.info: " + std::string(e.what()));
        return Project();
    }
}

bool ProjectManager::savePkgInfo(const Project& project) {
    std::string pkgInfoPath = Utils::joinPath(project.path, ".pkg.info");
    
    json data;
    data["id"] = project.id;
    data["name"] = project.name;
    data["path"] = project.path;
    data["language"] = project.language;
    data["default_dep_folder"] = project.defaultDepFolder;
    data["created_at"] = project.createdAt;
    
    if (!project.defaultEditor.empty()) {
        data["default_editor"] = project.defaultEditor;
    }
    
    if (!project.description.empty()) {
        data["description"] = project.description;
    }
    
    if (!project.tags.empty()) {
        data["tags"] = project.tags;
    }
    
    return Utils::writeFile(pkgInfoPath, data.dump(2));
}

std::string ProjectManager::getCurrentFolderName() const {
    std::string currentDir = Utils::getCurrentDir();
    return Utils::getFileName(currentDir);
}

} // namespace pkg
