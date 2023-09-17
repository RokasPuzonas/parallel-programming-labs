#!/bin/sh
mkdir -p build
cd build
cmake ..
make
./lab1a ../$1 ../$2
