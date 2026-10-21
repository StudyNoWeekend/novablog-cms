<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-header__left">
        <a-button type="text" @click="goBack">
          <template #icon><ArrowLeftOutlined /></template>
          {{ $t('common.actions.back') }}
        </a-button>
        <h1 class="page-title">{{ $t('menu.theme.detail') }}</h1>
      </div>
    </div>

    <!-- 加载骨架屏 -->
    <div v-if="store.detailLoading" class="detail-skeleton">
      <div class="detail-skeleton__cover"><a-skeleton active :title="false" :paragraph="{ rows: 3 }" /></div>
      <a-skeleton active :paragraph="{ rows: 6 }" />
    </div>

    <!-- 加载失败 -->
    <a-result
      v-else-if="store.detailError"
      status="warning"
      :title="$t('common.state.error')"
      :sub-title="$t('theme.error.detailLoad')"
    >
      <template #extra>
        <a-button type="primary" @click="loadAll">{{ $t('common.actions.retry') }}</a-button>
      </template>
    </a-result>

    <!-- 主题不存在 -->
    <a-empty v-else-if="!detail" :description="$t('theme.error.notFound')">
      <a-button type="primary" @click="goBack">{{ $t('theme.action.backToMarket') }}</a-button>
    </a-empty>

    <div v-else class="detail-layout">
      <!-- 主列：封面 / 标题 / 描述 / 特性 / 版本历史 -->
      <div class="detail-main">
        <div class="detail-card">
          <div class="detail-cover">
            <img
              v-if="coverURL && !coverFailed"
              :src="detail.preview"
              :alt="displayName"
              @error="coverFailed = true"
            />
            <div v-else class="detail-cover__placeholder" :style="{ background: themeGradient(detail.type) }">
              <span>{{ displayName.slice(0, 1) }}</span>
            </div>
          </div>

          <h2 class="detail-title">{{ displayName }}</h2>
          <div class="detail-tags">
            <a-tag color="blue">{{ typeLabel }}</a-tag>
            <a-tag v-for="s in detail.styles" :key="s">{{ s }}</a-tag>
            <a-tag v-if="detail.status !== 2" color="warning">{{ statusText(detail.status) }}</a-tag>
          </div>

          <p class="detail-desc">{{ detail.description || $t('theme.empty.noDescription') }}</p>

          <template v-if="detail.features?.length">
            <h4 class="detail-section-title">{{ $t('theme.section.features') }}</h4>
            <div class="detail-tags">
              <a-tag v-for="f in detail.features" :key="f" color="geekblue">
                <CheckCircleOutlined /> {{ f }}
              </a-tag>
            </div>
          </template>
        </div>

        <!-- 版本历史 -->
        <div class="detail-card">
          <h4 class="detail-section-title">{{ $t('theme.section.releases') }}</h4>

          <div v-if="store.releasesLoading" class="releases-loading">
            <a-skeleton active :paragraph="{ rows: 2 }" />
          </div>

          <div v-else-if="store.releasesError" class="releases-error">
            <span>{{ $t('theme.error.releasesLoad') }}</span>
            <a-button size="small" @click="retryReleases">{{ $t('common.actions.retry') }}</a-button>
          </div>

          <a-empty
            v-else-if="store.releases.length === 0"
            :image="Empty.PRESENTED_IMAGE_SIMPLE"
            :description="$t('theme.empty.noReleases')"
          />

          <div v-else class="release-list">
            <div v-for="r in store.releases" :key="`${r.version}-${r.tag}`" class="release-item">
              <div class="release-item__head">
                <a-tag color="processing">v{{ r.version }}</a-tag>
                <span class="release-item__tag" :title="r.tag">{{ r.tag }}</span>
                <span class="release-item__time">{{ r.published_at }}</span>
              </div>
              <div class="release-item__notes">{{ r.notes || $t('theme.empty.noNotes') }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 侧列：信息 / 评分 / 操作 -->
      <div class="detail-side">
        <div class="detail-card">
          <div class="detail-info">
            <div class="detail-info__item"><span>{{ $t('theme.info.author') }}</span><b>{{ detail.author }}</b></div>
            <div class="detail-info__item"><span>{{ $t('theme.info.latestVersion') }}</span><b>{{ detail.version || '—' }}</b></div>
            <div class="detail-info__item">
              <span>{{ $t('theme.info.price') }}</span><b>{{ isPaid ? `¥${detail.price_amount}` : $t('theme.price.free') }}</b>
            </div>
            <div class="detail-info__item"><span>{{ $t('theme.info.downloads') }}</span><b>{{ detail.downloads }}</b></div>
            <div class="detail-info__item"><span>{{ $t('theme.info.likes') }}</span><b>{{ detail.likes }}</b></div>
            <div class="detail-info__item">
              <span>{{ $t('theme.info.rating') }}</span><b>{{ detail.rating > 0 ? detail.rating.toFixed(1) : $t('theme.rate.none') }}</b>
            </div>
            <div class="detail-info__item"><span>{{ $t('theme.info.publishedAt') }}</span><b>{{ detail.created_at }}</b></div>
            <div class="detail-info__item"><span>slug</span><b>{{ detail.slug }}</b></div>
          </div>
        </div>

        <div class="detail-card">
          <h4 class="detail-section-title">{{ $t('theme.rate.myRating') }}</h4>
          <div class="detail-rating">
            <a-rate :value="myRating" :disabled="ratingLoading" @change="onRate" />
            <span class="detail-rating__hint">
              {{ myRating > 0 ? $t('theme.rate.score', { score: myRating }) : $t('theme.rate.hint') }}
            </span>
          </div>
        </div>

        <div class="detail-card">
          <div class="detail-actions">
            <a-button
              :type="detail.liked ? 'primary' : 'default'"
              :loading="likeLoading"
              @click="handleLike"
            >
              <template #icon><HeartFilled v-if="detail.liked" /><HeartOutlined v-else /></template>
              {{ detail.liked ? $t('theme.action.liked') : $t('theme.action.like') }}
            </a-button>
            <a-button
              :type="favorited ? 'primary' : 'default'"
              :loading="favLoading"
              @click="handleFavorite"
            >
              <template #icon><StarFilled v-if="favorited" /><StarOutlined v-else /></template>
              {{ favorited ? $t('theme.action.favorited') : $t('theme.action.favorite') }}
            </a-button>
          </div>
          <a-button type="primary" block :loading="downloadLoading" @click="handleDownload">
            <template #icon><DownloadOutlined /></template>
            {{ $t('theme.action.download') }}
          </a-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Empty } from 'ant-design-vue'
