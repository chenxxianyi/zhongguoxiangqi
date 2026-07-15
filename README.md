# 棋境（Xiangqi Lab）

中国象棋人机对战、棋谱学习与复盘项目。

## 当前内容

- `apps/web`：Vue 3 + TypeScript + Vite 前端
- `h5-demo`：已确认的纯 H5 视觉原型
- `前端开发方案.md`：前端实施方案
- `后端开发方案.md`：后端实施方案
- `xiangqi-ai-codex-prompt.md`：完整项目需求

## 前端启动

```powershell
npm.cmd install
npm.cmd run dev:web
```

访问 `http://localhost:5666/`。

## 前端验证

```powershell
npm.cmd run typecheck:web
npm.cmd run test:web
npm.cmd run lint:web
npm.cmd run build:web
```

当前 Vue 版本已完成八个核心页面、响应式布局、深浅主题、PWA、Router、Pinia、棋盘交互演示、棋谱导入演示和学习任务演示。真实规则、对局版本、AI 搜索与棋谱处理将在后端接入后替换演示逻辑。
