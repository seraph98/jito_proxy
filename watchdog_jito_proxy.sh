#!/bin/bash

# 定义需要监控的进程名称
PROCESS_NAME="jito_proxy"
# 定义重启命令
RESTART_COMMAND="~/jito_proxy/watchdog_jito_proxy.sh"

# 创建一个无限循环来监控进程
while true; do
    # 检查 node cs.js 是否存在
    if ! pgrep -f "$PROCESS_NAME" > /dev/null; then
        echo "$(date): $PROCESS_NAME is not running. Restarting..."
        # 如果进程不存在，执行重启命令
        cd ~/jito_proxy
        $RESTART_COMMAND
    	sleep 5
    else
        echo "$(date): $PROCESS_NAME is running."
    fi

    # 每隔 5 秒检查一次（可以根据需要调整时间）
    sleep 5
done
