<template>
  <a-spin :spinning="loading">
    <a-form layout="vertical" :model="form">
      <!-- 形象区：头像 / 博客 Icon / 页面背景图 横向排列 -->
      <div class="identity-row">
        <a-form-item label="头像">
          <div class="image-field">
            <div
              class="image-uploader avatar-uploader"
              :class="{ 'is-empty': !form.avatar || brokenImages.avatar }"
              role="button"
              tabindex="0"
              :aria-label="form.avatar && !brokenImages.avatar ? '更换头像' : '上传头像'"
              @click="triggerUpload('avatar')"
              @keydown.enter.prevent="triggerUpload('avatar')"
              @keydown.space.prevent="triggerUpload('avatar')"
            >
              <template v-if="form.avatar && !brokenImages.avatar">
                <img :src="form.avatar" alt="头像" @error="onImgError('avatar')" />
                <div class="uploader-mask">更换图片</div>
              </template>
              <div v-else class="uploader-empty">
                <PlusOutlined />
                <span class="uploader-hint">上传图片</span>
              </div>
              <div v-if="uploadingAvatar" class="uploader-loading">
                <a-spin size="small" />
              </div>
            </div>
            <a-button
              v-if="form.avatar"
              type="text"
              danger
              size="small"
              @click="clearImage('avatar')"
            >
              移除
            </a-button>
          </div>
        </a-form-item>

        <a-form-item label="博客 Icon">
          <div class="image-field">
            <div
              class="image-uploader icon-uploader"
              :class="{ 'is-empty': !form.blog_icon || brokenImages.icon }"
              role="button"
              tabindex="0"
              :aria-label="form.blog_icon && !brokenImages.icon ? '更换博客 Icon' : '上传博客 Icon'"
              @click="triggerUpload('icon')"
              @keydown.enter.prevent="triggerUpload('icon')"
              @keydown.space.prevent="triggerUpload('icon')"
            >
              <template v-if="form.blog_icon && !brokenImages.icon">
                <img :src="form.blog_icon" alt="博客 Icon" @error="onImgError('icon')" />
                <div class="uploader-mask">更换图片</div>
              </template>
              <div v-else class="uploader-empty">
                <PlusOutlined />
                <span class="uploader-hint">上传图片</span>
              </div>
              <div v-if="uploadingIcon" class="uploader-loading">
                <a-spin size="small" />
              </div>
            </div>
            <a-button
              v-if="form.blog_icon"
              type="text"
              danger
              size="small"
              @click="clearImage('icon')"
            >
              移除
            </a-button>
          </div>
        </a-form-item>

        <a-form-item label="页面背景图">
          <div class="image-field">
            <div
              class="image-uploader background-uploader"
              :class="{ 'is-empty': !form.page_background || brokenImages.background }"
              role="button"
              tabindex="0"
              :aria-label="form.page_background && !brokenImages.background ? '更换背景图' : '上传背景图'"
              @click="triggerUpload('background')"
              @keydown.enter.prevent="triggerUpload('background')"
              @keydown.space.prevent="triggerUpload('background')"
            >
              <template v-if="form.page_background && !brokenImages.background">
                <img :src="form.page_background" alt="背景图" @error="onImgError('background')" />
                <div class="uploader-mask">更换图片</div>
              </template>
              <div v-else class="uploader-empty">
                <PlusOutlined />
                <span class="uploader-hint">上传图片</span>
              </div>
              <div v-if="uploadingBackground" class="uploader-loading">
                <a-spin size="small" />
              </div>
            </div>
            <a-button
              v-if="form.page_background"
              type="text"
              danger
              size="small"
              @click="clearImage('background')"
            >
              移除
            </a-button>
          </div>
        </a-form-item>
      </div>

      <!-- 表单区：两列网格布局，长字段跨两列 -->
      <div class="form-grid">
        <a-form-item label="名称">
          <a-input
            v-model:value="form.nickname"
            placeholder="请输入名称"
            :maxlength="50"
          />
        </a-form-item>

        <a-form-item>
          <template #label>
            <span class="label-with-switch">
              邮箱
              <a-tooltip title="开启后将在博客对外展示">
                <a-switch
                  v-model:checked="form.show_email"
                  size="small"
                  aria-label="邮箱是否对外展示"
                />
              </a-tooltip>
            </span>
          </template>
          <a-input
            v-model:value="form.email"
            type="email"
            placeholder="请输入邮箱"
            :maxlength="100"
          />
        </a-form-item>

        <a-form-item>
          <template #label>
            <span class="label-with-switch">
              所在城市
              <a-tooltip title="开启后将在博客对外展示">
                <a-switch
                  v-model:checked="form.show_city"
                  size="small"
                  aria-label="城市是否对外展示"
                />
              </a-tooltip>
            </span>
          </template>
          <a-cascader
            v-model:value="cityValue"
            :options="CHINA_REGIONS"
            placeholder="请选择省/市"
            :allow-clear="true"
            :show-search="{ filter }"
            style="width: 100%"
          />
        </a-form-item>

        <a-form-item label="博客标题">
          <a-input
            v-model:value="form.blog_title"
            placeholder="请输入博客标题"
            :maxlength="100"
          />
        </a-form-item>

        <a-form-item>
          <template #label>
            <span class="label-with-switch">
              性格
              <a-tooltip title="开启后将在博客对外展示">
                <a-switch
                  v-model:checked="form.show_personality"
                  size="small"
                  aria-label="性格是否对外展示"
                />
              </a-tooltip>
            </span>
          </template>
          <CardGridPicker
            v-model="form.personality"
            :options="personalityOptions"
            placeholder="选择你的性格（MBTI）"
            search-placeholder="搜索性格类型"
            aria-label="选择性格"
          />
        </a-form-item>

        <a-form-item>
          <template #label>
            <span class="label-with-switch">
              星座
              <a-tooltip title="开启后将在博客对外展示">
                <a-switch
                  v-model:checked="form.show_zodiac"
                  size="small"
                  aria-label="星座是否对外展示"
                />
              </a-tooltip>
            </span>
          </template>
          <CardGridPicker
            v-model="form.zodiac"
            :options="zodiacOptions"
            placeholder="选择你的星座"
            search-placeholder="搜索星座"
            aria-label="选择星座"
          />
        </a-form-item>

        <a-form-item class="span-2" label="介绍">
          <a-textarea
            v-model:value="form.bio"
            placeholder="请输入个人介绍"
            :rows="4"
          />
        </a-form-item>

        <a-form-item class="span-2" label="博客描述">
          <a-textarea
            v-model:value="form.blog_description"
            placeholder="请输入博客描述"
            :rows="3"
          />
        </a-form-item>

        <a-form-item class="span-2" label="标签">
          <div class="tags-container">
            <a-tag
              v-for="(tag, index) in tags"
              :key="index"
              closable
              @close="removeTag(index)"
            >
              {{ tag }}
            </a-tag>
            <a-input
              v-if="inputVisible"
              ref="inputRef"
              v-model:value="inputValue"
              size="small"
              style="width: 120px"
              @keyup.enter="handleInputConfirm"
              @blur="handleInputConfirm"
            />
            <a-tag v-else style="background: #fafafa; border: 1px dashed #d9d9d9; cursor: pointer" @click="showInput">
              + 添加标签
            </a-tag>
          </div>
        </a-form-item>

        <a-form-item class="span-2" label="社交链接">
          <div class="social-links-container">
            <div
              v-for="(link, index) in socialLinks"
              :key="index"
              class="social-link-row"
            >
              <a-select
                v-model:value="link.platform"
                placeholder="选择平台"
                style="width: 180px"
                show-search
                :filter-option="filterPlatform"
              >
                <a-select-option
                  v-for="platform in SOCIAL_PLATFORMS"
                  :key="platform.key"
                  :value="platform.key"
                >
                  <div class="platform-option">
                    <span class="platform-option-icon" :style="{ color: platform.color }">
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
                        <path :d="platform.icon" />
                      </svg>
                    </span>
                    <span>{{ platform.name }}</span>
                  </div>
                </a-select-option>
              </a-select>
              <a-input
                v-model:value="link.url"
                placeholder="请输入个人主页 URL"
                allow-clear
                style="flex: 1"
              />
              <a-button danger @click="removeSocialLink(index)">
                删除
              </a-button>
            </div>
            <a-button type="dashed" block @click="addSocialLink">
              + 添加社交链接
            </a-button>
          </div>
        </a-form-item>

        <a-form-item class="span-2">
          <a-button type="primary" :loading="saving" @click="handleSave">保存</a-button>
        </a-form-item>
      </div>

      <input
        ref="backgroundInputRef"
        type="file"
        accept="image/*"
        style="display: none"
        aria-hidden="true"
        @change="(e: Event) => handleFileChange(e, 'background')"
      />
      <input
        ref="iconInputRef"
        type="file"
        accept="image/*"
        style="display: none"
        aria-hidden="true"
        @change="(e: Event) => handleFileChange(e, 'icon')"
      />
      <input
        ref="avatarInputRef"
        type="file"
        accept="image/*"
        style="display: none"
        aria-hidden="true"
        @change="(e: Event) => handleFileChange(e, 'avatar')"
      />

      <AvatarCropperModal
        v-model:open="avatarCropVisible"
        :file="avatarCropFile"
        :confirm-loading="uploadingAvatar"
        @confirm="handleAvatarCropped"
      />
    </a-form>
  </a-spin>
