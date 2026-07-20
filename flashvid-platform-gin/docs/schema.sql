-- FlashVid Platform 数据库表结构
-- MySQL 8.0+
-- 字符集: utf8mb4
-- 排序规则: utf8mb4_unicode_ci

-- 创建数据库
CREATE DATABASE IF NOT EXISTS `flashvid` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `flashvid`;

-- ==================== 用户相关表 ====================

-- 用户表
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

-- 关注关系表
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

-- 用户登录日志表
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

-- ==================== 视频相关表 ====================

-- 视频表
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

-- 话题表
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

-- 视频话题关联表
CREATE TABLE `video_topics` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `video_id` BIGINT UNSIGNED NOT NULL COMMENT '视频ID',
  `topic_id` BIGINT UNSIGNED NOT NULL COMMENT '话题ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_video_topic` (`video_id`, `topic_id`),
  KEY `idx_topic_id` (`topic_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='视频话题关联表';

-- 音乐表
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

-- 播放历史表
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

-- ==================== 互动相关表 ====================

-- 点赞表
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

-- 收藏表
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

-- 评论表
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

-- ==================== 消息相关表 ====================

-- 消息表
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
