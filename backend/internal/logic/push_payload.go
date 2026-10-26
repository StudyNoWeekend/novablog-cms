package logic

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"novablog/internal/model"
	"novablog/pkg/novapush"
)

// 推送契约约束（社区端强制校验，CMS 侧提前对齐避免 400）。
const (
	pushTitleMax     = 200
	pushSummaryMax   = 500
	pushContentMax   = 200 * 1024
	pushTagsMax      = 10
	pushTagMaxLen    = 30
	pushSourceURLMax = 500
)

// 内容类型标识（CMS 内部 slug，推送时 open_source 映射为社区字典的 opensource）。
const (
	TargetTypeArticle    = "article"
	TargetTypeTravel     = "travel"
	TargetTypeRecipe     = "recipe"
	TargetTypePortfolio  = "portfolio"
	TargetTypeOpenSource = "open_source"
)

// CommunityPushPayload 构建结果。
type CommunityPushPayload struct {
	Payload  *novapush.UpsertArticlePayload
	Warnings []string // 非致命提示（标签截断、摘要自动生成等），写入同步记录供博主查看
}

// BuildCommunityPayload 按内容类型加载内容并构建推送 payload：
// content_markdown 为合成保底正文（供搜索/摘要/回退渲染），extra 为 CMS 结构化参数直传（键名与 CMS 字段同名）。
func BuildCommunityPayload(ctx context.Context, targetType, targetID string) (*CommunityPushPayload, error) {
	switch targetType {
	case TargetTypeArticle:
		return buildArticlePushPayload(ctx, targetID)
	case TargetTypeTravel:
		return buildTravelPushPayload(ctx, targetID)
	case TargetTypeRecipe:
		return buildRecipePushPayload(ctx, targetID)
	case TargetTypePortfolio:
		return buildPortfolioPushPayload(ctx, targetID)
	case TargetTypeOpenSource:
		return buildOpenSourcePushPayload(ctx, targetID)
	default:
		return nil, fmt.Errorf("不支持推送的内容类型: %s", targetType)
	}
}

// —— 通用处理 ——

// normalizePushTags 标签归一化：去空白、去重、截断到 10 个且单个 ≤30 字符，截断行为记入 warnings。
func normalizePushTags(tags []string, warnings *[]string) []string {
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if utf8.RuneCountInString(tag) > pushTagMaxLen {
			runes := []rune(tag)
			tag = string(runes[:pushTagMaxLen])
		}
		if _, dup := seen[tag]; dup {
			continue
		}
		seen[tag] = struct{}{}
		if len(result) >= pushTagsMax {
			if warnings != nil {
				*warnings = append(*warnings, fmt.Sprintf("标签超过 %d 个，已截断", pushTagsMax))
			}
			break
		}
		result = append(result, tag)
	}
	return result
}

var markdownImagePattern = regexp.MustCompile(`!\[([^\]]*)\]\(([^)\s]+)\)`)
var markdownLinkPattern = regexp.MustCompile(`\]\(([^)\s]+)\)`)
var htmlSrcPattern = regexp.MustCompile(`(src=")([^"]+)(")`)

func isAbsoluteURL(raw string) bool {
	return strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://")
}

// resolveAbsolute 解析相对路径为绝对 URL，并修正 resolveURL 对前导斜杠路径产生的重复斜杠。
func resolveAbsolute(path string) string {
	resolved := model.ResolveURL(path)
	return strings.ReplaceAll(resolved, "/files//", "/files/")
}

// rewriteContentImages 把 markdown 正文中的相对路径图片/链接重写为绝对 URL（社区端外链直出）。
// 已是 http(s) 绝对地址的保持不变。
func rewriteContentImages(content string) string {
	if content == "" {
		return content
	}
	content = markdownImagePattern.ReplaceAllStringFunc(content, func(match string) string {
		sub := markdownImagePattern.FindStringSubmatch(match)
		if isAbsoluteURL(sub[2]) {
			return match
		}
		return "![" + sub[1] + "](" + resolveAbsolute(sub[2]) + ")"
	})
	content = htmlSrcPattern.ReplaceAllStringFunc(content, func(match string) string {
		sub := htmlSrcPattern.FindStringSubmatch(match)
		if isAbsoluteURL(sub[2]) {
			return match
		}
		return sub[1] + resolveAbsolute(sub[2]) + sub[3]
	})
	content = markdownLinkPattern.ReplaceAllStringFunc(content, func(match string) string {
		sub := markdownLinkPattern.FindStringSubmatch(match)
		if isAbsoluteURL(sub[1]) {
			return match
		}
		return "](" + resolveAbsolute(sub[1]) + ")"
	})
	return content
}

