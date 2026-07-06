# 验收检查清单

## 前端
- [x] 左侧面板已新增「显示 EXIF」`a-switch`，绑定 `frameConfig.showExif`
- [x] 开关切换为关闭时，预览 Canvas 立即重绘且不绘制底部信息条
- [x] 开关切换为开启时，预览 Canvas 恢复底部信息条绘制
- [x] `doRender` 中 `infoHeight` 在 `showExif` 为 false 时为 0，画布高度仅含边框 + 原图
- [x] `renderFinalBlob` 在 `showExif` 为 false 时跳过 `injectExif`，成品图不写入 EXIF
- [x] `renderFinalBlob` 在 `showExif` 为 true 时保持原有 EXIF 注入行为
- [x] 关闭开关时左侧「参数设定」字段与右侧只读 EXIF 面板仍可见
- [x] `showExif` 状态随 `frame_config` JSON 保存，应用已有预设时正确恢复
- [x] `npm run build` 通过

## 联调
- [x] 开关关闭 → 保存预设 → 成品图无底部信息条且未注入 EXIF
- [x] 开关开启 → 保存预设 → 成品图有底部信息条且 EXIF 已注入
- [x] 应用已有预设时开关状态正确还原
