# i18n 国际化规范

本目录是前端多语言（i18n）的唯一中心。支持 5 种语言：`zh-CN`（简体中文，默认/回退语言）、`zh-TW`（繁體中文）、`en`（English）、`ja`（日本語）、`ko`（한국어）。

## 目录结构

```
src/locales/
├── index.ts      # createI18n 装配、语言常量（SUPPORTED_LOCALES / LOCALE_META / AppLocale）
├── README.md     # 本规范
├── zh-CN/        # 基准语言（source of truth），每个 JSON 文件 = 一个命名空间
│   ├── common.json   # 全局共享文案（按钮/状态/提示/通用字段）
│   ├── menu.json     # 侧边栏、路由标题、面包屑
│   └── <域>.json     # 其余命名空间按业务域拆分（article/media/system...）
├── zh-TW/ en/ ja/ ko/   # 与 zh-CN 同名同构，key 必须严格一致
```

命名空间与业务域一一对应，新增页面优先放入已有命名空间，确属新域时新建 `<域>.json`（5 个语言目录同步创建）。

## Key 命名规范

- 完整 key = `命名空间.模块.语义`，如 `article.list.title`、`common.actions.confirm`。
- JSON 内嵌套最多三层，camelCase，按「模块 → 语义」组织，禁止用界面位置命名。
- **共享收敛原则**：确定/取消/新增/编辑/删除/操作/状态/创建时间/操作成功/暂无数据等通用文案一律使用 `common.*`，禁止在各命名空间重复定义；仅在该视图独有的文案放入业务命名空间。
- 插值用 `{name}`（如 `common.message.deleteConfirmNamed` 的 `{name}`）；**值中禁止出现裸的 `@` 与 `|`**（vue-i18n 链接消息/复数语法），需要时写成 `{'@'}`。
- 路由 `meta.title`、侧边栏菜单的 `labelKey`/`titleKey` 一律使用 `menu.*` key，两处共用同一批 key。

## 代码中使用规范

```ts
// script 中（推荐显式引入，保持组件自包含）
import { useI18n } from 'vue-i18n'
const { t } = useI18n()

// 模板中可直接用全局注入的 $t
```

1. **响应式要求**：凡 `t()` 出现在 script 常量中的场景（表格 `columns`、下拉 `options`、映射表），必须包成 `computed(() => ...)`，否则切换语言不刷新；禁止在模块顶层（组件外）调用 `t()`。
2. **命令式弹窗**：`message.success(t('common.message.operationSuccess'))`、`Modal.confirm({ title: t('...') })` 在调用时求值，天然跟随当前语言。
3. **日期/时间**：统一使用 `@/utils/datetime` 的 `formatDate / formatDateTime / formatTime`，禁止手写 `toLocaleString('zh-CN')` 等硬编码 locale。
4. **动态拼接 key**：形如 ``t(`article.status.${status}`)`` 允许，但状态枚举必须在 5 个语言包中同时存在，否则回退显示中文。
5. **后端动态数据**（文章标题、分类名等）不属于 i18n 范围，不翻译。

## 新增/修改文案流程

1. 在 `zh-CN/<命名空间>.json` 增改 key；
2. **同步**更新其余 4 个语言目录的同名文件（zh-TW 按台湾惯用词汇，如 設定/搜尋/資料/網路/影片）；
3. 运行 `npm run check:i18n`，确保 5 语言 key 完全一致；
4. 缺失的翻译会回退显示 zh-CN（`fallbackLocale`），dev 模式控制台有 missing/fallback 警告。

## 新增语言流程

1. `src/locales/index.ts`：`SUPPORTED_LOCALES`、`LOCALE_META` 各加一项（dayjs 无该语言包时 `dayjsLocale` 填 `'en'`）；
2. 复制 `zh-CN/` 全部 JSON 到新目录并翻译；
3. `src/App.vue` 的 `ANTD_LOCALES` 增加对应 antd locale（无对应包时用 `enUS`）；
4. `scripts/check-i18n.mjs` 的 `LOCALES` 数组加入新语言；
5. `src/components/common/LocaleSwitch.vue` 无需改动（自动遍历 `SUPPORTED_LOCALES`）。

## 校验

```bash
npm run check:i18n   # 5 语言文件与 key 一致性（CI/发版前必跑）
npm run build        # vue-tsc 类型检查 + vite 构建
```
