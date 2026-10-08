# NovaBlog 旅行攻略地图展示方案

> 状态：已实施（v1.1.1 起）　|　关联文档：[novablog-api.md](frontend/public/docs/novablog-api.md) 6.6 旅行攻略模块、[novablog-主题包规范与发布指南.md](novablog-主题包规范与发布指南.md)

---

## 一、背景与目标

管理后台的旅行攻略编辑器已支持为景点标注坐标（「选择定位」弹窗：中国区高德、海外 Google Maps），坐标随 `attractions` JSONB 原样透出给公开 API。本方案解决「最后一公里」：

1. **主题端展示**：novablog-web 在攻略详情页展示一张地图，标注全部含坐标的景点位置；
2. **访客跳转**：访客点击景点/标记，自动打开对应的地图平台（高德 / Google）并带上坐标，移动端可唤起原生 App。

## 二、现状与差距（方案设计前的约束）

| # | 约束 | 影响 |
|---|------|------|
| 1 | 地图 Key 是管理端编译期环境变量（`VITE_AMAP_KEY` 等），主题构建制品拿不到 | 主题内直接用商业地图 SDK 不可行 |
| 2 | 主题包规范 §3.4 要求壳页面自包含，仅允许引用 `/theme-config.js`，禁止第三方 `<script>` | 壳页内引 Leaflet/高德 SDK 违反规范 |
| 3 | 高德为 GCJ-02、Google 为 WGS-84，同一攻略内可能混存两种坐标系 | 展示与跳转都必须感知坐标系，否则国内点偏移约 500m |
| 4 | 历史景点数据可能只有 `location` 文本，无经纬度；JSONB 字段名可能存在驼峰/下划线差异 | 展示端必须容错降级 |
| 5 | 主题仓库有 11+ 风格各异的主题 | 逐主题实现地图的重复成本高、一致性差 |

## 三、架构选型

**采用：CMS 嵌入页收口（map-embed iframe）。**

```
novablog-web 主题详情页
      │  <iframe src="{CMS}/map-embed/travel/{id}">
      ▼
CMS 公开路由 /map-embed/travel/:id（自包含 HTML，go:embed 随二进制分发）
      │  客户端 fetch /api/v1/public/travels/:id（同源，免认证）
      ▼
Leaflet + OpenStreetMap 免 key 底图
   ├─ gcj02 点：先纠偏为 WGS-84 再上图
   └─ 点位 popup / 列表：按坐标系自动生成跳转外链（高德 / Google）
```

**理由**：

- **一处实现，全部主题受益**：11+ 主题只需一个 `<iframe>`，不重复实现地图、纠偏、跳转逻辑；
- **不触碰主题包规范**：SDK 与资产全部由 CMS 同源提供，壳页面零第三方引用；
- **免 key 免配额**：底图用 OSM 开放瓦片，跳转用各平台官方 URL Scheme（`uri.amap.com` / `google.com/maps`），无需任何 Key、不消耗商业配额；
- **坐标系收口**：GCJ-02 纠偏、平台匹配逻辑集中在 CMS 嵌入页，主题永远不感知坐标系差异。

**被否方案**：

- *主题内自行集成 Leaflet*——需修订壳页规范且 11+ 主题重复实现；
- *高德/Google JSAPI 直出*——Key 无法下发（见约束 1），且访客端配额成本失控；
- *CMS 生成静态图代理*——需要后端配置商业静态图 Key，且失去交互与跳转能力；
- *纯点位列表无底图*——不满足「地图式展示」的产品诉求。

## 四、数据契约

### 4.1 景点坐标字段（`attractions[]`，JSONB 原样透出）

| 字段 | 类型 | 必有 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 景点名称 |
| `location` | string | 选点后必有 | 定位文本地址 |
| `latitude` / `longitude` | number | 选点后必有 | 存储坐标，坐标系见 `coordType` |
| `coordType` | string | v1.2.0 起选点必写 | `gcj02`（高德选点）/ `wgs84`（Google 选点）；**历史数据可能缺失** |
| `image` / `duration` / `description` | string | 否 | 其他字段；注意 JSONB 内的 `image` **不做** URL 拼接 |