import {
  ArrowLeftOutlined,
  CheckCircleOutlined,
  DownloadOutlined,
  HeartFilled,
  HeartOutlined,
  StarFilled,
  StarOutlined,
} from '@ant-design/icons-vue'
import { MARKET_AUTH_EXPIRED_EVENT } from '@/api/theme'
import { useThemeMarketStore } from '@/stores/themeMarket'
import { isPreviewURL, themeDisplayName, themeGradient } from '@/utils/themeDisplay'

const route = useRoute()
const router = useRouter()
const store = useThemeMarketStore()
const { t } = useI18n()

// typeLabels 主题类型展示文案（i18n），未知类型回退原始值
const typeLabels = computed<Record<string, string>>(() => ({
  personal: t('theme.type.personal'),
  tech: t('theme.type.tech'),
  photo: t('theme.type.photo'),
  travel: t('theme.type.travel'),
  music: t('theme.type.music'),
  team: t('theme.type.team'),
}))
// statusTextMap 主题状态文案（1 审核中 / 2 已上架 / 3 已下架）
const statusTextMap = computed<Record<number, string>>(() => ({
  1: t('theme.status.reviewing'),
  2: t('theme.status.listed'),
  3: t('theme.status.offline'),
}))
function statusText(status: number): string {
  return statusTextMap.value[status] || t('common.status.unknown')
}

const coverFailed = ref(false)
const likeLoading = ref(false)
const favLoading = ref(false)
const downloadLoading = ref(false)
const ratingLoading = ref(false)
const myRating = ref(0)

const themeId = Number(route.params.id)
const detail = computed(() => store.currentDetail)
const coverURL = computed(() => isPreviewURL(detail.value?.preview || ''))
const displayName = computed(() =>
  detail.value ? themeDisplayName(detail.value.title, detail.value.slug) : '',
)
const typeLabel = computed(() => (detail.value ? typeLabels.value[detail.value.type] || detail.value.type : ''))
const isPaid = computed(() => detail.value?.price === 'paid')
const favorited = computed(() => (detail.value ? store.favoriteIds.has(detail.value.id) : false))

function goBack() {
  router.push('/themes')
}

