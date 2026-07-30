package music

import (
	"context"
	"flashvid-platform-gin/api"
	"flashvid-platform-gin/internal/dao/query"
	"flashvid-platform-gin/internal/model"
)

// GetMusicList 获取音乐列表
func GetMusicList(ctx context.Context, sort string, page, pageSize int) (*model.MusicListOutput, api.ResCode, error) {
	q := query.Music.WithContext(ctx).Where(query.Music.Status.Eq(1))

	if sort == "latest" {
		q = q.Order(query.Music.CreatedAt.Desc())
	} else {
		// 默认按热门（使用次数）排序
		q = q.Order(query.Music.UseCount.Desc())
	}

	// 1. 统计总数
	total, err := q.Count()
	if err != nil {
		return nil, api.CodeInternalError, err
	}

	// 2. 分页查询
	musics, err := q.Offset((page - 1) * pageSize).Limit(pageSize).Find()
	if err != nil {
		return nil, api.CodeInternalError, err
	}

	// 3. 封装成 MusicInfo
	var list []model.MusicInfo
	for _, m := range musics {
		list = append(list, model.MusicInfo{
			ID:        m.ID,
			Name:      m.Name,
			Artist:    m.Artist,
			Album:     m.Album,
			CoverURL:  m.CoverURL,
			MusicURL:  m.MusicURL,
			Duration:  m.Duration,
			UseCount:  m.UseCount,
			CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &model.MusicListOutput{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, api.CodeSuccess, nil
}

// SearchMusic 搜索音乐（按名称或艺术家）
func SearchMusic(ctx context.Context, keyword string, page, pageSize int) (*model.MusicListOutput, api.ResCode, error) {
	searchPattern := "%" + keyword + "%"

	// 1. 统计总数（gorm-gen 不支持字段级 OR，用 UnderlyingDB 追加原生条件）
	var total int64
	if err := query.Music.WithContext(ctx).
		Where(query.Music.Status.Eq(1)).
		UnderlyingDB().
		Where("name LIKE ? OR artist LIKE ?", searchPattern, searchPattern).
		Count(&total).Error; err != nil {
		return nil, api.CodeInternalError, err
	}

	// 2. 分页查询
	var musics []*model.Music
	if err := query.Music.WithContext(ctx).
		Where(query.Music.Status.Eq(1)).
		UnderlyingDB().
		Where("name LIKE ? OR artist LIKE ?", searchPattern, searchPattern).
		Order("use_count DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&musics).Error; err != nil {
		return nil, api.CodeInternalError, err
	}

	// 3. 封装成 MusicInfo
	var list []model.MusicInfo
	for _, m := range musics {
		list = append(list, model.MusicInfo{
			ID:        m.ID,
			Name:      m.Name,
			Artist:    m.Artist,
			Album:     m.Album,
			CoverURL:  m.CoverURL,
			MusicURL:  m.MusicURL,
			Duration:  m.Duration,
			UseCount:  m.UseCount,
			CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &model.MusicListOutput{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, api.CodeSuccess, nil
}
