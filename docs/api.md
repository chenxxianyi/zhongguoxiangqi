# HTTP API

默认地址：`http://127.0.0.1:8080`，统一前缀 `/api/v1`。

## 创建并进行对局

```http
POST /api/v1/matches
Idempotency-Key: create-demo-1
Content-Type: application/json

{"playerColor":"red","difficulty":4,"allowUndo":true}
```

```http
POST /api/v1/matches/{id}/moves
Idempotency-Key: move-demo-1
Content-Type: application/json

{"move":"a3a4","expectedMatchVersion":1}
```

写接口的 `expectedMatchVersion` 不匹配时返回 `409 MATCH_VERSION_CONFLICT`。
非法着返回 `422 ILLEGAL_MOVE`。玩家落子返回 `202`，不等待 AI 搜索。

实时事件：

- WebSocket：`ws://127.0.0.1:8080/api/v1/matches/{id}/stream`
- SSE 降级：同一 HTTP URL，使用 `EventSource`
- 重连游标：查询参数 `afterEventId`；SSE 也支持 `Last-Event-ID`

## 棋谱

JSON 导入：

```json
{
  "name": "练习棋谱",
  "format": "iccs",
  "content": "a3a4 a6a5 b0c2",
  "result": "1-0"
}
```

也支持 `multipart/form-data`，文件字段名为 `file`。MVP 支持 ICCS 坐标文本、包含坐标
着法的 PGN 文本和项目 JSON。中文纵线记谱不会被静默猜测。

## 端点清单

端点与 `后端开发方案.md` 第 15 节一致，并补充：

- `GET /api/v1/matches`
- `GET /api/v1/records/{id}/moves`
- `GET /api/v1/learning/versions/{id}`
- `GET /api/v1/about/licenses`
- `GET /health/live`
- `GET /health/ready`

错误信封：

```json
{"code":"MATCH_VERSION_CONFLICT","message":"...","requestId":"...","details":null}
```