/** 并行拉取详情与版本历史 */
async function loadAll() {
  if (!Number.isFinite(themeId)) return
  coverFailed.value = false
  await Promise.all([store.fetchDetail(themeId), store.fetchReleases(themeId)])
  // 直接刷新进入详情页时收藏集合可能为空，补拉一次以点亮收藏状态
  if (store.favoriteIds.size === 0) store.fetchFavoriteIds()
}

function retryReleases() {
  if (Number.isFinite(themeId)) store.fetchReleases(themeId)
}

function handleAuthExpired() {
  // 官方登录失效后详情接口不可用，退回主题页弹出登录遮罩
  router.replace('/themes')
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

onMounted(() => {
  window.addEventListener(MARKET_AUTH_EXPIRED_EVENT, handleAuthExpired)
  // 未连接官方市场时退回主题页（登录遮罩在该页展示）
  if (!store.loggedIn) {
    router.replace('/themes')
    return
  }
  loadAll()
})

onUnmounted(() => {
  window.removeEventListener(MARKET_AUTH_EXPIRED_EVENT, handleAuthExpired)
})
</script>

<style scoped>
.page-header__left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-header__left .page-title {
  margin: 0;
}

/* ===== 骨架屏 ===== */
.detail-skeleton__cover {
  height: 220px;
  margin-bottom: 20px;
  padding: 24px;
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #E5E9F2);
  border-radius: var(--border-radius-lg, 12px);
}

/* ===== 双列布局 ===== */
.detail-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  gap: 20px;
  align-items: start;
}

.detail-main {
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-width: 0;
}

.detail-side {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.detail-card {
  padding: 20px;
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #E5E9F2);
  border-radius: var(--border-radius-lg, 12px);
}

/* ===== 封面与标题 ===== */
.detail-cover {
  height: 260px;
  margin-bottom: 16px;
  border-radius: var(--border-radius, 8px);
  overflow: hidden;
}

.detail-cover img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.detail-cover__placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
}

.detail-cover__placeholder span {
  color: rgba(255, 255, 255, 0.92);
  font-size: 64px;
  font-weight: 600;
  user-select: none;
}

.detail-title {
  margin: 0 0 8px;
  font-size: 22px;
  font-weight: 600;
  color: var(--text-color, rgba(0, 0, 0, 0.88));
}

.detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 0;
  margin-bottom: 12px;
}

.detail-desc {
  margin: 0;
  line-height: 1.8;
  color: var(--text-color-secondary, rgba(0, 0, 0, 0.65));
  white-space: pre-wrap;
  word-break: break-word;
}

.detail-section-title {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color, rgba(0, 0, 0, 0.88));
}

.detail-card .detail-section-title:not(:first-child) {
  margin-top: 20px;
}

/* ===== 版本历史 ===== */
.releases-loading {
  padding: 4px 0;
}

.releases-error {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 13px;
  color: var(--text-color-secondary, rgba(0, 0, 0, 0.65));
}

.release-list {
  display: flex;
  flex-direction: column;
}

.release-item {
  padding: 14px 0;
  border-bottom: 1px solid var(--border-color, #E5E9F2);
}

.release-item:last-child {
  padding-bottom: 0;
  border-bottom: none;
}

.release-item__head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.release-item__tag {
  flex: 1;
  min-width: 0;
  font-size: 12px;
  color: var(--text-color-tertiary, rgba(0, 0, 0, 0.45));
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.release-item__time {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--text-color-tertiary, rgba(0, 0, 0, 0.45));
}

.release-item__notes {
  font-size: 13px;
  line-height: 1.7;
  color: var(--text-color-secondary, rgba(0, 0, 0, 0.65));
  white-space: pre-wrap;
  word-break: break-word;
}

/* ===== 信息网格 ===== */
.detail-info {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.detail-info__item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.detail-info__item span {
  color: var(--text-color-tertiary, rgba(0, 0, 0, 0.45));
  flex-shrink: 0;
}

.detail-info__item b {
  font-weight: 500;
  color: var(--text-color, rgba(0, 0, 0, 0.88));
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ===== 评分与操作 ===== */
.detail-rating {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-rating__hint {
  font-size: 12px;
  color: var(--text-color-tertiary, rgba(0, 0, 0, 0.45));
}

.detail-actions {
  display: flex;
  gap: 10px;
  margin-bottom: 12px;
}

.detail-actions .ant-btn {
  flex: 1;
}

/* ===== 响应式 ===== */
@media (max-width: 900px) {
  .detail-layout {
    grid-template-columns: 1fr;
  }
}
</style>
