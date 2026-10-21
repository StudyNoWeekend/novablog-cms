<template>
  <div class="config-json">
    <div class="config-meta">
      <span class="config-label">{{ $t('mcp.config.label') }}</span>
      <a-button type="link" size="small" class="copy-btn" @click="handleCopy">
        <template #icon><CopyOutlined /></template>
        {{ $t('mcp.config.copy') }}
      </a-button>
    </div>
    <pre class="config-code"><code>{{ json }}</code></pre>
    <p class="config-hint">{{ $t('mcp.config.hint') }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { message } from 'ant-design-vue'
import { CopyOutlined } from '@ant-design/icons-vue'
import { copyText } from '@/utils/clipboard'
import { buildMcpJsonConfig, MCP_KEY_PLACEHOLDER } from '@/constants/mcpSnippets'

const { t } = useI18n()

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
    message.success(t('mcp.message.configCopied'))
  } else {
    message.error(t('mcp.message.copyFailedSelect'))
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
