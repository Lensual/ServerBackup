#!/bin/bash

#程序入口
#$1 指定配置文件

workdir=$(dirname $0) #工作路径
logDir=$workdir/logs  #日志路径
mkdir -p $logDir
cpuNum=$(nproc) #cpu数量

#单一备份
if [ $1 ]; then
    if [ -s $workdir/conf/$1 ]; then
        source $workdir/backup.sh $workdir/conf/$1
        exit
    fi
    echo "invalid argument"
    exit
fi

#全部备份
configs=$(ls $workdir"/conf/"*.conf)
for conf in $configs; do
    source $workdir/backup.sh $conf
done
