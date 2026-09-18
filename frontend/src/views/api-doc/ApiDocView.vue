<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">API 文档</h1>
    </div>

    <!-- 筛选与搜索 -->
    <div class="filter-bar">
      <a-tabs v-model:activeKey="activeModule" class="module-tabs">
        <a-tab-pane v-for="tab in moduleTabs" :key="tab.value" :tab="tab.label" />
      </a-tabs>
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Empty, type TableColumnsType } from 'ant-design-vue'
import { getAPIDocsAPI } from '@/api/apiDoc'
import type { APIDocItem, APIDocParam, APIDocField } from '@/types/security'

const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE

const loading = ref(false)
const docs = ref<APIDocItem[]>([])
const activeModule = ref('all')
const searchKeyword = ref('')

const moduleTabs = [
  { label: '全部', value: 'all' },
  { label: '安装引导', value: '安装引导' },
  { label: '博主信息', value: '博主信息' },
  { label: '文章', value: '文章' },
  { label: '评论', value: '评论' },
  { label: '旅行攻略', value: '旅行攻略' },
  { label: '摄影作品集', value: '摄影作品集' },
  { label: '摄影器材', value: '摄影器材' },
  { label: '视频作品', value: '视频作品' },
  { label: '音乐', value: '音乐' },
  { label: '系统', value: '系统' },
]

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

async function fetchDocs() {
  loading.value = true
  try {
    const data = await getAPIDocsAPI()
    docs.value = data || []
  } catch {
    docs.value = []
  } finally {
    loading.value = false
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
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}

.module-tabs {
  flex: 1;
  min-width: 300px;
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
