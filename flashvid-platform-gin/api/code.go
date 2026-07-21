package api

// ResCode 定义返回码类型
type ResCode int64

// 错误码设计：
// 0: 成功
// 10000-19999: 用户相关错误
// 20000-29999: 视频相关错误
// 30000-39999: 评论相关错误
// 40000-49999: 系统错误
// 50000-59999: 第三方服务错误
const (
	CodeSuccess ResCode = 0

	// ==================== 用户相关错误 10000-19999 ====================
	CodeUserExist          ResCode = 10001 // 用户名已存在
	CodeUserNotExist       ResCode = 10002 // 用户不存在
	CodeInvalidPassword    ResCode = 10003 // 用户名或密码错误
	CodePhoneExist         ResCode = 10004 // 手机号已被注册
	CodeInvalidPhone       ResCode = 10005 // 手机号格式错误
	CodeInvalidCode        ResCode = 10006 // 验证码错误
	CodeUserBanned         ResCode = 10007 // 账号已被封禁
	CodeAlreadyFollowed    ResCode = 10008 // 已关注该用户
	CodeNotFollowed        ResCode = 10009 // 未关注该用户
	CodeCannotFollowSelf   ResCode = 10010 // 不能关注自己
	CodeInvalidUsername    ResCode = 10011 // 用户名格式错误
	CodeInvalidNickname    ResCode = 10012 // 昵称格式错误
	CodeTokenExpired       ResCode = 10013 // Token已过期
	CodeInvalidToken       ResCode = 10014 // 无效的Token
	CodeNeedLogin          ResCode = 10015 // 需要登录
	CodePermissionDenied   ResCode = 10016 // 无权限操作
	CodeInvalidBirthday    ResCode = 10017 // 生日格式错误
	CodeInvalidUserID        ResCode = 100018 // 用户ID无效

	// ==================== 视频相关错误 20000-29999 ====================
	CodeVideoNotExist      ResCode = 20001 // 视频不存在
	CodeVideoDeleted       ResCode = 20002 // 视频已删除
	CodeVideoUnderReview   ResCode = 20003 // 视频审核中
	CodeVideoNotApproved   ResCode = 20004 // 视频未通过审核
	CodeVideoTooLarge      ResCode = 20005 // 视频文件过大
	CodeVideoFormatInvalid ResCode = 20006 // 视频格式不支持
	CodeVideoDurationLimit ResCode = 20007 // 视频时长超限
	CodeAlreadyLiked       ResCode = 20008 // 已点赞
	CodeNotLiked           ResCode = 20009 // 未点赞
	CodeAlreadyFavorited   ResCode = 20010 // 已收藏
	CodeNotFavorited       ResCode = 20011 // 未收藏
	CodeTopicNotExist      ResCode = 20012 // 话题不存在
	CodeMusicNotExist      ResCode = 20013 // 音乐不存在

	// ==================== 评论相关错误 30000-39999 ====================
	CodeCommentNotExist    ResCode = 30001 // 评论不存在
	CodeCommentDeleted     ResCode = 30002 // 评论已删除
	CodeCommentTooLong     ResCode = 30003 // 评论内容过长
	CodeCommentTooShort    ResCode = 30004 // 评论内容过短
	CodeCommentSensitive   ResCode = 30005 // 评论包含敏感词
	CodeCommentTooFrequent ResCode = 30006 // 评论过于频繁
	CodeCannotReplyDeleted ResCode = 30007 // 无法回复已删除的评论

	// ==================== 消息相关错误 35000-35999 ====================
	CodeMessageNotExist    ResCode = 35001 // 消息不存在
	CodeCannotSendToSelf   ResCode = 35002 // 不能给自己发消息
	CodeMessageTooLong     ResCode = 35003 // 消息内容过长
	CodeMessageTooFrequent ResCode = 35004 // 消息发送过于频繁

	// ==================== 系统错误 40000-49999 ====================
	CodeInvalidParam       ResCode = 40001 // 请求参数错误
	CodeBindError          ResCode = 40002 // 参数绑定错误
	CodeValidationError    ResCode = 40003 // 参数验证失败
	CodeTooManyRequests    ResCode = 40004 // 请求过于频繁
	CodeNotFound           ResCode = 40005 // 资源不存在
	CodeMethodNotAllowed   ResCode = 40006 // 请求方法不允许
	CodeDatabaseError      ResCode = 40007 // 数据库错误
	CodeCacheError         ResCode = 40008 // 缓存错误
	CodeInternalError      ResCode = 40009 // 服务器内部错误
	CodeFileReadError      ResCode = 40010 // 文件读取错误
	CodeFileWriteError     ResCode = 40011 // 文件写入错误

	// ==================== 第三方服务错误 50000-59999 ====================
	CodeServerBusy         ResCode = 50001 // 服务繁忙
	CodeUploadFailed       ResCode = 50002 // 文件上传失败
	CodeSMSFailed          ResCode = 50003 // 短信发送失败
	CodeOSSError           ResCode = 50004 // 对象存储错误
	CodeThirdPartyError    ResCode = 50005 // 第三方服务错误
)