// ensureSummary 摘要为空时剥离 markdown 语法生成纯文本摘要（社区缺省会截 markdown 源码，观感差）。
func ensureSummary(summary, content string, warnings *[]string) string {
	if strings.TrimSpace(summary) != "" {
		return truncateRunes(summary, pushSummaryMax)
	}
	text := stripMarkdownSyntax(content)
	summary = truncateRunes(text, pushSummaryMax)
	if summary != "" && warnings != nil {
		*warnings = append(*warnings, "摘要在推送时由正文自动生成，建议在 CMS 中填写更精确的摘要")
	}
	return summary
}

var markdownSyntaxPatterns = []*regexp.Regexp{
	regexp.MustCompile("(?s)```.*?```"),         // 代码块
	regexp.MustCompile("`[^`]*`"),               // 行内代码
	regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)`),  // 图片
	regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`), // 链接保留文本
	regexp.MustCompile("(?m)^#{1,6}\\s+"),       // 标题井号
	regexp.MustCompile(`\*\*|\*|__|~~`),         // 强调
	regexp.MustCompile("(?m)^>\\s?"),            // 引用
	regexp.MustCompile("(?m)^\\s*[-*+]\\s+"),    // 无序列表符
}

func stripMarkdownSyntax(md string) string {
	for _, p := range markdownSyntaxPatterns {
		md = p.ReplaceAllString(md, "$1")
	}
	return strings.TrimSpace(md)
}

// pushSourceURL 按类型路由模板拼接原文链接。
func pushSourceURL(targetType, targetID, slug string) (string, []string) {
	site := strings.TrimRight(getCommunityPushSettings().SiteURL, "/")
	if site == "" {
		return "", []string{"未配置博客对外地址（community.site_url），推送内容将缺少「阅读原文」链接"}
	}
	var path string
	switch targetType {
	case TargetTypeArticle:
		path = "/articles/" + slug
	case TargetTypeTravel:
		path = "/travels/" + targetID
	case TargetTypeRecipe:
		path = "/recipes/" + targetID
	case TargetTypePortfolio:
		path = "/portfolio/" + targetID
	case TargetTypeOpenSource:
		path = "/open-source/" + targetID
	}
	url := site + path
	if len(url) > pushSourceURLMax {
		url = truncateRunes(url, pushSourceURLMax)
	}
	return url, nil
}

