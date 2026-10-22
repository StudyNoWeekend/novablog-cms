package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListCommentsIn list_comments 入参。
type ListCommentsIn struct {
	TargetType string `json:"target_type,omitempty" jsonschema:"按目标类型过滤: article(文章) / travel_guide(旅行攻略)，不传返回全部"`
	TargetID   string `json:"target_id,omitempty" jsonschema:"按目标内容 ID 过滤"`
	Keyword    string `json:"keyword,omitempty" jsonschema:"按评论内容模糊搜索"`
	Page       int    `json:"page,omitempty" jsonschema:"页码，默认 1"`
	PageSize   int    `json:"page_size,omitempty" jsonschema:"每页条数，默认 20，最大 100"`
}

// listComments 分页查询评论列表。
func (s *Service) listComments(ctx context.Context, _ *sdk.CallToolRequest, in ListCommentsIn) (*sdk.CallToolResult, res.PageRes[res.CommentRes], error) {
	if in.PageSize <= 0 {
		in.PageSize = 20
	}
	if in.PageSize > 100 {
		in.PageSize = 100
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.TargetType != "" && in.TargetType != "article" && in.TargetType != "travel_guide" {
		return nil, res.PageRes[res.CommentRes]{}, fmt.Errorf("非法的目标类型 %q，可选：article / travel_guide", in.TargetType)
	}

	listReq := &req.CommentListReq{
		PageReq:  req.PageReq{Page: in.Page, PageSize: in.PageSize},
		TargetID: stringPtr(in.TargetID),
		Keyword:  stringPtr(in.Keyword),
	}
	if in.TargetType != "" {
		listReq.TargetType = &in.TargetType
	}

	page, err := s.commentLogic.GetList(ctx, listReq)
	if err != nil {
		return nil, res.PageRes[res.CommentRes]{}, fmt.Errorf("查询评论列表失败: %w", err)
	}
	return nil, *page, nil
}

// ReplyCommentIn reply_comment 入参。
type ReplyCommentIn struct {
	CommentID string `json:"comment_id" jsonschema:"要回复的评论 ID"`
	Content   string `json:"content" jsonschema:"回复内容（1-2000 字，Markdown）"`
}

// ReplyCommentOut reply_comment 出参。
type ReplyCommentOut struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	IsBlogger bool      `json:"is_blogger"`
	CreatedAt time.Time `json:"created_at"`
}

// replyComment 以博主身份回复评论。
func (s *Service) replyComment(ctx context.Context, _ *sdk.CallToolRequest, in ReplyCommentIn) (*sdk.CallToolResult, ReplyCommentOut, error) {
	if in.CommentID == "" {
		return nil, ReplyCommentOut{}, fmt.Errorf("comment_id 不能为空")
	}
	content := strings.TrimSpace(in.Content)
	if content == "" {
		return nil, ReplyCommentOut{}, fmt.Errorf("content 不能为空")
	}
	if len([]rune(content)) > 2000 {
		return nil, ReplyCommentOut{}, fmt.Errorf("回复内容不能超过 2000 字")
	}

	comment, err := s.commentLogic.Reply(ctx, s.currentBloggerID(ctx), in.CommentID, &req.ReplyCommentReq{Content: content})
	if err != nil {
		return nil, ReplyCommentOut{}, fmt.Errorf("回复评论失败: %w", err)
	}
	return nil, ReplyCommentOut{
		ID:        comment.ID,
		Content:   comment.Content,
		IsBlogger: comment.IsBlogger,
		CreatedAt: comment.CreatedAt,
	}, nil
}

// DeleteCommentIn delete_comment 入参。
type DeleteCommentIn struct {
	CommentID string `json:"comment_id" jsonschema:"要删除的评论 ID"`
	Confirm   bool   `json:"confirm" jsonschema:"必须显式传 true 才会执行删除"`
}

// DeleteCommentOut delete_comment 出参。
type DeleteCommentOut struct {
	Deleted bool   `json:"deleted"`
	ID      string `json:"id"`
}

// deleteComment 删除评论（级联软删子回复）。
func (s *Service) deleteComment(ctx context.Context, _ *sdk.CallToolRequest, in DeleteCommentIn) (*sdk.CallToolResult, DeleteCommentOut, error) {
	if in.CommentID == "" {
		return nil, DeleteCommentOut{}, fmt.Errorf("comment_id 不能为空")
	}
	if !in.Confirm {
		return nil, DeleteCommentOut{}, fmt.Errorf("删除是不可逆操作（会同时删除该评论下的所有回复），请确认后显式传 confirm=true 再执行")
	}
	if err := s.commentLogic.Delete(ctx, in.CommentID); err != nil {
		return nil, DeleteCommentOut{}, fmt.Errorf("删除评论失败: %w", err)
	}
	return nil, DeleteCommentOut{Deleted: true, ID: in.CommentID}, nil
}

// stringPtr 空串转 nil（用于可选过滤参数）。
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// trimSpacePtr 去除首尾空白后转指针，空串返回 nil。
func trimSpacePtr(s string) *string {
	return stringPtr(strings.TrimSpace(s))
}
