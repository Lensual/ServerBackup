#!/bin/bash
function cleanArgs() {
    exclude=""
    host=""
    port=""
    user=""
    sshKey=""
    srcDir=""
    backupDir=""
    bwlimit=""
    excludeList=""
    option=""
    basename=""
    nice=""
}

workdir=/root/script/ServerBackup
logDir=$workdir/logs

configs=$(ls $workdir"/conf/"*.conf)
for conf in $configs
do
    cleanArgs
    source $conf

    basename=$(basename $conf)
    filename=${basename%%.*}

    for e in ${excludeList[@]}
    do
        exclude="$exclude --exclude '$e'"
    done

    timestamp=$(date +%Y-%m-%d_%T)

	#TODO 改变优先级
	#TODO 并行优化
	
	#奇怪问题：如果不使用bash单独运行rsync，在脚本中执行会忽略掉排除列表
    nice -n ${nice} bash -c "rsync $option --delete --bwlimit=${bwlimit} -e \"ssh -p ${port} -i ${sshKey} -o 'StrictHostKeyChecking no'\" ${exclude} ${user}@${host}:${srcDir} ${backupDir}/${filename}/ | tee ${logDir}/${filename}_$timestamp.log"

	#TODO 自动检测cpu核心数
	#TODO 参数使用变量配置
    nice -n ${nice} bash -c "tar cpf - ${backupDir}/${filename}/ | 7za a -t7z -m0=lzma2 -mx9 -mmt8 -si ${backupDir}/${filename}_$timestamp.7z"
done

