-- 摄影师视角示例数据种子脚本
-- 说明：
--   1. 本脚本为 NovaBlog 项目填充示例内容，涵盖文章、摄影器材、摄影作品、旅行攻略、个人资料与评论。
--   2. 图片使用 picsum.photos 占位图，便于本地快速预览；正式上线前请替换为自己的作品或真实图片。
--   3. 个人资料仅更新文字信息，不会覆盖已有的头像、博客图标和页面背景图。
--   4. 必须先完成数据库迁移（migrations），确保以下表已创建：
--        - categories, tags, articles, article_tags, media, media_presets
--        - portfolios, portfolio_items
--        - travel_guides
--        - comments
--        - photo_equipment (由 020_photography_equipment 迁移创建)
--        - bloggers
--
-- 前置迁移命令示例（使用 migrate 工具）：
--   migrate -database "postgres://postgres:密码@localhost:5432/novablog?sslmode=disable" -path backend/migrations up
--
-- 使用方法（二选一）：
--   1. 使用 psql 直接执行：
--        psql -U postgres -d novablog -f backend/scripts/seed_photographer_data.sql
--   2. 使用 Go 种子命令（需先创建 backend/config/config.yaml）：
--        cd backend && go run ./cmd/seed

BEGIN;

-- ============================================================
-- 1. 分类（文章 / 作品集 / 旅行攻略）
-- ============================================================
-- 兼容可能缺少唯一约束的历史库
CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_slug ON categories(slug);

INSERT INTO categories (id, name, slug, description, sort_order, type) VALUES
  ('10000000-0000-0000-0000-000000000001', '摄影技巧', 'photo-tips', '从构图到光线的实战心得', 1, 'article'),
  ('10000000-0000-0000-0000-000000000002', '器材评测', 'gear-review', '相机、镜头与配件的真实使用体验', 2, 'article'),
  ('10000000-0000-0000-0000-000000000003', '旅行笔记', 'travel-notes', '路上遇见的风景与故事', 3, 'article'),
  ('10000000-0000-0000-0000-000000000004', '风光作品', 'landscape-portfolio', '山川湖海，光影瞬间', 1, 'portfolio'),
  ('10000000-0000-0000-0000-000000000005', '街拍作品', 'street-portfolio', '城市街头的烟火与光影', 2, 'portfolio'),
  ('10000000-0000-0000-0000-000000000006', '人像作品', 'portrait-portfolio', '自然光下的人物故事', 3, 'portfolio'),
  ('10000000-0000-0000-0000-000000000007', '国内旅行', 'domestic-travel', '中国境内的摄影旅行路线', 1, 'travel'),
  ('10000000-0000-0000-0000-000000000008', '海外旅行', 'overseas-travel', '出境摄影旅行目的地', 2, 'travel'),
  ('10000000-0000-0000-0000-000000000009', '航拍作品', 'aerial-portfolio', '俯瞰大地的视角', 4, 'portfolio')
ON CONFLICT (slug) DO NOTHING;

-- ============================================================
-- 2. 文章标签
-- ============================================================
-- 兼容可能缺少唯一约束的历史库：先确保 name 唯一索引存在
CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_name ON tags(name);

INSERT INTO tags (id, name) VALUES
  ('20000000-0000-0000-0000-000000000001', '摄影技巧'),
  ('20000000-0000-0000-0000-000000000002', '旅行'),
  ('20000000-0000-0000-0000-000000000003', '器材评测'),
  ('20000000-0000-0000-0000-000000000004', '后期处理'),
  ('20000000-0000-0000-0000-000000000005', '风光摄影'),
  ('20000000-0000-0000-0000-000000000006', '人像摄影'),
  ('20000000-0000-0000-0000-000000000007', '街头摄影')
ON CONFLICT (name) DO NOTHING;

-- ============================================================
-- 3. 文章
-- ============================================================
-- 兼容可能缺少唯一约束的历史库
CREATE UNIQUE INDEX IF NOT EXISTS idx_articles_slug ON articles(slug);

