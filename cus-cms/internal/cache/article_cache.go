package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cus-cms/internal/model"

	"github.com/go-redis/redis/v8"
)

// ArticleCache 文章缓存操作结构体。
type ArticleCache struct {
	rdb *redis.Client
}

// NewArticleCache 创建 ArticleCache 实例。
func NewArticleCache(rdb *redis.Client) *ArticleCache {
	return &ArticleCache{rdb: rdb}
}

// 缓存 Key 前缀
const (
	articleDetailPrefix = "article:detail:"
	articleListPrefix   = "article:list:"
)

// 默认过期时间
const (
	articleDetailTTL = 30 * time.Minute
	articleListTTL   = 10 * time.Minute
)

// GetDetail 从缓存获取文章详情。
func (c *ArticleCache) GetDetail(ctx context.Context, id string) (*model.Article, error) {
	key := articleDetailPrefix + id
	data, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var article model.Article
	if err := json.Unmarshal([]byte(data), &article); err != nil {
		return nil, err
	}
	return &article, nil
}

// SetDetail 缓存文章详情。
func (c *ArticleCache) SetDetail(ctx context.Context, article *model.Article) error {
	key := articleDetailPrefix + article.ID
	data, err := json.Marshal(article)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, articleDetailTTL).Err()
}

// DeleteDetail 删除文章详情缓存。
func (c *ArticleCache) DeleteDetail(ctx context.Context, id string) error {
	return c.rdb.Del(ctx, articleDetailPrefix+id).Err()
}

// GetList 从缓存获取文章列表。
func (c *ArticleCache) GetList(ctx context.Context, page, pageSize int) ([]model.Article, error) {
	key := fmt.Sprintf("%s%d:%d", articleListPrefix, page, pageSize)
	data, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var articles []model.Article
	if err := json.Unmarshal([]byte(data), &articles); err != nil {
		return nil, err
	}
	return articles, nil
}

// SetList 缓存文章列表。
func (c *ArticleCache) SetList(ctx context.Context, page, pageSize int, articles []model.Article) error {
	key := fmt.Sprintf("%s%d:%d", articleListPrefix, page, pageSize)
	data, err := json.Marshal(articles)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, articleListTTL).Err()
}

// InvalidateList 使文章列表缓存全部失效（模糊删除）。
func (c *ArticleCache) InvalidateList(ctx context.Context) error {
	keys, err := c.rdb.Keys(ctx, articleListPrefix+"*").Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.rdb.Del(ctx, keys...).Err()
	}
	return nil
}