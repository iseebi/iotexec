#!/bin/bash
read line
current_dir=$(pwd)
file=$current_dir"/democmd.log"
echo "$(date) $line" >> $file