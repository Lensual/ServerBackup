#!/bin/bash

#清理旧备份
echo 备份存储路径：$1
echo 备份名：$2
echo 保留数量：$3 

backups=($(ls -tr ${1}/${2}?*)) #找到所有备份文件名时间升序排序

if [ ${#backups[@]} -gt $3 ]; then
    echo 大于备份保留数量，需要清理
    num=$(expr ${#backups[@]} - $3)
    for ((i = 0; i < num; i++)); do
        echo 删除："${backups[${i}]}"
        rm -rf "${backups[${i}]}"
    done
fi
