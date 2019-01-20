# ServerBackup
linux shell &amp; rsync for server backup

* 带权限备份
* tar.7z格式
* 使用rsync传输
* 支持排除列表
* 支持限速
* rsync备份日志

## HOW TO
1. 源服务器与目标服务器安装rsync `apt install rsync`
2. 目标服务器安装p7zip `apt install p7zip`
3. 拉取该项目到备份目标服务器 `git clone https://github.com/Lensual/ServerBackup`
4. 上传源服务器私钥到sshkey
5. 配置conf/example.conf
6. 运行backup.sh测试备份
7. 配置cron

## CRON
每天3:00备份
```
0 3 * * * /root/script/backup/backup.sh
```

