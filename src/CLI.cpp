#include "CLI.hpp"
#include "Utils.hpp"
#include <iostream>
#include <algorithm>

namespace pkg {

CLI::CLI(ProjectManager& projectMgr, DependencyManager& depMgr)
    : projectMgr(projectMgr), depMgr(depMgr) {}

int CLI::run(int argc, char* argv[]) {
    if (argc < 2) {
        printHelp();
        return 0;
    }
    
    std::string command = argv[1];
    std::vector<std::string> args;
    
    for (int i = 2; i < argc; ++i) {
        args.push_back(argv[i]);
    }
    
    // Route to appropriate handler
    if (command == "create") {
        return handleCreate(args);
    } else if (command == "init") {
        return handleInit(args);
    } else if (command == "open") {
        return handleOpen(args);
    } else if (command == "editor") {
        return handleEditor(args);
    } else if (command == "search") {
        return handleSearch(args);
    } else if (command == "projects") {
        return handleProjects(args);
    } else if (command == "delete") {
        return handleDelete(args);
    } else if (command == "add") {
        return handleAdd(args);
    } else if (command == "remove" || command == "rm") {
        return handleRemove(args);
    } else if (command == "update") {
        return handleUpdate(args);
    } else if (command == "deps") {
        return handleDeps(args);
    } else if (command == "help" || command == "--help" || command == "-h") {
        return handleHelp(args);
    } else if (command == "version" || command == "--version" || command == "-v") {
        return handleVersion(args);
    } else {
        Utils::logError("Unknown command: " + command);
        std::cout << "Run 'pkg help' for usage information\n";
        return 1;
    }
}

int CLI::handleCreate(const std::vector<std::string>& args) {
    std::string language = getOption(args, "--language", "");
    
    if (language.empty() && hasOption(args, "--lang")) {
        language = getOption(args, "--lang", "");
    }
    
    if (language.empty()) {
        Utils::logError("Language not specified");
        std::cout << "Usage: pkg create --language <language>\n";
        std::cout << "Supported: node, python, ruby, java, go\n";
        return 1;
    }
    
    return projectMgr.createProject(language) ? 0 : 1;
}

int CLI::handleInit(const std::vector<std::string>& args) {
    std::string language = getOption(args, "--language", "");
    
    if (language.empty() && hasOption(args, "--lang")) {
        language = getOption(args, "--lang", "");
    }
    
    if (language.empty()) {
        Utils::logError("Language not specified");
        std::cout << "Usage: pkg init --language <language>\n";
        return 1;
    }
    
    return projectMgr.initProject(language) ? 0 : 1;
}

int CLI::handleOpen(const std::vector<std::string>& args) {
    if (args.empty()) {
        Utils::logError("Project name or ID not specified");
        std::cout << "Usage: pkg open <project_name_or_id>\n";
        return 1;
    }
    
    return projectMgr.openProject(args[0]) ? 0 : 1;
}

int CLI::handleEditor(const std::vector<std::string>& args) {
    if (args.empty()) {
        Utils::logError("Editor command not specified");
        std::cout << "Usage: pkg editor set <command>\n";
        std::cout << "       pkg editor unset\n";
        return 1;
    }
    
    std::string subcommand = args[0];
    
    if (subcommand == "set") {
        if (args.size() < 2) {
            Utils::logError("Editor command not specified");
            std::cout << "Usage: pkg editor set <command>\n";
            return 1;
        }
        
        return projectMgr.setEditor(args[1]) ? 0 : 1;
    } else if (subcommand == "unset") {
        return projectMgr.unsetEditor() ? 0 : 1;
    } else {
        Utils::logError("Unknown editor subcommand: " + subcommand);
        std::cout << "Usage: pkg editor set <command>\n";
        std::cout << "       pkg editor unset\n";
        return 1;
    }
}

int CLI::handleSearch(const std::vector<std::string>& args) {
    if (args.empty()) {
        Utils::logError("Search query not specified");
        std::cout << "Usage: pkg search <query>\n";
        return 1;
    }
    
    projectMgr.searchProjects(args[0]);
    return 0;
}

int CLI::handleProjects(const std::vector<std::string>& args) {
    projectMgr.listProjects();
    return 0;
}

int CLI::handleDelete(const std::vector<std::string>& args) {
    if (args.empty()) {
        Utils::logError("Project name or ID not specified");
        std::cout << "Usage: pkg delete <project_name_or_id>\n";
        return 1;
    }
    
    return projectMgr.deleteProject(args[0]) ? 0 : 1;
}

int CLI::handleAdd(const std::vector<std::string>& args) {
    if (args.empty()) {
        Utils::logError("Package not specified");
        std::cout << "Usage: pkg add <package>[@<version>]\n";
        return 1;
    }
    
    return depMgr.addDependency(args[0]) ? 0 : 1;
}

int CLI::handleRemove(const std::vector<std::string>& args) {
    if (args.empty()) {
        Utils::logError("Package not specified");
        std::cout << "Usage: pkg remove <package>\n";
        return 1;
    }
    
    return depMgr.removeDependency(args[0]) ? 0 : 1;
}

int CLI::handleUpdate(const std::vector<std::string>& args) {
    if (args.empty()) {
        Utils::logError("Package not specified");
        std::cout << "Usage: pkg update <package>[@<version>]\n";
        return 1;
    }
    
    return depMgr.updateDependency(args[0]) ? 0 : 1;
}

int CLI::handleDeps(const std::vector<std::string>& args) {
    if (args.empty() || args[0] != "list") {
        Utils::logError("Unknown deps subcommand");
        std::cout << "Usage: pkg deps list [--global] [--lang <language>] [--all]\n";
        return 1;
    }
    
    bool isGlobal = hasOption(args, "--global");
    
    if (isGlobal) {
        bool allLanguages = hasOption(args, "--all");
        std::string language = getOption(args, "--lang", "node");
        
        depMgr.listGlobalDeps(language, allLanguages);
    } else {
        depMgr.listProjectDeps();
    }
    
    return 0;
}

int CLI::handleHelp(const std::vector<std::string>& args) {
    printHelp();
    return 0;
}

int CLI::handleVersion(const std::vector<std::string>& args) {
    printVersion();
    return 0;
}

void CLI::printHelp() {
    std::cout << R"(
PKT - Universal Package Manager

USAGE:
    pkt <command> [options]

PROJECT COMMANDS:
    create --language <lang>     Create a new project in current directory
    init --language <lang>       Initialize existing directory as project
    open <name_or_id>           Open project in default editor
    editor set <command>        Set default editor for current project
    editor unset                Unset default editor
    search <query>              Search for projects
    projects                    List all registered projects
    delete <name_or_id>         Delete a project

DEPENDENCY COMMANDS:
    add <package>[@<version>]   Add a dependency to current project
    remove <package>            Remove a dependency
    update <package>[@<version>] Update a dependency
    deps list                   List project dependencies
    deps list --global          List global dependencies
    deps list --global --all    List all global dependencies

SUPPORTED LANGUAGES:
    node, python, ruby, java, go

EXAMPLES:
    pkt create --language node
    pkt add react@18.3.0
    pkt add fastify
    pkt deps list
    pkt open my-project
    pkt editor set code

For more information, visit: https://github.com/yourusername/pkt
)";
}

void CLI::printVersion() {
    std::cout << "PKT version 1.0.0\n";
}

std::string CLI::getOption(const std::vector<std::string>& args, const std::string& option,
                           const std::string& defaultValue) {
    for (size_t i = 0; i < args.size(); ++i) {
        if (args[i] == option && i + 1 < args.size()) {
            return args[i + 1];
        }
    }
    return defaultValue;
}

bool CLI::hasOption(const std::vector<std::string>& args, const std::string& option) {
    return std::find(args.begin(), args.end(), option) != args.end();
}

} // namespace pkg
