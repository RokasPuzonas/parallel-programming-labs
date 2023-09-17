cmake_minimum_required(VERSION 3.12)
project(lab1a)

set(CMAKE_CXX_STANDARD 20)

add_executable(lab1a main.cpp)

TARGET_LINK_LIBRARIES(lab1a pthread)
