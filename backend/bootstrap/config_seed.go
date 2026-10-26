// 运行时配置启动种子：把存量 config.yaml 的运行时配置键幂等导入 system_configs 表。
// 数据库为运行时配置唯一主源——仅当对应键在 DB 中不存在时写入（存量升级一次性生效），
// 此后 config.yaml 中的旧键不再被读取（config.example.yaml 已移除这些键，
// deploy.sh 升级对比时会提示「废弃」并原样保留）。全新安装时 deploy.sh 生成的
// 首装默认值（upload.base_url 等）同样经此播种进数据库。
package bootstrap

import (
	"context"
	"strconv"
	"strings"
	"time"

	"novablog/internal/model"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// seedSystemConfigs 执行启动种子（幂等，失败仅记日志不阻塞启动——消费方有内置默认兜底）。
func seedSystemConfigs(cfg *viper.Viper, logger *zap.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// firstNonEmpty 依次取第一个非空配置值（多个旧代理键归一为单一通用代理）
	firstNonEmpty := func(keys ...string) string {
		for _, k := range keys {
			if v := strings.TrimSpace(cfg.GetString(k)); v != "" {
				return v
			}
		}
		return ""
	}
	intOrEmpty := func(key string, current int) string {
		if current <= 0 {
			return ""
		}
		return strconv.Itoa(current)
	}

	items := []model.SystemConfig{
		{Key: model.SystemConfigKeySiteURL, Value: firstNonEmpty("upload.base_url"),
			Description: "站点对外地址（媒体文件 URL 基址；修改后请检查跨域白名单是否包含该地址）"},
		{Key: model.SystemConfigKeyProxyURL, Value: firstNonEmpty("themes.proxy_url", "image_search.proxy_url", "clouddrive.proxy_url", "geoip.proxy_url"),
			Description: "出网 HTTP 代理（主题下载、图片搜索、云盘、IP 库更新共用；留空=直连）"},
		{Key: model.SystemConfigKeyGithubToken, Value: firstNonEmpty("themes.github_token"),
			Description: "GitHub 访问令牌（可选，提升主题与开源作品的接口限流额度）"},
		{Key: model.SystemConfigKeyMaxArtifactMB, Value: intOrEmpty("themes.max_artifact_mb", cfg.GetInt("themes.max_artifact_mb")),
			Description: "主题制品下载上限（MB）"},
		{Key: model.SystemConfigKeyCloudMaxFileSizeMB, Value: intOrEmpty("clouddrive.max_file_size_mb", cfg.GetInt("clouddrive.max_file_size_mb")),
			Description: "云盘导入单文件上限（MB）"},
		{Key: model.SystemConfigKeyCloudImportBatchLimit, Value: intOrEmpty("clouddrive.import_batch_limit", cfg.GetInt("clouddrive.import_batch_limit")),
			Description: "云盘导入单批文件数上限"},
		{Key: model.SystemConfigKeyGeoipDownloadURL, Value: firstNonEmpty("geoip.download_url"),
			Description: "IP 离线库下载地址（留空=内置默认源）"},
	}

	// 过滤空值（无旧值不播种，消费方走内置默认）
	filtered := make([]model.SystemConfig, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.Value) != "" {
			filtered = append(filtered, item)
		}
	}
	if len(filtered) == 0 {
		return
	}

	if err := model.NewSystemConfig().UpsertBatch(ctx, filtered, true); err != nil {
		logger.Warn("运行时配置种子导入失败（不影响启动，下次启动重试）", zap.Error(err))
		return
	}
	logger.Info("运行时配置种子导入完成（存量 config.yaml 键已收敛到数据库）")
}
