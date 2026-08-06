-- 清理关注/点赞/收藏三张表里的软删残留行
--
-- 背景：这三张表带 soft-delete（deleted_at），但唯一索引不含 deleted_at：
--   follows:   uk_follower_following (follower_id, following_id)
--   likes:     uk_user_target        (user_id, target_type, target_id)
--   favorites: uk_user_video         (user_id, video_id)
-- 旧代码取消关注/点赞时只软删（deleted_at 打标记，行仍占唯一键位置）。
-- 再次关注/点赞时 Count() 走默认作用域过滤掉软删行 → 以为没关注 → Create()
-- → 撞唯一键 duplicate key → 事务失败 → 返回 code:40009（服务器内部错误）。
--
-- 服务层已改为 Unscoped() 硬删除，不再产生新的软删行；
-- 本脚本清掉历史遗留的软删行。这些行逻辑上本就是"已取消"状态，物理删除无损业务数据。

DELETE FROM follows   WHERE deleted_at IS NOT NULL;
DELETE FROM likes     WHERE deleted_at IS NOT NULL;
DELETE FROM favorites WHERE deleted_at IS NOT NULL;
