# 第一阶段使用单实例 Goroutine 同步

第一阶段通过 HTTP 创建同步任务记录，并由 `SyncService` 拥有的后台 goroutine 执行任务。系统全局只允许一个 pending 或 running 同步任务，通过任务状态和进程内 mutex 保护；服务重启时会把 pending 或 running 任务视为失败。多实例同步执行、分布式锁、任务队列、worker pool 和取消能力都推迟实现。
