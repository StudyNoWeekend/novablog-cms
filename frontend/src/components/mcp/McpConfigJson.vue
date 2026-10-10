<template>
  <div class="config-json">
    <div class="config-meta">
      <span class="config-label">MCP 配置 JSON</span>
      <a-button type="link" size="small" class="copy-btn" @click="handleCopy">
        <template #icon><CopyOutlined /></template>
        复制配置
      </a-button>
    </div>
    <pre class="config-code"><code>{{ json }}</code></pre>
    <p class="config-hint">加入 mcpServers 配置并重启客户端后，即可让 AI 通过 MCP 工具发布内容</p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { message } from 'ant-design-vue'
import { CopyOutlined } from '@ant-design/icons-vue'
import { copyText } from '@/utils/clipboard'
import { buildMcpJsonConfig, MCP_KEY_PLACEHOLDER } from '@/constants/mcpSnippets'

const props = defineProps<{
  endpoint: string
  /** 密钥明文；为空时模板使用占位符 */
  keyValue: string
}>()

// 历史密钥明文不可回显，无真实密钥时用占位符渲染模板
const effectiveKey = computed(() => props.keyValue || MCP_KEY_PLACEHOLDER)
const json = computed(() => buildMcpJsonConfig(props.endpoint, effectiveKey.value))

async function handleCopy() {
  const ok = await copyText(json.value)
  if (ok) {
    message.success('MCP 配置已复制')
  } else {
    message.error('复制失败，请手动选择复制')
  }
}
</script>

<style scoped>
.config-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
}

.config-label {
  font-size: 12px;
  color: var(--text-secondary, #667085);
}

.copy-btn {
  padding: 0;
  flex-shrink: 0;
}

.config-code {
  margin: 0;
  padding: 14px 16px;
  background: #f6f8fb;
  border: 1px solid #e8ecf3;
  border-radius: 8px;
  overflow-x: auto;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12.5px;
  line-height: 1.7;
  color: #29365c;
  white-space: pre;
}

.config-hint {
  margin: 8px 0 0;
  font-size: 12px;
  color: var(--text-secondary, #667085);
}

@media (max-width: 768px) {
  .config-meta {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
}
</style>
