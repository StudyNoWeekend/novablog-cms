package logic

import (
	"context"
	"fmt"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"

	"github.com/google/uuid"
)

// CommentLogic 评论业务逻辑结构体。
type CommentLogic struct {
	commentModel     *model.CommentModel
	articleModel     *model.ArticleModel
	travelGuideModel *model.TravelGuideModel
	bloggerModel     *model.BloggerModel
}

// NewCommentLogic 创建 CommentLogic 实例。
func NewCommentLogic() *CommentLogic {
	return &CommentLogic{
		commentModel:     model.NewComment(),
		articleModel:     model.NewArticle(),
		travelGuideModel: model.NewTravelGuide(),
		bloggerModel:     model.NewBlogger(),
	}
}

// GetList 获取评论列表，关联查询目标标题。
func (l *CommentLogic) GetList(ctx context.Context, r *req.CommentListReq) (*res.PageRes[res.CommentRes], error) {
	comments, total, err := l.commentModel.GetList(
		ctx, r.GetPage(), r.GetPageSize(),
		r.TargetType, r.TargetID, r.Keyword,
	)
	if err != nil {
		return nil, fmt.Errorf("查询评论列表失败: %w", err)
	}

	// 预加载目标标题缓存，避免重复查询
	// TODO: 这里存在 N+1 查询问题，每条评论可能触发一次 getTargetTitle。
	// 当前 per-request titleCache 能避免对同一 (target_type, target_id) 的重复查询，
	// 覆盖了多数评论指向同一文章的常见场景。如需进一步优化，可改为批量查询所有目标标题。
	titleCache := make(map[string]string)

	var items []res.CommentRes
	for _, c := range comments {
		item := l.toCommentRes(ctx, &c, titleCache)
		items = append(items, item)
	}

	return res.NewPageRes(items, total, r.GetPage(), r.GetPageSize()), nil
}

// Reply 博主回复评论。
func (l *CommentLogic) Reply(ctx context.Context, bloggerID string, commentID string, r *req.ReplyCommentReq) (*res.CommentRes, error) {
	// 查询被回复的评论
	parent, err := l.commentModel.GetByID(ctx, commentID)
	if err != nil {
		return nil, fmt.Errorf("被回复的评论不存在")
	}

	// 获取博主信息
	blogger, err := l.bloggerModel.GetByID(ctx, bloggerID)
	if err != nil {
		return nil, fmt.Errorf("获取博主信息失败: %w", err)
	}

	// 创建回复评论
	reply := &model.Comment{
		ID:         uuid.New().String(),
		TargetType: parent.TargetType,
		TargetID:   parent.TargetID,
		ParentID:   &parent.ID,
		BloggerID:  &blogger.ID,
		Nickname:   blogger.Nickname,
		Content:    r.Content,
		IsBlogger:  true,
		Status:     2, // 博主回复默认已通过
	}

	if err := l.commentModel.Create(ctx, reply); err != nil {
		return nil, fmt.Errorf("创建回复失败: %w", err)
	}

	// 查询目标标题
	titleCache := make(map[string]string)
	result := l.toCommentRes(ctx, reply, titleCache)
	return &result, nil
}

// Delete 软删除评论及其所有子回复。
func (l *CommentLogic) Delete(ctx context.Context, id string) error {
	_, err := l.commentModel.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("评论不存在")
	}

	// 软删除评论本身
	if err := l.commentModel.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("删除评论失败: %w", err)
	}

	// 级联软删除该评论的所有子回复
	if err := l.commentModel.SoftDeleteByParentID(ctx, id); err != nil {
		return fmt.Errorf("删除子回复失败: %w", err)
	}

	return nil
}

// getTargetTitle 根据目标类型和 ID 查询标题，带缓存避免重复查询。
func (l *CommentLogic) getTargetTitle(ctx context.Context, targetType, targetID string, cache map[string]string) string {
	cacheKey := targetType + ":" + targetID
	if title, ok := cache[cacheKey]; ok {
		return title
	}

	var title string
	switch targetType {
	case "article":
		article, err := l.articleModel.GetByID(ctx, targetID)
		if err == nil {
			title = article.Title
		}
	case "travel_guide":
		guide, err := l.travelGuideModel.GetByID(ctx, targetID)
		if err == nil {
			title = guide.Title
		}
	}
	cache[cacheKey] = title
	return title
}

// toCommentRes 将 Comment 模型转换为响应结构体。
func (l *CommentLogic) toCommentRes(ctx context.Context, c *model.Comment, titleCache map[string]string) res.CommentRes {
	return res.CommentRes{
		ID:          c.ID,
		TargetType:  c.TargetType,
		TargetID:    c.TargetID,
		TargetTitle: l.getTargetTitle(ctx, c.TargetType, c.TargetID, titleCache),
		ParentID:    c.ParentID,
		BloggerID:   c.BloggerID,
		Nickname:    c.Nickname,
		Website:     c.Website,
		Content:     c.Content,
		IsBlogger:   c.IsBlogger,
		IPAddress:   c.IPAddress,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

// GetPublicList 获取公开评论列表（仅已通过审核的评论）。
func (l *CommentLogic) GetPublicList(ctx context.Context, targetType, targetID string, page, pageSize int) (*res.PageRes[res.CommentPublicRes], error) {
	comments, total, err := l.commentModel.GetPublicList(ctx, targetType, targetID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询评论列表失败: %w", err)
	}

	items := make([]res.CommentPublicRes, 0, len(comments))
	for _, c := range comments {
		items = append(items, res.CommentPublicRes{
			ID:         c.ID,
			TargetType: c.TargetType,
			TargetID:   c.TargetID,
			ParentID:   c.ParentID,
			Nickname:   c.Nickname,
			Website:    c.Website,
			Content:    c.Content,
			IsBlogger:  c.IsBlogger,
			CreatedAt:  c.CreatedAt,
		})
	}

	return res.NewPageRes(items, total, page, pageSize), nil
}

// CreatePublic 创建公开评论，状态默认为已通过，记录 IP 地址。
func (l *CommentLogic) CreatePublic(ctx context.Context, r *req.CreatePublicCommentReq, ip string) (*res.CommentPublicRes, error) {
	comment := &model.Comment{
		ID:         uuid.New().String(),
		TargetType: r.TargetType,
		TargetID:   r.TargetID,
		ParentID:   r.ParentID,
		Nickname:   r.Nickname,
		Website:    r.Website,
		Content:    r.Content,
		IsBlogger:  false,
		Status:     2,
		IPAddress:  ip,
	}

	if err := l.commentModel.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("创建评论失败: %w", err)
	}

	return &res.CommentPublicRes{
		ID:         comment.ID,
		TargetType: comment.TargetType,
		TargetID:   comment.TargetID,
		ParentID:   comment.ParentID,
		Nickname:   comment.Nickname,
		Website:    comment.Website,
		Content:    comment.Content,
		IsBlogger:  comment.IsBlogger,
		CreatedAt:  comment.CreatedAt,
	}, nil
}
