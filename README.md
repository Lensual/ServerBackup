# ServerBackup

linux shell &amp; rsync for server backup

- 带权限备份
- tar.7z 格式
- 使用 rsync 传输
- 支持排除列表
- 支持限速
- rsync 备份日志
- 支持优先级

## HOW TO

### 源服务器与目标服务器安装 rsync

RH/CentOS

`# yum install rsync`

Debian/Ubuntu

`# apt install rsync`

### （可选）目标服务器安装 p7zip、pigz 用于备份压缩

RH/CentOS

`# yum install p7zip pigz`

Debian/Ubuntu

`# apt install p7zip-full pigz`

### 拉取该项目到备份目标服务器

`git clone https://github.com/Lensual/ServerBackup`

### 上传源服务器私钥到 sshkey

### 配置 conf/example.conf

可选用 windows.conf 和 linux.conf 模板

### 为脚本添加可执行权限

`# chmod +x ./*.sh`

运行 main.sh 测试备份

### 配置 crontab

`# crontab -e`

## 压缩参数说明

`compType`：压缩格式，支持 `7z` `xz` `gz` `pigz` `bz` `bz2` `z` `zip` `rar` `tar`。需要安装相应的软件包。

p7z_mx：压缩质量，7z 格式下有效 `0-9`
p7z_m0: 压缩算法，7z 格式下有效 `lzma` `lzma2` `ppmd` `bzip2` `deflate` `copy`

pigz_level：压缩质量， pigz 格式下有效 `0-9`

## CRON 示例

每天 3:00 备份

```
0 3 * * * /root/script/backup/main.sh
```

## TODO

- 完善压缩参数
- 增量备份
- 并行优化
