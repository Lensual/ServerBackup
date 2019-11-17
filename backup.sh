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

workdir=$(dirname $0)	#工作路径
logDir=$workdir/logs	#日志路径
mkdir -p $logDir
cpuNum=$(nproc)		#cpu数量

compType=7z		#压缩格式 还可以用xz gz等
p7z_mx=9		#7z压缩质量
p7z_m0=lzma2		#7z压缩算法


configs=$(ls $workdir"/conf/"*.conf)
for conf in $configs
do
    cleanArgs
    source $conf

    #获取配置文件的名字，用这个名字存档
    basename=$(basename $conf)
    filename=${basename%%.*}

    for e in ${excludeList[@]}
    do
        exclude="$exclude --exclude '$e'"
    done

    timestamp=$(date +%Y-%m-%d_%T)

    #TODO 并行优化
	
    #奇怪问题：如果不使用bash单独运行rsync，在脚本中执行会忽略掉排除列表
    nice -n ${nice} bash -c "rsync $option --delete --bwlimit=${bwlimit} -e \"ssh -p ${port} -i ${sshKey} -o 'StrictHostKeyChecking no'\" ${exclude} ${user}@${host}:${srcDir} ${backupDir}/${filename}/ | tee ${logDir}/${filename}_$timestamp.log"

    if [ $compType == "7z" ]; then
        nice -n ${nice} bash -c "tar cpf - ${backupDir}/${filename}/ | 7za a -t7z -m0=${p7z_m0} -mx${p7z_mx} -mmt${cpuNum} -si ${backupDir}/${filename}_$timestamp.7z"
    elif [ $compType == "xz" ]; then
        nice -n ${nice} bash -c "tar Jcpf ${backupDir}/${filename}_$timestamp.tar.xz ${backupDir}/${filename}/"
    elif [ $compType == "gz" ]; then
        nice -n ${nice} bash -c "tar zcpf ${backupDir}/${filename}_$timestamp.tar.gz ${backupDir}/${filename}/"
    elif [ $compType == "bz" ]; then
        nice -n ${nice} bash -c "tar jcpf ${backupDir}/${filename}_$timestamp.tar.bz ${backupDir}/${filename}/"
    elif [ $compType == "bz2" ]; then
        nice -n ${nice} bash -c "tar jcpf ${backupDir}/${filename}_$timestamp.tar.bz2 ${backupDir}/${filename}/"
    elif [ $compType == "z" ]; then
        nice -n ${nice} bash -c "tar Zcpf ${backupDir}/${filename}_$timestamp.tar.z ${backupDir}/${filename}/"
    elif [ $compType == "zip" ]; then
        nice -n ${nice} bash -c "zip -qrSV ${backupDir}/${filename}_$timestamp.zip ${backupDir}/${filename}/"
    elif [ $compType == "rar" ]; then
        nice -n ${nice} bash -c "rar a -r ${backupDir}/${filename}_$timestamp.rar ${backupDir}/${filename}/"
    elif [ $compType == "tar" ]; then
        nice -n ${nice} bash -c "tar cpf ${backupDir}/${filename}_$timestamp.tar ${backupDir}/${filename}/"
    fi
done

