# ServerBackup 

## Design

### Worker

执行备份任务，同时协调各组件的运行

### Backuper

实现特定方式的备份逻辑，输出文件

### Comprossor

负责对Backuper输出的文件进行压缩

Comprossor可能对输出的文件进行重命名

### Storage

提供存储方式

## Roadmap

- [x] Worker
    - [] lock
- [x] Backuper
    - [x] rsyncmd 调用rsync命令同步数据，以tar形式打包输出
    - [] rsync 使用rsync协议进行备份
    - [] vss 对windows使用vss方式进行备份
- [x] Comprossor
    - [x] zstd
    - [] gz
    - [] lzma
- [x] Storage
    - [x] local 支持本地存储
    - [] s3 支持S3存储
- [] Plugin 支持加载go plugin