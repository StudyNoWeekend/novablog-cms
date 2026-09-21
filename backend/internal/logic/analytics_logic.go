package logic

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"novablog/internal/cache"
	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
)

// AnalyticsLogic 工作台统计业务逻辑结构体。
type AnalyticsLogic struct {
	model *model.AnalyticsModel
	cache *cache.AnalyticsCache
}

// NewAnalyticsLogic 创建 AnalyticsLogic 实例。
func NewAnalyticsLogic() *AnalyticsLogic {
	return &AnalyticsLogic{
		model: model.NewAnalytics(),
		cache: cache.NewAnalyticsCache(),
	}
}

// GetOverview 获取工作台概览统计数据（带缓存）。
func (l *AnalyticsLogic) GetOverview(ctx context.Context) (*res.OverviewRes, error) {
	// 优先读取缓存
	cached, err := l.cache.GetOverview(ctx)
	if err != nil {
		return nil, fmt.Errorf("读取概览缓存失败: %w", err)
	}
	if cached != nil {
		return cached, nil
	}

	// 缓存未命中，并行查询各项数据
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		firstErr error

		artTotal, artPub, artDraft int64
		portTotal, portPub         int64
		videoTotal, videoPub       int64
		travelTotal, travelPub     int64
		songTotal                  int64
		commentTotal               int64
		travelViews                int64
		lastPublishAt              *time.Time
	)

	// 文章统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, p, d, e := l.model.CountArticles(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		artTotal, artPub, artDraft = t, p, d
	}()

	// 作品集统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, p, e := l.model.CountPortfolios(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		portTotal, portPub = t, p
	}()

	// 视频统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, p, e := l.model.CountVideos(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		videoTotal, videoPub = t, p
	}()

	// 旅行攻略统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, p, e := l.model.CountTravelGuides(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		travelTotal, travelPub = t, p
	}()

	// 歌曲统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, e := l.model.CountSongs(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		songTotal = t
	}()

	// 评论统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, e := l.model.CountComments(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		commentTotal = t
	}()

	// 攻略浏览量统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		v, e := l.model.SumTravelViews(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		travelViews = v
	}()

	// 最近发布时间
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, e := l.model.GetLastPublishDate(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		lastPublishAt = t
	}()

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	// 计算距上次发布天数
	daysSinceLastPublish := 0
	if lastPublishAt != nil {
		daysSinceLastPublish = int(time.Since(*lastPublishAt).Hours() / 24)
	}

	overview := &res.OverviewRes{
		ArticleTotal:         artTotal,
		ArticlePublished:     artPub,
		ArticleDraft:         artDraft,
		PortfolioTotal:       portTotal,
		PortfolioPublished:   portPub,
		VideoTotal:           videoTotal,
		VideoPublished:       videoPub,
		TravelTotal:          travelTotal,
		TravelPublished:      travelPub,
		SongTotal:            songTotal,
		CommentTotal:         commentTotal,
		TravelViews:          travelViews,
		LastPublishAt:        lastPublishAt,
		DaysSinceLastPublish: daysSinceLastPublish,
	}

	// 写入缓存（失败不影响主流程）
	_ = l.cache.SetOverview(ctx, overview)

	return overview, nil
}

// GetContentTrend 获取内容产出趋势数据。
func (l *AnalyticsLogic) GetContentTrend(ctx context.Context, r req.ContentTrendReq) (*res.ContentTrendRes, error) {
	rangeStr := r.GetRange()
	var days int
	switch rangeStr {
	case "7d":
		days = 7
	case "90d":
		days = 90
	default:
		days = 30
	}

	rows, err := l.model.GetContentTrend(ctx, days)
	if err != nil {
		return nil, fmt.Errorf("查询内容趋势失败: %w", err)
	}

	items := make([]res.ContentTrendItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, res.ContentTrendItem{
			Date:         row.Date.Format("2006-01-02"),
			ArticleCount: row.ArticleCount,
			TravelCount:  row.TravelCount,
		})
	}

	return &res.ContentTrendRes{
		Range: rangeStr,
		Items: items,
	}, nil
}

