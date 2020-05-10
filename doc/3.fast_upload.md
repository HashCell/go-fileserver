### 实现文件秒传

#### 数据库表

```
CREATE TABLE `tbl_user_file` (
  `id` int(11) NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `user_name` varchar(64) NOT NULL,
  `file_sha1` varchar(64) NOT NULL DEFAULT '' COMMENT '文件hash',
  `file_size` bigint(20) DEFAULT '0' COMMENT '文件大小',
  `file_name` varchar(256) NOT NULL DEFAULT '' COMMENT '文件名',
  `upload_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '上传时间',
  `last_update` datetime DEFAULT CURRENT_TIMESTAMP
          ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '文件状态(0正常1已删除2禁用)',
  UNIQUE KEY `idx_user_file` (`user_name`, `file_sha1`),
  KEY `idx_status` (`status`),
  KEY `idx_user_id` (`user_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

```

#### 秒传原理

１）上传文件时，对文件进行计算其hash值,先向服务端发送待传文件的hash值，服务端查找数据库是否存在相同的文件，
如果存在，就更新用户文件表，将一条新的文件记录插入到用户的文件表，然后通知客户端文件上传已完成，至此，文件不需要发送到服务端．

user_name和file_sha1组成了一个唯一键,file_sha1关联文件表中的唯一一条记录
UNIQUE KEY `idx_user_file` (`user_name`, `file_sha1`) 客户端无法重复上传同一文件，`去掉这个唯一键`

２）当然，如果服务端不存在hash值对应的文件，那么客户端就需要上传整个文件到服务端，即调用TryFastUploadHandler接口后，客户端得到秒传失败
的通知，然后使用普通接口再传一次．