<template>
  <a-modal
    :open="visible"
    :footer="null"
    :width="1200"
    :mask-closable="false"
    :destroy-on-close="true"
    @cancel="handleClose"
  >
    <div class="workbench">
      <!-- 左侧：模板与参数 -->
      <div class="panel left-panel">
        <div class="section">
          <div class="section-title">模板风格</div>
          <a-radio-group v-model:value="frameConfig.template" class="template-group">
            <a-radio-button value="gallery">画廊</a-radio-button>
            <a-radio-button value="movie">电影</a-radio-button>
            <a-radio-button value="floating">悬浮</a-radio-button>
          </a-radio-group>
        </div>

        <div class="section">
          <div class="section-title">样式参数</div>
          <div class="param-row">
            <span>字体大小</span>
            <span>{{ frameConfig.fontScale.toFixed(1) }}x</span>
          </div>
          <a-slider v-model:value="frameConfig.fontScale" :min="0.5" :max="2" :step="0.1" />

          <div class="param-row">
            <span>边框尺寸</span>
            <span>{{ frameConfig.borderScale.toFixed(1) }}x</span>
          </div>
          <a-slider v-model:value="frameConfig.borderScale" :min="0.5" :max="2" :step="0.1" />

          <div class="param-row">
            <span>边框背景</span>
            <a-radio-group v-model:value="frameConfig.borderColor" size="small">
              <a-radio-button value="#ffffff">白</a-radio-button>
              <a-radio-button value="#000000">黑</a-radio-button>
            </a-radio-group>
          </div>

          <div class="param-row">
            <span>文字颜色</span>
            <a-radio-group v-model:value="frameConfig.textColor" size="small">
              <a-radio-button value="auto">自动</a-radio-button>
              <a-radio-button value="black">黑</a-radio-button>
              <a-radio-button value="white">白</a-radio-button>
            </a-radio-group>
          </div>

          <div class="param-row">
            <span>字体</span>
            <a-select v-model:value="frameConfig.fontFamily" size="small" style="width: 120px">
              <a-select-option value="Inter">Inter</a-select-option>
              <a-select-option value="Arial">Arial</a-select-option>
              <a-select-option value="serif">Serif</a-select-option>
            </a-select>
          </div>
        </div>

        <div class="section">
          <div class="section-title">品牌/型号展示</div>
          <a-radio-group v-model:value="frameConfig.logoMode" size="small">
            <a-radio-button value="text">文字</a-radio-button>
            <a-radio-button value="image">图片</a-radio-button>
          </a-radio-group>
          <div v-if="frameConfig.logoMode === 'image'" class="placeholder-tip">
            Logo 图片模式暂占位
          </div>
        </div>

        <div class="section">
          <div class="param-row">
            <span>显示 EXIF</span>
            <a-switch v-model:checked="frameConfig.showExif" />
          </div>
        </div>

        <div class="section">
          <div class="section-title">参数设定</div>
          <a-form layout="vertical" size="small">
            <a-form-item label="品牌">
              <a-input v-model:value="displayParams.make" />
            </a-form-item>
            <a-form-item label="型号">
              <a-input v-model:value="displayParams.model" />
            </a-form-item>
            <a-form-item label="镜头">
              <a-input v-model:value="displayParams.lens" />
            </a-form-item>
            <a-row :gutter="8">
              <a-col :span="12">
                <a-form-item label="焦距">
                  <a-input v-model:value="displayParams.focalLength" />
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item label="光圈">
                  <a-input v-model:value="displayParams.aperture" />
                </a-form-item>
              </a-col>
            </a-row>
            <a-row :gutter="8">
              <a-col :span="12">
                <a-form-item label="快门">
                  <a-input v-model:value="displayParams.shutter" />
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item label="ISO">
                  <a-input v-model:value="displayParams.iso" />
                </a-form-item>
              </a-col>
            </a-row>
            <a-form-item label="日期">
              <a-input v-model:value="displayParams.date" />
            </a-form-item>
          </a-form>
        </div>
      </div>

      <!-- 中间：预览 + 预设栏 -->
      <div class="preview-area">
        <div class="canvas-wrapper">
          <a-spin :spinning="loading" tip="加载原图中...">
            <canvas ref="previewCanvas" class="preview-canvas" />
          </a-spin>
        </div>

        <div class="preset-bar">
          <div class="preset-bar-header">
            <span>图库 ({{ presets.length }})</span>
            <a-button type="text" size="small" @click="fetchPresets">
              <ReloadOutlined />
            </a-button>
          </div>
          <div class="preset-thumbnails">
            <div
              v-for="preset in presets"
              :key="preset.id"
              class="preset-thumb"
              @click="applyPreset(preset)"
            >
              <img :src="preset.output_url" :alt="preset.name" />
              <span class="thumb-name" :title="preset.name">{{ preset.name }}</span>
              <a-button
                type="text"
                danger
                size="small"
                class="thumb-delete"
                @click.stop="handleDeletePreset(preset.id)"
              >
                <DeleteOutlined />
              </a-button>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧：EXIF + 预设库 + 导出 -->
      <div class="panel right-panel">
        <div class="section">
          <div class="section-title">EXIF 信息</div>
          <div class="exif-list">
            <div class="exif-item">
              <span class="exif-label">文件</span>
              <span class="exif-value">{{ exifInfo.filename }}</span>
            </div>
            <div class="exif-item">
              <span class="exif-label">分辨率</span>
              <span class="exif-value">{{ exifInfo.width && exifInfo.height ? `${exifInfo.width} x ${exifInfo.height}` : '-' }}</span>
            </div>
            <div class="exif-item">
              <span class="exif-label">大小</span>
              <span class="exif-value">{{ formatBytes(exifInfo.size) }}</span>
            </div>
            <div class="exif-item">
              <span class="exif-label">厂商</span>
              <span class="exif-value">{{ exifInfo.make || '--' }}</span>
            </div>
            <div class="exif-item">
              <span class="exif-label">型号</span>
              <span class="exif-value">{{ exifInfo.model || '--' }}</span>
            </div>
            <div class="exif-item">
              <span class="exif-label">镜头</span>
              <span class="exif-value">{{ exifInfo.lens || '--' }}</span>
            </div>
            <div class="exif-item">
              <span class="exif-label">软件</span>
              <span class="exif-value">{{ exifInfo.software || '--' }}</span>
            </div>
          </div>
          <div class="exif-cards">
            <div class="exif-card">
              <div class="exif-card-label">FOCAL</div>
              <div class="exif-card-value">{{ exifInfo.focalLength || '--' }}</div>
            </div>
            <div class="exif-card">
              <div class="exif-card-label">AP</div>
              <div class="exif-card-value">{{ exifInfo.aperture || '--' }}</div>
            </div>
            <div class="exif-card">
              <div class="exif-card-label">SHTR</div>
              <div class="exif-card-value">{{ exifInfo.shutter || '--' }}</div>
            </div>
            <div class="exif-card">
              <div class="exif-card-label">ISO</div>
              <div class="exif-card-value">{{ exifInfo.iso || '--' }}</div>
            </div>
          </div>
        </div>

        <div class="section">
          <div class="section-title">预设库</div>
          <a-input v-model:value="presetName" placeholder="预设名称..." />
          <a-button type="primary" block style="margin-top: 8px" :loading="savingPreset" @click="handleSave">
            <SaveOutlined /> 保存
          </a-button>
        </div>

        <div class="section">
          <div class="section-title">导出</div>
          <a-input v-model:value="exportName" placeholder="文件名(默认自动)">
            <template #addonAfter>.JPG</template>
          </a-input>
          <a-button type="primary" block style="margin-top: 8px" :loading="savingExport" @click="handleExport">
            <DownloadOutlined /> 下载
          </a-button>
        </div>
      </div>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, reactive, watch, nextTick, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import {
  ReloadOutlined,
  DeleteOutlined,
  SaveOutlined,
  DownloadOutlined,
} from '@ant-design/icons-vue'
import exifr from 'exifr'
import piexif from 'piexifjs'
import { mediaApi } from '@/api/media'
import type { MediaPreset, FrameConfig, DisplayParams, ExifInfo } from '@/types/api'

