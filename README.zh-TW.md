<div align="center">

<img src="docs/assets/novablog-logo.png" alt="NovaBlog CMS" width="280" />

# NovaBlog CMS

**依你的興趣與職業預先配置好一切的開源部落格系統 —— CMS 管內容，主題管顏值，AI 替你寫**

<p>
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Vue-3-4FC08D?style=flat-square&logo=vue.js&logoColor=white" alt="Vue 3" />
  <img src="https://img.shields.io/badge/TypeScript-5-3178C6?style=flat-square&logo=typescript&logoColor=white" alt="TypeScript" />
  <img src="https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat-square&logo=postgresql&logoColor=white" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/Redis-7+-DC382D?style=flat-square&logo=redis&logoColor=white" alt="Redis" />
  <img src="https://img.shields.io/badge/Docker-ready-2496ED?style=flat-square&logo=docker&logoColor=white" alt="Docker" />
</p>

[官方網站](http://www.novablog.cn) · [免費註冊](http://www.novablog.cn/register) · [作者部落格](http://czf.ditancafebar.cn) · [主題倉庫](https://github.com/StudyNoWeekend/novablog-web) · [Docker 映像](https://github.com/StudyNoWeekend/novablog-cms/pkgs/container/novablog-cms) · [Skill 文件](https://github.com/StudyNoWeekend/novablog-cms/blob/main/frontend/public/docs/novablog-api.md)

</div>

<div align="center">

**語言**：[简体中文](README.md) · 繁體中文 · [English](README.en.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

</div>

---

## 🎯 三步擁有你的部落格

不需要懂程式，也不需要懂伺服器：

1. **開通部落格** —— 到[官方託管申請](http://www.novablog.cn/hosting#apply)即可使用官方託管，零部署開箱即用；想完全自己掌控，就依照下方 [部署](#-部署兩種方式) 章節用 Docker 自建；
2. **首裝精靈** —— 建立帳號 → **選擇創作方向** → 設定儲存 → 自動裝好預設主題，全程點幾下滑鼠；
3. **開始創作** —— 寫文章、傳照片、發食譜、記旅遊……你選的方向決定了後台預先開啟哪些模組，之後隨時可調。

首裝完成後你就擁有一個能存取的完整部落格：CMS 直接託管部落格頁面，訪客存取的是你選的主題，內容資料分毫不動。

### 依興趣 / 職業客製化：創作方向

首裝精靈第 2 步會展示 **17 張「創作方向」卡片**（可多選），選中所愛，系統自動開啟對應的模組；一個都不選則預設開啟全部。登入後也能在後台頂端隨時修改創作方向。

| 分組 | 創作方向 |
|------|----------|
| 技術 · 數位 | 技術開發者 · 數位測評 |
| 影像 · 聲音 | 攝影愛好者 · 影片創作者 · 音樂人 · 旅遊 Vlogger |
| 旅遊 · 生活 | 旅遊部落客 · 美食愛好者 · 健身運動 · 時尚穿搭 · 寵物日常 |
| 興趣 · 收藏 | 讀書 · 遊戲玩家 · 設計師/插畫師 · 手作 DIY |
| 特殊 | 生活方式 · 全能創作者（一鍵開啟全部模組） |

---

## 🧩 模組一覽

後台模組隨時可在「系統設定 → 模組管理」開關，關閉後前台與後台選單同步隱藏（文章模組恆開，媒體庫為基礎設施恆可用）。

### ✍️ 內容創作（依你的興趣選裝）

| 模組 | 給你什麼 |
|------|----------|
| 📝 文章管理 | Markdown / 富文字雙編輯器，分類標籤、草稿/發佈/下架、SEO（自訂 slug、摘要、封面、關鍵字）、閱讀統計，一鍵推送官方社群 |
| 📸 攝影作品集 | 依主題組織相片集，自動擷取 EXIF 拍攝參數，多種封面樣式與浮水印保護 |
| 🗺️ 旅遊攻略 | 結構化多日行程編輯器，高德 / Google 地圖選點與詳情頁地圖展示，按讚與熱門推薦 |
| 🍳 美食食譜 | 食材與烹飪步驟管理，一鍵推送官方社群 |
| 📖 讀書書架 | 藏書、評分與書評，支援關聯豆瓣圖書資訊 |
| 🎮 遊戲庫 | 遊戲收藏、遊玩時數與短評，支援關聯 Bangumi 條目 |
| 💪 健身訓練 | 訓練紀錄管理與前台展示 |
| 🎵 音樂播放清單 | 上傳歌曲建歌單，貼上 B 站連結自動解析播放，匯入網易雲 / QQ 音樂 / Spotify 公開歌單 |
| 🎥 影片作品 | 影片與多對多專輯管理，B 站連結解析嵌入 |

### 👤 關於我（打造個人名片）

| 模組 | 給你什麼 |
|------|----------|
| 個人資料 | 暱稱頭像簡介、部落格標題/描述/背景、20+ 社群平台連結（內建品牌色圖示） |
| 個人設備 | 數位/攝影裝備清單，上傳相片自動辨識拍攝參數 |
| 專案經歷 | 線上履歷式專案經歷：領域、角色、技術棧、起訖時間 |
| 技術棧 | 技能標籤管理，前台用官方品牌圖示展示 |
| 開源作品 | GitHub 開源專案展示卡，一鍵推送官方社群 |

### 🧰 基礎設施（開箱自帶）

| 模組 | 給你什麼 |
|------|----------|
| 工作台 | 造訪量/內容量指標、各模組內容卡片、ECharts 趨勢圖 |
| 留言管理 | 彙整文章、旅遊攻略等來源的訪客留言，三級審核、樓中樓回覆、IP 黑名單 |
| 媒體庫 | 圖片/影片/音訊上傳（大檔案分片）、資料夾管理、尺寸預設工作台、外部圖片匯入與搜尋 |
| 主題市場 | 官方主題瀏覽、預覽、一鍵安裝啟用（見下方主題章節） |
| 系統設定 | 模組開關、物件儲存（本機/MinIO/阿里雲 OSS/騰訊雲 COS 熱切換與遷移）、跨網域、地圖 Key、IP 黑名單、社群推播 |
| 存取統計 | 前台訪客 IP 統計，ip2region 離線解析歸屬地，不依賴外部服務 |
| API 文件 / MCP 服務 | 免認證公開 API 線上瀏覽、Skill 文件下載、MCP 金鑰管理（見下方 AI 章節） |

---

## 🤖 讓 AI 替你寫：MCP + Skill

NovaBlog CMS 內建 **MCP Server** 與 **Skill 格式介面文件**，Codex、ZCode、Claude Code、Cursor 等 AI Agent 都能直接在你的部落格上創作內容——你動口，它動手。

### 方式一：MCP 接入（推薦，可寫可發佈可管理）

後台「AI → MCP 服務」三步接入：

1. **建立金鑰**：自助開通 MCP API Key（`nbt_mcp_` 前綴，可設有效期，內建限流保護）；
2. **貼上設定**：複製頁面自動產生的 JSON，貼進 AI 用戶端的 MCP 設定：

```json
{
  "mcpServers": {
    "novablog": {
      "type": "http",
      "url": "https://你的部落格網域/api/v1/mcp",
      "headers": {
        "Authorization": "Bearer nbt_mcp_你的金鑰"
      }
    }
  }
}
```

3. **開口創作**：對 AI 說「幫我寫一篇 XX 主題的文章並發佈」即可。

MCP 端點為 Streamable HTTP（`/api/v1/mcp`），目前提供 **93 個工具**，覆蓋管理後台全部內容模組（刪除操作均需顯式 `confirm=true` 防誤刪）：

| 域 | 工具 |
|------|------|
| 文章管理 | `list_articles` / `get_article` 查詢；`create_article` 寫文章（Markdown 正文、摘要、標籤/分類自動建立、封面、狀態）；`update_article` 修改；`set_article_status` 發佈 / 轉草稿 / 下架；`delete_article` 刪除 |
| 媒體庫 | `upload_image` 上傳圖片（http 連結自動轉存或 base64，`module` 按內容模組歸檔）；`list_media` / `list_media_folders` 查詢檔案與資料夾；`move_media` 移動；`delete_media` 刪除（被引用需 `force=true`） |
| 分類與標籤 | `list_categories`（支援 12 種內容模組類型）/ `create_category` / `update_category` / `delete_category`；`list_tags` / `create_tag` / `update_tag` / `delete_tag` |
| 關於我 | `get_profile` / `update_profile` 部落客資料；個人設備 `list_/create_/update_/delete_equipment`；專案經歷 `*_project`；技術棧 `*_tech_stack`；開源作品 `*_open_source` + `refresh_open_source` 重新拉取倉庫資訊 |
| 內容創作 | 第三方歌單 `*_playlist`；影片作品 `*_video` + `parse_video_url` 預解析；旅行攻略 `*_travel` + `set_travel_status`；攝影作品集 `*_portfolio` + 作品項 `add_/update_/delete_portfolio_item` / `sort_portfolio_items`；美食食譜 `*_recipe`；讀書書架 `*_book`；遊戲庫 `*_game`；健身訓練 `*_fitness_record` |
| 評論管理 | `list_comments` 查詢；`reply_comment` 以部落客身份回覆；`delete_comment` 刪除（級聯刪除回覆） |
| 工作台（唯讀） | `get_dashboard_overview` 總覽與流量摘要；`get_dashboard_trend` 產出/流量趨勢（7d/30d/90d）；`get_top_content` 熱門內容排行 |

### 方式二：Skill 文件（讓 AI 掌握全部公開 API）

後台「AI → API 文件」頁可下載 **SKILL.md 格式的介面參考**（[novablog-api.md](frontend/public/docs/novablog-api.md)），放進 Claude Code / Cursor / ZCode 等 AI 助手的工作區，它就能掌握全部免認證公開 API 的路徑、參數與回應範例，幫你寫主題、做整合、批次取數。

---

## 🏠 官方生態：NovaBlog 官方站與 CMS 的關係

一句話：**CMS 是你自己的部落格核心，官方 NovaBlog 站是你的內容社群。**

- **NovaBlog CMS（本專案）**：自架的開源部落格系統，管內容、管外觀、直接託管部落格頁面；
- **NovaBlog 官方站**：社群入口——官方帳號體系、主題市場、內容社群；CMS 的主題市場與社群推播都連向官方站；
- **官方位址統一託管**：出廠內建、後台「系統設定 → 社群推播」可查改，主題市場與社群推播共用同一入口，伺服器直連。

| 入口 | 網址 | 用途 |
|------|------|------|
| 官方網站 | <http://www.novablog.cn> | 社群入口 · 主題市場 |
| 免費註冊 | <http://www.novablog.cn/register> | 開通官方帳號：市場登入、託管部署、推播金鑰 |
| 官方 API | <http://api.novablog.cn> | 社群推播與主題市場的統一服務入口 |

> 後台「系統設定 → 社群推播」中的官方網址以實際設定為準。

**社群推播**：文章、旅遊攻略、美食食譜、攝影作品集、開源作品 5 類內容可一鍵同步到官方社群獲得曝光——內容建立/更新後自動非同步推送、刪除自動下架、無變化自動略過、失敗自動重試，編輯頁內的推送按鈕也支援手動重推與查看同步狀態。

---

## 🎨 主題：換外觀像換衣服

部落格前端是可整體替換的「主題」：主題負責外觀與體驗，CMS 負責內容與託管。切換主題只更新一個指標，秒級生效、不重啟服務、內容分毫不動。

### 方式一：官方市場一鍵安裝（推薦）

後台「主題市場」瀏覽、搜尋、評分、收藏 11+ 風格各異的官方主題（極客風、鏡頭生活、漫遊世界、旋律筆記……），選中即可：

- **啟用前預覽**真實渲染效果（`/preview`）；
- **一鍵安裝**：自動完成下載 → SHA256 校驗 → 安全校驗解壓 → `theme.json` 清單校驗；
- **版本管理與一鍵回滾**，主題透過 `theme.json` 宣告設定項，後台自動產生視覺化設定面板，改設定不碰程式。

### 方式二：本地掛載自備前端

想用自己的前端？把含 `theme.json` 與 `dist/` 的目錄掛載給 CMS 即可，零改造接入：裸機部署設 `themes.frontend_dir`，Docker 部署用 `./deploy.sh --frontend-dir <目錄>`。

### 🌍 主題開源開放，和我們一起共建

全部官方主題開源在主題倉庫 [StudyNoWeekend/novablog-web](https://github.com/StudyNoWeekend/novablog-web)，採用**官方 + 社群共同維護**的開放模式：官方持續產出與維護官方主題，社群作者按規範開發自己的主題提交 Issue / PR，經官方市場分發給所有部落客。

- 主題包規範：[novablog-web/docs/theme-spec.md](https://github.com/StudyNoWeekend/novablog-web/blob/main/docs/theme-spec.md)
- 打包與發布指南：[novablog-主题包规范与发布指南.md](novablog-主题包规范与发布指南.md)

---

## 🚀 部署：兩種方式

### 方式一：官方託管（零部署，新手推薦）

到[官方網站免費註冊](http://www.novablog.cn/register)即可開通託管部落格，無需伺服器、無需 Docker，註冊後進入與自部署完全一致的首裝精靈。

### 方式二：Docker 自部署

單一映像內建 nginx 雙入口、Go 後端與管理後台前端：`ghcr.io/studynoweekend/novablog-cms`（`linux/amd64` / `linux/arm64`）。需自備 PostgreSQL 15+ 與 Redis 7+ 實例（腳本會做連通性預檢）：

```bash
cd deploy
./deploy.sh --name myblog                        # 互動問答補全其餘設定
./deploy.sh --name myblog --yes \
  --db-host 10.0.0.5 --db-password '***' \
  --redis-host 10.0.0.5 --redis-password '***'   # 非互動全預設部署
```

啟動後：部落格入口 `http://<域名>:80`，後台入口 `http://<域名>:8080/admin/`。

| 入口 | 預設連接埠 | 內容 |
|------|----------|------|
| 部落格入口 | `80` | 已啟用主題頁面（或本地掛載前端）· `/preview` · `/api/v1` · `/files` |
| 後台入口 | `8080` | `/admin/` 管理後台 · `/api/v1` · `/files` |

常用維運：

```bash
./deploy.sh --name myblog --status    # 查看容器狀態
./deploy.sh --name myblog --logs      # 追蹤應用日誌
./deploy.sh --name myblog --down      # 停止並移除容器（掛載資料保留）
```

- `--name` 必填，作為同機多部落格的隔離鍵，不同名稱與連接埠即可同機共存；
- 升級重跑 `./deploy.sh --name myblog --version vX.Y.Z`，掛載資料保留；
- 完整參數、外部資料庫、宿主機 nginx 反代說明見 [deploy/README.md](deploy/README.md)。

### 本地開發

```bash
# 1. 啟動後端
cd backend
cp config/config.example.yaml config/config.yaml   # 編輯 postgres/redis 連線與密鑰
make run                                            # :8111，首次啟動自動遷移

# 2. 啟動管理後台前端
cd ../frontend
pnpm install && pnpm dev                            # :5173，/api 代理到 :8111
```

瀏覽器開啟 `http://localhost:5173`，自動進入首裝精靈。

---

## 🤝 貢獻

歡迎提交 Issue 與 Pull Request：

1. 後端提交前執行 `make fmt && make vet && make lint`；
2. 前端提交前執行 `pnpm lint`；
3. 新增 API 時同步更新 DTO 與 API 文件。

---

## 📜 開源協議

本專案基於 [MIT](LICENSE) 協議開源。

---

## 🙏 致謝

- [Gin](https://github.com/gin-gonic/gin) · [GORM](https://gorm.io) · [Vue.js](https://vuejs.org) · [Ant Design Vue](https://antdv.com) · [PostgreSQL](https://www.postgresql.org) · [Redis](https://redis.io)
- 以及所有為本專案提供靈感與支援的社群貢獻者

---

<p align="center">
  <sub>Made with ❤️ for independent bloggers</sub>
</p>
