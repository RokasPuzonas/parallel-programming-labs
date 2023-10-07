#!/bin/sh
mkdir -p build
cd build
cmake ..
make
./lab1b ../$1 ../$2