const props = defineProps<{
  mediaId: string
  visible: boolean
  originalUrl: string
  filename?: string
  size?: number
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'saved'): void
}>()

const loading = ref(false)
const savingPreset = ref(false)
const savingExport = ref(false)
const previewCanvas = ref<HTMLCanvasElement | null>(null)
const originalImage = ref<HTMLImageElement | null>(null)
const originalBuffer = ref<ArrayBuffer | null>(null)
const originalMime = ref('image/jpeg')
const presets = ref<MediaPreset[]>([])
const presetName = ref('')
const exportName = ref('')

const frameConfig = reactive<FrameConfig>({
  template: 'gallery',
  fontScale: 1.0,
  borderScale: 1.0,
  borderColor: '#ffffff',
  textColor: 'auto',
  fontFamily: 'Inter',
  logoMode: 'text',
  showExif: true,
})

const displayParams = reactive<DisplayParams>({
  make: '',
  model: '',
  lens: '',
  focalLength: '',
  aperture: '',
  shutter: '',
  iso: '',
  date: '',
})

const exifInfo = reactive<ExifInfo>({
  filename: '',
  width: 0,
  height: 0,
  size: 0,
  make: '',
  model: '',
  lens: '',
  software: '',
  focalLength: '',
  aperture: '',
  shutter: '',
  iso: '',
})

