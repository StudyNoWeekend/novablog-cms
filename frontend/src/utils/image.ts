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

/**
 * 构建歌曲封面候选地址列表，按序降级尝试。
 * 1. 缩略处理后的 URL（COS 追加 imageMogr2 参数）
 * 2. 原始 URL 原样（自托管 http:// 地址不做协议改写，B 站外链由 no-referrer 绕过防盗链）
 * 3. 后端图片代理（仅外部 URL，服务端取图绕过防盗链）
 *
 * 注意：不要对 http:// 做 https 强转——本站封面转存后的地址跟随 upload.base_url，
 * 强转会导致无 TLS 的自托管地址加载失败。
 */
export function buildSongCoverCandidates(url: string, width: number): string[] {
  if (!url) {
    return []
  }
  const candidates = [getThumbUrl(url, width), url]
  if (/^https?:\/\//i.test(url)) {
    candidates.push(`/api/v1/image-search/thumbnail?url=${encodeURIComponent(url)}`)
  }
  return candidates
}
