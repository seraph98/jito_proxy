#!/bin/bash

git stash
git reset --hard origin/master
git pull origin master


# 获取包含关键字的进程列表
processes=$(ps -ef | grep jito_proxy | grep -v "grep" | awk '{print $2}')

# 遍历进程列表并终止进程
for pid in $processes; do
    echo "Terminating process with PID: $pid"
    kill -9 $pid
done

nohup ./jito_proxy > /dev/null 2>&1 &
