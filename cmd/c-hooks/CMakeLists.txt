cmake_minimum_required(VERSION 3.10)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON CACHE INTERNAL "") # works
set(CMAKE_CXX_FLAGS "-Werror -Wall -Wextra")

include_directories(".")
include_directories("C:/Users/alexw/repos/nats.c/src")
link_directories("C:/Users/alexw/repos/nats.c/ninja/src")

# include = C:/Users/alexw/repos/nats.c/src

# lib = C:/Users/alexw/repos/nats.c/ninja/src/libnats_static.a
# lib = C:/Users/alexw/repos/nats.c/ninja/src/libnats.dll
# lib = C:/Users/alexw/repos/nats.c/ninja/src/libnats.dll.a

project(c-hooks)

add_executable(c-hooks main.cpp)
