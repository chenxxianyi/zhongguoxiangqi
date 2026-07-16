# 棋境 Web 前端

Vue 3 + TypeScript + Vite 前端。生产运行时数据通过领域 API 模块读取 Go 后端，
不包含 Demo 棋局、固定统计或静态业务兜底。

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
- `/match/:id` 对弈棋盘
- `/records` 棋谱库
- `/learning` 学习中心
- `/analysis/:matchId` 复盘分析
- `/history` 历史对局
- `/settings` 设置与诊断

## 说明

棋盘规则、局面、对局版本、棋谱、学习任务、复盘结果、难度档位、引擎状态和
许可证信息均来自后端。接口失败时页面显示明确错误，不生成静态业务数据作为兜底。
