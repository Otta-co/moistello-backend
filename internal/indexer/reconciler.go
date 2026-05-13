package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

// Reconciler detects gaps in processed events by comparing the cursor position
// with the current chain height, and replays any missed ledgers.
type Reconciler struct {
	cursor    *CursorTracker
	poller    *Poller
	processor *EventProcessor
	dedup     *Deduplicator
	interval  time.Duration
}

// NewReconciler creates a Reconciler with the given dependencies.
func NewReconciler(
	cursor *CursorTracker,
	poller *Poller,
	processor *EventProcessor,
	dedup *Deduplicator,
) *Reconciler {
	return &Reconciler{
		cursor:    cursor,
		poller:    poller,
		processor: processor,
		dedup:     dedup,
		interval:  5 * time.Minute,
	}
}

// StartReconciliation runs reconciliation on a periodic ticker until the
// context is cancelled.
func (r *Reconciler) StartReconciliation(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = r.interval
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Info().Dur("interval", interval).Msg("reconciler started")

	for {
		select {
		case <-ticker.C:
			if err := r.Reconcile(ctx); err != nil {
				log.Error().Err(err).Msg("reconciliation failed")
			}
		case <-ctx.Done():
			log.Info().Msg("reconciler stopped")
			return
		}
	}
}

// Reconcile checks for gaps between the stored cursor and the latest chain
// ledger, and replays any missed ledgers in batches.
func (r *Reconciler) Reconcile(ctx context.Context) error {
	cursor, err := r.cursor.GetCurrent(ctx)
	if err != nil {
		return fmt.Errorf("reading cursor: %w", err)
	}

	// Fetch the latest ledger to determine the current chain height
	ledgers, err := r.poller.FetchLedgers(ctx, cursor.LastLedger, 1)
	if err != nil {
		return fmt.Errorf("fetching latest ledger: %w", err)
	}
	if len(ledgers) == 0 {
		return nil
	}

	latestChain := ledgers[0].Sequence + int64(len(ledgers)) - 1
	gap := latestChain - cursor.LastLedger

	if gap > 1000 {
		log.Warn().
			Int64("gap", gap).
			Int64("cursor", cursor.LastLedger).
			Int64("chain", latestChain).
			Msg("large gap detected — indexer may be behind")
	}

	// Nothing to catch up
	if gap <= 0 {
		return nil
	}

	// Fetch missed ledgers in batches
	missedLedgers, err := r.poller.FetchLedgers(ctx, cursor.LastLedger, 50)
	if err != nil {
		return fmt.Errorf("fetching missed ledgers: %w", err)
	}

	processed := 0
	skipped := 0
	for _, ledger := range missedLedgers {
		txns, err := r.poller.FetchTransactions(ctx, ledger.Sequence)
		if err != nil {
			log.Warn().Err(err).Int64("ledger", ledger.Sequence).Msg("skipping ledger during reconciliation")
			continue
		}

		filtered := r.poller.FilterByContract(txns)
		for _, txn := range filtered {
			if r.dedup.Has(txn.Hash) {
				skipped++
				continue
			}
			r.dedup.Add(txn.Hash)

			if err := r.processor.ProcessTransaction(ctx, &txn); err != nil {
				log.Warn().Err(err).
					Str("hash", txn.Hash).
					Msg("reconciler process error")
				continue
			}
			processed++
		}

		// Update the cursor after each successfully processed ledger
		if err := r.cursor.Update(ctx, ledger.Sequence); err != nil {
			return fmt.Errorf("updating cursor during reconciliation: %w", err)
		}
	}

	if processed > 0 || skipped > 0 {
		log.Info().
			Int64("gap", gap).
			Int("replayed", processed).
			Int("skipped", skipped).
			Msg("reconciliation complete")
	} else if gap < 100 {
		log.Debug().Int64("gap", gap).Msg("reconciliation — no new events")
	}

	return nil
}
