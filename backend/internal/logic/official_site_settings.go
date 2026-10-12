package logic

import "strings"

// officialSiteBaseURL NovaBlog 官方站点地址（社区推送与主题市场共用的统一出口）。
// 权威来源为 DB 单行表 community_push_configs.BaseURL（后台「社区推送配置」页维护）；
// config.yaml 的 community.base_url 与旧版主题市场配置仅作启动兜底，运行期以本值为准。
var officialSiteBaseURL string

// SetOfficialSiteBaseURL 设置官方站点地址运行时值（bootstrap 兜底链解析后注入 / 后台修改后热更新）。
func SetOfficialSiteBaseURL(base string) {
	officialSiteBaseURL = strings.TrimSpace(base)
}

// OfficialSiteBaseURL 读取官方站点地址（空=尚未配置）。
func OfficialSiteBaseURL() string {
	return officialSiteBaseURL
}
