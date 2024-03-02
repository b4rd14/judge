#!/bin/bash

check_python_memory_usage() {
    local pid=$1
    local python_program=$2


    local mem_usage=$(ps -o rss= -p $pid)

    local mem_usage_mb=$(( mem_usage / 1024 ))

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
while true; do
    python_pids=$(pgrep -f "python3")
    for pid in $python_pids; do
        cmd=$(ps -o cmd= -p $pid)
        if echo "$cmd" | grep -q "$python_program"; then
             check_python_memory_usage $pid "$python_program"
        fi
    done
done
