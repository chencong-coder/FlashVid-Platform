# FlashVid Platform - 数据库设计文档

## 1. 数据库设计原则

- **三范式**：消除数据冗余，保证数据一致性
- **软删除**：使用 `deleted_at` 字段标记删除，便于数据恢复和审计
- **索引优化**：为高频查询字段添加索引
- **字段命名**：使用 snake_case 命名规范
- **时间字段**：统一使用 `created_at`、`updated_at`、`deleted_at`
- **分表策略**：对于评论、点赞等高频写入表考虑分表

## 2. 核心表设计

### 2.1 用户表 (users)

```sql
CREATE TABLE `users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `username` VARCHAR(32) NOT NULL COMMENT '用户名',
  `password` VARCHAR(255) NOT NULL COMMENT '密码（bcrypt加密）',
  `nickname` VARCHAR(64) NOT NULL COMMENT '昵称',
  `avatar` VARCHAR(500) DEFAULT '' COMMENT '头像URL',
  `phone` VARCHAR(20) DEFAULT NULL COMMENT '手机号',
  `email` VARCHAR(100) DEFAULT NULL COMMENT '邮箱',
  `gender` TINYINT DEFAULT 0 COMMENT '性别：0-未知，1-男，2-女',
  `birthday` DATE DEFAULT NULL COMMENT '生日',
  `bio` VARCHAR(500) DEFAULT '' COMMENT '个人简介',
  `province` VARCHAR(50) DEFAULT '' COMMENT '省份',
  `city` VARCHAR(50) DEFAULT '' COMMENT '城市',
  `ip_address` VARCHAR(45) DEFAULT '' COMMENT '注册IP',
  `status` TINYINT DEFAULT 1 COMMENT '状态：1-正常，2-封禁',
  `follower_count` INT UNSIGNED DEFAULT 0 COMMENT '粉丝数',
  `following_count` INT UNSIGNED DEFAULT 0 COMMENT '关注数',
  `video_count` INT UNSIGNED DEFAULT 0 COMMENT '作品数',
  `like_count` INT UNSIGNED DEFAULT 0 COMMENT '获赞数',
  `last_login_at` DATETIME DEFAULT NULL COMMENT '最后登录时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  UNIQUE KEY `uk_phone` (`phone`),
  UNIQUE KEY `uk_email` (`email`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
```

### 2.2 视频表 (videos)

```sql
CREATE TABLE `videos` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '视频ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '作者ID',
  `title` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '标题',
  `description` TEXT COMMENT '视频描述',
  `cover_url` VARCHAR(500) NOT NULL COMMENT '封面URL',
  `video_url` VARCHAR(500) NOT NULL COMMENT '视频URL',
  `duration` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '视频时长（秒）',
  `width` INT UNSIGNED DEFAULT 0 COMMENT '视频宽度',
  `height` INT UNSIGNED DEFAULT 0 COMMENT '视频高度',
  `file_size` BIGINT UNSIGNED DEFAULT 0 COMMENT '文件大小（字节）',
  `music_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '背景音乐ID',
  `province` VARCHAR(50) DEFAULT '' COMMENT '省份',
  `city` VARCHAR(50) DEFAULT '' COMMENT '城市',
  `location` VARCHAR(255) DEFAULT '' COMMENT '具体位置',
  `latitude` DECIMAL(9,6) DEFAULT NULL COMMENT '纬度 (-90 to 90)',
  `longitude` DECIMAL(10,6) DEFAULT NULL COMMENT '经度 (-180 to 180)',
  `status` TINYINT DEFAULT 1 COMMENT '状态：1-审核中，2-已发布，3-未通过，4-已下架',
  `view_count` INT UNSIGNED DEFAULT 0 COMMENT '播放次数',
  `like_count` INT UNSIGNED DEFAULT 0 COMMENT '点赞数',
  `comment_count` INT UNSIGNED DEFAULT 0 COMMENT '评论数',
  `share_count` INT UNSIGNED DEFAULT 0 COMMENT '分享数',
  `favorite_count` INT UNSIGNED DEFAULT 0 COMMENT '收藏数',
  `published_at` DATETIME DEFAULT NULL COMMENT '发布时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_published_at` (`published_at`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_city` (`city`),
  KEY `idx_music_id` (`music_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='视频表';
```

### 2.3 视频话题关联表 (video_topics)

```sql
CREATE TABLE `video_topics` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `video_id` BIGINT UNSIGNED NOT NULL COMMENT '视频ID',
  `topic_id` BIGINT UNSIGNED NOT NULL COMMENT '话题ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_video_topic` (`video_id`, `topic_id`),
  KEY `idx_topic_id` (`topic_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='视频话题关联表';
```

### 2.4 话题表 (topics)

```sql
CREATE TABLE `topics` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '话题ID',
  `name` VARCHAR(100) NOT NULL COMMENT '话题名称',
  `description` VARCHAR(500) DEFAULT '' COMMENT '话题描述',
  `cover_url` VARCHAR(500) DEFAULT '' COMMENT '话题封面',
  `view_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '浏览量',
  `video_count` INT UNSIGNED DEFAULT 0 COMMENT '视频数',
  `status` TINYINT DEFAULT 1 COMMENT '状态：1-正常，2-禁用',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`),
  KEY `idx_video_count` (`video_count`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='话题表';
```

### 2.5 关注关系表 (follows)

```sql
CREATE TABLE `follows` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `follower_id` BIGINT UNSIGNED NOT NULL COMMENT '粉丝ID',
  `following_id` BIGINT UNSIGNED NOT NULL COMMENT '被关注者ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_follower_following` (`follower_id`, `following_id`),
  KEY `idx_following_id` (`following_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='关注关系表';
```

### 2.6 点赞表 (likes)

```sql
CREATE TABLE `likes` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `target_type` TINYINT NOT NULL COMMENT '点赞类型：1-视频，2-评论',
  `target_id` BIGINT UNSIGNED NOT NULL COMMENT '目标ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_target` (`user_id`, `target_type`, `target_id`),
  KEY `idx_target` (`target_type`, `target_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='点赞表';
```

### 2.7 收藏表 (favorites)

```sql
CREATE TABLE `favorites` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `video_id` BIGINT UNSIGNED NOT NULL COMMENT '视频ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_video` (`user_id`, `video_id`),
  KEY `idx_video_id` (`video_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='收藏表';
```

### 2.8 评论表 (comments)

```sql
CREATE TABLE `comments` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '评论ID',
  `video_id` BIGINT UNSIGNED NOT NULL COMMENT '视频ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '评论者ID',
  `parent_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '父评论ID，0为一级评论',
  `reply_to_user_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '回复的用户ID',
  `content` TEXT NOT NULL COMMENT '评论内容',
  `like_count` INT UNSIGNED DEFAULT 0 COMMENT '点赞数',
  `reply_count` INT UNSIGNED DEFAULT 0 COMMENT '回复数',
  `status` TINYINT DEFAULT 1 COMMENT '状态：1-正常，2-已删除，3-审核中',
  `ip_address` VARCHAR(45) DEFAULT '' COMMENT '评论IP',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_video_id` (`video_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论表';
```

### 2.9 消息表 (messages)

```sql
CREATE TABLE `messages` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '消息ID',
  `from_user_id` BIGINT UNSIGNED NOT NULL COMMENT '发送者ID',
  `to_user_id` BIGINT UNSIGNED NOT NULL COMMENT '接收者ID',
  `message_type` TINYINT NOT NULL COMMENT '消息类型：1-文本，2-图片，3-视频',
  `content` TEXT COMMENT '消息内容',
  `media_url` VARCHAR(500) DEFAULT '' COMMENT '媒体文件URL',
  `is_read` TINYINT DEFAULT 0 COMMENT '是否已读：0-未读，1-已读',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_from_user` (`from_user_id`),
  KEY `idx_to_user` (`to_user_id`),
  KEY `idx_conversation` (`from_user_id`, `to_user_id`, `created_at`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息表';
```

### 2.10 播放历史表 (view_history)

```sql
CREATE TABLE `view_history` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `video_id` BIGINT UNSIGNED NOT NULL COMMENT '视频ID',
  `duration` INT UNSIGNED DEFAULT 0 COMMENT '观看时长（秒）',
  `progress` DECIMAL(5,2) DEFAULT 0.00 COMMENT '观看进度（百分比）',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_video` (`user_id`, `video_id`),
  KEY `idx_video_id` (`video_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='播放历史表';
```

### 2.11 音乐表 (music)

```sql
CREATE TABLE `music` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '音乐ID',
  `name` VARCHAR(255) NOT NULL COMMENT '音乐名称',
  `artist` VARCHAR(255) DEFAULT '' COMMENT '艺术家',
  `album` VARCHAR(255) DEFAULT '' COMMENT '专辑',
  `cover_url` VARCHAR(500) DEFAULT '' COMMENT '封面URL',
  `music_url` VARCHAR(500) NOT NULL COMMENT '音乐URL',
  `duration` INT UNSIGNED DEFAULT 0 COMMENT '时长（秒）',
  `use_count` INT UNSIGNED DEFAULT 0 COMMENT '使用次数',
  `status` TINYINT DEFAULT 1 COMMENT '状态：1-正常，2-禁用',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_use_count` (`use_count`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='音乐表';
```

### 2.12 用户登录日志表 (login_logs)

```sql
CREATE TABLE `login_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `ip_address` VARCHAR(45) NOT NULL COMMENT '登录IP',
  `device_type` VARCHAR(50) DEFAULT '' COMMENT '设备类型',
  `device_name` VARCHAR(100) DEFAULT '' COMMENT '设备名称',
  `os` VARCHAR(50) DEFAULT '' COMMENT '操作系统',
  `browser` VARCHAR(50) DEFAULT '' COMMENT '浏览器',
  `location` VARCHAR(255) DEFAULT '' COMMENT '地理位置',
  `login_type` TINYINT DEFAULT 1 COMMENT '登录方式：1-密码，2-短信，3-第三方',
  `status` TINYINT DEFAULT 1 COMMENT '登录状态：1-成功，2-失败',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='登录日志表';
```

## 3. 索引优化策略

### 3.1 复合索引

- `videos` 表：`idx_user_status_published` (user_id, status, published_at)
- `comments` 表：`idx_video_status_created` (video_id, status, created_at)
- `follows` 表：`idx_follower_status` (follower_id, status)

### 3.2 分表策略

**评论表分表**（按视频ID取模）：
```sql
-- comments_0 ~ comments_9
CREATE TABLE `comments_0` LIKE `comments`;
CREATE TABLE `comments_1` LIKE `comments`;
-- ... 共10张表
```

**点赞表分表**（按用户ID取模）：
```sql
-- likes_0 ~ likes_9
CREATE TABLE `likes_0` LIKE `likes`;
CREATE TABLE `likes_1` LIKE `likes`;
-- ... 共10张表
```

## 4. 性能优化建议

### 4.1 读写分离
- 主库：处理写操作
- 从库：处理读操作（列表查询、详情查询）

### 4.2 缓存策略
- **Redis 缓存热点数据**：
  - 用户信息：`user:{id}`
  - 视频详情：`video:{id}`
  - 视频点赞数：`video:like:{id}`
  - 用户关注列表：`user:following:{id}`
  
### 4.3 计数器优化
- 使用 Redis 计数器，定期同步到 MySQL
- 避免频繁更新 MySQL 计数字段

### 4.4 分页优化
- 推荐流使用游标分页（cursor-based）
- 避免深度分页，超过一定页数后使用时间戳+ID筛选

## 5. 数据安全

- **敏感信息加密**：密码使用 bcrypt，手机号部分脱敏
- **软删除**：重要数据使用 `deleted_at` 标记
- **审计日志**：记录关键操作日志
- **定期备份**：每日全量备份+增量备份
