<template>
  <a-drawer :open="store.detailVisible" title="主题详情" width="460" @close="store.closeDetail()">
    <a-spin v-if="store.detailLoading" class="detail-spin" />

    <a-result
      v-else-if="store.detailError"
      status="warning"
      title="加载失败"
      sub-title="获取主题详情时出错，请重试"
    >
      <template #extra>
        <a-button type="primary" @click="retry">重试</a-button>
      </template>
    </a-result>

    <div v-else-if="detail" class="detail">
      <div class="detail__cover">
        <img
          v-if="coverURL && !coverFailed"
          :src="detail.preview"
          :alt="detail.title"
          @error="coverFailed = true"
        />
        <div v-else class="detail__placeholder" :style="{ background: themeGradient(detail.type) }">
          <span>{{ detail.title.slice(0, 1) }}</span>
        </div>
      </div>

      <h2 class="detail__title">{{ detail.title }}</h2>
      <div class="detail__tags">
        <a-tag color="blue">{{ typeLabel }}</a-tag>
        <a-tag v-for="s in detail.styles" :key="s">{{ s }}</a-tag>
        <a-tag v-if="detail.status !== 2" color="warning">{{ themeStatusText(detail.status) }}</a-tag>
      </div>

      <p class="detail__desc">{{ detail.description || '暂无描述' }}</p>

      <template v-if="detail.features?.length">
        <h4 class="detail__section">功能特性</h4>
        <div class="detail__tags">
          <a-tag v-for="f in detail.features" :key="f" color="geekblue">
            <CheckCircleOutlined /> {{ f }}
          </a-tag>
        </div>
      </template>

      <div class="detail__info">
        <div class="detail__info-item"><span>作者</span><b>{{ detail.author }}</b></div>
        <div class="detail__info-item"><span>版本</span><b>{{ detail.version || '—' }}</b></div>
        <div class="detail__info-item"><span>价格</span><b>{{ isPaid ? `¥${detail.price_amount}` : '免费' }}</b></div>
        <div class="detail__info-item"><span>下载量</span><b>{{ detail.downloads }}</b></div>
        <div class="detail__info-item"><span>点赞</span><b>{{ detail.likes }}</b></div>
        <div class="detail__info-item">
          <span>综合评分</span><b>{{ detail.rating > 0 ? detail.rating.toFixed(1) : '暂无' }}</b>
        </div>
        <div class="detail__info-item"><span>发布时间</span><b>{{ detail.created_at }}</b></div>
        <div class="detail__info-item"><span>slug</span><b>{{ detail.slug }}</b></div>
      </div>

      <div class="detail__rating">
        <span class="detail__rating-label">我的评分：</span>
        <a-rate :value="myRating" :disabled="ratingLoading" @change="onRate" />
        <span class="detail__rating-hint">
          {{ myRating > 0 ? `${myRating} 分` : '点击评分（1-5 分，可覆盖）' }}
        </span>
      </div>

      <div class="detail__actions">
        <a-button
          :type="detail.liked ? 'primary' : 'default'"
          :loading="likeLoading"
          @click="handleLike"
        >
          <template #icon><HeartFilled v-if="detail.liked" /><HeartOutlined v-else /></template>
          {{ detail.liked ? '已点赞' : '点赞' }}
        </a-button>
        <a-button
          :type="favorited ? 'primary' : 'default'"
          :loading="favLoading"
          @click="handleFavorite"
        >
          <template #icon><StarFilled v-if="favorited" /><StarOutlined v-else /></template>
          {{ favorited ? '已收藏' : '收藏' }}
        </a-button>
        <a-button type="primary" :loading="downloadLoading" @click="handleDownload">
          <template #icon><DownloadOutlined /></template>
          下载主题
        </a-button>
      </div>
    </div>
  </a-drawer>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  CheckCircleOutlined,
  DownloadOutlined,
  HeartFilled,
  HeartOutlined,
  StarFilled,
  StarOutlined,
} from '@ant-design/icons-vue'
import { useThemeMarketStore } from '@/stores/themeMarket'
import { isPreviewURL, themeGradient, themeStatusText, THEME_TYPE_LABELS } from '@/utils/themeDisplay'

