<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">API 文档</h1>
      <a-button type="primary" :icon="h(DownloadOutlined)" :loading="downloading" @click="downloadDoc">
        下载文档
      </a-button>
    </div>

    <!-- 筛选与搜索 -->
    <div class="doc-layout">
      <!-- 左侧模块导航 -->
      <aside class="doc-sider">
        <div class="sider-title">模块</div>
        <ul class="module-list">
          <li
            v-for="tab in moduleTabs"
            :key="tab.value"
            class="module-item"
            :class="{ active: activeModule === tab.value }"
            @click="activeModule = tab.value"
          >
            {{ tab.label }}
          </li>
        </ul>
      </aside>

      <!-- 右侧内容区 -->
      <main class="doc-main">
        <div class="filter-bar">
          <a-input-search
            v-model:value="searchKeyword"
            placeholder="搜索路径或描述..."
            style="width: 280px"
            allow-clear
          />
        </div>

        <!-- 文档列表 -->
        <a-spin :spinning="loading">
          <div v-if="filteredDocs.length > 0" class="doc-list">
            <a-card
              v-for="(item, idx) in filteredDocs"
              :key="idx"
              class="doc-card"
              :bordered="true"
            >
              <!-- 接口标题行 -->
              <div class="doc-header">
                <a-tag :color="methodColor(item.method)" class="method-tag">
                  {{ item.method.toUpperCase() }}
                </a-tag>
                <span class="doc-path">/api/v1/public{{ item.path }}</span>
                <a-tooltip :title="'所属模块：' + item.module">
                  <span class="doc-module">{{ item.module }}</span>
                </a-tooltip>
              </div>
              <p class="doc-desc">{{ item.description }}</p>

              <!-- 请求参数表 -->
              <div class="doc-section">
                <h4 class="section-title">请求参数</h4>
                <a-table
                  v-if="item.params && item.params.length > 0"
                  :columns="paramColumns"
                  :data-source="item.params"
                  :pagination="false"
                  row-key="name"
                  size="small"
                  :scroll="{ x: 500 }"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'required'">
                      <a-tag :color="record.required ? 'red' : 'default'">
                        {{ record.required ? '是' : '否' }}
                      </a-tag>
                    </template>
                    <template v-else>
                      {{ record[column.dataIndex] }}
                    </template>
                  </template>
                </a-table>
                <a-empty v-else description="无请求参数" :image="simpleImage" />
              </div>

              <!-- 响应结构表 -->
              <div class="doc-section">
                <h4 class="section-title">响应结构</h4>
                <a-table
                  v-if="item.response && item.response.length > 0"
                  :columns="responseColumns"
                  :data-source="item.response"
                  :pagination="false"
                  row-key="name"
                  size="small"
                  :scroll="{ x: 500 }"
                />
                <a-empty v-else description="无响应结构" :image="simpleImage" />
              </div>
            </a-card>
          </div>
          <a-empty v-else-if="!loading" description="未找到匹配的 API" />
        </a-spin>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import { Empty, message, type TableColumnsType } from 'ant-design-vue'
import { DownloadOutlined } from '@ant-design/icons-vue'
import { parseApiDocMarkdown, type APIDocItem, type APIDocParam, type APIDocField } from '@/utils/markdownApiDoc'

const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE

const loading = ref(false)
const downloading = ref(false)
const docs = ref<APIDocItem[]>([])
const activeModule = ref('all')
const searchKeyword = ref('')

// 模块 tab：全部 + 解析结果中出现的所有模块（去重，保持出现顺序）
const moduleTabs = computed(() => {
  const modules = Array.from(new Set(docs.value.map((d) => d.module).filter(Boolean)))
  return [
    { label: '全部', value: 'all' },
    ...modules.map((m) => ({ label: m, value: m })),
  ]
})

const paramColumns: TableColumnsType<APIDocParam> = [
  { title: '参数名', dataIndex: 'name', key: 'name', width: 180 },
  { title: '类型', dataIndex: 'type', key: 'type', width: 120 },
  { title: '必填', key: 'required', width: 80 },
  { title: '说明', dataIndex: 'desc', key: 'desc', ellipsis: true },
]

const responseColumns: TableColumnsType<APIDocField> = [
  { title: '字段名', dataIndex: 'name', key: 'name', width: 180 },
  { title: '类型', dataIndex: 'type', key: 'type', width: 120 },
  { title: '说明', dataIndex: 'desc', key: 'desc', ellipsis: true },
]

