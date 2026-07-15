# 棋境（Xiangqi Lab）

中国象棋人机对战、棋谱学习与复盘项目。

## 当前内容

- `apps/web`：Vue 3 + TypeScript + Vite 前端
- `cmd/api`：Go HTTP、WebSocket/SSE API 服务
- `cmd/worker`：生产持久化队列接入前的独立 worker 入口
- `internal/domain/xiangqi`：服务端权威中国象棋规则核心
- `internal/engine/builtin`：可取消、受预算约束的基础 Alpha-Beta 引擎
- `internal/game`：对局状态机、乐观版本、幂等和异步 AI
- `internal/records`、`learning`、`analysis`：棋谱、学习版本和复盘闭环
- `migrations`：MySQL 8 数据模型基线
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

## 后端启动

```powershell
$env:GOCACHE = "$PWD\.cache\go-build"
go run ./cmd/api
```

API 默认访问 `http://localhost:8080/`。前端 Vite 已代理 `/api`、`/health` 和
WebSocket 到该端口。

健康检查：

```powershell
Invoke-RestMethod http://localhost:8080/health/live
Invoke-RestMethod http://localhost:8080/health/ready
```

也可运行：

```powershell
docker compose up --build
```

Compose 中前端仍使用 `http://localhost:5666/`。

## 后端验证

```powershell
$env:GOCACHE = "$PWD\.cache\go-build"
gofmt -w ./cmd ./internal
go vet ./...
go test ./...
go build ./cmd/api ./cmd/worker ./cmd/tools/perft ./cmd/tools/enginebench
```

## 前端验证

```powershell
npm.cmd run typecheck:web
npm.cmd run test:web
npm.cmd run lint:web
npm.cmd run build:web
```

当前后端默认采用内存权威仓库，适合直接联调；进程重启会丢失数据。MySQL schema 已提供，
但 MySQL/Redis 生产适配器和外部 Pikafish 进程适配器尚未伪装成已完成。具体接口、坐标、
架构边界和许可证说明见 `docs/`。