const store = useThemeMarketStore()

const coverFailed = ref(false)
const likeLoading = ref(false)
const favLoading = ref(false)
const downloadLoading = ref(false)
const ratingLoading = ref(false)
const myRating = ref(0)

const detail = computed(() => store.currentDetail)
const coverURL = computed(() => isPreviewURL(detail.value?.preview || ''))
const typeLabel = computed(() => (detail.value ? THEME_TYPE_LABELS[detail.value.type] || detail.value.type : ''))
const isPaid = computed(() => detail.value?.price === 'paid')
const favorited = computed(() => (detail.value ? store.favoriteIds.has(detail.value.id) : false))

watch(
  () => store.detailVisible,
  (open) => {
    if (open) coverFailed.value = false
  },
)

// 详情加载完成后同步「我的评分」
watch(
  () => store.currentDetail,
  (d) => {
    if (d) myRating.value = d.user_rating || 0
  },
)

function retry() {
  if (detail.value) store.fetchDetail(detail.value.id)
}

async function handleLike() {
  if (!detail.value) return
  likeLoading.value = true
  try {
    await store.toggleLike(detail.value.id)
  } finally {
    likeLoading.value = false
  }
}

async function handleFavorite() {
  if (!detail.value) return
  favLoading.value = true
  try {
    await store.toggleFavorite(detail.value.id)
  } finally {
    favLoading.value = false
  }
}

async function onRate(score: number) {
  if (!detail.value || !score) return
  ratingLoading.value = true
  try {
    myRating.value = score
    await store.rate(detail.value.id, score)
  } catch {
    // 提交失败回滚到服务端已有评分
    myRating.value = detail.value.user_rating || 0
  } finally {
    ratingLoading.value = false
  }
}

async function handleDownload() {
  if (!detail.value) return
  downloadLoading.value = true
  try {
    await store.download(detail.value.id)
  } finally {
    downloadLoading.value = false
  }
}
</script>

<style scoped>
.detail-spin {
  display: flex;
  justify-content: center;
  padding: 60px 0;
}

.detail__cover {
  height: 200px;
  margin-bottom: 16px;
  border-radius: var(--border-radius-lg, 12px);
  overflow: hidden;
}

.detail__cover img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.detail__placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
}

.detail__placeholder span {
  color: rgba(255, 255, 255, 0.92);
  font-size: 56px;
  font-weight: 600;
  user-select: none;
}

.detail__title {
  margin: 0 0 8px;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color, rgba(0, 0, 0, 0.88));
}

.detail__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 0;
  margin-bottom: 12px;
}

.detail__desc {
  margin: 0 0 16px;
  line-height: 1.7;
  color: var(--text-color-secondary, rgba(0, 0, 0, 0.65));
  white-space: pre-wrap;
  word-break: break-word;
}

.detail__section {
  margin: 0 0 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-color, rgba(0, 0, 0, 0.88));
}

.detail__info {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px 16px;
  margin: 16px 0;
  padding: 12px 14px;
  background: var(--bg-layout, rgba(0, 0, 0, 0.02));
  border-radius: var(--border-radius, 8px);
}

.detail__info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.detail__info-item span {
  color: var(--text-color-tertiary, rgba(0, 0, 0, 0.45));
  flex-shrink: 0;
}

.detail__info-item b {
  font-weight: 500;
  color: var(--text-color, rgba(0, 0, 0, 0.88));
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.detail__rating {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  padding: 10px 14px;
  border: 1px dashed var(--border-color, #f0f0f0);
  border-radius: var(--border-radius, 8px);
}

.detail__rating-label {
  font-size: 13px;
  color: var(--text-color, rgba(0, 0, 0, 0.88));
}

.detail__rating-hint {
  font-size: 12px;
  color: var(--text-color-tertiary, rgba(0, 0, 0, 0.45));
}

.detail__actions {
  display: flex;
  gap: 10px;
}

.detail__actions .ant-btn {
  flex: 1;
}
</style>
