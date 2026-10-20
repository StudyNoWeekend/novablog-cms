#!/usr/bin/env node
// check-i18n 校验各语言包与 zh-CN（基准语言）的命名空间文件和 key 完全一致，
// 并扫描源码中的静态 t('...') / $t('...') 引用，确认 key 在 zh-CN 中存在。
// 用法：node scripts/check-i18n.mjs（或 npm run check:i18n）
import { readdirSync, readFileSync, existsSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

const LOCALES_DIR = join(dirname(fileURLToPath(import.meta.url)), '..', 'src', 'locales')
const SRC_DIR = join(dirname(fileURLToPath(import.meta.url)), '..', 'src')
const BASE_LOCALE = 'zh-CN'
const LOCALES = ['zh-CN', 'zh-TW', 'en', 'ja', 'ko']

// flatten 将嵌套 JSON 的叶子节点展平为「a.b.c」形式的 key 集合
function flatten(obj, prefix = '', out = new Set()) {
  for (const [key, value] of Object.entries(obj)) {
    const fullKey = prefix ? `${prefix}.${key}` : key
    if (value !== null && typeof value === 'object') {
      flatten(value, fullKey, out)
    } else {
      out.add(fullKey)
    }
  }
  return out
}

function collectErrors() {
  const errors = []
  const baseDir = join(LOCALES_DIR, BASE_LOCALE)
  if (!existsSync(baseDir)) return [`[目录缺失] 基准语言 ${BASE_LOCALE}/ 不存在`]
  const baseFiles = readdirSync(baseDir).filter((f) => f.endsWith('.json'))

  for (const locale of LOCALES) {
    const dir = join(LOCALES_DIR, locale)
    if (!existsSync(dir)) {
      errors.push(`[${locale}] 语言目录不存在`)
      continue
    }
    const files = readdirSync(dir).filter((f) => f.endsWith('.json'))

    for (const file of baseFiles) {
      if (!files.includes(file)) {
        errors.push(`[${locale}] 缺少命名空间文件 ${file}`)
        continue
      }
      let baseKeys
      let keys
      try {
        baseKeys = flatten(JSON.parse(readFileSync(join(baseDir, file), 'utf8')))
        keys = flatten(JSON.parse(readFileSync(join(dir, file), 'utf8')))
      } catch (err) {
        errors.push(`[${locale}/${file}] JSON 解析失败: ${err.message}`)
        continue
      }
      for (const key of baseKeys) {
        if (!keys.has(key)) errors.push(`[${locale}/${file}] 缺少 key: ${key}`)
      }
      for (const key of keys) {
        if (!baseKeys.has(key)) errors.push(`[${locale}/${file}] 多余 key: ${key}`)
      }
    }

    for (const file of files) {
      if (!baseFiles.includes(file)) {
        errors.push(`[${locale}] 多余命名空间文件 ${file}（${BASE_LOCALE} 中不存在）`)
      }
    }
  }
  return errors
}

// listSourceFiles 递归收集 src 下的 .ts/.vue 源文件（排除语言包自身）
function listSourceFiles(dir, out = []) {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const full = join(dir, entry.name)
    if (entry.isDirectory()) {
      if (full === LOCALES_DIR) continue
      listSourceFiles(full, out)
    } else if (/\.(ts|vue)$/.test(entry.name)) {
      out.push(full)
    }
  }
  return out
}

// collectStaticKeys 扫描源码中 t('...') / $t('...') 的字面量 key（跳过模板字符串与变量）
function collectStaticKeys() {
  const keyPattern = /[$\s({[,!:]t\(\s*'([^']+)'/g
  const usages = new Map()
  for (const file of listSourceFiles(SRC_DIR)) {
    const content = readFileSync(file, 'utf8')
    let match
    while ((match = keyPattern.exec(content)) !== null) {
      const key = match[1]
      if (!usages.has(key)) usages.set(key, [])
      usages.set(key, [...usages.get(key), file])
    }
  }
  return usages
}

const errors = collectErrors()
if (errors.length > 0) {
  console.error(`✗ i18n 校验失败，共 ${errors.length} 处与 ${BASE_LOCALE} 不一致：`)
  for (const error of errors) console.error(`  ${error}`)
  process.exit(1)
}
console.log(`✓ i18n 校验通过：${LOCALES.join(' / ')} 的命名空间文件与 key 和 ${BASE_LOCALE} 完全一致`)

// 静态 key 引用检查：t() 使用的字面量 key 必须存在于 zh-CN（否则运行期回退会掩盖拼写错误）
const baseDir = join(LOCALES_DIR, BASE_LOCALE)
const baseKeys = new Set()
for (const file of readdirSync(baseDir).filter((f) => f.endsWith('.json'))) {
  for (const key of flatten(JSON.parse(readFileSync(join(baseDir, file), 'utf8')))) {
    baseKeys.add(`${file.replace(/\.json$/, '')}.${key}`)
  }
}
const usages = collectStaticKeys()
const missingUsages = [...usages.entries()].filter(([key]) => !baseKeys.has(key))
if (missingUsages.length > 0) {
  console.error(`✗ 源码引用了 ${missingUsages.length} 个 zh-CN 中不存在的 i18n key：`)
  for (const [key, files] of missingUsages) {
    console.error(`  ${key}  ←  ${files.map((f) => f.replace(SRC_DIR + '/', '')).join(', ')}`)
  }
  process.exit(1)
}
console.log(`✓ 源码静态 t() 引用检查通过（共 ${usages.size} 个 key）`)
