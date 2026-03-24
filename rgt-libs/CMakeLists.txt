cmake_minimum_required(VERSION 3.20)
project(rgt-library
        VERSION 1.16.0 LANGUAGES C)

set(CMAKE_C_STANDARD 99)

if(EXISTS "${CMAKE_BINARY_DIR}/generators/conan_toolchain.cmake")
    set(CMAKE_TOOLCHAIN_FILE "${CMAKE_BINARY_DIR}/generators/conan_toolchain.cmake")
elseif(EXISTS "${CMAKE_BINARY_DIR}/build/generators/conan_toolchain.cmake")
    set(CMAKE_TOOLCHAIN_FILE "${CMAKE_BINARY_DIR}/build/generators/conan_toolchain.cmake")
elseif(EXISTS "${CMAKE_BINARY_DIR}/build/Release/generators/conan_toolchain.cmake")
    set(CMAKE_TOOLCHAIN_FILE "${CMAKE_BINARY_DIR}/build/Release/generators/conan_toolchain.cmake")
endif()

if(CMAKE_TOOLCHAIN_FILE)
    message(STATUS "Using Conan toolchain: ${CMAKE_TOOLCHAIN_FILE}")
    include(${CMAKE_TOOLCHAIN_FILE})
endif()

# Include the Harbour compiler configuration
include(cmake/HarbourCompiler.cmake)
include_directories(${HB_INC})

find_package(cfl-library CONFIG REQUIRED)
include_directories(${cfl-library_INCLUDE_DIRS})

add_definitions(-D__HBR__)

add_subdirectory(rgt-common-lib)
add_subdirectory(rgt-app-lib)
add_subdirectory(rgt-terminal-lib)

install(DIRECTORY "${PROJECT_SOURCE_DIR}/rgt-common-headers/src/main/headers/"
    DESTINATION include
    FILES_MATCHING PATTERN "*.ch"
)

install(DIRECTORY "${PROJECT_SOURCE_DIR}/rgt-app-headers/src/main/headers/"
    DESTINATION include
    FILES_MATCHING PATTERN "*.ch"
)

install(DIRECTORY "${PROJECT_SOURCE_DIR}/rgt-terminal-headers/src/main/headers/"
    DESTINATION include
    FILES_MATCHING PATTERN "*.ch"
)

install(TARGETS rgt-common-lib rgt-app-lib rgt-terminal-lib
    ARCHIVE DESTINATION lib
)

set(CPACK_GENERATOR "ZIP")
set(CPACK_PACKAGE_FILE_NAME "${CMAKE_PROJECT_NAME}")
set(CPACK_INCLUDE_TOPLEVEL_DIRECTORY OFF)
set(CPACK_VERBATIM_VARIABLES YES)
include(CPack)
