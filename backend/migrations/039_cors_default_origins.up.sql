-- CORS 默认白名单补充生产域名（https://www.novablog.cn / https://api.novablog.cn）

-- 1. 仅调整列默认值：影响全新安装的播种行，不影响已部署环境的存量数据
ALTER TABLE cors_configs
    ALTER COLUMN allowed_origins SET DEFAULT 'http://localhost:5173,http://localhost:5174,https://www.novablog.cn,https://api.novablog.cn';

-- 2. 仅回填从未被修改过的播种行（updated_at = created_at 为"未修改"标记，
--    与启动引导 InitFromFallback 的判定一致），用户自定义过的配置不受影响。
--    不显式改动 updated_at，保持其仍等于 created_at，维持未修改语义。
UPDATE cors_configs
SET allowed_origins = 'http://localhost:5173,http://localhost:5174,https://www.novablog.cn,https://api.novablog.cn'
WHERE updated_at = created_at;
