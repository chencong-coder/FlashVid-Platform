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
      "pageSize": 20,
      "total": 100,
      "totalPages": 5
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
| 60000-69999 | 私信相关错误 |

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
    "userId": 123456,
    "username": "testuser"
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
  "account": "string",     // 用户名或手机号
  "password": "string"     // 密码
}
```

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "userId": 123456,
    "username": "testuser",
    "nickname": "测试用户",
    "avatar": "https://...",
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
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
  "refreshToken": "string"
}
```

#### 2.1.4 获取用户信息
```
GET /api/v1/user/:id
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
    "followerCount": 1000,
    "followingCount": 500,
    "videoCount": 50,
    "likeCount": 10000,
    "isFollowing": false
  }
}
```

#### 2.1.5 更新用户信息
```
PUT /api/v1/user/:id
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
GET /api/v1/user/:id/videos
```

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| page | int | 否 | 页码，默认1 |
| pageSize | int | 否 | 每页数量，默认20 |

#### 2.1.7 获取用户喜欢列表
```
GET /api/v1/user/profile/likes
```

#### 2.1.8 获取用户收藏列表
```
GET /api/v1/user/profile/favorites
```

### 2.2 关注模块

#### 2.2.1 关注用户
```
POST /api/v1/user/:id/follow
```

**响应**：
```json
{
  "code": 0,
  "message": "关注成功",
  "data": {
    "isFollowing": true
  }
}
```

#### 2.2.2 取消关注
```
DELETE /api/v1/user/:id/follow
```

#### 2.2.3 获取关注列表
```
GET /api/v1/user/:id/followings
```

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| page | int | 否 | 页码 |
| pageSize | int | 否 | 每页数量 |

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
        "isFollowing": true,
        "followerCount": 5000
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 100
    }
  }
}
```

#### 2.2.4 获取粉丝列表
```
GET /api/v1/user/:id/followers
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
        "coverUrl": "https://...",
        "videoUrl": "https://...",
        "duration": 30,
        "width": 1080,
        "height": 1920,
        "musicId": 456,
        "city": "深圳",
        "topics": ["话题1", "话题2"],
        "author": {
          "id": 123,
          "username": "author",
          "nickname": "作者",
          "avatar": "https://..."
        },
        "stats": {
          "viewCount": 10000,
          "likeCount": 1000,
          "commentCount": 100,
          "shareCount": 50,
          "favoriteCount": 200
        },
        "userInteraction": {
          "isLiked": false,
          "isFavorited": false,
          "isFollowing": false
        },
        "createdAt": "2024-01-01T12:00:00Z"
      }
    ],
    "nextCursorToken": "...",
    "hasMore": true
  }
}
```

#### 2.3.2 获取关注视频流
```
GET /api/v1/feed/follow
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
    "coverUrl": "https://...",
    "videoUrl": "https://...",
    "duration": 30,
    "width": 1080,
    "height": 1920,
    "musicId": 456,
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
      "viewCount": 10000,
      "likeCount": 1000,
      "commentCount": 100,
      "shareCount": 50,
      "favoriteCount": 200
    },
    "userInteraction": {
      "isLiked": false,
      "isFavorited": false,
      "isFollowing": false
    },
    "publishedAt": "2024-01-01T12:00:00Z",
    "createdAt": "2024-01-01T12:00:00Z"
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
  "coverUrl": "string",       // 封面URL（已上传）
  "videoUrl": "string",       // 视频URL（已上传）
  "duration": 30,
  "width": 1080,
  "height": 1920,
  "musicId": 456,
  "city": "深圳",
  "location": "南山区",
  "latitude": 22.5431,
  "longitude": 114.0579,
  "topicNames": ["话题1", "话题2"]   // 话题名称，后端自动查询/创建话题
}
```

**响应**：
```json
{
  "code": 0,
  "message": "发布成功",
  "data": {
    "videoId": 12345,
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
| pageSize | int | 否 | 每页数量 |
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
    "isLiked": true,
    "likeCount": 1001
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
    "shareUrl": "https://...",
    "shareCount": 51
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
| pageSize | int | 否 | 每页数量 |
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
        "likeCount": 10,
        "replyCount": 2,
        "isLiked": false,
        "isAuthor": false,
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
            "replyToUser": {
              "id": 123,
              "nickname": "评论者"
            },
            "likeCount": 5,
            "isLiked": false,
            "createdAt": "2024-01-01T12:05:00Z"
          }
        ],
        "createdAt": "2024-01-01T12:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 20,
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
  "parentId": 0,             // 父评论ID，0为一级评论
  "replyToUserId": 0       // 回复的用户ID
}
```

**响应**：
```json
{
  "code": 0,
  "message": "评论成功",
  "data": {
    "commentId": 789,
    "content": "评论内容",
    "createdAt": "2024-01-01T12:00:00Z"
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
| pageSize | int | 否 | 每页数量 |
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
    "coverUrl": "https://...",
    "viewCount": 1000000,
    "videoCount": 5000,
    "createdAt": "2024-01-01T12:00:00Z"
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
| pageSize | int | 否 | 每页数量 |

### 2.7 文件上传模块

#### 2.7.1 上传文件
```
POST /api/v1/upload
```

> 需要登录（Authorization: Bearer \<token\>）

**请求参数**（multipart/form-data）：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| file | file | 是 | 文件 |
| file_type | string | 是 | 文件类型：image、video、audio |

**文件大小限制**：
- image：10 MB
- audio：50 MB
- video：500 MB

**响应**：
```json
{
  "code": 0,
  "message": "上传成功",
  "data": {
    "file_url": "http://localhost:8089/static/video/1234567890.mp4",
    "file_size": 5242880,
    "file_type": "video",
    "duration": 0
  }
}
```

> `duration` 当前本地存储固定返回 0，不做媒体分析。

### 2.8 音乐模块

#### 2.8.1 获取音乐列表
```
GET /api/v1/music
```

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| page | int | 否 | 页码，默认 1 |
| pageSize | int | 否 | 每页数量，默认 20，最大 100 |
| sort | string | 否 | 排序：hot(最热，默认)、latest(最新) |

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "曲名",
        "artist": "艺术家",
        "album": "专辑名",
        "coverUrl": "https://...",
        "musicUrl": "https://...",
        "duration": 180,
        "useCount": 5000,
        "createdAt": "2024-01-01T12:00:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "pageSize": 20
  }
}
```

#### 2.8.2 搜索音乐
```
GET /api/v1/music/search
```

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| keyword | string | 是 | 搜索关键词（匹配曲名或艺术家） |
| page | int | 否 | 页码，默认 1 |
| pageSize | int | 否 | 每页数量，默认 20，最大 100 |

**响应结构同 2.8.1**。

### 2.9 私信模块

采用「会话 + 消息」两层模型：

- **会话（Conversation）**：两个用户之间的私信关系，聚合出对方用户、最后一条消息、未读数。会话列表用 offset 分页（数量有限，按最后消息时间排序）。
- **消息（Message）**：会话内的单条私信。对话内消息用游标分页（数量大，向上翻加载历史）。

> 全部接口需要登录（Authorization: Bearer \<token\>）。

#### 2.9.1 获取会话列表
```
GET /api/v1/conversations
```

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| page | int | 否 | 页码，默认 1 |
| pageSize | int | 否 | 每页数量，默认 20，最大 100 |

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "targetUser": {
          "id": 456,
          "username": "friend",
          "nickname": "好友",
          "avatar": "https://..."
        },
        "lastMessage": {
          "id": 1001,
          "messageType": 1,
          "content": "在吗？",
          "mediaUrl": "",
          "createdAt": "2024-01-01 12:00:00"
        },
        "unreadCount": 3,
        "updatedAt": "2024-01-01 12:00:00"
      }
    ],
    "total": 12,
    "page": 1,
    "pageSize": 20
  }
}
```

