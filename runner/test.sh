#!/bin/bash

for i in {1..3}; do
    echo "step " $i
    sleep 1s
    echo >&2 "stderr!!"
done
