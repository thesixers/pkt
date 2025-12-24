cmake_minimum_required(VERSION 3.15)
project(pkt VERSION 1.0.0 LANGUAGES CXX)

# Set C++ standard
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# Enable testing
enable_testing()

# Fetch nlohmann/json
include(FetchContent)
FetchContent_Declare(
    json
    GIT_REPOSITORY https://github.com/nlohmann/json.git
    GIT_TAG v3.11.3
)
FetchContent_MakeAvailable(json)

# Find libcurl
find_package(CURL REQUIRED)

# Include directories
include_directories(${PROJECT_SOURCE_DIR}/include)

# Source files
set(SOURCES
    src/main.cpp
    src/CLI.cpp
    src/Utils.cpp
    src/GlobalRegistry.cpp
    src/GlobalStore.cpp
    src/ProjectManager.cpp
    src/DependencyManager.cpp
    src/RegistryClient.cpp
)

# Create executable
add_executable(pkt ${SOURCES})

# Link libraries
target_link_libraries(pkt PRIVATE nlohmann_json::nlohmann_json CURL::libcurl)

if(WIN32)
    target_link_libraries(pkt PRIVATE ws2_32 crypt32 wldap32 normaliz)
endif()

# Compiler warnings
if(MSVC)
    target_compile_options(pkt PRIVATE /W4)
    target_compile_definitions(pkt PRIVATE _CRT_SECURE_NO_WARNINGS)
else()
    target_compile_options(pkt PRIVATE -Wall -Wextra -Wpedantic)
endif()

# Install target
install(TARGETS pkt DESTINATION bin)

# Tests (optional, can be expanded later)
add_subdirectory(tests EXCLUDE_FROM_ALL)