const templateDefaults: Record<FrameConfig['template'], Partial<FrameConfig>> = {
  gallery: { borderColor: '#ffffff', textColor: 'auto' },
  movie: { borderColor: '#000000', textColor: 'white' },
  floating: { borderColor: '#ffffff', textColor: 'black' },
}

watch(() => frameConfig.template, (tpl) => {
  const defaults = templateDefaults[tpl]
  if (defaults) {
    Object.assign(frameConfig, defaults)
  }
})

watch([frameConfig, displayParams], () => {
  renderPreview()
}, { deep: true })

watch(() => props.visible, (val) => {
  if (val) {
    loadOriginal()
    fetchPresets()
  }
})

onMounted(() => {
  if (props.visible) {
    loadOriginal()
    fetchPresets()
  }
})

// 工作台原图需以 crossOrigin='anonymous' 加载，要求响应携带 ACAO 头。
// 媒体网格缩略图以 no-cors 方式加载并可能缓存了不含 ACAO 的响应（未带 Vary: Origin 的历史缓存），
// 追加 wbv 参数更换缓存键，确保工作台始终回源拿到带 ACAO 的响应。
function workbenchUrl(): string {
  try {
    const u = new URL(props.originalUrl)
    u.searchParams.set('wbv', '1')
    return u.toString()
  } catch {
    return props.originalUrl
  }
}

async function loadOriginal() {
  if (!props.originalUrl) return
  loading.value = true
  try {
    const img = new Image()
    img.crossOrigin = 'anonymous'
    img.onload = () => {
      originalImage.value = img
      exifInfo.width = img.width
      exifInfo.height = img.height
      renderPreview()
      loading.value = false
    }
    img.onerror = () => {
      message.error('原图加载失败，请检查对象存储 CORS 配置')
      loading.value = false
    }
    img.src = workbenchUrl()

    exifInfo.filename = props.filename || ''
    exifInfo.size = props.size || 0

    // 异步读取 EXIF，失败不影响预览
    readExif().catch(() => {})
  } catch (err) {
    message.error('加载原图失败')
    loading.value = false
  }
}

async function readExif() {
  if (!props.originalUrl || originalMime.value !== 'image/jpeg') return
  try {
    const res = await fetch(workbenchUrl(), { mode: 'cors' })
    if (!res.ok) return
    const buffer = await res.arrayBuffer()
    originalBuffer.value = buffer
    const data = new Uint8Array(buffer)
    const parsed = await exifr.parse(data)
    if (!parsed) return

    exifInfo.make = parsed.Make || ''
    exifInfo.model = parsed.Model || ''
    exifInfo.lens = parsed.LensModel || parsed.Lens || ''
    exifInfo.software = parsed.Software || ''
    exifInfo.focalLength = parsed.FocalLength ? `${parsed.FocalLength}mm` : ''
    exifInfo.aperture = parsed.FNumber ? `f/${parsed.FNumber}` : ''
    exifInfo.shutter = parsed.ExposureTime ? formatShutter(parsed.ExposureTime) : ''
    exifInfo.iso = parsed.ISO ? String(parsed.ISO) : ''

    displayParams.make = exifInfo.make
    displayParams.model = exifInfo.model
    displayParams.lens = exifInfo.lens
    displayParams.focalLength = exifInfo.focalLength
    displayParams.aperture = exifInfo.aperture
    displayParams.shutter = exifInfo.shutter
    displayParams.iso = exifInfo.iso
    displayParams.date = parsed.DateTimeOriginal
      ? formatDate(parsed.DateTimeOriginal)
      : ''
  } catch {
    // EXIF 读取失败不影响主流程
  }
}

function getTextColor(): string {
  if (frameConfig.textColor !== 'auto') return frameConfig.textColor
  return frameConfig.borderColor === '#000000' ? 'white' : 'black'
}

function renderPreview() {
  nextTick(() => {
    doRender(false)
  })
}