var codeMsgMap = map[ResCode]string{
	CodeSuccess: "success",

	// ==================== 用户相关错误 ====================
	CodeUserExist:          "用户名已存在",
	CodeUserNotExist:       "用户不存在",
	CodeInvalidPassword:    "用户名或密码错误",
	CodePhoneExist:         "手机号已被注册",
	CodeInvalidPhone:       "手机号格式错误",
	CodeInvalidCode:        "验证码错误",
	CodeUserBanned:         "账号已被封禁",
	CodeAlreadyFollowed:    "已关注该用户",
	CodeNotFollowed:        "未关注该用户",
	CodeCannotFollowSelf:   "不能关注自己",
	CodeInvalidUsername:    "用户名格式错误",
	CodeInvalidNickname:    "昵称格式错误",
	CodeTokenExpired:       "Token已过期",
	CodeInvalidToken:       "无效的Token",
	CodeNeedLogin:          "需要登录",
	CodePermissionDenied:   "无权限操作",
	CodeInvalidBirthday:    "生日格式错误",
	CodeInvalidUserID:        "用户ID无效",

	// ==================== 视频相关错误 ====================
	CodeVideoNotExist:      "视频不存在",
	CodeVideoDeleted:       "视频已删除",
	CodeVideoUnderReview:   "视频审核中",
	CodeVideoNotApproved:   "视频未通过审核",
	CodeVideoTooLarge:      "视频文件过大",
	CodeVideoFormatInvalid: "视频格式不支持",
	CodeVideoDurationLimit: "视频时长超限",
	CodeAlreadyLiked:       "已点赞",
	CodeNotLiked:           "未点赞",
	CodeAlreadyFavorited:   "已收藏",
	CodeNotFavorited:       "未收藏",
	CodeTopicNotExist:      "话题不存在",
	CodeMusicNotExist:      "音乐不存在",

	// ==================== 评论相关错误 ====================
	CodeCommentNotExist:    "评论不存在",
	CodeCommentDeleted:     "评论已删除",
	CodeCommentTooLong:     "评论内容过长",
	CodeCommentTooShort:    "评论内容过短",
	CodeCommentSensitive:   "评论包含敏感词",
	CodeCommentTooFrequent: "评论过于频繁",
	CodeCannotReplyDeleted: "无法回复已删除的评论",

	// ==================== 消息相关错误 ====================
	CodeMessageNotExist:    "消息不存在",
	CodeCannotSendToSelf:   "不能给自己发消息",
	CodeMessageTooLong:     "消息内容过长",
	CodeMessageTooFrequent: "消息发送过于频繁",

	// ==================== 系统错误 ====================
	CodeInvalidParam:      "请求参数错误",
	CodeBindError:         "参数绑定错误",
	CodeValidationError:   "参数验证失败",
	CodeTooManyRequests:   "请求过于频繁，请稍后再试",
	CodeNotFound:          "资源不存在",
	CodeMethodNotAllowed:  "请求方法不允许",
	CodeDatabaseError:     "数据库错误",
	CodeCacheError:        "缓存错误",
	CodeInternalError:     "服务器内部错误",
	CodeFileReadError:     "文件读取错误",
	CodeFileWriteError:    "文件写入错误",

	// ==================== 第三方服务错误 ====================
	CodeServerBusy:      "服务繁忙",
	CodeUploadFailed:    "文件上传失败",
	CodeSMSFailed:       "短信发送失败",
	CodeOSSError:        "对象存储错误",
	CodeThirdPartyError: "第三方服务错误",
}

func (c ResCode) Msg() string {
	msg, ok := codeMsgMap[c]
	if !ok {
		msg = codeMsgMap[CodeServerBusy]
	}
	return msg
}
