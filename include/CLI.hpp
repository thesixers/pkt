#pragma once

#include "ProjectManager.hpp"
#include "DependencyManager.hpp"
#include <string>
#include <vector>

namespace pkg {

/**
 * Command-line interface parser and router
 * Handles argument parsing and command execution
 */
class CLI {
public:
    CLI(ProjectManager& projectMgr, DependencyManager& depMgr);
    
    // Main entry point
    int run(int argc, char* argv[]);
    
private:
    ProjectManager& projectMgr;
    DependencyManager& depMgr;
    
    // Command handlers
    int handleCreate(const std::vector<std::string>& args);
    int handleInit(const std::vector<std::string>& args);
    int handleOpen(const std::vector<std::string>& args);
    int handleEditor(const std::vector<std::string>& args);
    int handleSearch(const std::vector<std::string>& args);
    int handleProjects(const std::vector<std::string>& args);
    int handleDelete(const std::vector<std::string>& args);
    int handleAdd(const std::vector<std::string>& args);
    int handleRemove(const std::vector<std::string>& args);
    int handleUpdate(const std::vector<std::string>& args);
    int handleDeps(const std::vector<std::string>& args);
    int handleHelp(const std::vector<std::string>& args);
    int handleVersion(const std::vector<std::string>& args);
    
    // Utilities
    void printHelp();
    void printVersion();
    std::string getOption(const std::vector<std::string>& args, const std::string& option, 
                         const std::string& defaultValue = "");
    bool hasOption(const std::vector<std::string>& args, const std::string& option);
};

} // namespace pkg
