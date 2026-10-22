<div align="center">

<img src="docs/assets/novablog-logo.png" alt="NovaBlog CMS" width="280" />

# NovaBlog CMS

**The open-source blogging system that presets everything around your passions and profession — CMS manages content, themes manage looks, and AI writes for you**

<p>
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Vue-3-4FC08D?style=flat-square&logo=vue.js&logoColor=white" alt="Vue 3" />
  <img src="https://img.shields.io/badge/TypeScript-5-3178C6?style=flat-square&logo=typescript&logoColor=white" alt="TypeScript" />
  <img src="https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat-square&logo=postgresql&logoColor=white" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/Redis-7+-DC382D?style=flat-square&logo=redis&logoColor=white" alt="Redis" />
  <img src="https://img.shields.io/badge/Docker-ready-2496ED?style=flat-square&logo=docker&logoColor=white" alt="Docker" />
</p>

[Official Site](http://www.novablog.cn) · [Free Sign-up](http://www.novablog.cn/register) · [Author's Blog](http://czf.ditancafebar.cn) · [Theme Repo](https://github.com/StudyNoWeekend/novablog-web) · [Docker Image](https://github.com/StudyNoWeekend/novablog-cms/pkgs/container/novablog-cms) · [Skill Docs](https://github.com/StudyNoWeekend/novablog-cms/blob/main/frontend/public/docs/novablog-api.md)

</div>

<div align="center">

**Language**: [简体中文](README.md) · [繁體中文](README.zh-TW.md) · English · [日本語](README.ja.md) · [한국어](README.ko.md)

</div>

---

## 🎯 Own Your Blog in Three Steps

No coding and no server admin knowledge required:

1. **Create your blog** — Apply at [Official Hosting](http://www.novablog.cn/hosting#apply) for zero-deployment, out-of-the-box hosting; if you prefer full control, self-host with Docker following the [Deployment](#-deployment-two-ways) section below;
2. **Setup wizard** — Create an account → **choose your creator focus** → configure storage → the default theme installs automatically, all in a few clicks;
3. **Start creating** — Write articles, upload photos, publish recipes, keep travel journals… The focus you choose determines which modules come pre-enabled, adjustable anytime.

Once setup completes, you own a fully working blog: the CMS hosts the blog pages directly, visitors see the theme you picked, and your content stays untouched.

### Tailored to Your Passion / Profession: Creator Focus

Step 2 of the setup wizard shows **17 "Creator Focus" cards** (multi-select). Pick what you love and the system enables the matching modules automatically; pick none and everything is enabled by default. You can change your focus anytime from the dashboard.

| Group | Creator Focus |
|------|----------|
| Tech · Digital | Tech Developer · Gadget Reviewer |
| Photo · Sound | Photography Lover · Video Creator · Musician · Travel Vlogger |
| Travel · Lifestyle | Travel Blogger · Food Lover · Fitness · Fashion Outfits · Pet Life |
| Hobby · Collection | Reading · Gamer · Designer/Illustrator · Handmade DIY |
| Special | Lifestyle · All-round Creator (enables every module at once) |

---

## 🧩 Modules at a Glance

Modules can be toggled anytime in the dashboard under "System Settings → Module Management"; disabled modules disappear from both the blog and the dashboard menu (the article module is always on, and the media library is always available as core infrastructure).

### ✍️ Content Creation (install what you love)

| Module | What you get |
|------|----------|
| 📝 Articles | Dual Markdown / rich-text editors, categories & tags, draft/publish/unpublish, SEO (custom slug, summary, cover, keywords), read stats, one-click push to the official community |
| 📸 Photo Portfolios | Photo sets organized by theme, automatic EXIF extraction, multiple cover styles and watermark protection |
| 🗺️ Travel Guides | Structured multi-day itinerary editor, Amap / Google Maps location picking with maps on detail pages, likes and trending picks |
| 🍳 Recipes | Ingredients and cooking steps management, one-click push to the official community |
| 📖 Bookshelf | Book collection, ratings and reviews, with Douban book info support |
| 🎮 Game Library | Game collection, playtime and mini reviews, with Bangumi entry support |
| 💪 Fitness | Workout log management with public display |
| 🎵 Music Playlists | Upload songs to build playlists, paste a Bilibili link to auto-resolve playback, import public playlists from NetEase Cloud / QQ Music / Spotify |
| 🎥 Videos | Video and many-to-many album management, Bilibili link parsing and embedding |

### 👤 About Me (build your personal card)

| Module | What you get |
|------|----------|
| Profile | Nickname, avatar and bio, blog title/description/background, 20+ social platform links (with brand-colored icons) |
| My Devices | Digital/photography gear list; uploaded photos get their shooting parameters recognized automatically |
| Projects | Resume-style project history: domain, role, tech stack, start/end dates |
| Tech Stack | Skill tag management, displayed on the blog with official brand icons |
| Open Source | GitHub project showcase cards, one-click push to the official community |

### 🧰 Infrastructure (batteries included)

| Module | What you get |
|------|----------|
| Dashboard | Traffic/content metrics, per-module content cards, ECharts trend charts |
| Comments | Aggregates visitor comments from articles, travel guides and more, three-level moderation, threaded replies, IP blacklist |
| Media Library | Image/video/audio upload (chunked for large files), folder management, size-preset workbench, external image import and search |
| Theme Market | Browse, preview and one-click install official themes (see the theme section below) |
| System Settings | Module switches, object storage (local/MinIO/Aliyun OSS/Tencent COS with hot switching and migration), CORS, map keys, IP blacklist, community push |
| Traffic Statistics | Visitor IP stats with offline ip2region geolocation, no external services required |
| API Docs / MCP Service | Browse public auth-free APIs online, download the Skill doc, manage MCP keys (see the AI section below) |

---

## 🤖 Let AI Write for You: MCP + Skill

NovaBlog CMS ships with a built-in **MCP Server** and **Skill-format API docs**, so AI agents like Codex, ZCode, Claude Code and Cursor can create content directly on your blog — you talk, it types.

### Option 1: MCP Integration (recommended — write, publish and manage)

Three steps in the dashboard under "AI → MCP Service":

1. **Create a key**: self-service MCP API Key (`nbt_mcp_` prefix, optional expiry, built-in rate limiting);
2. **Paste the config**: copy the auto-generated JSON into your AI client's MCP config:

```json
{
  "mcpServers": {
    "novablog": {
      "type": "http",
      "url": "https://your-blog-domain/api/v1/mcp",
      "headers": {
        "Authorization": "Bearer nbt_mcp_your-key"
      }
    }
  }
}
```

3. **Start creating**: just tell the AI "write an article about XX and publish it".

The MCP endpoint is Streamable HTTP (`/api/v1/mcp`) and currently offers **93 tools** covering every content module of the admin console (all delete operations require an explicit `confirm=true` to prevent accidents):

| Domain | Tools |
|------|------|
| Articles | `list_articles` / `get_article` query; `create_article` write (Markdown body, summary, auto-created tags/categories, cover, status); `update_article` edit; `set_article_status` publish / draft / unpublish; `delete_article` delete |
| Media library | `upload_image` upload (auto-transfer from an http URL or base64, `module` archives by content module); `list_media` / `list_media_folders` browse files and folders; `move_media` move; `delete_media` delete (`force=true` when referenced) |
| Categories & tags | `list_categories` (12 content-module types supported) / `create_category` / `update_category` / `delete_category`; `list_tags` / `create_tag` / `update_tag` / `delete_tag` |
| About me | `get_profile` / `update_profile` blogger profile; equipment `list_/create_/update_/delete_equipment`; projects `*_project`; tech stack `*_tech_stack`; open source `*_open_source` + `refresh_open_source` to re-fetch repo info |
| Content creation | Playlists `*_playlist`; videos `*_video` + `parse_video_url` pre-parse; travel guides `*_travel` + `set_travel_status`; photography portfolios `*_portfolio` + items `add_/update_/delete_portfolio_item` / `sort_portfolio_items`; recipes `*_recipe`; books `*_book`; games `*_game`; fitness `*_fitness_record` |
| Comments | `list_comments` query; `reply_comment` reply as the blogger; `delete_comment` delete (cascades to replies) |
| Dashboard (read-only) | `get_dashboard_overview` overview and traffic summary; `get_dashboard_trend` output/traffic trends (7d/30d/90d); `get_top_content` top content ranking |

### Option 2: Skill Doc (let AI master every public API)

The "AI → API Docs" page offers a **SKILL.md-format API reference** ([novablog-api.md](frontend/public/docs/novablog-api.md)). Drop it into the workspace of an AI assistant like Claude Code / Cursor / ZCode and it masters every auth-free public API — paths, parameters and response examples — to build themes, integrations and batch exports for you.

---

## 🏠 Official Ecosystem: NovaBlog Site vs. the CMS

In one sentence: **the CMS is your own blogging core; the official NovaBlog site is your content community.**

- **NovaBlog CMS (this project)**: self-hosted open-source blogging system — content, looks, and hosting of blog pages;
- **NovaBlog official site**: community portal — official accounts, theme market, content community; the CMS's theme market and community push both connect to it;
- **Unified official endpoints**: pre-configured out of the box, viewable and editable in the dashboard under "System Settings → Community Push"; the theme market and community push share the same entry, connected server-side.

| Entry | URL | Purpose |
|------|------|------|
| Official site | <http://www.novablog.cn> | Community portal · Theme market |
| Free sign-up | <http://www.novablog.cn/register> | Create an official account: market login, hosted deployment, push keys |
| Official API | <http://api.novablog.cn> | Unified service entry for community push and the theme market |

> The official endpoints in the dashboard under "System Settings → Community Push" follow the actual runtime configuration.

**Community push**: articles, travel guides, recipes, photo portfolios and open-source projects (5 content types) can be synced to the official community with one click for exposure — auto async push after create/update, auto unlist on delete, auto skip when unchanged, auto retry on failure; the push button on the edit page also supports manual re-push and sync status.

---

## 🎨 Themes: Change Looks Like Changing Clothes

The blog frontend is a fully swappable "theme": themes own looks and experience, the CMS owns content and hosting. Switching themes only updates a pointer — takes effect in seconds, no restart, content untouched.

### Option 1: One-click Install from the Official Market (recommended)

Browse, search, rate and favorite 11+ official themes of all styles (Tech Geek, Lens Life, Wanderlust, Melody Notes…) in the dashboard "Theme Market":

- **Preview before activation** with real rendering (`/preview`);
- **One-click install**: download → SHA256 verification → safe-extract validation → `theme.json` manifest check, all automated;
- **Version management and one-click rollback**; themes declare settings via `theme.json` and the dashboard auto-generates a visual settings panel — change settings without touching code.

### Option 2: Mount Your Own Frontend Locally

Want to use your own frontend? Just mount a directory containing `theme.json` and `dist/` to the CMS with zero modification: set `themes.frontend_dir` on bare metal, or use `./deploy.sh --frontend-dir <dir>` with Docker.

### 🌍 Themes Are Open Source — Build With Us

All official themes are open-sourced in the theme repo [StudyNoWeekend/novablog-web](https://github.com/StudyNoWeekend/novablog-web), under an **official + community co-maintenance** model: the official team keeps producing and maintaining official themes, while community authors develop their own themes per the spec and submit Issues / PRs for distribution to every blogger through the official market.

- Theme package spec: [novablog-web/docs/theme-spec.md](https://github.com/StudyNoWeekend/novablog-web/blob/main/docs/theme-spec.md)
- Packaging & release guide: [novablog-主题包规范与发布指南.md](novablog-主题包规范与发布指南.md)

---

## 🚀 Deployment: Two Ways

### Option 1: Official Hosting (zero deployment, recommended for beginners)

Sign up free at the [official site](http://www.novablog.cn/register) to get a hosted blog — no server, no Docker — and enter the same setup wizard as self-hosting.

### Option 2: Self-host with Docker

A single image bundles nginx dual entry, the Go backend and the admin frontend: `ghcr.io/studynoweekend/novablog-cms` (`linux/amd64` / `linux/arm64`). Bring your own PostgreSQL 15+ and Redis 7+ (the script runs connectivity pre-checks):

```bash
cd deploy
./deploy.sh --name myblog                        # interactive Q&A fills the rest
./deploy.sh --name myblog --yes \
  --db-host 10.0.0.5 --db-password '***' \
  --redis-host 10.0.0.5 --redis-password '***'   # non-interactive defaults
```

Once started: blog at `http://<domain>:80`, dashboard at `http://<domain>:8080/admin/`.

| Entry | Default Port | Contents |
|------|----------|------|
| Blog | `80` | Active theme pages (or locally mounted frontend) · `/preview` · `/api/v1` · `/files` |
| Dashboard | `8080` | `/admin/` admin panel · `/api/v1` · `/files` |

Common operations:

```bash
./deploy.sh --name myblog --status    # container status
./deploy.sh --name myblog --logs      # follow app logs
./deploy.sh --name myblog --down      # stop and remove the container (mounted data is kept)
```

- `--name` is required and acts as the isolation key for multiple blogs on one host — different names and ports can coexist;
- To upgrade, re-run `./deploy.sh --name myblog --version vX.Y.Z`; mounted data is kept;
- Full parameters, external databases and host nginx reverse-proxy notes: [deploy/README.md](deploy/README.md).

### Local Development

```bash
# 1. Start the backend
cd backend
cp config/config.example.yaml config/config.yaml   # edit postgres/redis connection and keys
make run                                            # :8111, auto-migrates on first run

# 2. Start the admin frontend
cd ../frontend
pnpm install && pnpm dev                            # :5173, /api proxied to :8111
```

Open `http://localhost:5173` in a browser and the setup wizard starts automatically.

---

## 🤝 Contributing

Issues and Pull Requests are welcome:

1. Run `make fmt && make vet && make lint` before committing backend changes;
2. Run `pnpm lint` before committing frontend changes;
3. When adding APIs, update the DTOs and the API doc together.

---

## 📜 License

This project is open-sourced under the [MIT](LICENSE) license.

---

## 🙏 Acknowledgements

- [Gin](https://github.com/gin-gonic/gin) · [GORM](https://gorm.io) · [Vue.js](https://vuejs.org) · [Ant Design Vue](https://antdv.com) · [PostgreSQL](https://www.postgresql.org) · [Redis](https://redis.io)
- And every community contributor who inspired and supported this project

---

<p align="center">
  <sub>Made with ❤️ for independent bloggers</sub>
</p>
