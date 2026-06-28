# 使用分层测试和表驱动单元测试

第一阶段默认测试套件使用快速单元测试，并将 MySQL repository 集成测试放在 `integration` build tag 后面。单元测试在合适时应使用表驱动形式，provider 测试使用构造的原始响应而不进行真实 Tushare 网络调用；测试命令默认不禁用 Go 内联，`-gcflags=all="-N -l"` 仅保留给本地调试使用。
