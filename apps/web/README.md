# 棋境 Web 前端

Vue 3 + TypeScript + Vite 转换版本，视觉基准来自根目录 `h5-demo`。

## 开发命令

在仓库根目录执行：

```powershell
npm.cmd install
npm.cmd run dev:web
npm.cmd run typecheck:web
npm.cmd run test:web
npm.cmd run lint:web
npm.cmd run build:web
```

也可以在 `apps/web` 目录直接运行对应 npm script。

## 页面路由

- `/` 首页
- `/new-game` 新对局
- `/match` 对弈棋盘
- `/records` 棋谱库
- `/learning` 学习中心
- `/analysis` 复盘分析
- `/history` 历史对局
- `/settings` 设置与诊断

## 说明

当前版本完成了 H5 Demo 的 Vue 组件化、Router 路由化和 Pinia 状态化。棋盘交互、文件导入和学习任务是前端演示逻辑；接入后端后，规则、局面、对局版本和引擎结果必须以服务端为权威。