func rfc3339(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// —— article：直映射 ——

func buildArticlePushPayload(ctx context.Context, targetID string) (*CommunityPushPayload, error) {
	article, err := model.NewArticle().GetByID(ctx, targetID)
	if err != nil {
		return nil, fmt.Errorf("文章不存在: %w", err)
	}

	warnings := make([]string, 0)
	tags := make([]string, 0, len(article.Tags))
	for _, tag := range article.Tags {
		tags = append(tags, tag.Name)
	}
	content := rewriteContentImages(article.Content)
	summary := ensureSummary(article.Summary, content, &warnings)
	sourceURL, srcWarnings := pushSourceURL(TargetTypeArticle, article.ID, article.Slug)
	warnings = append(warnings, srcWarnings...)

	status := "published"
	if article.Status != 2 {
		status = "hidden"
	}

	// 封面：媒体引用动态解析（preset_id → media_id → 历史 URL）
	cover := model.ResolveMediaRefURL(ctx, refStr(article.CoverMediaID), refStr(article.CoverPresetID), article.CoverImage)

	return &CommunityPushPayload{
		Payload: &novapush.UpsertArticlePayload{
			ExternalID:      article.ID,
			Type:            TargetTypeArticle,
			Title:           truncateRunes(article.Title, pushTitleMax),
			Summary:         summary,
			ContentMarkdown: content,
			CoverImage:      resolveAbsolute(cover),
			Tags:            normalizePushTags(tags, &warnings),
			SourceURL:       sourceURL,
			PublishedAt:     rfc3339(publishedAtOrDefault(article.PublishedAt, article.CreatedAt)),
			Status:          status,
		},
		Warnings: warnings,
	}, nil
}

func publishedAtOrDefault(publishedAt *time.Time, createdAt time.Time) time.Time {
	if publishedAt != nil && !publishedAt.IsZero() {
		return *publishedAt
	}
	return createdAt
}

// —— travel：合成 markdown + extra 直传 ——

func buildTravelPushPayload(ctx context.Context, targetID string) (*CommunityPushPayload, error) {
	guide, err := model.NewTravelGuide().GetByID(ctx, targetID)
	if err != nil {
		return nil, fmt.Errorf("旅行攻略不存在: %w", err)
	}

	warnings := make([]string, 0)
	var md strings.Builder
	infoParts := make([]string, 0, 4)
	if guide.Destination != "" {
		infoParts = append(infoParts, "目的地 "+guide.Destination)
	}
	if guide.Region != "" {
		infoParts = append(infoParts, "地区 "+guide.Region)
	}
	if guide.Days > 0 {
		infoParts = append(infoParts, fmt.Sprintf("%d 天行程", guide.Days))
	}
	if guide.BestMonth != "" {
		infoParts = append(infoParts, "最佳月份 "+guide.BestMonth)
	}
	if len(infoParts) > 0 {
		md.WriteString("> " + strings.Join(infoParts, " · ") + "\n\n")
	}

	attractions := guide.Attractions
	if len(attractions) > 0 {
		md.WriteString("## 景点推荐\n\n")
		for i, a := range attractions {
			name, _ := a["name"].(string)
			if name == "" {
				name = fmt.Sprintf("景点 %d", i+1)
			}
			md.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, name))
			if duration, _ := a["duration"].(string); duration != "" {
				md.WriteString(fmt.Sprintf("建议游玩：%s\n\n", duration))
			}
			if desc, _ := a["description"].(string); desc != "" {
				md.WriteString(desc + "\n\n")
			}
			if image, _ := a["image"].(string); image != "" {
				md.WriteString(fmt.Sprintf("![%s](%s)\n\n", name, resolveAbsolute(image)))
			}
		}
	}

	itinerary := guide.Itinerary
	if len(itinerary) > 0 {
		md.WriteString("## 行程安排\n\n")
		for i, d := range itinerary {
			day, _ := d["day"].(float64)
			title, _ := d["title"].(string)
			if title != "" {
				md.WriteString(fmt.Sprintf("### Day %d：%s\n\n", intOrFallback(day, i+1), title))
			} else {
				md.WriteString(fmt.Sprintf("### Day %d\n\n", intOrFallback(day, i+1)))
			}
			if desc, _ := d["description"].(string); desc != "" {
				md.WriteString(desc + "\n\n")
			}
		}
	}

	reviews := guide.Reviews
	if len(reviews) > 0 {
		md.WriteString("## 游记随笔\n\n")
		for _, r := range reviews {
			username, _ := r["username"].(string)
			content, _ := r["content"].(string)
			if content == "" {
				continue
			}
			rating, _ := r["rating"].(float64)
			head := username
			if head == "" {
				head = "旅人"
			}
			if rating > 0 {
				head += fmt.Sprintf("（%.1f 分）", rating)
			}
			md.WriteString(fmt.Sprintf("> **%s**：%s\n\n", head, content))
		}
	}

	if strings.TrimSpace(md.String()) == "" {
		md.WriteString(guide.Summary)
	}

	sourceURL, srcWarnings := pushSourceURL(TargetTypeTravel, guide.ID, "")
	warnings = append(warnings, srcWarnings...)

	cover := model.ResolveMediaRefURL(ctx, refStr(guide.CoverMediaID), refStr(guide.CoverPresetID), guide.CoverImage)

	return &CommunityPushPayload{
		Payload: &novapush.UpsertArticlePayload{
			ExternalID:      guide.ID,
			Type:            TargetTypeTravel,
			Title:           truncateRunes(guide.Title, pushTitleMax),
			Summary:         ensureSummary(guide.Summary, md.String(), &warnings),
			ContentMarkdown: md.String(),
			CoverImage:      resolveAbsolute(cover),
			Tags:            normalizePushTags(nil, &warnings),
			SourceURL:       sourceURL,
			PublishedAt:     rfc3339(guide.CreatedAt),
			Status:          binaryStatusToPush(int(guide.Status), 2),
			Extra: map[string]any{
				"destination": guide.Destination,
				"region":      guide.Region,
				"days":        guide.Days,
				"best_month":  guide.BestMonth,
				"attractions": attractionsOrEmpty(guide.Attractions),
				"itinerary":   itineraryOrEmpty(guide.Itinerary),
				"reviews":     reviewsOrEmpty(guide.Reviews),
			},
		},
		Warnings: warnings,
	}, nil
}

// —— recipe：合成 markdown（食材表格 + 步骤）+ extra 直传 ——

