package market

import "testing"

var (
	_ StockRepository         = (*mysqlStockRepository)(nil)
	_ KLineRepository         = (*mysqlKLineRepository)(nil)
	_ TradeCalendarRepository = (*mysqlTradeCalendarRepository)(nil)
	_ SyncTaskRepository      = (*mysqlSyncTaskRepository)(nil)
	_ TxManager               = (*mysqlTxManager)(nil)
	_ TxRepositories          = mysqlRepositories{}
)

func TestSplitMySQLUpsertBatches(t *testing.T) {
	records := make([]int, mysqlUpsertBatchSize*2+1)
	for i := range records {
		records[i] = i
	}

	got := splitMySQLUpsertBatches(records)
	if len(got) != 3 {
		t.Fatalf("len(batches) = %d, want 3", len(got))
	}
	if len(got[0]) != mysqlUpsertBatchSize || len(got[1]) != mysqlUpsertBatchSize || len(got[2]) != 1 {
		t.Fatalf("batch sizes = [%d %d %d], want [%d %d 1]", len(got[0]), len(got[1]), len(got[2]), mysqlUpsertBatchSize, mysqlUpsertBatchSize)
	}
	if got[2][0] != mysqlUpsertBatchSize*2 {
		t.Fatalf("last item = %d, want %d", got[2][0], mysqlUpsertBatchSize*2)
	}
}
