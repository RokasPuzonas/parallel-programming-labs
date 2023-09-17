#!/bin/sh
COUNT=40
go run gen-data.go IF-1-1_PuzonasR_L1_dat_1.json $COUNT 1.0
go run gen-data.go IF-1-1_PuzonasR_L1_dat_2.json $COUNT 0.5
go run gen-data.go IF-1-1_PuzonasR_L1_dat_3.json $COUNT 0.0