> **容错约定**：读取端必须兼容 `coordType`（驼峰）与 `coord_type`（下划线）两种键名，以及 `latitude/longitude` 缺失的情况（见 API 文档 6.6）。CMS 嵌入页已内置该容错。

### 4.2 坐标系语义

- `coordType: "gcj02"`：坐标适用于高德/腾讯系底图与跳转；在 WGS-84 底图（OSM/Google）上展示前必须纠偏；
- `coordType: "wgs84"`：坐标适用于 OSM/Google 底图与跳转；
- **缺失时**（旧数据）：无法判定坐标系，嵌入页按原始值直接上图与跳转（宁可接受可能的偏移，不做猜测转换）。

## 五、嵌入页设计（`GET /map-embed/travel/:id`）

- **响应**：自包含 HTML（Leaflet 资产由 `/map-embed/assets/*` 同源提供，`go:embed` 打包，不依赖 CDN）；
- **取数**：页面脚本 fetch 同源 `/api/v1/public/travels/:id`，仅消费公开只读接口；
- **上图规则**：仅渲染含合法经纬度的景点；`gcj02` 点先做 GCJ-02→WGS-84 纠偏（公开近似算法，内联约 40 行，误差 1-2m）；自动 `fitBounds`；
- **跳转规则**：popup 与列表中的外链使用**原始存储坐标**（跳转目标平台消费自己的坐标系），按 `coordType` 自动选平台，见下方 URI 约定；
- **降级**：无任何带坐标景点时展示提示与纯列表；无坐标景点在列表中提供「按地名搜索」跳转；
- **查询参数**：`h`（嵌入高度提示 px，默认 420，范围 200-2000）；`lang`（预留）；
- **高度自适应**：渲染后与窗口尺寸变化时 `postMessage({ type: 'novablog-map-embed:height', height }, '*')`，主题可监听调整 iframe 高度；
- **安全**：页面为纯静态模板，动态内容全部经 `textContent` 注入；外链仅 `https` 且 `rel="noopener noreferrer"`；
- **瓦片合规**：保留 OSM attribution（Leaflet 默认图层自带）。

### 跳转 URI 约定

| 坐标系 | 平台 | URI 模板 |
|--------|------|----------|
| `gcj02` | 高德（marker，移动端 `callnative=1` 可唤起 App） | `https://uri.amap.com/marker?position={lng},{lat}&name={name}&src=novablog&callnative=1` |
| `gcj02` | 腾讯（备选，主题自治场景可用） | `https://apis.map.qq.com/uri/v1/marker?marker=coord:{lat},{lng};title:{name}&referer=novablog` |
| `wgs84` | Google Maps | `https://www.google.com/maps/search/?api=1&query={lat},{lng}` |
| `wgs84` | Apple 地图（iOS 访客友好，备选） | `https://maps.apple.com/?ll={lat},{lng}&q={name}` |
| 无坐标 | 按地名搜索 | 高德：`https://uri.amap.com/search?keyword={location}&src=novablog`；Google：`https://www.google.com/maps/search/?api=1&query={location}` |

## 六、主题接入指南（novablog-web）

1. **iframe 地址拼接**：与取数 apiBase 同源规则——`apiBase` 为空（同域部署）时直接用相对路径；非空时剥掉尾部 `/api/v1` 后拼接：

```js
const cmsBase = (window.__NOVA_CONFIG__?.apiBase || '').replace(/\/api\/v1\/?$/, '');
const src = `${cmsBase}/map-embed/travel/${guideId}?h=420`;
```

2. **嵌入示例**：

```html
<iframe
  src="/map-embed/travel/550e8400-e29b-41d4-a716-446655440000?h=420"
  style="width:100%;border:0;border-radius:12px"
  loading="lazy"
  title="景点地图"
></iframe>
```

3. **（可选）高度自适应**：

