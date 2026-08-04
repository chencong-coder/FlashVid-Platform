package api_test

import (
	"encoding/json"
	"testing"

	authv1 "flashvid-platform-gin/api/auth/v1"
	commentv1 "flashvid-platform-gin/api/comment/v1"
	messagev1 "flashvid-platform-gin/api/message/v1"
	playlistv1 "flashvid-platform-gin/api/playlist/v1"
	userv1 "flashvid-platform-gin/api/user/v1"
	videov1 "flashvid-platform-gin/api/video/v1"
	"flashvid-platform-gin/internal/model"

	"github.com/gin-gonic/gin/binding"
)

const largeSnowflakeID int64 = 9_007_199_254_740_993

func TestSnowflakeResponseIDsAreJSONStrings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
		field string
	}{
		{name: "register user", value: authv1.RegisterResp{UserID: largeSnowflakeID}, field: "userId"},
		{name: "login user", value: authv1.LoginResp{UserID: largeSnowflakeID}, field: "userId"},
		{name: "user profile", value: userv1.UserInfoResp{UserId: largeSnowflakeID}, field: "userId"},
		{name: "updated user profile", value: userv1.UpdateUserInfoResp{UserId: largeSnowflakeID}, field: "userId"},
		{name: "follow list user", value: model.UserInfo{UserId: largeSnowflakeID}, field: "userId"},
		{name: "created video", value: videov1.CreateVideoResp{VideoID: largeSnowflakeID}, field: "video_id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("marshal response: %v", err)
			}

			var body map[string]any
			if err := json.Unmarshal(encoded, &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			got, ok := body[tt.field].(string)
			if !ok {
				t.Fatalf("%s must be a JSON string, got %T (%v)", tt.field, body[tt.field], body[tt.field])
			}
			if got != "9007199254740993" {
				t.Fatalf("unexpected %s: %q", tt.field, got)
			}
		})
	}
}

func TestSnowflakeRequestIDsRequireJSONStrings(t *testing.T) {
	t.Parallel()

	t.Run("message recipient", func(t *testing.T) {
		var req messagev1.SendMessageReq
		assertStringIDRequest(
			t,
			`{"toUserId":"9007199254740993","messageType":1}`,
			`{"toUserId":9007199254740993,"messageType":1}`,
			&req,
		)
		if req.ToUserID != "9007199254740993" {
			t.Fatalf("unexpected toUserId: %q", req.ToUserID)
		}
		assertValidBinding(t, &req)
	})

	t.Run("comment relations", func(t *testing.T) {
		var req commentv1.CreateCommentReq
		assertStringIDRequest(
			t,
			`{"content":"reply","parentId":"9007199254740993","replyToUserId":"9007199254740992"}`,
			`{"content":"reply","parentId":9007199254740993,"replyToUserId":9007199254740992}`,
			&req,
		)
		if req.ParentID != "9007199254740993" || req.ReplyToUserID != "9007199254740992" {
			t.Fatalf("unexpected comment IDs: parent=%q replyTo=%q", req.ParentID, req.ReplyToUserID)
		}
		assertValidBinding(t, &req)
	})

	t.Run("video music", func(t *testing.T) {
		var req videov1.CreateVideoReq
		assertStringIDRequest(
			t,
			`{"title":"title","coverUrl":"cover","videoUrl":"video","duration":1,"musicId":"9007199254740993"}`,
			`{"title":"title","coverUrl":"cover","videoUrl":"video","duration":1,"musicId":9007199254740993}`,
			&req,
		)
		if req.MusicId == nil || *req.MusicId != "9007199254740993" {
			t.Fatalf("unexpected musicId: %v", req.MusicId)
		}
		assertValidBinding(t, &req)
	})

	t.Run("playlist video", func(t *testing.T) {
		var req playlistv1.AddVideoToPlaylistReq
		assertStringIDRequest(t, `{"videoId":"9007199254740993"}`, `{"videoId":9007199254740993}`, &req)
		if req.VideoID != largeSnowflakeID {
			t.Fatalf("unexpected videoId: %d", req.VideoID)
		}
		req.PlaylistID = 1
		assertValidBinding(t, &req)
	})
}

func assertValidBinding(t *testing.T, target any) {
	t.Helper()

	if err := binding.Validator.ValidateStruct(target); err != nil {
		t.Fatalf("request binding validation failed: %v", err)
	}
}

func assertStringIDRequest(t *testing.T, validBody string, numericBody string, target any) {
	t.Helper()

	if err := json.Unmarshal([]byte(numericBody), target); err == nil {
		t.Fatal("numeric JSON ID must be rejected")
	}
	if err := json.Unmarshal([]byte(validBody), target); err != nil {
		t.Fatalf("quoted ID must decode: %v", err)
	}
}