#### 2.9.2 获取对话消息
```
GET /api/v1/conversations/:userId/messages
```

`:userId` 为对方用户 ID。消息按创建时间倒序返回（最新在前），游标分页向上翻历史。

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| cursor | string | 否 | 游标（上一页最后一条消息的创建时间，格式 `2006-01-02 15:04:05`），首次不传 |
| count | int | 否 | 每页数量，默认 20，最大 50 |

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "messages": [
      {
        "id": 1001,
        "fromUserId": 456,
        "toUserId": 123,
        "messageType": 1,
        "content": "在吗？",
        "mediaUrl": "",
        "isRead": true,
        "createdAt": "2024-01-01 12:00:00"
      }
    ],
    "nextCursorToken": "2024-01-01 11:00:00",
    "hasMore": true
  }
}
```

> `messageType`：1-文本，2-图片，3-视频。图片/视频类型时 `content` 可为空，`mediaUrl` 为已上传的资源地址。

#### 2.9.3 发送私信
```
POST /api/v1/messages
```

**请求参数**：
```json
{
  "toUserId": 456,
  "messageType": 1,
  "content": "在吗？",
  "mediaUrl": ""
}
```

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| toUserId | int64 | 是 | 接收方用户 ID |
| messageType | int | 是 | 消息类型：1-文本 2-图片 3-视频 |
| content | string | 条件 | 文本内容，messageType=1 时必填 |
| mediaUrl | string | 条件 | 媒体地址，messageType=2/3 时必填（先经上传接口获取） |

**响应**：
```json
{
  "code": 0,
  "message": "发送成功",
  "data": {
    "id": 1002,
    "fromUserId": 123,
    "toUserId": 456,
    "messageType": 1,
    "content": "在吗？",
    "mediaUrl": "",
    "isRead": false,
    "createdAt": "2024-01-01 12:05:00"
  }
}
```

#### 2.9.4 标记会话已读
```
PUT /api/v1/conversations/:userId/read
```

将与 `:userId` 的会话中所有未读消息标记为已读。

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "readCount": 3
  }
}
```

#### 2.9.5 删除消息
```
DELETE /api/v1/messages/:id
```

删除自己发送或接收的某条消息（仅对自己隐藏）。

#### 2.9.6 获取未读总数
```
GET /api/v1/messages/unread-count
```

用于消息入口红点提示。

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "unreadCount": 8
  }
}
```

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
