cmake_minimum_required(VERSION 3.1)
project(HE)

set(CMAKE_CXX_STANDARD 14)

set(CMAKE_MODULE_PATH ${CMAKE_MODULE_PATH} "${CMAKE_SOURCE_DIR}/cmake_modules")



find_package(GMP REQUIRED)

set(CMAKE_RUNTIME_OUTPUT_DIRECTORY "${CMAKE_CURRENT_SOURCE_DIR}/out")

INCLUDE_DIRECTORIES(
        "${GMP_INCLUDES}"
        "${CMAKE_SOURCE_DIR}/src"
)

file(GLOB SOURCES "*.c")

add_executable(HE ${SOURCES})

target_link_libraries(HE ${GMP_LIBRARIES})