/**
 * 文件上传工具：accept 写法与前端校验统一入口
 */

/**
 * 移动端兼容要求 accept 必须使用通配 MIME（image/*、video/*），
 * 具体扩展名/ MIME 列表（如 "image/png, image/jpeg"）会导致部分安卓
 * 浏览器（微信内置等）只弹文件管理器不弹图库，iOS 相册中 HEIC 照片置灰。
 */
export const IMAGE_ACCEPT = 'image/*'
export const VIDEO_ACCEPT = 'video/*'

/** 上传大小上限，与后端 media_controller 的 100MB 限制保持一致 */
export const MAX_UPLOAD_SIZE = 100 * 1024 * 1024

/**
 * 取文件小写扩展名（不含点），无扩展名时回退解析 MIME subtype。
 */
export function getFileExtension(file: File): string {
  const name = file.name || ''
  const dotIndex = name.lastIndexOf('.')
  if (dotIndex >= 0 && dotIndex < name.length - 1) {
    return name.slice(dotIndex + 1).toLowerCase()
  }
  const type = file.type || ''
  if (type.includes('/')) {
    return type.split('/')[1]?.toLowerCase() || ''
  }
  return ''
}

/**
 * 按扩展名白名单校验图片文件（accept 放宽为通配后，前端接管格式拦截）。
 *
 * @param file 待校验文件
 * @param allowedExts 允许的扩展名列表，如 ['jpg', 'jpeg', 'png', 'webp']
 * @returns 校验通过返回 null，否则返回错误提示文案
 */
export function validateImageFile(file: File, allowedExts: string[]): string | null {
  if (file.size > MAX_UPLOAD_SIZE) {
    return '文件大小不能超过 100MB'
  }
  const ext = getFileExtension(file)
  if (!allowedExts.includes(ext)) {
    return `仅支持 ${allowedExts.map((e) => e.toUpperCase()).join('/')} 格式`
  }
  return null
}
