# Use single-instance goroutine sync in Phase 1

Phase 1 executes sync tasks by creating a sync task record from HTTP and running the work in a background goroutine owned by `SyncService`. The system allows only one pending or running sync task globally, guarded by task state plus an in-process mutex, and treats pending or running tasks as failed on service restart; multi-instance sync execution, distributed locks, task queues, worker pools, and cancellation are deferred.
