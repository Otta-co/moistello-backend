package production

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/moistello/backend/internal/indexer"
	"github.com/moistello/backend/pkg/stellar"
)

func Benchmark_TransactionBuilder(b *testing.B) {
	pubKey := "GAX23V3WWDPPR5WRER3KTEUTDLSCGZYMSJY5FDRRKKCIQ4JADF5T27RC"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := stellar.NewTransactionBuilder(pubKey)
		builder.SetFee(100)
		builder.AddPayment(
			"GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H",
			stellar.Asset{Code: "XLM"},
			"1",
		)
		tx := builder.Build(int64(i))
		tx.ToJSON()
	}
}

func Benchmark_AccountManager_NextSequence(b *testing.B) {
	client := stellar.NewClient(
		"https://horizon-testnet.stellar.org",
		"https://soroban-testnet.stellar.org",
		"Test SDF Network ; September 2015",
	)
	mgr := stellar.NewAccountManager(client, "GAX23V3WWDPPR5WRER3KTEUTDLSCGZYMSJY5FDRRKKCIQ4JADF5T27RC")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mgr.NextSequence(ctx)
	}
}

func Benchmark_Deduplicator_Add(b *testing.B) {
	d := indexer.NewDeduplicator(1 * time.Hour)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Add(fmt.Sprintf("hash-%d", i%b.N))
	}
}

func Benchmark_Deduplicator_Has(b *testing.B) {
	d := indexer.NewDeduplicator(1 * time.Hour)
	for i := 0; i < 100000; i++ {
		d.Add(fmt.Sprintf("hash-%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Has(fmt.Sprintf("hash-%d", i%100000))
	}
}
