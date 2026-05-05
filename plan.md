# VitePress 文档网站 404 和 Logo 异常修复

## 问题分析

### 问题 1：Tab 按钮点击 404
**原因**：当前配置启用了 `cleanUrls: true`，VitePress 使用基于 History 的路由，但 GitHub Pages 默认只支持