```js
window.addEventListener('message', (e) => {
  if (e.data?.type === 'novablog-map-embed:height') {
    iframe.style.height = e.data.height + 'px';
  }
});
```

4. **容错**：iframe 加载失败或超时应隐藏整个地图区块（API 层容错降级为规范 §3.5 通用要求）；攻略无坐标景点时嵌入页会自行展示提示与列表降级，主题无需重复判断。

## 七、实施记录

### v1.2.0

| 变更 | 位置 |
|------|------|
| 选点输出增加 `coordType`（高德→gcj02、Google→wgs84） | `frontend/src/composables/mapTypes.ts`、`LocationPickerModal.vue`、`TravelAttractionPool.vue`、`types/travel.ts` |
| 地图嵌入页 + 资产（Leaflet 1.9.4 vendored） | `backend/assets/mapembed/`（`travel.html` + `embed.go`）、`backend/internal/controller/map_embed_controller.go` |
| 根级公开路由 `/map-embed/travel/:id`、`/map-embed/assets/:file` | `backend/internal/router/map_embed.go`（注册于 `router.go`，显式路由优先于主题托管 NoRoute 兜底） |
| 单元测试 | `backend/internal/controller/map_embed_controller_test.go` |

### v1.2.x（地图服务配置化）

| 变更 | 位置 |
|------|------|
| 新表 `map_configs`（单行；`amap_key`、`amap_security_code`〔AES-GCM 加密〕、`google_key`） | `migrations/033_map_config.{up,down}.sql`、`backend/internal/model/map_config.go` |
| 管理端接口 `GET/PUT /api/v1/map-config`（authMiddleware；GET/PUT 返回真实值，安全密钥读取时解密） | `map_config_req.go`/`res`、`map_config_logic.go`、`map_config_controller.go`、`map_config.go` |
| 后台「系统配置 → 地图配置」页面（三个 Key 输入 + 保存） | `frontend/src/views/map/MapConfigView.vue`、`src/api/map.ts`（带模块级缓存）、`src/router/routes.ts` |
| 侧边栏「系统」分组内新增二级子菜单「系统配置」（模块管理/对象存储/跨域配置/地图配置） | `frontend/src/components/layout/AppSidebar.vue`、`src/utils/storage.ts` |
| 选点器改运行时读 Key：后端配置优先、编译期 env 回退；未配置/加载失败降级 OSM 免 Key 模式 | `useAmap.ts`/`useGoogleMap.ts`/`LocationPickerModal.vue` |

**说明**：地图 Key（`amap_key`/`google_key`）明文落库并返回给登录后台（仅后台使用，不向公开接口/主题下发）；安全密钥 `amap_security_code` 加密存储、读时解密返回明文——暴露面与原编译期 `.env` 一致。博客端 map-embed 展示页始终使用免 Key OSM 方案，不受此配置影响。

**已知限制（已加固）**：高德 Key 无效或安全密钥（securityJsCode）缺失时，JSAPI 能加载但地图永远渲染不出首屏瓦片——选点器现已监听 `complete` 事件并设 6 秒超时，超时自动降级 OSM 免 Key 模式（实测有效）；要获得高德底图与 POI 搜索，仍需在「系统配置 → 地图配置」补全与 Key 配套的安全密钥，并确认 Key 类型为「Web端(JS API)」。

## 八、演进路线（仍待实施）

1. **（已实施，见上）后台地图服务可视化配置**；
2. **（已实施，见上）无效 Key 渲染失败自动降级**：监听 `complete` 事件 + 6 秒超时自动切 OSM；
3. **静态图代理**：`GET /api/v1/public/travels/:id/map-static` 由后端生成并缓存静态图，供列表页/卡片等轻量场景免 iframe 使用；
4. **多平台选择菜单**：点击标记弹出高德/腾讯/Google/Apple 多选（含百度 BD-09 转换）；
5. **行程分组渲染**：按 `itinerary` 天次分组着色与连线（Polyline 轨迹）；
6. **瓦片源可配置**：国内访问慢时可配置自建/国内瓦片源。
