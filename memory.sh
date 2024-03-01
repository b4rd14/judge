#!/bin/bash

# Function to check memory usage of a Python process
check_python_memory_usage() {
    local pid=$1
    local python_program=$2

    # Get memory usage of the Python process in KB
    local mem_usage=$(ps -o rss= -p $pid)

    # Convert memory usage to MB (rounded to nearest integer)
    local mem_usage_mb=$(( mem_usage / 1024 ))

    # Check if memory usage exceeds threshold (e.g., 256 MB)
    if [ "$mem_usage_mb" -gt 256 ]; then
        echo "Memory usage of Python process with PID $pid: $mem_usage_mb MB"
        echo "Memory usage exceeded threshold. Killing Python process with PID $pid..."
        kill -15 $pid
        return 0
    fi

    return 1
}


if [ $# -ne 1 ]; then
    echo "Usage: $0 <python_program>"
    exit 1
fi

python_program="$1"
# Infinite loop to continuously monitor Python processes
while true; do
    # Get the PID of all Python processes
    python_pids=$(pgrep -f "python3")


    # Iterate over each Python process and check its memory usage
    for pid in $python_pids; do
        # Get the command associated with the process
        cmd=$(ps -o cmd= -p $pid)
        # Check if the command contains the name of your Python program
        if echo "$cmd" | grep -q "$python_program"; then
             check_python_memory_usage $pid "$python_program"
        fi
    done
done
