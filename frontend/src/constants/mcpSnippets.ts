/**
 * MCP 客户端配置 JSON 生成器。
 * 输出标准的 mcpServers 结构，支持 MCP 协议的 AI 客户端均可直接粘贴使用。
 * 密钥参数为空时使用占位符（历史密钥明文不可回显）。
 */

/** MCP 服务的端点路径（与后端路由保持一致） */
export const MCP_ENDPOINT_PATH = '/api/v1/mcp'

/** MCP 端点完整地址（管理后台与 API 同源部署） */
export function getMcpEndpoint(): string {
  return `${window.location.origin}${MCP_ENDPOINT_PATH}`
}

/** 页面常驻展示用的密钥占位符 */
export const MCP_KEY_PLACEHOLDER = 'nbt_mcp_你的APIKey'

/** 生成可直接粘贴使用的 mcpServers 配置 JSON */
export function buildMcpJsonConfig(endpoint: string, key: string): string {
  return JSON.stringify(
    {
      mcpServers: {
        novablog: {
          type: 'http',
          url: endpoint,
          headers: { Authorization: `Bearer ${key}` },
        },
      },
    },
    null,
    2,
  )
}
