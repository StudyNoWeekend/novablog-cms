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
	model     *model.AnalyticsModel
	viewStats *model.ViewStatsModel
	cache     *cache.AnalyticsCache
}

// NewAnalyticsLogic 创建 AnalyticsLogic 实例。
func NewAnalyticsLogic() *AnalyticsLogic {
	return &AnalyticsLogic{
		model:     model.NewAnalytics(),
		viewStats: model.NewViewStats(),
		cache:     cache.NewAnalyticsCache(),
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

		equipTotal                   int64
		projTotal, projPub           int64
		osTotal, osPub               int64
		recipeTotal, recipePub       int64
		bookTotal, bookPub           int64
		gameTotal, gamePub           int64
		fitnessTotal, fitnessPub     int64
		techStackTotal, techStackPub int64
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

	// 个人设备统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, e := l.model.CountEquipment(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		equipTotal = t
	}()

	// 项目经历统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, p, e := l.model.CountProjects(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		projTotal, projPub = t, p
	}()

	// 开源作品统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, p, e := l.model.CountOpenSourceWorks(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		osTotal, osPub = t, p
	}()

	// 菜谱统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, p, e := l.model.CountRecipes(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		recipeTotal, recipePub = t, p
	}()

	// 读书统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, p, e := l.model.CountBooks(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		bookTotal, bookPub = t, p
	}()

	// 游戏统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, p, e := l.model.CountGames(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		gameTotal, gamePub = t, p
	}()

	// 健身训练统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, p, e := l.model.CountFitnessRecords(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		fitnessTotal, fitnessPub = t, p
	}()

	// 技术栈统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, p, e := l.model.CountTechStackItems(ctx)
		if e != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = e
			}
			mu.Unlock()
			return
		}
		techStackTotal, techStackPub = t, p
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
		EquipmentTotal:       equipTotal,
		ProjectTotal:         projTotal,
		ProjectPublished:     projPub,
		OpenSourceTotal:      osTotal,
		OpenSourcePublished:  osPub,
		RecipeTotal:          recipeTotal,
		RecipePublished:      recipePub,
		BookTotal:            bookTotal,
		BookPublished:        bookPub,
		GameTotal:            gameTotal,
		GamePublished:        gamePub,
		FitnessTotal:         fitnessTotal,
		FitnessPublished:     fitnessPub,
		TechStackTotal:       techStackTotal,
		TechStackPublished:   techStackPub,
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
			Date:   row.Date.Format("2006-01-02"),
			Counts: row.Counts,
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
	case "all":
		rows, err := l.viewStats.GetTopAllContent(ctx, limit)
		if err != nil {
			return nil, fmt.Errorf("查询全模块热门内容失败: %w", err)
		}
		items = make([]res.TopContentItem, 0, len(rows))
		for _, row := range rows {
			items = append(items, res.TopContentItem{
				ID:           row.ID,
				Type:         row.Type,
				Title:        row.Title,
				ViewCount:    row.ViewCount,
				CommentCount: row.CommentCount,
			})
		}
	case "travel":
		rows, err := l.model.GetTopTravelGuides(ctx, sort, limit)
		if err != nil {
			return nil, fmt.Errorf("查询热门旅行攻略失败: %w", err)
		}
		items = make([]res.TopContentItem, 0, len(rows))
		for _, row := range rows {
			items = append(items, res.TopContentItem{
				ID:           row.ID,
				Type:         "travel",
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
				Type:         "article",
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

	// 参与内容分布统计的模块，新增模块时在此维护
	type distSource struct {
		typ   string
		name  string
		count func(ctx context.Context) (int64, error)
	}
	sources := []distSource{
		{"article", "文章", func(ctx context.Context) (int64, error) { t, _, _, e := l.model.CountArticles(ctx); return t, e }},
		{"portfolio", "摄影作品", func(ctx context.Context) (int64, error) { t, _, e := l.model.CountPortfolios(ctx); return t, e }},
		{"video", "视频作品", func(ctx context.Context) (int64, error) { t, _, e := l.model.CountVideos(ctx); return t, e }},
		{"travel", "旅行攻略", func(ctx context.Context) (int64, error) { t, _, e := l.model.CountTravelGuides(ctx); return t, e }},
		{"song", "音乐", func(ctx context.Context) (int64, error) { return l.model.CountSongs(ctx) }},
		{"equipment", "个人设备", func(ctx context.Context) (int64, error) { return l.model.CountEquipment(ctx) }},
		{"project", "项目经历", func(ctx context.Context) (int64, error) { t, _, e := l.model.CountProjects(ctx); return t, e }},
		{"open_source", "开源作品", func(ctx context.Context) (int64, error) { t, _, e := l.model.CountOpenSourceWorks(ctx); return t, e }},
		{"recipe", "美食菜谱", func(ctx context.Context) (int64, error) { t, _, e := l.model.CountRecipes(ctx); return t, e }},
		{"book", "读书书架", func(ctx context.Context) (int64, error) { t, _, e := l.model.CountBooks(ctx); return t, e }},
		{"game", "游戏库", func(ctx context.Context) (int64, error) { t, _, e := l.model.CountGames(ctx); return t, e }},
		{"fitness", "健身训练", func(ctx context.Context) (int64, error) { t, _, e := l.model.CountFitnessRecords(ctx); return t, e }},
		{"tech_stack", "技术栈", func(ctx context.Context) (int64, error) { t, _, e := l.model.CountTechStackItems(ctx); return t, e }},
	}

	// 并行查询各内容类型总数
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		firstErr error
	)
	counts := make(map[string]int64, len(sources))
	for _, s := range sources {
		wg.Add(1)
		go func(s distSource) {
			defer wg.Done()
			t, e := s.count(ctx)
			if e != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = e
				}
				mu.Unlock()
				return
			}
			mu.Lock()
			counts[s.typ] = t
			mu.Unlock()
		}(s)
	}

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	items := make([]res.DistributionItem, 0, len(sources))
	total := int64(0)
	for _, s := range sources {
		total += counts[s.typ]
		items = append(items, res.DistributionItem{Type: s.typ, Name: s.name, Count: counts[s.typ]})
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