// GetTopContent 获取热门内容排行数据。
func (l *AnalyticsLogic) GetTopContent(ctx context.Context, r req.TopContentReq) (*res.TopContentRes, error) {
	contentType := r.GetType()
	sort := r.GetSort()
	limit := r.GetLimit()

	var items []res.TopContentItem

	switch contentType {
	case "travel":
		rows, err := l.model.GetTopTravelGuides(ctx, sort, limit)
		if err != nil {
			return nil, fmt.Errorf("查询热门旅行攻略失败: %w", err)
		}
		items = make([]res.TopContentItem, 0, len(rows))
		for _, row := range rows {
			items = append(items, res.TopContentItem{
				ID:           row.ID,
				Title:        row.Title,
				ViewCount:    row.ViewCount,
				CommentCount: row.CommentCount,
				PublishedAt:  row.PublishedAt,
			})
		}
	default:
		rows, err := l.model.GetTopArticles(ctx, sort, limit)
		if err != nil {
			return nil, fmt.Errorf("查询热门文章失败: %w", err)
		}
		items = make([]res.TopContentItem, 0, len(rows))
		for _, row := range rows {
			items = append(items, res.TopContentItem{
				ID:           row.ID,
				Title:        row.Title,
				Slug:         row.Slug,
				ViewCount:    row.ViewCount,
				CommentCount: row.CommentCount,
				PublishedAt:  row.PublishedAt,
			})
		}
	}

	return &res.TopContentRes{
		Type:  contentType,
		Sort:  sort,
		Items: items,
	}, nil
}

// GetDistribution 获取内容类型分布数据（带缓存）。
func (l *AnalyticsLogic) GetDistribution(ctx context.Context) (*res.DistributionRes, error) {
	// 优先读取缓存
	cached, err := l.cache.GetDistribution(ctx)
	if err != nil {
		return nil, fmt.Errorf("读取分布缓存失败: %w", err)
	}
	if cached != nil {
		return cached, nil
	}

	// 缓存未命中，并行查询各内容类型总数
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		firstErr error

		artTotal    int64
		portTotal   int64
		videoTotal  int64
		travelTotal int64
		songTotal   int64
	)

	// 文章总数
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, _, _, e := l.model.CountArticles(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		artTotal = t
	}()

	// 作品集总数
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, _, e := l.model.CountPortfolios(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		portTotal = t
	}()

	// 视频总数
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, _, e := l.model.CountVideos(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		videoTotal = t
	}()

	// 旅行攻略总数
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, _, e := l.model.CountTravelGuides(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		travelTotal = t
	}()

	// 歌曲总数
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, e := l.model.CountSongs(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		songTotal = t
	}()

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	total := artTotal + portTotal + videoTotal + travelTotal + songTotal

	items := []res.DistributionItem{
		{Type: "article", Name: "文章", Count: artTotal},
		{Type: "portfolio", Name: "摄影作品", Count: portTotal},
		{Type: "video", Name: "视频作品", Count: videoTotal},
		{Type: "travel", Name: "旅行攻略", Count: travelTotal},
		{Type: "song", Name: "音乐", Count: songTotal},
	}

	// 计算百分比
	if total > 0 {
		for i := range items {
			p := float64(items[i].Count) / float64(total) * 100
			// 保留一位小数
			items[i].Percentage = math.Round(p*10) / 10
		}
	}

	dist := &res.DistributionRes{
		Items: items,
		Total: total,
	}

	// 写入缓存（失败不影响主流程）
	_ = l.cache.SetDistribution(ctx, dist)

	return dist, nil
}

// GetRecentComments 获取最近评论数据。
func (l *AnalyticsLogic) GetRecentComments(ctx context.Context, r req.RecentCommentsReq) (*res.RecentCommentsRes, error) {
	rows, err := l.model.GetRecentComments(ctx, r.GetLimit())
	if err != nil {
		return nil, fmt.Errorf("查询最近评论失败: %w", err)
	}

	items := make([]res.RecentCommentItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, res.RecentCommentItem{
			ID:          row.ID,
			Nickname:    row.Nickname,
			Content:     row.Content,
			TargetType:  row.TargetType,
			TargetID:    row.TargetID,
			TargetTitle: row.TargetTitle,
			IsBlogger:   row.IsBlogger,
			CreatedAt:   row.CreatedAt,
		})
	}

	return &res.RecentCommentsRes{
		Items: items,
	}, nil
}
