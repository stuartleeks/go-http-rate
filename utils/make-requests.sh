#!/bin/bash

for INDEX in {1..300}; do curl -s http://localhost:8080/api; done | awk '{arr[$1]+=1}END {for (i in arr) print i,arr[i]}' | sort

