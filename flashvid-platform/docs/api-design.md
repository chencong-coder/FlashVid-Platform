# FlashVid Platform - RESTful API 设计文档

## 1. API 设计原则

### 1.1 RESTful 规范
- 使用 HTTP 方法表示操作：GET（查询）、POST（创建）、PUT（更新）、DELETE（删除）
- 使用名词复数表示资源：`/users`、`/videos`、`/comments`
- 使用嵌套路由表示资源关系：`/videos/:id/comments`
- 使用 HTTP 状态码表示结果

### 1.2 版本控制
- URL 版本控制：`/api/v1/...`
- 便于 API 迭代升级

### 1.3 统一响应格式

**成功响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    // 业务数据
  }
}
```

**错误响应**：
```json
{
  "code": 10001,
  "message": "用户不存在",
  "data": null
}
```

**分页响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

### 1.4 错误码设计

| 错误码范围 | 说明 |
|---------|------|
| 0 | 成功 |
| 10000-19999 | 用户相关错误 |
| 20000-29999 | 视频相关错误 |
| 30000-39999 | 评论相关错误 |
| 40000-49999 | 系统错误 |
| 50000-59999 | 第三方服务错误 |

## 2. API 接口设计

### 2.1 用户模块

#### 2.1.1 用户注册
```
POST /api/v1/auth/register
```

**请求参数**：
```json
{
  "username": "string",    // 用户名，4-32字符
  "password": "string",    // 密码，6-20字符
  "phone": "string",       // 手机号
  "code": "string"         // 短信验证码
}
```

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 123456,
    "username": "testuser",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

#### 2.1.2 用户登录
```
POST /api/v1/auth/login
```

**请求参数**：
```json
{
  "username": "string",    // 用户名或手机号
  "password": "string"     // 密码
}
```

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 123456,
    "username": "testuser",
    "nickname": "测试用户",
    "avatar": "https://...",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

#### 2.1.3 刷新 Token
```
POST /api/v1/auth/refresh
```

**请求参数**：
```json
{
  "refresh_token": "string"
}
```

#### 2.1.4 获取用户信息
```
GET /api/v1/users/:id
```

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 123456,
    "username": "testuser",
    "nickname": "测试用户",
    "avatar": "https://...",
    "bio": "个人简介",
    "gender": 1,
    "city": "深圳",
    "follower_count": 1000,
    "following_count": 500,
    "video_count": 50,
    "like_count": 10000,
    "is_following": false
  }
}
```

#### 2.1.5 更新用户信息
```
PUT /api/v1/users/:id
```

**请求参数**：
```json
{
  "nickname": "string",
  "avatar": "string",
  "bio": "string",
  "gender": 1,
  "birthday": "2000-01-01",
  "city": "string"
}
```

#### 2.1.6 获取用户作品列表
```
GET /api/v1/users/:id/videos
```

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认20 |

#### 2.1.7 获取用户喜欢列表
```
GET /api/v1/users/:id/likes
```

#### 2.1.8 获取用户收藏列表
```
GET /api/v1/users/:id/favorites
```

### 2.2 关注模块

#### 2.2.1 关注用户
```
POST /api/v1/users/:id/follow
```

**响应**：
```json
{
  "code": 0,
  "message": "关注成功",
  "data": {
    "is_following": true
  }
}
```

#### 2.2.2 取消关注
```
DELETE /api/v1/users/:id/follow
```

#### 2.2.3 获取关注列表
```
GET /api/v1/users/:id/following
```

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 789,
        "username": "user789",
        "nickname": "用户789",
        "avatar": "https://...",
        "bio": "简介",
        "is_following": true,
        "follower_count": 5000
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 100
    }
  }
}
```

#### 2.2.4 获取粉丝列表
```
GET /api/v1/users/:id/followers
```

### 2.3 视频模块

#### 2.3.1 获取推荐视频流
```
GET /api/v1/feed/recommend
```

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| cursor | string | 否 | 游标，用于分页 |
| count | int | 否 | 数量，默认10 |

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 12345,
        "title": "视频标题",
        "description": "视频描述",
        "cover_url": "https://...",
        "video_url": "https://...",
        "duration": 30,
        "width": 1080,
        "height": 1920,
        "music_name": "背景音乐",
        "city": "深圳",
        "topics": ["话题1", "话题2"],
        "author": {
          "id": 123,
          "username": "author",
          "nickname": "作者",
          "avatar": "https://..."
        },
        "stats": {
          "view_count": 10000,
          "like_count": 1000,
          "comment_count": 100,
          "share_count": 50,
          "favorite_count": 200
        },
        "user_interaction": {
          "is_liked": false,
          "is_favorited": false,
          "is_following": false
        },
        "created_at": "2024-01-01T12:00:00Z"
      }
    ],
    "cursor": "next_cursor_token",
    "has_more": true
  }
}
```

#### 2.3.2 获取关注视频流
```
GET /api/v1/feed/following
```