func buildRecipePushPayload(ctx context.Context, targetID string) (*CommunityPushPayload, error) {
	recipe, err := model.NewRecipe().GetByID(ctx, targetID)
	if err != nil {
		return nil, fmt.Errorf("食谱不存在: %w", err)
	}

	warnings := make([]string, 0)
	var md strings.Builder
	difficultyLabels := map[int]string{1: "简单", 2: "中等", 3: "困难"}
	infoParts := make([]string, 0, 3)
	if recipe.Difficulty > 0 {
		infoParts = append(infoParts, "难度 "+difficultyLabels[recipe.Difficulty])
	}
	if recipe.Minutes > 0 {
		infoParts = append(infoParts, fmt.Sprintf("耗时 %d 分钟", recipe.Minutes))
	}
	if recipe.Servings > 0 {
		infoParts = append(infoParts, fmt.Sprintf("%d 人份", recipe.Servings))
	}
	if len(infoParts) > 0 {
		md.WriteString("> " + strings.Join(infoParts, " · ") + "\n\n")
	}

	if len(recipe.Ingredients) > 0 {
		md.WriteString("## 食材清单\n\n")
		md.WriteString("| 食材 | 用量 |\n|---|---|\n")
		for _, ing := range recipe.Ingredients {
			name, _ := ing["name"].(string)
			amount, _ := ing["amount"].(string)
			md.WriteString(fmt.Sprintf("| %s | %s |\n", strings.ReplaceAll(name, "|", "\\|"), strings.ReplaceAll(amount, "|", "\\|")))
		}
		md.WriteString("\n")
	}

	if len(recipe.Steps) > 0 {
		md.WriteString("## 做法步骤\n\n")
		for i, step := range recipe.Steps {
			text, _ := step["text"].(string)
			md.WriteString(fmt.Sprintf("%d. %s\n", i+1, text))
			if image, _ := step["image"].(string); image != "" {
				md.WriteString(fmt.Sprintf("\n   ![步骤 %d](%s)\n", i+1, resolveAbsolute(image)))
			}
		}
		md.WriteString("\n")
	}

	if strings.TrimSpace(md.String()) == "" {
		md.WriteString(recipe.Summary)
	}

	items := make([]any, 0, len(recipe.Ingredients))
	for _, ing := range recipe.Ingredients {
		items = append(items, ing)
	}

	sourceURL, srcWarnings := pushSourceURL(TargetTypeRecipe, recipe.ID, "")
	warnings = append(warnings, srcWarnings...)

	cover := model.ResolveMediaRefURL(ctx, refStr(recipe.CoverMediaID), refStr(recipe.CoverPresetID), recipe.Cover)
	return &CommunityPushPayload{
		Payload: &novapush.UpsertArticlePayload{
			ExternalID:      recipe.ID,
			Type:            TargetTypeRecipe,
			Title:           truncateRunes(recipe.Title, pushTitleMax),
			Summary:         ensureSummary(recipe.Summary, md.String(), &warnings),
			ContentMarkdown: md.String(),
			CoverImage:      resolveAbsolute(cover),
			Tags:            normalizePushTags(splitCommaTags(recipe.Tags), &warnings),
			SourceURL:       sourceURL,
			PublishedAt:     rfc3339(recipe.CreatedAt),
			Status:          binaryStatusToPush(recipe.Status, 1),
			Extra: map[string]any{
				"ingredients": items,
				"steps":       stepsOrEmpty(recipe.Steps),
				"difficulty":  recipe.Difficulty,
				"minutes":     recipe.Minutes,
				"servings":    recipe.Servings,
			},
		},
		Warnings: warnings,
	}, nil
}

// —— portfolio：图集合成 + extra 直传 ——

