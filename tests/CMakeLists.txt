cmake_minimum_required(VERSION 3.15)

# Simple test placeholder
# In a full implementation, you would use a testing framework like Google Test

add_executable(test_basic test_basic.cpp)
target_link_libraries(test_basic PRIVATE nlohmann_json::nlohmann_json)

add_test(NAME BasicTest COMMAND test_basic)
