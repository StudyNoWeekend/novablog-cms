/**
 * 图片处理工具函数
 */

/**
 * 判断 URL 是否为腾讯云 COS 域名。
 * COS 域名格式：xxx.cos.xx.myqcloud.com 或自定义域名指向 COS。
 */
function isCosUrl(url: string): boolean {
  return url.includes('.myqcloud.com')
}

/**
 * 为腾讯云 COS 图片 URL 生成缩略图 URL。
 * 利用 COS 数据万象 imageMogr2 接口，在 URL 后追加参数实现服务端动态缩放。
 * 非 COS URL 原样返回。
 *
 * @param url 原始图片 URL
 * @param width 缩略图目标宽度（像素）
 * @returns 缩略图 URL
 */
export function getThumbUrl(url: string, width: number): string {
  if (!url || !isCosUrl(url)) {
    return url
  }
  const separator = url.includes('?') ? '&' : '?'
  return `${url}${separator}imageMogr2/thumbnail/${width}x/format/webp/q/80`
}