</template>

<script setup lang="ts">
import { computed, reactive, ref, onMounted, nextTick, watch } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { authApi } from '@/api/auth'
import type { UpdateProfileReq, SocialLink, ProfileMeta } from '@/types/api'
import { SOCIAL_PLATFORMS } from '@/components/profile/socialPlatforms'
import AvatarCropperModal from '@/components/profile/AvatarCropperModal.vue'
import CardGridPicker from '@/components/profile/CardGridPicker.vue'
import type { PickerCardOption } from '@/components/profile/CardGridPicker.vue'
import { CHINA_REGIONS } from '@/assets/data/china-regions'
import type { CascaderOptionType } from 'ant-design-vue/es/cascader'

const loading = ref(false)
const saving = ref(false)
const uploadingIcon = ref(false)
const uploadingBackground = ref(false)
const uploadingAvatar = ref(false)

const backgroundInputRef = ref<HTMLInputElement | null>(null)
const iconInputRef = ref<HTMLInputElement | null>(null)
const avatarInputRef = ref<HTMLInputElement | null>(null)
// 头像 QQ 式裁剪：选图后先弹圆形裁剪框，确认裁剪后再上传
const avatarCropVisible = ref(false)
const avatarCropFile = ref<File | null>(null)
// 图片加载失败标记：key 为 'icon' | 'background' | 'avatar'
const brokenImages = reactive<Record<string, boolean>>({})

