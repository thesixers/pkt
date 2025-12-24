#include "CLI.hpp"
#include "GlobalRegistry.hpp"
#include "GlobalStore.hpp"
#include "RegistryClient.hpp"
#include "ProjectManager.hpp"
#include "DependencyManager.hpp"
#include "Utils.hpp"
#include <iostream>

int main(int argc, char* argv[]) {
    try {
        // Initialize core components
        pkg::GlobalRegistry registry;
        pkg::GlobalStore store;
        pkg::RegistryClient client;
        pkg::ProjectManager projectMgr(registry, store);
        pkg::DependencyManager depMgr(registry, store, client);
        pkg::CLI cli(projectMgr, depMgr);
        
        // Run CLI
        return cli.run(argc, argv);
    } catch (const std::exception& e) {
        pkg::Utils::logError("Fatal error: " + std::string(e.what()));
        return 1;
    }
}
