# Tasks

- [x] Task 1: 左侧面板新增「显示 EXIF」开关
  - [x] SubTask 1.1: 在 `MediaPresetWorkbench.vue` 左侧面板「参数设定」区块之上新增一个 section，包含「显示 EXIF」`a-switch`，绑定 `frameConfig.showExif`，开关变化触发已有 `watch` 重绘预览

- [x] Task 2: 修复预览渲染 `doRender` 遵循开关
  - [x] SubTask 2.1: 修改 `doRender`，根据 `frameConfig.showExif` 计算 `infoHeight`（false 时为 0），并在 `showExif` 为 false 时跳过信息条绘制逻辑，画布高度相应调整

- [x] Task 3: 修复保存合成 `renderFinalBlob` 与 EXIF 注入遵循开关
  - [x] SubTask 3.1: 修改 `renderFinalBlob` 中 `toBlob` 回调，当 `showExif` 为 false 时直接返回 blob，跳过 `injectExif` 调用
  - [x] SubTask 3.2: 验证 `showExif` 为 true 时保持原有 EXIF 注入行为不变

- [x] Task 4: 验证
  - [x] SubTask 4.1: 手动验证：开关关闭 → 预览无信息条 → 保存的成品图无信息条且无注入 EXIF
  - [x] SubTask 4.2: 手动验证：开关开启 → 预览有信息条 → 保存的成品图有信息条且 EXIF 已注入
  - [x] SubTask 4.3: 执行前端 `npm run build`，确保无编译错误

# Task Dependencies
- Task 2、Task 3 可与 Task 1 并行（均为同一文件不同逻辑块）
- Task 4 依赖 Task 1-3
