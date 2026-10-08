<template>
  <a-modal
    :open="open"
    title="裁剪头像"
    :width="440"
    :mask-closable="false"
    :confirm-loading="cropping || confirmLoading"
    ok-text="确认"
    cancel-text="取消"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <div class="avatar-cropper">
      <div class="crop-container">
        <img ref="imgRef" :src="objectUrl" alt="头像裁剪预览" />
      </div>
      <div class="crop-actions">
        <a-tooltip title="缩小">
          <a-button size="small" @click="zoom(-0.1)"><ZoomOutOutlined /></a-button>
        </a-tooltip>
        <a-tooltip title="放大">
          <a-button size="small" @click="zoom(0.1)"><ZoomInOutlined /></a-button>
        </a-tooltip>
        <a-tooltip title="向左旋转">
          <a-button size="small" @click="rotate(-90)"><UndoOutlined /></a-button>
        </a-tooltip>
        <a-tooltip title="向右旋转">
          <a-button size="small" @click="rotate(90)"><RedoOutlined /></a-button>
        </a-tooltip>
        <a-tooltip title="重置">
          <a-button size="small" @click="reset"><ReloadOutlined /></a-button>
        </a-tooltip>
      </div>
      <p class="crop-hint">拖动图片调整位置，滚轮或按钮缩放</p>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import { nextTick, ref, watch, onBeforeUnmount } from 'vue'
import { message } from 'ant-design-vue'
import Cropper from 'cropperjs'
import 'cropperjs/dist/cropper.css'
import {
  ZoomInOutlined,
  ZoomOutOutlined,
  UndoOutlined,
  RedoOutlined,
  ReloadOutlined,
} from '@ant-design/icons-vue'

const props = defineProps<{
  open: boolean
  /** 待裁剪的原始图片文件 */
  file: File | null
  /** 父组件上传期间的加载态，透传到确认按钮 */
  confirmLoading?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'confirm', blob: Blob): void
}>()

/** 裁剪输出尺寸：1:1 正方形，展示时由样式渲染为圆形 */
const OUTPUT_SIZE = 512

const imgRef = ref<HTMLImageElement | null>(null)
const objectUrl = ref('')
const cropping = ref(false)
let cropper: Cropper | null = null

watch(
  () => [props.open, props.file] as const,
  ([open, file]) => {
    if (open && file) {
      initCropper()
    } else if (!open) {
      destroyCropper()
    }
  },
  { flush: 'post' },
)

async function initCropper() {
  if (!props.file) return
  destroyCropper()
  objectUrl.value = URL.createObjectURL(props.file)
  // 等待 <img> 的 src 渲染完成后再初始化，否则 Cropper 读到空 src
  await nextTick()
  if (!props.open || !imgRef.value) return
  cropper = new Cropper(imgRef.value, {
    aspectRatio: 1,
    viewMode: 1,
    dragMode: 'move',
    autoCropArea: 1,
    cropBoxResizable: false,
    cropBoxMovable: false,
    toggleDragModeOnDblclick: false,
    background: false,
  })
}

function destroyCropper() {
  if (cropper) {
    cropper.destroy()
    cropper = null
  }
  if (objectUrl.value) {
    URL.revokeObjectURL(objectUrl.value)
    objectUrl.value = ''
  }
}

function zoom(ratio: number) {
  cropper?.zoom(ratio)
}

function rotate(deg: number) {
  cropper?.rotate(deg)
}

function reset() {
  cropper?.reset()
}

function handleOk() {
  if (!cropper || cropping.value || props.confirmLoading) return
  cropping.value = true
  const canvas = cropper.getCroppedCanvas({
    width: OUTPUT_SIZE,
    height: OUTPUT_SIZE,
    imageSmoothingEnabled: true,
    imageSmoothingQuality: 'high',
  })
  if (!canvas) {
    message.error('图片裁剪失败，请重试')
    cropping.value = false
    return
  }
  canvas.toBlob((blob) => {
    cropping.value = false
    if (!blob) {
      message.error('图片导出失败，请重试')
      return
    }
    emit('confirm', blob)
  }, 'image/png')
}

function handleCancel() {
  emit('update:open', false)
}

onBeforeUnmount(destroyCropper)
</script>

<style scoped>
.avatar-cropper {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* 固定高度容器，cropperjs 依据其尺寸初始化画布 */
.crop-container {
  height: 300px;
  overflow: hidden;
  background: var(--bg-card, #f5f5f5);
}

.crop-container :deep(img) {
  display: block;
  max-width: 100%;
}

/* QQ 头像风格：1:1 裁剪框渲染为圆形预览 */
.crop-container :deep(.cropper-view-box),
.crop-container :deep(.cropper-face) {
  border-radius: 50%;
}

.crop-actions {
  display: flex;
  justify-content: center;
  gap: 8px;
}

.crop-hint {
  margin: 0;
  font-size: 12px;
  color: var(--text-color-secondary, #999);
  text-align: center;
}
</style>