INSERT INTO articles (id, title, slug, summary, content, cover_image, category_id, status, type, is_top, is_comment, published_at) VALUES
  ('30000000-0000-0000-0000-000000000001',
   '用好黄金时刻，拍出通透风光照',
   'golden-hour-landscape',
   '日出日落前后的黄金时刻是风光摄影师最珍贵的拍摄窗口。本文分享如何判断光线、控制曝光与构图。',
   $$## 黄金时刻的魅力

![黄金时刻的山脉](https://picsum.photos/seed/golden-mountain/1200/800)

日出后和日落前的一小时，太阳角度低、光线柔和、色温温暖，是拍摄风光的最佳时机。这个时段的光线有三个特点：

1. **方向性强**：侧光能勾勒出地形起伏，增强立体感
2. **色温温暖**：金色的光线让大地披上一层暖色外衣
3. **反差适中**：相比正午的硬光，阴影更柔和，高光不过曝

### 如何判断黄金时刻

![使用App预判太阳方位](https://picsum.photos/seed/sun-app/1200/800)

提前踩点是风光摄影的基本功。推荐几个实用工具：

- **巧摄（中国版）**：实时显示太阳方位角和高度角
- **Sun Surveyor**：可视化太阳轨迹，支持 AR 预览
- **PlanIt! Pro**：综合规划工具，可预判机位与构图

> 摄影不是记录场景，而是记录光与情绪。提前踩点能让你在最佳光线到来时从容不迫。

### 实战技巧：控制大光比

![包围曝光示意](https://picsum.photos/seed/bracket-hdr/1200/800)

黄金时刻的天空与地面往往存在 6-10 档光差，单张曝光很难兼顾。我的做法是：

**方法一：包围曝光 + 后期合成**

拍摄 3-5 张不同曝光的照片（如 -2、0、+2 EV），在 Lightroom 或 Photomatix 中合成 HDR。注意保持相机稳定，建议使用快门线或间隔拍摄。

**方法二：渐变灰镜（GND）**

如果你习惯直出，一块 3 档软边 GND 能有效压暗天空。软边适合山体交界处，硬边适合海平面。

**方法三：摇黑卡**

没有滤镜时，在曝光过程中用黑卡遮挡天空部分，手动平衡光比。需要多次尝试。

### 构图要点：善用前景

![有前景的风光照片](https://picsum.photos/seed/landscape-foreground/1200/800)

前景是风光照片的"锚点"，能让观众一眼找到视觉入口。常见的前景元素包括：

- 石头、岩石纹理
- 花草、枯木
- 水面倒影
- 道路、栈道

构图时把前景放在画面下三分之一处，配合广角镜头贴近拍摄，能获得强烈的纵深感。

### 参数参考

以下是我在川西高原拍摄日出时的常用参数：

| 参数 | 数值 | 说明 |
|------|------|------|
| 光圈 | f/11 - f/16 | 保证景深 |
| ISO | 64 - 100 | 原生ISO画质最佳 |
| 快门 | 1/15s - 1s | 根据光线调整 |
| 滤镜 | CPL + GND 0.9 | 压暗天空，消除反光 |

### 总结

黄金时刻是风光摄影师最珍贵的拍摄窗口。提前踩点、控制光比、善用前景，是拍出通透风光照的三个关键。下次出行前，记得先用 App 预判光线，带上滤镜和快门线，耐心等待那一刻的到来。

---

*你有哪些黄金时刻的拍摄心得？欢迎在评论区分享。*$$,
   'https://picsum.photos/seed/golden-hour/1600/900',
   '10000000-0000-0000-0000-000000000001',
   2, 1, true, true, NOW()),

  ('30000000-0000-0000-0000-000000000002',
   '我的 35mm 街头摄影一年回顾',
   '35mm-street-photography-year',
   '一枚 35mm 镜头，一台相机，记录城市一年的脉搏。街头摄影教会我的是观察与等待。',
   $$## 为什么我选择 35mm

![街头的35mm视角](https://picsum.photos/seed/street-35mm-view/1200/800)

2024 年初，我做了一个决定：只带一枚 35mm 镜头，一台相机，用一年时间记录城市。这个选择源于对"摄影减法"的思考——当你手里的工具越少，你会越专注于观察本身。

35mm 介于广角与标头之间，既能容纳环境，又不会像 28mm 那样产生强烈透视变形。它是一种"讲故事"的焦段，能同时呈现人物与环境的关系。

### 一年的收获

![街头摄影作品集锦](https://picsum.photos/seed/street-collage/1200/800)

这一年，我走过了十几个城市，按下数万次快门，最终留下的不过百张。街头摄影教会我三件事：

**学会等待**

好照片是等出来的，而不是追出来的。看到一个有趣的场景，不要急着按下快门——等等看有没有人走进画面，等光线变化，等一个故事发生。

![等待的瞬间](https://picsum.photos/seed/waiting-moment/1200/800)

> 罗伯特·卡帕说，拍得不够好是因为靠得不够近。但"近"不等于追着人跑，有时候是站在那里，等世界走向你。

**靠近主体**

35mm 不像 85mm 那样能远距离"偷拍"。如果你想拍出有感染力的照片，必须靠近，再靠近。这需要勇气，但也是街拍的乐趣所在。

**保持低调**

黑色机身、静音快门、不打扰被摄者，是街拍的底线。我习惯把相机调到静音模式，预对焦后快速按下，大部分时候对方根本没注意到。

### 街拍的器材建议

![我的街拍装备](https://picsum.photos/seed/street-gear/1200/800)

| 器材 | 推荐 | 理由 |
|------|------|------|
| 相机 | 索尼 A7C II / 富士 X-T5 | 轻便、低调、对焦快 |
| 镜头 | 35mm f/1.8 或 f/2 | 轻巧、大光圈、够用 |
| 背带 | Peak Design Slide | 快速取机 |
| 设置 | 快门优先 1/500s | 保证动态清晰 |

### 一个月的城市，一个月的光影

我把这一年分成十二个月，每个月专注一座城市。从成都的老街到重庆的梯坎，从上海的法租界到广州的骑楼，每一座城市都有独特的视觉节奏。

![不同城市的街头光影](https://picsum.photos/seed/city-street-grid/1200/800)

最难忘的是十一月的重庆。那天傍晚，解放碑附近的巷子里，一个卖烤红薯的老人弯腰添炭，背后的霓虹灯把蒸汽染成粉色。我没多想，按下快门，那一刻城市烟火与光影交织。

### 给新手的建议

如果你刚开始尝试街拍，这几点或许有帮助：

1. **先拍光影，再拍人**：从建筑、影子、反光入手，熟悉街头视觉语言
2. **固定一个焦段**：用熟一个镜头，比换着镜头拍更有收获
3. **多去同一个地方**：熟悉的环境能让你更专注于观察
4. **不要怕拍烂片**：街头摄影的成片率极低，十张里有一张满意就很好了

---

*街拍最迷人的，是那种"在路上"的感觉。下次出门，试着不带预设，让城市告诉你故事。*$$,
   'https://picsum.photos/seed/street-35mm/1600/900',
   '10000000-0000-0000-0000-000000000001',
   2, 1, false, true, NOW()),

  ('30000000-0000-0000-0000-000000000003',
   '旅行摄影轻量化打包指南',
   'lightweight-travel-photo-packing',
   '如何在长途旅行中只带最必要的器材，却依然拍出满意的作品？这是我的打包心得。',
   $$## 轻量化的核心思路

![轻装上阵的旅行摄影](https://picsum.photos/seed/lightweight-travel/1200/800)

旅行摄影的敌人是重量。背得越少，走得越远，拍得越多。

我曾经背着两机四镜、三脚架、滤镜套装出门，结果是肩膀酸痛、行动迟缓，错过很多精彩瞬间。后来我开始反思：那些"以防万一"带的器材，真的用上了吗？

答案是否定的。大部分时候，我只用了一机两镜。

### 我的常用组合

![我的旅行器材组合](https://picsum.photos/seed/travel-kit/1200/800)

经过多次调整，现在的组合是：

**机身：索尼 A7R V**

高像素、强防抖、翻转屏，兼顾风光与人像。

**镜头一：FE 24-70mm f/2.8 GM II**

万金油标准变焦，覆盖大部分场景。画质与便携兼顾，旅行挂机首选。

**镜头二：FE 85mm f/1.4 GM**

人像与弱光环境下的得力工具，焦外柔美、锐度出众。

**无人机：DJI Mavic 3 Pro**

上帝视角利器，三摄组合覆盖广角与长焦。

**滤镜：NiSi V7 套装**

包含 CPL、ND64、ND1000、GND 0.9，控制反光与光比。

**背包：Peak Design Everyday Backpack 30L**

分区灵活、取机方便，能装下以上全部。

### 打包清单

![打包清单可视化](https://picsum.photos/seed/packing-list/1200/800)

| 类别 | 物品 | 数量 | 备注 |
|------|------|------|------|
| 机身 | A7R V + 电池 | 2块 | +1块备用 |
| 镜头 | 24-70mm GM II | 1 | 防水袋包裹 |
| 镜头 | 85mm GM | 1 | 竖放节省空间 |
| 无人机 | Mavic 3 Pro | 1 | 电池放随身行李 |
| 滤镜 | NiSi V7 | 1套 | 滤镜包贴背板 |
| 存储 | CFexpress + SD | 各3张 | 分开存放 |
| 配件 | 快门线、读卡器 | 各1 | 小包收纳 |
| 配件 | 清洁套装 | 1 | 气吹+镜头布 |

### 打包原则

![打包原则图解](https://picsum.photos/seed/packing-rules/1200/800)

**原则一：每件器材必须有明确用途**

出发前列出"必拍清单"，根据场景选择器材。比如这次要去沙漠拍星空，那三脚架必带；如果只是城市街拍，三脚架可以省掉。

**原则二：电池、存储卡按天数翻倍**

一天行程至少准备两天的电量和存储。长时间外出，建议带充电宝和读卡器，每天晚上备份。

**原则三：重要文件双备份**

拍摄结束后，第一时间备份到本地硬盘和云盘。数据无价，不要因为一张卡损坏而痛失心血。

### 不同场景的打包建议

![不同场景的器材选择](https://picsum.photos/seed/scene-gear/1200/800)

**城市旅行（3-5天）**

- 一机一镜（24-70mm）
- 备用电池两块
- CPL 滤镜
- 随身小包即可

**风光摄影（7-15天）**

- 一机两镜（24-70mm + 超广）
- 三脚架
- 滤镜套装
- 快门线、间隔拍摄器

**人像旅拍（7-15天）**

- 一机两镜（24-70mm + 85mm）
- 反光板或便携灯
- 备用电池三块
- 大容量存储卡

### 轻量化的心理建设

最难的其实不是"带什么"，而是"不带什么"。克服"以防万一"的心态需要练习：

- 想象最坏情况：没带那颗镜头，真的拍不了吗？
- 回顾过往：上次带的东西，用了几成？
- 设定限制：背包重量不超过体重的 15%

---

*轻装上阵，才能走得更远。祝你下次旅行满载而归。*$$,
   'https://picsum.photos/seed/travel-packing/1600/900',
   '10000000-0000-0000-0000-000000000003',
   2, 1, false, true, NOW()),

  ('30000000-0000-0000-0000-000000000004',
   '索尼 A7R V 一年使用体验',
   'sony-a7r-v-one-year',
   '高像素、强对焦、AI 芯片，这台相机如何改变我的拍摄方式？一年深度体验报告。',
   $$## 6100 万像素的代价与红利

![索尼 A7R V 正面](https://picsum.photos/seed/a7rv-front/1200/800)

2023 年底，我从 A7R IV 升级到了 A7R V。一年的使用下来，这台机器陪我走过了雪山、沙漠与城市，是我创作路上可靠的伙伴。

先说结论：如果你以风光、建筑、静物和商业人像为主，A7R V 是绝佳选择；如果更偏向视频或体育追焦，A7S III 或 A1 会更合适。

### 最让我惊喜的三点

**AI 对焦：生态摄影事半功倍**

![鸟类眼部识别](https://picsum.photos/seed/bird-focus/1200/800)

A7R V 搭载了索尼最新的 AI 对焦芯片，鸟类、动物眼部识别成功率极高。以前拍鸟要手动对焦或连拍盲选，现在大部分时候一次就中。

> 在川西高原拍黑颈鹤，AI 对焦在 400mm 焦段下依然稳稳锁住眼睛，成片率从 30% 提升到 80%。

**8 级防抖：手持慢门不再是奢望**

![手持慢门效果](https://picsum.photos/seed/ibis-test/1200/800)

官方宣称 8 级防抖，实测在 24mm 焦段，1/4 秒手持成功率超过 80%。这意味着很多夜景场景可以不用三脚架，拍摄更自由。

**四轴翻转屏：高低角度、竖拍都游刃有余**

![翻转屏示意](https://picsum.photos/seed/flip-screen/1200/800)

终于，索尼的翻转屏可以翻转了。低角度拍花、高角度俯拍、竖构图取景，都变得轻松自然。

### 高像素的代价

![存储卡示意](https://picsum.photos/seed/storage-cards/1200/800)

6100 万像素带来惊人细节，但也意味着更大的存储和计算压力：

| 项目 | 挑战 | 我的解决方案 |
|------|------|------|
| 存储 | 一张 RAW 约 120MB | CFexpress Type A + 256GB SD |
| 后期 | 电脑卡顿 | 升级 M3 Max MacBook Pro |
| 镜头 | 解析力要求高 | 优先使用 GM 镜头 |
| 防抖 | 放大抖动 | 开启电子快门、提高 ISO |

### 实战参数参考

![参数设置界面](https://picsum.photos/seed/a7rv-settings/1200/800)

以下是我在不同场景下的常用设置：

**风光摄影**

- 光圈：f/8 - f/11
- ISO：64 - 100
- 快门：根据光线
- 文件格式：RAW + JPEG（备用）
- 防抖：开启（手持时）

**人像摄影**

- 光圈：f/1.4 - f/2.8
- ISO：100 - 800
- 对焦模式：实时追踪 + 眼部识别
- 文件格式：RAW（后期空间大）

**生态摄影**

- 光圈：f/5.6 - f/8
- ISO：自动（上限 3200）
- 对焦模式：鸟类/动物识别
- 快门：1/1000s 以上

### 与前代对比

![A7R IV vs A7R V](https://picsum.photos/seed/a7rv-compare/1200/800)

| 项目 | A7R IV | A7R V |
|------|--------|-------|
| 像素 | 6100 万 | 6100 万 |
| 对焦 | 相位检测 | AI 驱动 |
| 防抖 | 5.5 级 | 8 级 |
| 屏幕 | 翻转 | 四轴翻转 |
| 视频 | 4K 30p | 4K 60p + 8K 25p |
| 价格 | 约 ¥16,000 | 约 ¥21,000 |

### 适合谁？

![用户画像](https://picsum.photos/seed/a7rv-user/1200/800)

**强烈推荐**

- 风光摄影师：高像素、强防抖、高动态范围
- 建筑摄影师：像素密度高，后期裁剪空间大
- 商业人像：解析力足够应付大幅输出

**可以考虑**

- 纪实摄影师：如果预算充足，AI 对焦很香

**不太适合**

- 体育摄影师：连拍速度一般
- 视频创作者：建议 A7S III 或 FX3
- 旅行爱好者：重量和存储压力较大

---

*一年下来，A7R V 是我用过最满意的风光机身。如果你追求画质和细节，它不会让你失望。*$$,
   'https://picsum.photos/seed/sony-a7rv/1600/900',
   '10000000-0000-0000-0000-000000000002',
   2, 1, false, true, NOW()),

  ('30000000-0000-0000-0000-000000000005',
   '无人机航拍入门：从法规到构图',
   'drone-beginner-guide',
   '第一次飞无人机需要注意什么？这篇指南带你安全起飞，拍出稳定大气的航拍作品。',
   $$## 起飞前必读

![无人机安全飞行](https://picsum.photos/seed/drone-safety/1200/800)

在拿到无人机之前，有几件事必须了解：

**实名登记**

在中国大陆，250g 以上无人机需在 UOM 系统实名登记。登记后会在机身贴上识别码，飞行时必须携带登记证明。

**了解禁飞区**

![禁飞区地图](https://picsum.photos/seed/no-fly-zone/1200/800)

机场、军事区、政府机关、人群密集区严禁飞行。大部分无人机 App 会自动提示禁飞区，但偏远地区需要自行判断。

> 建议：起飞前在 App 内查看当前区域是否允许飞行，不确定就不要飞。

**保险**

建议购买第三者责任险。无人机坠落伤人、损坏财物，保险能帮你分担风险。DJI 官方就有 Care 随心换，包含意外保障。

### 新手第一飞

![第一次飞行](https://picsum.photos/seed/first-flight/1200/800)

拿到无人机后，不要急着去复杂环境。建议：

1. **先在开阔草地练习**：远离人群、建筑、树木
2. **熟悉遥控器手感**：小幅推杆，感受响应
3. **练习基本动作**：悬停、前进后退、左右平移、原地旋转
4. **尝试智能模式**：一键短片、智能跟随，体验自动化

### 航拍构图建议

![航拍构图示例](https://picsum.photos/seed/drone-composition/1200/800)

无人机最大的价值不是飞得高，而是提供一个全新的观看世界的角度。以下是几个构图技巧：

**寻找线条**

![线条引导](https://picsum.photos/seed/drone-lines/1200/800)

道路、河流、海岸线都是天然引导线。从空中俯瞰，线条会把观众视线引向画面焦点。

**利用阴影**

日出日落的长影能让画面更立体。尽量选择侧光时段，避免正午顶光。

**保持简洁**

空中视角容易杂乱。学会做减法：只保留一到两个视觉重点，其余留白。

**尝试不同高度**

![不同高度的效果](https://picsum.photos/seed/drone-height/1200/800)

| 高度 | 视觉效果 | 适用场景 |
|------|----------|----------|
| 10-30m | 环境细节丰富 | 建筑、街景 |
| 50-100m | 线条感强 | 道路、河流 |
| 100-200m | 抽象化 | 沙漠、草原、海岸 |

### 常见问题与解决

**信号丢失**

无人机超出遥控范围时会自动返航。建议提前设置返航高度（高于周围建筑），避免返航途中撞楼。

**电池续航短**

官方续航数据通常是理想条件。实际飞行时，建议 15 分钟内返航，预留 30% 电量。

**风大不敢飞**

一般无人机抗风等级为 5 级（约 10m/s）。起飞前观察树叶摇摆程度，如果"树弯腰"，就不要飞了。

### 器材推荐

![航拍器材推荐](https://picsum.photos/seed/drone-gear/1200/800)

| 需求 | 推荐 | 特点 |
|------|------|------|
| 入门 | DJI Mini 4 Pro | 249g 免登记，便携，画质够用 |
| 进阶 | DJI Air 3 | 双摄，性价比高 |
| 专业 | DJI Mavic 3 Pro | 三摄，哈苏主摄，画质最佳 |

### 飞行安全清单

每次起飞前，检查以下事项：

- [ ] 电池已充满
- [ ] 存储卡已插入
- [ ] 桨叶无损伤
- [ ] GPS 信号良好（≥8 颗星）
- [ ] 禁飞区已确认
- [ ] 天气良好（无雨、风小）
- [ ] 返航点已记录

---

*无人机打开了摄影的新维度。安全飞行，理性航拍，让这片天空更友好。*$$,
   'https://picsum.photos/seed/drone-aerial/1600/900',
   '10000000-0000-0000-0000-000000000001',
   2, 1, false, true, NOW())
ON CONFLICT (slug) DO NOTHING;

-- 文章标签关联
INSERT INTO article_tags (article_id, tag_id) VALUES
  ('30000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001'),
  ('30000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000005'),
  ('30000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000007'),
  ('30000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002'),
  ('30000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000002'),
  ('30000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000001'),
  ('30000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000003'),
  ('30000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000001'),
  ('30000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000002')
ON CONFLICT (article_id, tag_id) DO NOTHING;

-- ============================================================
-- 4. 摄影器材
-- ============================================================
INSERT INTO photo_equipment (id, name, image_url, brand, description, sort_order) VALUES
  ('40000000-0000-0000-0000-000000000001',
   'Sony A7R V 机身',
   'https://picsum.photos/seed/sony-a7rv/800/800',
   'Sony',
   '6100 万像素全画幅微单，搭载 AI 对焦芯片，适合风光、人像与商业摄影。', 1),
  ('40000000-0000-0000-0000-000000000002',
   'FE 24-70mm f/2.8 GM II',
   'https://picsum.photos/seed/sony-2470gm2/800/800',
   'Sony',
   '万金油标准变焦镜头，画质与便携兼顾，旅行挂机首选。', 2),
  ('40000000-0000-0000-0000-000000000003',
   'FE 85mm f/1.4 GM',
   'https://picsum.photos/seed/sony-85gm/800/800',
   'Sony',
   '焦外柔美、锐度出众，是人像与弱光环境下的得力工具。', 3),
  ('40000000-0000-0000-0000-000000000004',
   'DJI Mavic 3 Pro',
   'https://picsum.photos/seed/dji-mavic3/800/800',
   'DJI',
   '三摄旗舰航拍无人机，哈苏主摄与双长焦组合，上帝视角利器。', 4),
  ('40000000-0000-0000-0000-000000000005',
   'Peak Design Everyday Backpack 30L',
   'https://picsum.photos/seed/peak-design-bag/800/800',
   'Peak Design',
   '分区灵活的摄影背包，取机方便，城市与户外兼顾。', 5),
  ('40000000-0000-0000-0000-000000000006',
   'NiSi V7 滤镜套装',
   'https://picsum.photos/seed/nisi-v7/800/800',
   'NiSi',
   '包含 CPL、ND 与 GND，有效控制反光、延长曝光与平衡天空光比。', 6)
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- 5. 媒体与媒体预设（供作品集调用）
-- ============================================================
INSERT INTO media (id, filename, file_type, mime_type, size, url, storage_type, width, height) VALUES
  ('50000000-0000-0000-0000-000000000001', 'landscape-sunrise.jpg', 1, 'image/jpeg', 2048000, 'https://picsum.photos/seed/landscape-sunrise/1600/900', 'local', 1600, 900),
  ('50000000-0000-0000-0000-000000000002', 'landscape-milkyway.jpg', 1, 'image/jpeg', 1980000, 'https://picsum.photos/seed/landscape-milkyway/1600/900', 'local', 1600, 900),
  ('50000000-0000-0000-0000-000000000003', 'landscape-lake.jpg', 1, 'image/jpeg', 2100000, 'https://picsum.photos/seed/landscape-lake/1600/900', 'local', 1600, 900),
  ('50000000-0000-0000-0000-000000000004', 'street-rainy-night.jpg', 1, 'image/jpeg', 1600000, 'https://picsum.photos/seed/street-rainy-night/1600/900', 'local', 1600, 900),
  ('50000000-0000-0000-0000-000000000005', 'street-market.jpg', 1, 'image/jpeg', 1750000, 'https://picsum.photos/seed/street-market/1600/900', 'local', 1600, 900),
  ('50000000-0000-0000-0000-000000000006', 'street-silhouette.jpg', 1, 'image/jpeg', 1650000, 'https://picsum.photos/seed/street-silhouette/1600/900', 'local', 1600, 900),
  ('50000000-0000-0000-0000-000000000007', 'portrait-window.jpg', 1, 'image/jpeg', 1800000, 'https://picsum.photos/seed/portrait-window/1600/900', 'local', 1600, 900),
  ('50000000-0000-0000-0000-000000000008', 'portrait-golden.jpg', 1, 'image/jpeg', 1900000, 'https://picsum.photos/seed/portrait-golden/1600/900', 'local', 1600, 900),
  ('50000000-0000-0000-0000-000000000009', 'portrait-environmental.jpg', 1, 'image/jpeg', 1850000, 'https://picsum.photos/seed/portrait-environmental/1600/900', 'local', 1600, 900),
  ('50000000-0000-0000-0000-000000000010', 'aerial-coastline.jpg', 1, 'image/jpeg', 2200000, 'https://picsum.photos/seed/aerial-coastline/1600/900', 'local', 1600, 900),
  ('50000000-0000-0000-0000-000000000011', 'aerial-city-night.jpg', 1, 'image/jpeg', 2150000, 'https://picsum.photos/seed/aerial-city-night/1600/900', 'local', 1600, 900),
  ('50000000-0000-0000-0000-000000000012', 'aerial-countryside.jpg', 1, 'image/jpeg', 2050000, 'https://picsum.photos/seed/aerial-countryside/1600/900', 'local', 1600, 900)
ON CONFLICT (id) DO NOTHING;

INSERT INTO media_presets (id, media_id, name, frame_config, display_params, output_url, output_storage_path, output_size, mime_type) VALUES
  ('60000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000001', '原图', '{}', '{}', 'https://picsum.photos/seed/landscape-sunrise/1600/900', 'seed/landscape-sunrise.jpg', 2048000, 'image/jpeg'),
  ('60000000-0000-0000-0000-000000000002', '50000000-0000-0000-0000-000000000002', '原图', '{}', '{}', 'https://picsum.photos/seed/landscape-milkyway/1600/900', 'seed/landscape-milkyway.jpg', 1980000, 'image/jpeg'),
  ('60000000-0000-0000-0000-000000000003', '50000000-0000-0000-0000-000000000003', '原图', '{}', '{}', 'https://picsum.photos/seed/landscape-lake/1600/900', 'seed/landscape-lake.jpg', 2100000, 'image/jpeg'),
  ('60000000-0000-0000-0000-000000000004', '50000000-0000-0000-0000-000000000004', '原图', '{}', '{}', 'https://picsum.photos/seed/street-rainy-night/1600/900', 'seed/street-rainy-night.jpg', 1600000, 'image/jpeg'),
  ('60000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000005', '原图', '{}', '{}', 'https://picsum.photos/seed/street-market/1600/900', 'seed/street-market.jpg', 1750000, 'image/jpeg'),
  ('60000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000006', '原图', '{}', '{}', 'https://picsum.photos/seed/street-silhouette/1600/900', 'seed/street-silhouette.jpg', 1650000, 'image/jpeg'),
  ('60000000-0000-0000-0000-000000000007', '50000000-0000-0000-0000-000000000007', '原图', '{}', '{}', 'https://picsum.photos/seed/portrait-window/1600/900', 'seed/portrait-window.jpg', 1800000, 'image/jpeg'),
  ('60000000-0000-0000-0000-000000000008', '50000000-0000-0000-0000-000000000008', '原图', '{}', '{}', 'https://picsum.photos/seed/portrait-golden/1600/900', 'seed/portrait-golden.jpg', 1900000, 'image/jpeg'),
  ('60000000-0000-0000-0000-000000000009', '50000000-0000-0000-0000-000000000009', '原图', '{}', '{}', 'https://picsum.photos/seed/portrait-environmental/1600/900', 'seed/portrait-environmental.jpg', 1850000, 'image/jpeg'),
  ('60000000-0000-0000-0000-000000000010', '50000000-0000-0000-0000-000000000010', '原图', '{}', '{}', 'https://picsum.photos/seed/aerial-coastline/1600/900', 'seed/aerial-coastline.jpg', 2200000, 'image/jpeg'),
  ('60000000-0000-0000-0000-000000000011', '50000000-0000-0000-0000-000000000011', '原图', '{}', '{}', 'https://picsum.photos/seed/aerial-city-night/1600/900', 'seed/aerial-city-night.jpg', 2150000, 'image/jpeg'),
  ('60000000-0000-0000-0000-000000000012', '50000000-0000-0000-0000-000000000012', '原图', '{}', '{}', 'https://picsum.photos/seed/aerial-countryside/1600/900', 'seed/aerial-countryside.jpg', 2050000, 'image/jpeg')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- 6. 摄影作品集与作品项
-- ============================================================
INSERT INTO portfolios (id, name, description, status, sort_order, category_id, cover_mode) VALUES
  ('70000000-0000-0000-0000-000000000001',
   '川西风光集',
   '雪山、海子与星空，记录川西高原的壮丽与宁静。',
   1, 1, '10000000-0000-0000-0000-000000000004', 0),
  ('70000000-0000-0000-0000-000000000002',
   '城市光影集',
   '雨夜霓虹、市井烟火与逆光剪影，关于城市的视觉日记。',
   1, 2, '10000000-0000-0000-0000-000000000005', 0),
  ('70000000-0000-0000-0000-000000000003',
   '自然光人像习作',
   '窗边、黄昏与环境人像，探索自然光下的人物情绪。',
   1, 3, '10000000-0000-0000-0000-000000000006', 0),
  ('70000000-0000-0000-0000-000000000004',
   '航拍中国',
   '从空中俯瞰大地， coastline、城市脉络与田园肌理。',
   1, 4, '10000000-0000-0000-0000-000000000009', 0)
ON CONFLICT (id) DO NOTHING;

INSERT INTO portfolio_items (id, portfolio_id, preset_id, title, description, sort_order) VALUES
  ('80000000-0000-0000-0000-000000000001', '70000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000001', '日照金山', '清晨第一缕阳光洒向雪山之巅。', 1),
  ('80000000-0000-0000-0000-000000000002', '70000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000002', '高原星河', '海拔四千米处的银河拱桥。', 2),
  ('80000000-0000-0000-0000-000000000003', '70000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000003', '湖面倒影', '无风时，雪山完整映入湖中。', 3),
  ('80000000-0000-0000-0000-000000000004', '70000000-0000-0000-0000-000000000002', '60000000-0000-0000-0000-000000000004', '雨夜霓虹', '湿漉漉的街道反射着城市灯光。', 1),
  ('80000000-0000-0000-0000-000000000005', '70000000-0000-0000-0000-000000000002', '60000000-0000-0000-0000-000000000005', '菜市场烟火', '清晨市场里忙碌的人群与色彩。', 2),
  ('80000000-0000-0000-0000-000000000006', '70000000-0000-0000-0000-000000000002', '60000000-0000-0000-0000-000000000006', '逆光剪影', '黄昏时分街头的匆匆过客。', 3),
  ('80000000-0000-0000-0000-000000000007', '70000000-0000-0000-0000-000000000003', '60000000-0000-0000-0000-000000000007', '窗边人像', '柔和的自然光勾勒出面部轮廓。', 1),
  ('80000000-0000-0000-0000-000000000008', '70000000-0000-0000-0000-000000000003', '60000000-0000-0000-0000-000000000008', '黄昏人像', '金色时刻下温暖的肤色与眼神。', 2),
  ('80000000-0000-0000-0000-000000000009', '70000000-0000-0000-0000-000000000003', '60000000-0000-0000-0000-000000000009', '环境人像', '人物与场景的叙事关系。', 3),
  ('80000000-0000-0000-0000-000000000010', '70000000-0000-0000-0000-000000000004', '60000000-0000-0000-0000-000000000010', '蜿蜒海岸线', '从空中看海与陆的交界。', 1),
  ('80000000-0000-0000-0000-000000000011', '70000000-0000-0000-0000-000000000004', '60000000-0000-0000-0000-000000000011', '城市夜景', '灯光组成的城市血管。', 2),
  ('80000000-0000-0000-0000-000000000012', '70000000-0000-0000-0000-000000000004', '60000000-0000-0000-0000-000000000012', '田园肌理', '大地色块如抽象画一般。', 3)
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- 7. 旅行攻略
-- ============================================================
INSERT INTO travel_guides (id, title, summary, cover_image, status, destination, region, days, best_month, category_id, rating, review_count, attractions, itinerary, reviews) VALUES
  ('90000000-0000-0000-0000-000000000001',
   '川西七日摄影自驾：雪山、海子与星空',
   '从成都出发，经四姑娘山、新都桥、稻城亚丁，一路向西，记录雪山、草原与高原星河。',
   'https://picsum.photos/seed/sichuan-west/1600/900',
   2, '四川川西', 'china', 7, '9月-10月',
   '10000000-0000-0000-0000-000000000007',
   4.8, 3,
   $$[
    {"name":"四姑娘山","description":"东方阿尔卑斯，雪山与森林交织","image":"https://picsum.photos/seed/siguniang/800/600","duration":"半天","location":"四川阿坝小金县"},
    {"name":"新都桥","description":"摄影师的天堂，光影与藏寨","image":"https://picsum.photos/seed/xinduqiao/800/600","duration":"1天","location":"四川甘孜康定市"},
    {"name":"稻城亚丁","description":"三神山与牛奶海，蓝色星球最后净土","image":"https://picsum.photos/seed/daocheng/800/600","duration":"2天","location":"四川甘孜稻城县"}
  ]$$::jsonb,
   $$[
    {"day":1,"title":"成都集合","description":"检查装备，采购补给","attractions":[],"attractionIds":[]},
    {"day":2,"title":"四姑娘山","description":"双桥沟拍摄雪山倒影","attractions":[{"name":"四姑娘山","description":"东方阿尔卑斯","image":"https://picsum.photos/seed/siguniang/800/600","duration":"半天","location":"四川阿坝小金县"}],"attractionIds":[]},
    {"day":3,"title":"新都桥","description":"黄昏时分拍摄藏寨与杨树","attractions":[{"name":"新都桥","description":"摄影师的天堂","image":"https://picsum.photos/seed/xinduqiao/800/600","duration":"1天","location":"四川甘孜康定市"}],"attractionIds":[]},
    {"day":4,"title":"稻城亚丁长线","description":"徒步牛奶海与五色海","attractions":[{"name":"稻城亚丁","description":"蓝色星球最后净土","image":"https://picsum.photos/seed/daocheng/800/600","duration":"2天","location":"四川甘孜稻城县"}],"attractionIds":[]},
    {"day":5,"title":"亚丁短线与星空","description":"冲古寺与珍珠海，夜间拍摄银河","attractions":[],"attractionIds":[]},
    {"day":6,"title":"返程新都桥","description":"沿途补拍日照金山","attractions":[],"attractionIds":[]},
    {"day":7,"title":"返回成都","description":"整理照片，结束行程","attractions":[],"attractionIds":[]}
  ]$$::jsonb,
   $$[
    {"username":"雪山控","rating":5,"content":"亚丁长线虽然累，但牛奶海的蓝真的值得。","avatar":"https://picsum.photos/seed/reviewer1/100/100","date":"2024-10-15T08:00:00Z"},
    {"username":"自驾达人","rating":4,"content":"路上海拔变化大，注意高反。风景无敌。","avatar":"https://picsum.photos/seed/reviewer2/100/100","date":"2024-10-20T10:30:00Z"},
    {"username":"风光狗","rating":5,"content":"新都桥的光影每天都不一样，待了三天没拍够。","avatar":"https://picsum.photos/seed/reviewer3/100/100","date":"2024-10-22T18:00:00Z"}
  ]$$::jsonb),

  ('90000000-0000-0000-0000-000000000002',
   '日本北海道：雪国摄影十日漫游',
   '从札幌到小樽、富良野、函馆，雪国北海道是冬季摄影师的天堂。',
   'https://picsum.photos/seed/hokkaido-winter/1600/900',
   2, '日本北海道', 'japan', 10, '12月-2月',
   '10000000-0000-0000-0000-000000000008',
   4.9, 2,
   $$[
    {"name":"小樽运河","description":"煤油灯与雪景的浪漫组合","image":"https://picsum.photos/seed/otaru/800/600","duration":"2小时","location":"日本北海道小樽市"},
    {"name":"美瑛丘陵","description":"雪原与孤树，极简风光","image":"https://picsum.photos/seed/biei/800/600","duration":"半天","location":"日本北海道上川郡美瑛町"},
    {"name":"函馆山夜景","description":"世界三大夜景之一","image":"https://picsum.photos/seed/hakodate/800/600","duration":"2小时","location":"日本北海道函馆市"}
  ]$$::jsonb,
   $$[
    {"day":1,"title":"抵达札幌","description":"适应时差，拍摄城市夜景","attractions":[],"attractionIds":[]},
    {"day":2,"title":"小樽一日游","description":"运河、仓库与寿司街","attractions":[{"name":"小樽运河","description":"煤油灯与雪景","image":"https://picsum.photos/seed/otaru/800/600","duration":"2小时","location":"日本北海道小樽市"}],"attractionIds":[]},
    {"day":3,"title":"旭川动物园","description":"企鹅散步，极地动物","attractions":[],"attractionIds":[]},
    {"day":4,"title":"美瑛雪原","description":"拍摄雪地里的孤树与丘陵","attractions":[{"name":"美瑛丘陵","description":"雪原与孤树","image":"https://picsum.photos/seed/biei/800/600","duration":"半天","location":"日本北海道上川郡美瑛町"}],"attractionIds":[]},
    {"day":5,"title":"富良野滑雪","description":"粉雪与度假村","attractions":[],"attractionIds":[]},
    {"day":6,"title":"登别温泉","description":"地狱谷与温泉街","attractions":[],"attractionIds":[]},
    {"day":7,"title":"洞爷湖","description":"火山湖与雪中游船","attractions":[],"attractionIds":[]},
    {"day":8,"title":"函馆","description":"五棱郭与夜景","attractions":[{"name":"函馆山夜景","description":"世界三大夜景之一","image":"https://picsum.photos/seed/hakodate/800/600","duration":"2小时","location":"日本北海道函馆市"}],"attractionIds":[]},
    {"day":9,"title":"函馆朝市","description":"海鲜早餐与人文扫街","attractions":[],"attractionIds":[]},
    {"day":10,"title":"返程","description":"从函馆机场离开","attractions":[],"attractionIds":[]}
  ]$$::jsonb,
   $$[
    {"username":"雪景控","rating":5,"content":"冬天的小樽像童话世界，函馆夜景太美了。","avatar":"https://picsum.photos/seed/reviewer4/100/100","date":"2024-02-10T09:00:00Z"},
    {"username":"北海道常客","rating":5,"content":"美瑛的雪地极简风非常出片，建议自驾。","avatar":"https://picsum.photos/seed/reviewer5/100/100","date":"2024-02-18T11:20:00Z"}
  ]$$::jsonb),

  ('90000000-0000-0000-0000-000000000003',
   '新疆伊犁：草原花海十五日自驾',
   '独库公路、赛里木湖、那拉提草原，用车轮丈量天山脚下的壮美画卷。',
   'https://picsum.photos/seed/xinjiang-yili/1600/900',
   2, '新疆伊犁', 'china', 15, '6月-8月',
   '10000000-0000-0000-0000-000000000007',
   4.7, 2,
   $$[
    {"name":"赛里木湖","description":"大西洋最后一滴眼泪","image":"https://picsum.photos/seed/sayram/800/600","duration":"全天","location":"新疆博尔塔拉州"},
    {"name":"那拉提草原","description":"空中草原，哈萨克牧歌","image":"https://picsum.photos/seed/nalati/800/600","duration":"全天","location":"新疆伊犁新源县"},
    {"name":"独库公路","description":"中国最美景观公路","image":"https://picsum.photos/seed/duku/800/600","duration":"全天","location":"新疆独山子至库车"}
  ]$$::jsonb,
   $$[
    {"day":1,"title":"乌鲁木齐集合","description":"取车、采购、召开行程会","attractions":[],"attractionIds":[]},
    {"day":2,"title":"S101省道","description":"百里丹霞地貌","attractions":[],"attractionIds":[]},
    {"day":3,"title":"赛里木湖","description":"环湖自驾与星空","attractions":[{"name":"赛里木湖","description":"大西洋最后一滴眼泪","image":"https://picsum.photos/seed/sayram/800/600","duration":"全天","location":"新疆博尔塔拉州"}],"attractionIds":[]},
    {"day":4,"title":"霍城薰衣草","description":"东方普罗旺斯的紫色花海","attractions":[],"attractionIds":[]},
    {"day":5,"title":"伊宁老城","description":"喀赞其民俗村","attractions":[],"attractionIds":[]},
    {"day":6,"title":"夏塔古道","description":"雪山森林徒步","attractions":[],"attractionIds":[]},
    {"day":7,"title":"特克斯八卦城","description":"世界最大八卦布局城市","attractions":[],"attractionIds":[]},
    {"day":8,"title":"喀拉峻草原","description":"人体草原摄影","attractions":[],"attractionIds":[]},
    {"day":9,"title":"那拉提草原","description":"空中草原与牧民生活","attractions":[{"name":"那拉提草原","description":"空中草原","image":"https://picsum.photos/seed/nalati/800/600","duration":"全天","location":"新疆伊犁新源县"}],"attractionIds":[]},
    {"day":10,"title":"独库公路北段","description":"翻越天山","attractions":[{"name":"独库公路","description":"中国最美景观公路","image":"https://picsum.photos/seed/duku/800/600","duration":"全天","location":"新疆独山子至库车"}],"attractionIds":[]},
    {"day":11,"title":"巴音布鲁克","description":"九曲十八弯日落","attractions":[],"attractionIds":[]},
    {"day":12,"title":"独库公路南段","description":"天山神秘大峡谷","attractions":[],"attractionIds":[]},
    {"day":13,"title":"库车老城","description":"龟兹文化与馕坑肉","attractions":[],"attractionIds":[]},
    {"day":14,"title":"返程乌鲁木齐","description":"沿途补拍","attractions":[],"attractionIds":[]},
    {"day":15,"title":"结束行程","description":"还车返程","attractions":[],"attractionIds":[]}
  ]$$::jsonb,
   $$[
    {"username":"自驾狂人","rating":5,"content":"新疆自驾天花板，风景太壮阔了！","avatar":"https://picsum.photos/seed/reviewer6/100/100","date":"2024-07-05T08:00:00Z"},
    {"username":"草原控","rating":4,"content":"路程比较长，但风景绝对值得。","avatar":"https://picsum.photos/seed/reviewer7/100/100","date":"2024-07-12T14:30:00Z"}
  ]$$::jsonb)
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- 8. 评论
-- ============================================================
INSERT INTO comments (id, target_type, target_id, parent_id, nickname, email, avatar, content, is_blogger, status) VALUES
  ('a0000000-0000-0000-0000-000000000001', 'article', '30000000-0000-0000-0000-000000000001', NULL, '追光者', 'chaser@example.com', 'https://picsum.photos/seed/chaser/100/100', '黄金时刻的拍摄思路很受用，下次去高原试试包围曝光。', false, 2),
  ('a0000000-0000-0000-0000-000000000002', 'article', '30000000-0000-0000-0000-000000000004', NULL, '器材党小王', 'gear@example.com', 'https://picsum.photos/seed/gearwang/100/100', 'A7R V 的存储压力确实大，请问平时用什么卡？', false, 2),
  ('a0000000-0000-0000-0000-000000000003', 'article', '30000000-0000-0000-0000-000000000005', NULL, '飞手阿杰', 'drone@example.com', 'https://picsum.photos/seed/dronejie/100/100', '航拍构图那部分讲得很清楚，新手友好。', false, 2),
  ('a0000000-0000-0000-0000-000000000004', 'travel_guide', '90000000-0000-0000-0000-000000000001', NULL, '川西常客', 'west@example.com', 'https://picsum.photos/seed/west/100/100', '亚丁长线建议带够氧气，风景真的值得。', false, 2),
  ('a0000000-0000-0000-0000-000000000005', 'travel_guide', '90000000-0000-0000-0000-000000000002', NULL, '雪国旅人', 'snow@example.com', 'https://picsum.photos/seed/snow/100/100', '北海道冬天路面滑，租车一定要换雪胎。', false, 2),
  ('a0000000-0000-0000-0000-000000000006', 'travel_guide', '90000000-0000-0000-0000-000000000003', NULL, '新疆自驾老李', 'xinjiang@example.com', 'https://picsum.photos/seed/xinjiangli/100/100', '独库公路每年开放时间不固定，出发前一定要查官方公告。', false, 2)
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- 9. 个人资料（仅更新文字，不覆盖已有图片）
-- ============================================================
UPDATE bloggers
SET
  nickname = '林夕（旅行摄影师）',
  bio = '自由摄影师，热爱风光、旅行与人文记录。相信每一束光都有故事，每一次快门都是与世界的对话。',
  blog_title = '光影行记',
  blog_description = '一个关于摄影、旅行与生活的个人博客，记录路上的风景与心中的光影。',
  social_links = $$[
    {"platform":"weibo","url":"https://weibo.com/example","sort_order":1},
    {"platform":"xiaohongshu","url":"https://xiaohongshu.com/example","sort_order":2},
    {"platform":"instagram","url":"https://instagram.com/example","sort_order":3},
    {"platform":"bilibili","url":"https://bilibili.com/example","sort_order":4}
  ]$$,
  tags = $$["风光摄影","旅行","器材","后期处理"]$$,
  updated_at = NOW()
WHERE id = (SELECT id FROM bloggers ORDER BY created_at LIMIT 1);

COMMIT;