#### 2.3.3 获取附近视频流
```
GET /api/v1/feed/nearby
```

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| latitude | float | 是 | 纬度 |
| longitude | float | 是 | 经度 |
| radius | int | 否 | 半径（km），默认10 |
| cursor | string | 否 | 游标 |
| count | int | 否 | 数量 |

#### 2.3.4 获取视频详情
```
GET /api/v1/videos/:id
```

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 12345,
    "title": "视频标题",
    "description": "视频描述",
    "cover_url": "https://...",
    "video_url": "https://...",
    "duration": 30,
    "width": 1080,
    "height": 1920,
    "music_id": 456,
    "music_name": "背景音乐",
    "city": "深圳",
    "location": "南山区",
    "topics": [
      {
        "id": 1,
        "name": "话题1"
      }
    ],
    "author": {
      "id": 123,
      "username": "author",
      "nickname": "作者",
      "avatar": "https://...",
      "bio": "简介"
    },
    "stats": {
      "view_count": 10000,
      "like_count": 1000,
      "comment_count": 100,
      "share_count": 50,
      "favorite_count": 200
    },
    "user_interaction": {
      "is_liked": false,
      "is_favorited": false,
      "is_following": false
    },
    "published_at": "2024-01-01T12:00:00Z",
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

#### 2.3.5 发布视频
```
POST /api/v1/videos
```

**请求参数**：
```json
{
  "title": "string",
  "description": "string",
  "cover_url": "string",       // 封面URL（已上传）
  "video_url": "string",       // 视频URL（已上传）
  "duration": 30,
  "width": 1080,
  "height": 1920,
  "file_size": 5242880,
  "music_id": 456,
  "music_name": "背景音乐",
  "city": "深圳",
  "location": "南山区",
  "latitude": 22.5431,
  "longitude": 114.0579,
  "topic_ids": [1, 2, 3]
}
```

**响应**：
```json
{
  "code": 0,
  "message": "发布成功",
  "data": {
    "video_id": 12345,
    "status": 1  // 1-审核中
  }
}
```

#### 2.3.6 删除视频
```
DELETE /api/v1/videos/:id
```

#### 2.3.7 搜索视频
```
GET /api/v1/videos/search
```

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| keyword | string | 是 | 搜索关键词 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |
| sort | string | 否 | 排序：latest(最新)、popular(最热) |

### 2.4 互动模块

#### 2.4.1 点赞视频
```
POST /api/v1/videos/:id/like
```

**响应**：
```json
{
  "code": 0,
  "message": "点赞成功",
  "data": {
    "is_liked": true,
    "like_count": 1001
  }
}
```

#### 2.4.2 取消点赞
```
DELETE /api/v1/videos/:id/like
```

#### 2.4.3 收藏视频
```
POST /api/v1/videos/:id/favorite
```

#### 2.4.4 取消收藏
```
DELETE /api/v1/videos/:id/favorite
```

#### 2.4.5 分享视频
```
POST /api/v1/videos/:id/share
```

**请求参数**：
```json
{
  "platform": "string"  // wechat, qq, weibo, link
}
```

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "share_url": "https://...",
    "share_count": 51
  }
}
```

### 2.5 评论模块

#### 2.5.1 获取视频评论列表
```
GET /api/v1/videos/:id/comments
```

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |
| sort | string | 否 | 排序：latest(最新)、hot(最热) |

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 789,
        "content": "评论内容",
        "user": {
          "id": 123,
          "username": "commenter",
          "nickname": "评论者",
          "avatar": "https://..."
        },
        "like_count": 10,
        "reply_count": 2,
        "is_liked": false,
        "is_author": false,
        "replies": [
          {
            "id": 790,
            "content": "回复内容",
            "user": {
              "id": 456,
              "username": "replier",
              "nickname": "回复者",
              "avatar": "https://..."
            },
            "reply_to_user": {
              "id": 123,
              "nickname": "评论者"
            },
            "like_count": 5,
            "is_liked": false,
            "created_at": "2024-01-01T12:05:00Z"
          }
        ],
        "created_at": "2024-01-01T12:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 100
    }
  }
}
```

#### 2.5.2 发表评论
```
POST /api/v1/videos/:id/comments
```

**请求参数**：
```json
{
  "content": "string",         // 评论内容
  "parent_id": 0,             // 父评论ID，0为一级评论
  "reply_to_user_id": 0       // 回复的用户ID
}
```

