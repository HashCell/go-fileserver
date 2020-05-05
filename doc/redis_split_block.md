## 文件分块上传

### redis安装
```
//1.更新软件源
sudo apt-get update

//2.　安装redis服务
sudo apt-get install redis-server

//3. 启动redis服务
redis-server

//4. 查看redis-server状态
ps -ef | grep redis
netstat -an | grep redis

//5. 使用redis-cli
redis-cli

//6. 设置密码
➜  ~ redis-cli
127.0.0.1:6379> config get requirepass
1) "requirepass"
2) ""
127.0.0.1:6379> config set requirepass 123456
OK
127.0.0.1:6379> config get requirepass
(error) NOAUTH Authentication required.
127.0.0.1:6379> auth 123456
OK
127.0.0.1:6379> config get requirepass
1) "requirepass"
2) "123456"
127.0.0.1:6379> exit

```

###分块上传的原理

１）分块上传：把文件切成多块，每一块独立传输，上传后，云端完成合并
２）断点续传：传输暂停或者异常中断后，可基于原来的进度继续重传

#### 流程

```
client --> request for split-block upload （客户端发送文件信息，由服务端初始化分块信息）
--> server initialize multi-part block according to file size　
--> return initialization information to client　（将分块初始化信息返回给客户端）
--> client upload multi-part by parallel　（客户端开始分块传输，可以并行，服务端保存每一个文件块，redis缓存记录当前收到哪些文件块）
--> client notify server when multi-part all uploaded　（客户端完成分块传输，通知服务端合并全部分块）
```

#### 服务架构
```
 　　　　　　用户文件表          唯一文件表
                ｜               ｜
                ｜               ｜
hash计算　－－－分块上传－－－－－－－｜
redis缓存　－－｜　｜　   ｜　－－－－本地存储
                 ｜
                 用户
```

