#!/bin/bash

#清理旧备份
#$1 备份存储路径 backupDir
#$2 备份名 filename
#$3 保留数量

backups=($(ls -tr ${1}/${2}?*)) #找到所有备份文件名时间升序排序

if [ ${#backups[@]} -gt $3 ]; then
    num=$(expr ${#backups[@]} - $3)
    for ((i = 0; i < num; i++)); do
        rm -rf "${1}/${backups[${i}]}"
    done
fi