**响应**：
```json
{
  "code": 0,
  "message": "评论成功",
  "data": {
    "comment_id": 789,
    "content": "评论内容",
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

#### 2.5.3 删除评论
```
DELETE /api/v1/comments/:id
```

#### 2.5.4 点赞评论
```
POST /api/v1/comments/:id/like
```

#### 2.5.5 取消点赞评论
```
DELETE /api/v1/comments/:id/like
```

### 2.6 话题模块

#### 2.6.1 获取话题列表
```
GET /api/v1/topics
```

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |
| sort | string | 否 | 排序：hot(最热)、latest(最新) |

#### 2.6.2 获取话题详情
```
GET /api/v1/topics/:id
```

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "话题名称",
    "description": "话题描述",
    "cover_url": "https://...",
    "view_count": 1000000,
    "video_count": 5000,
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

#### 2.6.3 获取话题视频列表
```
GET /api/v1/topics/:id/videos
```

#### 2.6.4 搜索话题
```
GET /api/v1/topics/search
```

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| keyword | string | 是 | 搜索关键词 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

### 2.7 消息模块

#### 2.7.1 获取消息列表
```
GET /api/v1/messages
```

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "from_user": {
          "id": 123,
          "username": "sender",
          "nickname": "发送者",
          "avatar": "https://..."
        },
        "message_type": 1,
        "content": "消息内容",
        "media_url": "",
        "is_read": false,
        "created_at": "2024-01-01T12:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 50
    },
    "unread_count": 10
  }
}
```

#### 2.7.2 获取对话消息
```
GET /api/v1/messages/conversations/:user_id
```

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| cursor | string | 否 | 游标 |
| count | int | 否 | 数量 |

#### 2.7.3 发送消息
```
POST /api/v1/messages
```

**请求参数**：
```json
{
  "to_user_id": 456,
  "message_type": 1,
  "content": "消息内容",
  "media_url": ""
}
```

#### 2.7.4 标记消息已读
```
PUT /api/v1/messages/:id/read
```

#### 2.7.5 删除消息
```
DELETE /api/v1/messages/:id
```

### 2.8 文件上传模块

#### 2.8.1 获取上传凭证
```
GET /api/v1/upload/token
```

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| file_type | string | 是 | 文件类型：image、video、audio |
| file_name | string | 是 | 文件名 |

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "upload_token",
    "upload_url": "https://upload.example.com",
    "expires_at": "2024-01-01T13:00:00Z"
  }
}
```

#### 2.8.2 上传文件
```
POST /api/v1/upload
```

**请求参数**（multipart/form-data）：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| file | file | 是 | 文件 |
| file_type | string | 是 | 文件类型 |

**响应**：
```json
{
  "code": 0,
  "message": "上传成功",
  "data": {
    "file_url": "https://cdn.example.com/xxx.mp4",
    "file_size": 5242880,
    "file_type": "video",
    "duration": 30,
    "width": 1080,
    "height": 1920
  }
}
```

### 2.9 音乐模块

#### 2.9.1 获取音乐列表
```
GET /api/v1/music
```

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |
| sort | string | 否 | 排序：hot(最热)、latest(最新) |

#### 2.9.2 搜索音乐
```
GET /api/v1/music/search
```

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| keyword | string | 是 | 搜索关键词 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

## 3. 认证与鉴权

### 3.1 JWT Token
- **Access Token**：有效期 2 小时
- **Refresh Token**：有效期 30 天
- Token 放在 HTTP Header 中：`Authorization: Bearer <token>`

### 3.2 权限控制
- 公开接口：无需登录
- 用户接口：需要登录
- 管理接口：需要管理员权限

## 4. 限流策略

### 4.1 接口限流
- 登录接口：5次/分钟/IP
- 发布视频：10次/小时/用户
- 评论接口：30次/分钟/用户
- 点赞接口：100次/分钟/用户

### 4.2 限流响应
```json
{
  "code": 40003,
  "message": "请求过于频繁，请稍后再试",
  "data": {
    "retry_after": 60  // 秒
  }
}
```

HTTP 状态码：`429 Too Many Requests`

## 5. HTTP 状态码

| 状态码 | 说明 |
|-------|------|
| 200 | 请求成功 |
| 201 | 创建成功 |
| 204 | 删除成功（无返回内容） |
| 400 | 请求参数错误 |
| 401 | 未登录或 Token 失效 |
| 403 | 无权限访问 |
| 404 | 资源不存在 |
| 409 | 资源冲突（如重复关注） |
| 422 | 请求参数验证失败 |
| 429 | 请求过于频繁 |
| 500 | 服务器内部错误 |
| 503 | 服务暂时不可用 |

## 6. 安全建议

### 6.1 HTTPS
- 生产环境强制使用 HTTPS
- 敏感信息加密传输

### 6.2 CORS
- 配置允许的域名白名单
- 限制跨域请求方法

### 6.3 防刷
- 接口限流
- 验证码机制
- 设备指纹识别

### 6.4 敏感操作
- 修改密码：需要验证旧密码或短信验证码
- 删除视频：需要二次确认
- 提现操作：需要支付密码

## 7. 监控与日志

### 7.1 接口监控
- 请求量统计
- 响应时间监控
- 错误率监控

### 7.2 日志记录
- 请求日志：记录所有请求
- 错误日志：记录所有错误
- 业务日志：记录关键业务操作