function doRender(fullSize: boolean) {
  const img = originalImage.value
  const canvas = previewCanvas.value
  if (!img || !canvas) return

  const maxSide = fullSize ? Math.max(img.width, img.height) : 4096
  let width = img.width
  let height = img.height
  if (width > maxSide || height > maxSide) {
    if (width > height) {
      height = Math.round((height * maxSide) / width)
      width = maxSide
    } else {
      width = Math.round((width * maxSide) / height)
      height = maxSide
    }
  }

  const base = Math.min(width, height)
  const borderWidth = base * 0.04 * frameConfig.borderScale
  const fontSize = base * 0.022 * frameConfig.fontScale
  const infoHeight = frameConfig.showExif ? fontSize * 3.5 : 0

  canvas.width = Math.round(width + borderWidth * 2)
  canvas.height = Math.round(height + borderWidth * 2 + infoHeight)

  const ctx = canvas.getContext('2d')
  if (!ctx) return

  // 背景
  ctx.fillStyle = frameConfig.borderColor
  ctx.fillRect(0, 0, canvas.width, canvas.height)

  // 图片
  ctx.drawImage(img, borderWidth, borderWidth, width, height)

  // 信息条
  if (frameConfig.showExif) {
    const textColor = getTextColor()
    const infoY = borderWidth + height
    ctx.fillStyle = frameConfig.borderColor
    ctx.fillRect(borderWidth, infoY, width, infoHeight)

    ctx.fillStyle = textColor
    ctx.font = `${fontSize}px ${frameConfig.fontFamily}, sans-serif`
    ctx.textBaseline = 'middle'

    const padX = fontSize * 1.5
    const centerY = infoY + infoHeight / 2

    // 左侧品牌/型号
    const brandText = frameConfig.logoMode === 'text'
      ? `${displayParams.make || ''} ${displayParams.model || ''}`.trim()
      : displayParams.make || ''
    if (brandText) {
      ctx.fillText(brandText, borderWidth + padX, centerY)
    }

    // 中间参数
    const params: string[] = []
    if (displayParams.focalLength) params.push(displayParams.focalLength)
    if (displayParams.aperture) params.push(displayParams.aperture)
    if (displayParams.shutter) params.push(displayParams.shutter)
    if (displayParams.iso) params.push(displayParams.iso)
    const paramsText = params.join('    ')
    if (paramsText) {
      const paramsWidth = ctx.measureText(paramsText).width
      ctx.fillText(paramsText, borderWidth + width / 2 - paramsWidth / 2, centerY)
    }

    // 右侧镜头+日期
    const rightText = [displayParams.lens, displayParams.date].filter(Boolean).join('  ')
    if (rightText) {
      const rightWidth = ctx.measureText(rightText).width
      ctx.fillText(rightText, borderWidth + width - padX - rightWidth, centerY)
    }
  }

  return { width: canvas.width, height: canvas.height }
}

async function fetchPresets() {
  try {
    const res = await mediaApi.getPresets(props.mediaId)
    presets.value = res.list || []
  } catch {
    presets.value = []
  }
}

function applyPreset(preset: MediaPreset) {
  try {
    const fc = JSON.parse(preset.frame_config) as FrameConfig
    const dp = JSON.parse(preset.display_params) as DisplayParams
    Object.assign(frameConfig, fc)
    frameConfig.showExif = fc.showExif ?? true
    Object.assign(displayParams, dp)
    presetName.value = preset.name
  } catch {
    message.error('应用预设失败')
  }
}

async function handleDeletePreset(id: string) {
  try {
    await mediaApi.deletePreset(id)
    message.success('删除成功')
    fetchPresets()
  } catch {
    message.error('删除失败')
  }
}

async function handleSave() {
  const name = presetName.value.trim() || '未命名预设'
  savingPreset.value = true
  try {
    const blob = await renderFinalBlob()
    if (!blob) throw new Error('合成图片失败')

    await mediaApi.createPreset({
      media_id: props.mediaId,
      name,
      frame_config: JSON.stringify(frameConfig),
      display_params: JSON.stringify(displayParams),
      file: blob,
    })
    message.success('保存成功')
    presetName.value = ''
    emit('saved')
    fetchPresets()
  } catch (err) {
    message.error(`保存失败: ${err instanceof Error ? err.message : '未知错误'}`)
  } finally {
    savingPreset.value = false
  }
}

