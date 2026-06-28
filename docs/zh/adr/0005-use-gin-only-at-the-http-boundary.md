# 仅在 HTTP 边界使用 Gin

第一阶段使用 Gin 处理路由、handler 和 HTTP middleware，但 `gin.Context` 不得进入应用 service 或 repository。这样可以保持 market service 易于测试，并避免 HTTP 框架关注点泄漏到同步任务、repository 代码或未来的非 HTTP 入口。
