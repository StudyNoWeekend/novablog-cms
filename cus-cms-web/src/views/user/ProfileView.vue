<template>
  <a-spin :spinning="loading">
    <a-form layout="vertical" :model="form">
      <a-form-item label="页面背景图">
        <div class="image-field">
          <a-input
            v-model:value="form.page_background"
            placeholder="请输入背景图 URL"
            allow-clear
          />
          <a-upload
            :show-upload-list="false"
            :before-upload="(file: File) => handleUpload(file, 'background')"
            accept="image/*"
          >
            <a-button :loading="uploadingBackground">上传</a-button>
          </a-upload>
          <div v-if="form.page_background" class="image-preview background-preview">
            <img :src="form.page_background" alt="背景图预览" @error="onImgError" />
          </div>
        </div>
      </a-form-item>

      <a-form-item label="博客 Icon">
        <div class="image-field">
          <a-input
            v-model:value="form.blog_icon"
            placeholder="请输入博客 Icon 图 URL"
            allow-clear
          />
          <a-upload
            :show-upload-list="false"
            :before-upload="(file: File) => handleUpload(file, 'icon')"
            accept="image/*"
          >
            <a-button :loading="uploadingIcon">上传</a-button>
          </a-upload>
          <div v-if="form.blog_icon" class="image-preview icon-preview">
            <img :src="form.blog_icon" alt="Icon 预览" @error="onImgError" />
          </div>
        </div>
      </a-form-item>

      <a-form-item label="头像">
        <div class="image-field">
          <a-input
            v-model:value="form.avatar"
            placeholder="请输入头像 URL"
            allow-clear
          />
          <div v-if="form.avatar" class="image-preview avatar-preview">
            <img :src="form.avatar" alt="头像预览" @error="onImgError" />
          </div>
        </div>
      </a-form-item>

      <a-form-item label="名称">
        <a-input
          v-model:value="form.nickname"
          placeholder="请输入名称"
          :maxlength="50"
        />
      </a-form-item>

      <a-form-item label="介绍">
        <a-textarea
          v-model:value="form.bio"
          placeholder="请输入个人介绍"
          :rows="4"
        />
      </a-form-item>

      <a-form-item label="博客标题">
        <a-input
          v-model:value="form.blog_title"
          placeholder="请输入博客标题"
          :maxlength="100"
        />
      </a-form-item>

      <a-form-item label="博客描述">
        <a-textarea
          v-model:value="form.blog_description"
          placeholder="请输入博客描述"
          :rows="3"
        />
      </a-form-item>

      <a-form-item label="社交链接">
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

      <a-form-item>
        <a-button type="primary" :loading="saving" @click="handleSave">保存</a-button>
      </a-form-item>
    </a-form>
  </a-spin>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { authApi } from '@/api/auth'
import type { UpdateProfileReq, SocialLink } from '@/types/api'
import { SOCIAL_PLATFORMS } from '@/components/profile/socialPlatforms'

const loading = ref(false)
const saving = ref(false)
const uploadingIcon = ref(false)
const uploadingBackground = ref(false)

const form = reactive<UpdateProfileReq>({
  nickname: '',
  avatar: '',
  bio: '',
  page_background: '',
  blog_icon: '',
  blog_title: '',
  blog_description: '',
})

const socialLinks = ref<SocialLink[]>([])

function addSocialLink() {
  socialLinks.value.push({ platform: '', url: '', sort_order: socialLinks.value.length })
}

function removeSocialLink(index: number) {
  socialLinks.value.splice(index, 1)
}

function filterPlatform(input: string, option: { value: string }) {
  const platform = SOCIAL_PLATFORMS.find((p) => p.key === option.value)
  return platform ? platform.name.toLowerCase().includes(input.toLowerCase()) : false
}

async function loadProfile() {
  loading.value = true
  try {
    const data = await authApi.getProfile()
    form.nickname = data.nickname || ''
    form.avatar = data.avatar || ''
    form.bio = data.bio || ''
    form.page_background = data.page_background || ''
    form.blog_icon = data.blog_icon || ''
    form.blog_title = data.blog_title || ''
    form.blog_description = data.blog_description || ''
    socialLinks.value = data.social_links || []
  } catch {
    message.error('加载个人资料失败')
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    await authApi.updateProfile({ ...form, social_links: socialLinks.value })
    message.success('保存成功')
  } catch {
    message.error('保存失败')
  } finally {
    saving.value = false
  }
}

async function handleUpload(file: File, type: 'icon' | 'background') {
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
    message.success('上传成功')
  } catch {
    message.error('上传失败')
  } finally {
    uploading.value = false
  }
  return false
}

function onImgError(e: Event) {
  const img = e.target as HTMLImageElement
  img.style.display = 'none'
}

onMounted(loadProfile)
</script>

<style scoped>
.image-field {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.image-preview {
  flex-shrink: 0;
  border: 1px solid #e8e8e8;
  border-radius: 6px;
  overflow: hidden;
  background: #fafafa;
}

.image-preview img {
  display: block;
  object-fit: cover;
}

.background-preview {
  width: 200px;
  height: 80px;
}

.background-preview img {
  width: 100%;
  height: 100%;
}

.avatar-preview {
  width: 64px;
  height: 64px;
  border-radius: 50%;
}

.avatar-preview img {
  width: 100%;
  height: 100%;
}

.icon-preview {
  width: 48px;
  height: 48px;
  border-radius: 8px;
}

.icon-preview img {
  width: 100%;
  height: 100%;
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
</style>