async function handleExport() {
  savingExport.value = true
  try {
    const blob = await renderFinalBlob()
    if (!blob) throw new Error('合成图片失败')

    const filename = (exportName.value.trim() || presetName.value.trim() || '未命名') + '.jpg'
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    message.success('导出成功')
  } catch (err) {
    message.error(`导出失败: ${err instanceof Error ? err.message : '未知错误'}`)
  } finally {
    savingExport.value = false
  }
}

async function renderFinalBlob(): Promise<Blob | null> {
  return new Promise((resolve, reject) => {
    const img = originalImage.value
    if (!img) {
      reject(new Error('原图未加载'))
      return
    }

    // 使用离屏 canvas 按原图尺寸合成
    const canvas = document.createElement('canvas')
    const width = img.width
    const height = img.height
    const base = Math.min(width, height)
    const borderWidth = base * 0.04 * frameConfig.borderScale
    const fontSize = base * 0.022 * frameConfig.fontScale
    const infoHeight = frameConfig.showExif ? fontSize * 3.5 : 0

    canvas.width = Math.round(width + borderWidth * 2)
    canvas.height = Math.round(height + borderWidth * 2 + infoHeight)

    const ctx = canvas.getContext('2d')
    if (!ctx) {
      reject(new Error('创建 canvas 失败'))
      return
    }

    ctx.fillStyle = frameConfig.borderColor
    ctx.fillRect(0, 0, canvas.width, canvas.height)
    ctx.drawImage(img, borderWidth, borderWidth, width, height)

    if (frameConfig.showExif) {
      const textColor = getTextColor()
      const infoY = borderWidth + height
      ctx.fillStyle = frameConfig.borderColor
      ctx.fillRect(borderWidth, infoY, width, infoHeight)

      ctx.fillStyle = textColor
      ctx.font = `${fontSize}px ${frameConfig.fontFamily}, sans-serif`
      ctx.textBaseline = 'middle'

      const padX = fontSize * 1.5
      const centerY = infoY + infoHeight / 2

      const brandText = frameConfig.logoMode === 'text'
        ? `${displayParams.make || ''} ${displayParams.model || ''}`.trim()
        : displayParams.make || ''
      if (brandText) {
        ctx.fillText(brandText, borderWidth + padX, centerY)
      }

      const params: string[] = []
      if (displayParams.focalLength) params.push(displayParams.focalLength)
      if (displayParams.aperture) params.push(displayParams.aperture)
      if (displayParams.shutter) params.push(displayParams.shutter)
      if (displayParams.iso) params.push(displayParams.iso)
      const paramsText = params.join('    ')
      if (paramsText) {
        const paramsWidth = ctx.measureText(paramsText).width
        ctx.fillText(paramsText, borderWidth + width / 2 - paramsWidth / 2, centerY)
      }

      const rightText = [displayParams.lens, displayParams.date].filter(Boolean).join('  ')
      if (rightText) {
        const rightWidth = ctx.measureText(rightText).width
        ctx.fillText(rightText, borderWidth + width - padX - rightWidth, centerY)
      }
    }

    const mimeType = 'image/jpeg'
    const quality = 0.92

    canvas.toBlob(async (blob) => {
      if (!blob) {
        reject(new Error('导出 Blob 失败'))
        return
      }
      if (!frameConfig.showExif || originalMime.value !== 'image/jpeg' || !originalBuffer.value) {
        resolve(blob)
        return
      }
      try {
        const withExif = await injectExif(blob)
        resolve(withExif)
      } catch {
        resolve(blob)
      }
    }, mimeType, quality)
  })
}

async function injectExif(blob: Blob): Promise<Blob> {
  const originalBinary = arrayBufferToBinary(originalBuffer.value!)
  const exifObj = piexif.load(originalBinary)

  // 覆盖/填充 EXIF 字段
  if (!exifObj['0th']) exifObj['0th'] = {}
  if (!exifObj.Exif) exifObj.Exif = {}

  const setTag = (ifd: Record<number, unknown>, tag: string, value: string | number) => {
    const tagNum = piexif.ImageIFD[tag] || piexif.ExifIFD[tag]
    if (!tagNum) return
    if (typeof value === 'string' && value) {
      ifd[tagNum] = value
    }
  }

  setTag(exifObj['0th'], 'Make', displayParams.make)
  setTag(exifObj['0th'], 'Model', displayParams.model)
  setTag(exifObj['0th'], 'LensModel', displayParams.lens)
  setTag(exifObj.Exif, 'FocalLength', displayParams.focalLength)
  setTag(exifObj.Exif, 'FNumber', displayParams.aperture)
  setTag(exifObj.Exif, 'ExposureTime', displayParams.shutter)
  setTag(exifObj.Exif, 'ISOSpeedRatings', displayParams.iso)

  const exifDump = piexif.dump(exifObj)
  const reader = new FileReader()
  return new Promise((resolve, reject) => {
    reader.onload = () => {
      const base64 = reader.result as string
      const dataUrl = piexif.insert(exifDump, base64)
      const newBlob = dataURLToBlob(dataUrl)
      if (!newBlob) {
        reject(new Error('写入 EXIF 失败'))
        return
      }
      resolve(newBlob)
    }
    reader.onerror = reject
    reader.readAsDataURL(blob)
  })
}

