// 解析 SKILL.md 格式的 novablog API 文档，输出与后端 /api/v1/api-docs 相同的结构。
// skill.md 的典型段落结构：
//   ### 6.3 文章模块（公开）                  -> 模块分组
//   #### GET /api/v1/public/articles           -> 接口（方法 + 路径）
//   获取已发布文章列表                          -> description
//   **查询参数:**                              -> 参数表格
//   | 参数 | 类型 | 必填 | 默认值 | 说明 |      -> 表格
//   **响应 data 字段（分页）:**                -> 响应字段表格
//   | 字段 | 类型 | 说明 |                    -> 表格

export interface APIDocParam {
  name: string
  type: string
  required: boolean
  desc: string
}

export interface APIDocField {
  name: string
  type: string
  desc: string
}

export interface APIDocItem {
  module: string
  method: string
  path: string
  description: string
  params: APIDocParam[]
  response: APIDocField[]
}

/** 拆分 markdown 表格行，去掉首尾管道与反引号。空行或无管道返回空数组。 */
function splitRow(line: string): string[] {
  if (!line.trim().startsWith('|')) return []
  const cells = line.trim().replace(/^\||\|$/g, '').split('|')
  return cells.map((c) => c.trim().replace(/`/g, ''))
}

/** 表格分隔行（表头下的 --- 行）。 */
function isSeparatorRow(cells: string[]): boolean {
  if (!cells.length) return false
  return cells.every((c) => /^:?-{2,}:?$/.test(c))
}

type Table = { rows: string[][]; next: number }

/** 解析表格：start 指向表头行（可跳过其前的空行）。返回数据行与下一条语句行号（表格结束后第一个非表格行）。 */
function parseTable(lines: string[], start: number): Table {
  // 跳过表头前的空行
  while (start < lines.length && !splitRow(lines[start]).length) start++
  const header = splitRow(lines[start] ?? '')
  const empty: Table = { rows: [], next: start }
  if (!header.length) return empty

  let i = start + 1
  const sep = splitRow(lines[i] ?? '')
  if (isSeparatorRow(sep)) i++

  const rows: string[][] = []
  while (i < lines.length && splitRow(lines[i]).length) {
    rows.push(splitRow(lines[i]))
    i++
  }
  return { rows, next: i }
}

/** 构建参数列表。必填列缺失时统一视为非必填（查询参数）；路径参数表列头无"必填"列为 3 列）。 */
function buildParams(rows: string[][]): APIDocParam[] {
  const out: APIDocParam[] = []
  for (const r of rows) {
    const name = r[0]?.trim()
    if (!name) continue
    // 参数表列：参数 类型 必填 默认值 说明  /  参数 类型 必填 说明
    // 必填列存在且为"是"则必填；否则路径参数（说明在第 3 列）按"是"处理
    const hasRequiredCol = r.length >= 5
    const required = hasRequiredCol ? r[2] === '是' : r[2] === '是' || r.length === 4
    const desc = hasRequiredCol ? (r[4] ?? '') : (r[3] ?? '')
    out.push({ name, type: r[1] ?? '', required, desc: desc.replace(/^[（(]?是[)）]?,?/, '') })
  }
  return out
}

/** 构建响应字段列表。 */
function buildFields(rows: string[][]): APIDocField[] {
  const out: APIDocField[] = []
  for (const r of rows) {
    const name = r[0]?.trim()
    if (!name) continue
    out.push({ name, type: r[1] ?? '', desc: r[2] ?? '' })
  }
  return out
}

/** 解析整份 skill.md，返回与 /api/v1/api-docs 一致的 APIDocItem 数组。 */
export function parseApiDocMarkdown(md: string): APIDocItem[] {
  const lines = md.split(/\r?\n/)
  const items: APIDocItem[] = []
  let module = ''

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i]
    const trimmed = line.trim()

    // 模块标题：### 6.x 模块名（公开）
    const m = trimmed.match(/^###\s+\d+\.[\d-]+\s+(.+)$/)
    if (m) {
      module = m[1].replace(/（公开）$/, '').replace(/[：:]\s*$/, '').trim()
      continue
    }

    // 接口标题：#### GET /api/v1/public/xxx
    const a = trimmed.match(/^####\s+([A-Z]+)\s+(\/api\/v1\/public\/\S+)/)
    if (!a) continue

    const item: APIDocItem = {
      module,
      method: a[1].toUpperCase(),
      path: a[2].replace(/^\/api\/v1\/public/, ''),
      description: '',
      params: [],
      response: [],
    }
    items.push(item)

    // 逐行扫描接口内容，直到下一个 #####/####/### 或文件尾。
    let j = i + 1
    while (j < lines.length) {
      const t = lines[j].trim()
      if (/^###/.test(t) || /^####/.test(t)) break

      // description：接口标题后的第一个普通文本行（非表、非标题、非加粗小节、非代码块、非列表）
      const looksLikeDesc =
        t && !t.startsWith('|') && !t.startsWith('#') && !t.startsWith('**') &&
        !t.startsWith('```') && !t.startsWith('>') && !t.startsWith('-') && !t.startsWith(':::')
      if (!item.description && looksLikeDesc) {
        item.description = t
      }

      // 参数表格
      const pH = t.match(/^\*\*[\s\S]*(查询参数|路径参数)[\s\S]*[：:]\s*\*\*$/)
      if (pH) {
        const { rows, next } = parseTable(lines, j + 1)
        if (rows.length) item.params = item.params.concat(buildParams(rows))
        j = next
        continue
      }

      // 响应表格：**响应 data 字段（分页）:** / **响应 data 字段（数组）:** / **list 中每项字段:** /
      // **items 中每项字段:** / **platforms 中每项字段:** / **响应 data:**（后跟 null 或文本）。
      // 排除 **响应示例:** 与 **响应 data 示例:**。
      const rH = t.match(/^\*\*(?:响应|[\s\S]+?\s+中每项字段)[\s\S]*?[：:]\s*\*\*$/)
      const rHAlt = t.match(/^\*\*响应（[\s\S]+?）[：:]\s*\*\*$/)
      const isRespExample = /^\*\*响应[^：:]*示例\s*[：:]\s*\*\*$/.test(t)
      if ((rH || rHAlt) && !isRespExample) {
        const { rows, next } = parseTable(lines, j + 1)
        if (rows.length) {
          const fields = buildFields(rows)
          // 子表（xxx 中每项字段）并入响应字段：在主字段上追加结构说明
          const subTitle = t.match(/^\*\*(\S+)\s+中每项字段/)
          if (subTitle && item.response.length) {
            const last = item.response[item.response.length - 1]
            last.desc = `${last.desc}：${fields.map((f) => f.name).join(', ')}`
          } else {
            item.response = item.response.concat(fields)
          }
          j = next
        } else {
          // 无表格（`**响应 data:** `null`` 等）：跳过后继非语句行即可
          j = next
        }
        continue
      }

      // 纯 "**响应:** 描述"（无表格）
      const rN = t.match(/^\*\*响应[：:]\s*\*\*\s*([\s\S]*)$/)
      if (rN && !item.response.length) {
        item.description = item.description || rN[1].trim()
        j++
        continue
      }

      j++
    }
    i = j - 1
  }

  return items
}