func buildPortfolioPushPayload(ctx context.Context, targetID string) (*CommunityPushPayload, error) {
	portfolioModel := model.NewPortfolio()
	portfolio, err := portfolioModel.GetByID(ctx, targetID)
	if err != nil {
		return nil, fmt.Errorf("作品集不存在: %w", err)
	}
	items, err := model.NewPortfolioItem().GetByPortfolioID(ctx, targetID)
	if err != nil {
		return nil, fmt.Errorf("读取作品项失败: %w", err)
	}

	warnings := make([]string, 0)
	covers, err := portfolioModel.ResolveCoverURLs(ctx, []model.Portfolio{*portfolio})
	var cover string
	if err == nil {
		cover = covers[portfolio.ID]
	}

	var md strings.Builder
	md.WriteString("## 作品一览\n\n")
	extraItems := make([]map[string]any, 0, len(items))
	for i, item := range items {
		title := item.Title
		if title == "" {
			title = fmt.Sprintf("作品 %d", i+1)
		}
		image := resolveAbsolute(item.OutputURL)
		md.WriteString(fmt.Sprintf("### %s\n\n", title))
		if item.Description != "" {
			md.WriteString(item.Description + "\n\n")
		}
		if image != "" {
			md.WriteString(fmt.Sprintf("![%s](%s)\n\n", title, image))
		}
		extraItems = append(extraItems, map[string]any{
			"title":       item.Title,
			"description": item.Description,
			"image":       image,
		})
	}

	sourceURL, srcWarnings := pushSourceURL(TargetTypePortfolio, portfolio.ID, "")
	warnings = append(warnings, srcWarnings...)

	return &CommunityPushPayload{
		Payload: &novapush.UpsertArticlePayload{
			ExternalID:      portfolio.ID,
			Type:            TargetTypePortfolio,
			Title:           truncateRunes(portfolio.Name, pushTitleMax),
			Summary:         ensureSummary(firstNonEmpty(portfolio.Description, portfolio.Name), md.String(), &warnings),
			ContentMarkdown: md.String(),
			CoverImage:      resolveAbsolute(cover),
			SourceURL:       sourceURL,
			PublishedAt:     rfc3339(portfolio.CreatedAt),
			Status:          binaryStatusToPush(portfolio.Status, 1),
			Extra:           map[string]any{"items": extraItems},
		},
		Warnings: warnings,
	}, nil
}

// —— open_source：Readme 直用 + 仓库信息卡 extra ——

func buildOpenSourcePushPayload(ctx context.Context, targetID string) (*CommunityPushPayload, error) {
	work, err := model.NewOpenSourceWork().GetByID(ctx, targetID)
	if err != nil {
		return nil, fmt.Errorf("开源作品不存在: %w", err)
	}

	warnings := make([]string, 0)
	var md strings.Builder
	infoParts := make([]string, 0, 2)
	if work.Language != "" {
		infoParts = append(infoParts, "主语言 "+work.Language)
	}
	if work.Stars > 0 {
		infoParts = append(infoParts, fmt.Sprintf("%d Stars", work.Stars))
	}
	if len(infoParts) > 0 {
		md.WriteString("> " + strings.Join(infoParts, " · "))
		if work.Homepage != "" {
			md.WriteString(" · [项目主页](" + work.Homepage + ")")
		}
		md.WriteString("\n\n")
	}
	md.WriteString(rewriteContentImages(work.Readme))

	sourceURL, srcWarnings := pushSourceURL(TargetTypeOpenSource, work.ID, "")
	// 开源作品的原文链接语义优先指向仓库本身
	if work.RepoURL != "" {
		sourceURL = work.RepoURL
	}
	warnings = append(warnings, srcWarnings...)

	return &CommunityPushPayload{
		Payload: &novapush.UpsertArticlePayload{
			ExternalID:      work.ID,
			Type:            "opensource", // 社区字典 slug（CMS 内部为 open_source）
			Title:           truncateRunes(work.Name, pushTitleMax),
			Summary:         ensureSummary(work.Summary, work.Readme, &warnings),
			ContentMarkdown: md.String(),
			Tags:            normalizePushTags(splitCommaTags(work.Topics), &warnings),
			SourceURL:       sourceURL,
			PublishedAt:     rfc3339(work.CreatedAt),
			Status:          binaryStatusToPush(work.Status, 1),
			Extra: map[string]any{
				"repo_url": work.RepoURL,
				"language": work.Language,
				"topics":   splitCommaTags(work.Topics),
				"stars":    work.Stars,
				"homepage": work.Homepage,
			},
		},
		Warnings: warnings,
	}, nil
}

// —— 小工具 ——

// binaryStatusToPush 两态内容（0=草稿,1=已发布）→ 社区推送状态。
func binaryStatusToPush(status, publishedValue int) string {
	if status == publishedValue {
		return "published"
	}
	return "hidden"
}

func splitCommaTags(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	return strings.Split(raw, ",")
}

func intOrFallback(value float64, fallback int) int {
	if value > 0 {
		return int(value)
	}
	return fallback
}

func attractionsOrEmpty(items []map[string]any) []map[string]any {
	if items == nil {
		return []map[string]any{}
	}
	return items
}

func itineraryOrEmpty(items []map[string]any) []map[string]any {
	if items == nil {
		return []map[string]any{}
	}
	return items
}

func reviewsOrEmpty(items []map[string]any) []map[string]any {
	if items == nil {
		return []map[string]any{}
	}
	return items
}

func stepsOrEmpty(items []map[string]any) []map[string]any {
	if items == nil {
		return []map[string]any{}
	}
	return items
}

// firstNonEmpty 返回第一个非空字符串。
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