const filteredDocs = computed(() => {
  let result = docs.value
  if (activeModule.value !== 'all') {
    result = result.filter((d) => d.module === activeModule.value)
  }
  const kw = searchKeyword.value.trim().toLowerCase()
  if (kw) {
    result = result.filter(
      (d) =>
        d.path.toLowerCase().includes(kw) ||
        d.description.toLowerCase().includes(kw),
    )
  }
  return result
})

function methodColor(method: string): string {
  switch (method.toUpperCase()) {
    case 'GET':
      return 'green'
    case 'POST':
      return 'blue'
    case 'PUT':
      return 'orange'
    case 'DELETE':
      return 'red'
    default:
      return 'default'
  }
}

const docUrl = `${import.meta.env.BASE_URL}docs/novablog-api.md`

async function fetchDocs() {
  loading.value = true
  try {
    const res = await fetch(docUrl)
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const md = await res.text()
    docs.value = parseApiDocMarkdown(md)
  } catch (e) {
    console.error('加载 API 文档失败:', e)
    message.error('加载 API 文档失败')
    docs.value = []
  } finally {
    loading.value = false
  }
}

async function downloadDoc() {
  downloading.value = true
  try {
    const res = await fetch(docUrl)
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const blob = await res.blob()
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = 'novablog-api.md'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(a.href)
  } catch {
    message.error('下载文档失败')
  } finally {
    downloading.value = false
  }
}

onMounted(() => {
  fetchDocs()
})
</script>

<style scoped>
.page-container {
  padding: 0;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 24px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  margin: 0;
}

.filter-bar {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  margin-bottom: 24px;
}

.module-tabs {
  flex: 1;
  min-width: 300px;
}

.doc-layout {
  display: flex;
  gap: 24px;
  align-items: flex-start;
}

.doc-sider {
  width: 180px;
  flex-shrink: 0;
  position: sticky;
  top: 16px;
  max-height: calc(100vh - 120px);
  overflow-y: auto;
  background: var(--bg-subtle, #f8fafc);
  border-radius: var(--border-radius-lg, 12px);
  padding: 12px 0;
}

.sider-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary, #64748b);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 8px 16px 12px;
}

.module-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.module-item {
  padding: 8px 16px;
  font-size: 13px;
  color: var(--text-primary, #1e293b);
  cursor: pointer;
  transition: all 0.15s ease;
  border-left: 3px solid transparent;
  user-select: none;
}

.module-item:hover {
  background: var(--bg-hover, #e2e8f0);
}

.module-item.active {
  background: var(--bg-active, #dbeafe);
  color: var(--color-primary, #3b82f6);
  border-left-color: var(--color-primary, #3b82f6);
  font-weight: 600;
}

.doc-main {
  flex: 1;
  min-width: 0;
}

@media (max-width: 768px) {
  .doc-layout {
    flex-direction: column;
  }

  .doc-sider {
    width: 100%;
    position: static;
    max-height: none;
    overflow-x: auto;
    overflow-y: hidden;
    padding: 8px 0;
    border-radius: var(--border-radius-lg, 12px);
  }

  .sider-title {
    display: none;
  }

  .module-list {
    display: flex;
    gap: 4px;
    white-space: nowrap;
    padding: 4px 12px;
  }

  .module-item {
    border-left: none;
    border-radius: 999px;
    padding: 4px 14px;
    font-size: 13px;
  }

  .module-item.active {
    border-left: none;
  }
}

.doc-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.doc-card {
  border-radius: var(--border-radius-lg, 12px);
}

.doc-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.method-tag {
  font-weight: 600;
  font-size: 13px;
  min-width: 60px;
  text-align: center;
}

.doc-path {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 14px;
  color: var(--text-primary, #1e293b);
  font-weight: 500;
  word-break: break-all;
}

.doc-module {
  font-size: 12px;
  color: var(--text-secondary, #64748b);
  background: var(--bg-subtle, #f1f5f9);
  border-radius: 999px;
  padding: 2px 10px;
  cursor: default;
}

.doc-desc {
  font-size: 14px;
  color: var(--text-secondary, #64748b);
  margin: 0 0 16px 0;
  line-height: 1.6;
}

.doc-section {
  margin-top: 16px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  margin: 0 0 12px 0;
}

:deep(.ant-card-body) {
  padding: 20px 24px;
}

:deep(.ant-tabs-nav) {
  margin-bottom: 0;
}
</style>
