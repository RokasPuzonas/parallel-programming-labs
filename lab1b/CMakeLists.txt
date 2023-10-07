cmake_minimum_required(VERSION 3.12)
project(lab1b)

set(CMAKE_CXX_STANDARD 20)

add_executable(lab1b main.cpp)
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fopenmp")