function arrayBufferToBinary(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer)
  let binary = ''
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i])
  }
  return binary
}

function dataURLToBlob(dataUrl: string): Blob | null {
  const arr = dataUrl.split(',')
  if (arr.length !== 2) return null
  const mimeMatch = arr[0].match(/:(.*?);/)
  const mime = mimeMatch ? mimeMatch[1] : 'image/jpeg'
  const bstr = atob(arr[1])
  let n = bstr.length
  const u8arr = new Uint8Array(n)
  while (n--) {
    u8arr[n] = bstr.charCodeAt(n)
  }
  return new Blob([u8arr], { type: mime })
}

function formatShutter(sec: number): string {
  if (sec >= 1) return `${sec}s`
  const denom = Math.round(1 / sec)
  return `1/${denom}`
}

function formatDate(dateInput: string | Date): string {
  const d = typeof dateInput === 'string' ? new Date(dateInput) : dateInput
  if (Number.isNaN(d.getTime())) return String(dateInput)
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}.${month}.${day}`
}

function formatBytes(bytes?: number): string {
  if (bytes == null || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

function handleClose() {
  emit('update:visible', false)
}
</script>

<style scoped>
.workbench {
  display: grid;
  grid-template-columns: 300px 1fr 300px;
  gap: 16px;
  min-height: 600px;
}

.panel {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #f0f0f0);
  border-radius: 8px;
  padding: 16px;
  overflow-y: auto;
  max-height: 720px;
}

.section {
  margin-bottom: 20px;
}

.section-title {
  font-weight: 600;
  margin-bottom: 10px;
  color: var(--text-primary, #262626);
}

.template-group {
  display: flex;
  width: 100%;
}

.template-group :deep(.ant-radio-button-wrapper) {
  flex: 1;
  text-align: center;
}

.param-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
  font-size: 13px;
  color: var(--text-secondary, #595959);
}

.placeholder-tip {
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-secondary, #999);
}

.preview-area {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}

.canvas-wrapper {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-secondary, #f5f5f5);
  border-radius: 8px;
  overflow: auto;
  min-height: 480px;
}

.preview-canvas {
  max-width: 100%;
  max-height: 620px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
}

.preset-bar {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #f0f0f0);
  border-radius: 8px;
  padding: 12px;
}

.preset-bar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-weight: 500;
}

.preset-thumbnails {
  display: flex;
  gap: 8px;
  overflow-x: auto;
}

.preset-thumb {
  position: relative;
  flex-shrink: 0;
  width: 80px;
  height: 80px;
  border-radius: 4px;
  overflow: hidden;
  cursor: pointer;
  border: 1px solid var(--border-color, #f0f0f0);
}

.preset-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.thumb-name {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 2px 4px;
  font-size: 11px;
  line-height: 1.4;
  color: #fff;
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  background: linear-gradient(to top, rgba(0, 0, 0, 0.75), rgba(0, 0, 0, 0));
  pointer-events: none;
}

.thumb-delete {
  position: absolute;
  top: 2px;
  right: 2px;
  padding: 2px 4px;
  background: rgba(255, 255, 255, 0.8);
  border-radius: 4px;
}

.exif-list {
  background: var(--bg-secondary, #f5f5f5);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 12px;
}

.exif-item {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  line-height: 2;
}

.exif-label {
  color: var(--text-secondary, #8c8c8c);
}

.exif-value {
  color: var(--text-primary, #262626);
  max-width: 60%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.exif-cards {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.exif-card {
  background: var(--bg-secondary, #f5f5f5);
  border-radius: 8px;
  padding: 10px;
  text-align: center;
}

.exif-card-label {
  font-size: 11px;
  color: var(--text-secondary, #8c8c8c);
  margin-bottom: 4px;
}

.exif-card-value {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary, #262626);
}
</style>