const form = reactive<UpdateProfileReq>({
  nickname: '',
  avatar: '',
  bio: '',
  email: '',
  personality: '',
  zodiac: '',
  show_email: true,
  show_city: true,
  show_zodiac: true,
  show_personality: true,
  page_background: '',
  blog_icon: '',
  blog_title: '',
  blog_description: '',
})

// 星座/性格选项元数据（公开接口下发，选择器与展示共用一份数据源）
const profileMeta = ref<ProfileMeta | null>(null)

const personalityOptions = computed<PickerCardOption[]>(() =>
  (profileMeta.value?.personality ?? []).map((item) => ({
    key: item.key,
    name: item.name,
    image: item.image,
    description: item.description,
  })),
)

const zodiacOptions = computed<PickerCardOption[]>(() =>
  (profileMeta.value?.zodiac ?? []).map((item) => ({
    key: item.key,
    name: item.name,
    image: item.image,
    dateRange: item.date_range,
    element: item.element,
  })),
)

// 城市级联选择器值：['广东省', '广州市']
const cityValue = ref<string[]>([])

const socialLinks = ref<SocialLink[]>([])
const tags = ref<string[]>([])
const inputVisible = ref(false)
const inputValue = ref('')
const inputRef = ref()

// 城市值变化时同步到 form.city
watch(cityValue, (val) => {
  form.city = val.length > 0 ? val.join('/') : ''
})

// 级联选择器搜索
function filter(inputValue: string, path: CascaderOptionType[]): boolean {
  return path.some((option) => option.label.toLowerCase().includes(inputValue.toLowerCase()))
}

function addSocialLink() {
  socialLinks.value.push({ platform: '', url: '', sort_order: socialLinks.value.length })
}

function removeSocialLink(index: number) {
  socialLinks.value.splice(index, 1)
}

function removeTag(index: number) {
  tags.value.splice(index, 1)
}

function showInput() {
  inputVisible.value = true
  nextTick(() => {
    inputRef.value?.focus()
  })
}

function handleInputConfirm() {
  const val = inputValue.value.trim()
  if (val && !tags.value.includes(val)) {
    tags.value.push(val)
  }
  inputVisible.value = false
  inputValue.value = ''
}

function filterPlatform(input: string, option: { value: string }) {
  const platform = SOCIAL_PLATFORMS.find((p) => p.key === option.value)
  return platform ? platform.name.toLowerCase().includes(input.toLowerCase()) : false
}

async function loadProfileMeta() {
  try {
    profileMeta.value = await authApi.getProfileMeta()
  } catch {
    // 选项元数据加载失败不阻塞资料编辑，选择器仅展示占位符
  }
}

async function loadProfile() {
  loading.value = true
  try {
    const data = await authApi.getProfile()
    form.nickname = data.nickname || ''
    form.avatar = data.avatar || ''
    form.bio = data.bio || ''
    form.email = data.email || ''
    form.personality = data.personality || ''
    form.zodiac = data.zodiac || ''
    form.show_email = data.show_email ?? true
    form.show_city = data.show_city ?? true
    form.show_zodiac = data.show_zodiac ?? true
    form.show_personality = data.show_personality ?? true
    form.page_background = data.page_background || ''
    form.blog_icon = data.blog_icon || ''
    form.blog_title = data.blog_title || ''
    form.blog_description = data.blog_description || ''
    // 城市：拆分 "广东省/广州市" → ['广东省', '广州市']
    cityValue.value = data.city ? data.city.split('/') : []
    socialLinks.value = data.social_links || []
    tags.value = data.tags || []
    brokenImages.background = false
    brokenImages.icon = false
    brokenImages.avatar = false
  } catch {
    message.error('加载个人资料失败')
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    await authApi.updateProfile({ ...form, social_links: socialLinks.value, tags: tags.value })
    message.success('保存成功')
  } catch {
    message.error('保存失败')
  } finally {
    saving.value = false
  }
}

type ImageField = 'icon' | 'background' | 'avatar'

function triggerUpload(type: ImageField) {
  const uploading = type === 'icon' ? uploadingIcon : type === 'background' ? uploadingBackground : uploadingAvatar
  if (uploading.value) return
  const input = type === 'icon' ? iconInputRef.value : type === 'background' ? backgroundInputRef.value : avatarInputRef.value
  input?.click()
}

async function handleFileChange(e: Event, type: ImageField) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  target.value = ''
  if (type === 'avatar') {
    // 头像走 QQ 式裁剪流程：先弹圆形裁剪框，确认裁剪后再上传
    avatarCropFile.value = file
    avatarCropVisible.value = true
    return
  }
  const uploading = type === 'icon' ? uploadingIcon : uploadingBackground
  uploading.value = true
  try {
    const res = type === 'icon'
      ? await authApi.uploadIcon(file)
      : await authApi.uploadBackground(file)
    if (type === 'icon') {
      form.blog_icon = res.url
    } else {
      form.page_background = res.url
    }
    brokenImages[type] = false
    message.success('上传成功')
  } catch {
    message.error('上传失败')
  } finally {
    uploading.value = false
  }
}

/** 圆形裁剪确认：裁剪结果以 PNG 上传（后端按扩展名校验，.png 在白名单内） */
async function handleAvatarCropped(blob: Blob) {
  uploadingAvatar.value = true
  try {
    const file = new File([blob], 'avatar.png', { type: 'image/png' })
    const res = await authApi.uploadAvatar(file)
    form.avatar = res.url
    brokenImages.avatar = false
    message.success('上传成功')
    avatarCropVisible.value = false
    avatarCropFile.value = null
  } catch {
    message.error('上传失败')
  } finally {
    uploadingAvatar.value = false
  }
}

function clearImage(type: ImageField) {
  if (type === 'icon') {
    form.blog_icon = ''
  } else if (type === 'background') {
    form.page_background = ''
  } else {
    form.avatar = ''
  }
  brokenImages[type] = false
}

function onImgError(type: ImageField) {
  brokenImages[type] = true
}

onMounted(() => {
  loadProfile()
  loadProfileMeta()
})
</script>

<style scoped>
/* 形象区：三个上传项横向排列 */
.identity-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0 32px;
  align-items: flex-end;
}

/* 表单区：两列网格，窄屏降为单列 */
.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 24px;
}

.form-grid .span-2 {
  grid-column: 1 / -1;
}

@media (max-width: 767px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
}

/* 带对外展示开关的表单 label */
.label-with-switch {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.image-field {
  display: flex;
  gap: 12px;
  align-items: center;
}

.image-uploader {
  position: relative;
  flex-shrink: 0;
  cursor: pointer;
  overflow: hidden;
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius);
  background: var(--bg-card);
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.image-uploader:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-sm);
}

.image-uploader:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.image-uploader.is-empty {
  border-style: dashed;
}

.image-uploader img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.uploader-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  width: 100%;
  height: 100%;
  color: var(--text-secondary);
  transition: color var(--transition-fast);
}

.image-uploader:hover .uploader-empty {
  color: var(--color-primary);
}

.uploader-empty :deep(.anticon) {
  font-size: 20px;
}

.uploader-hint {
  font-size: 12px;
  line-height: 1;
}

.uploader-mask {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(15, 23, 42, 0.45);
  color: #fff;
  font-size: 12px;
  opacity: 0;
  transition: opacity var(--transition-fast);
}

.image-uploader:hover .uploader-mask,
.image-uploader:focus-visible .uploader-mask {
  opacity: 1;
}

.uploader-loading {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.65);
  cursor: not-allowed;
}

.background-uploader {
  width: 200px;
  height: 80px;
}

.avatar-uploader {
  width: 72px;
  height: 72px;
  border-radius: 50%;
}

.icon-uploader {
  width: 56px;
  height: 56px;
}

.social-links-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
}

.social-link-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.platform-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

.platform-option-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}
</style>
