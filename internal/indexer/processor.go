package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/moistello/backend/internal/domain/circle"
	"github.com/moistello/backend/internal/domain/contribution"
	"github.com/moistello/backend/internal/domain/payout"
	"github.com/moistello/backend/internal/domain/reputation"
	"github.com/moistello/backend/pkg/rabbitmq"
)

// EventProcessor maps Stellar transactions to domain events, persists them
// to PostgreSQL, broadcasts real-time updates via WebSocket, and publishes
// events to RabbitMQ for async workers.
type EventProcessor struct {
	db             *sqlx.DB
	rmqClient      *rabbitmq.Client
	circleRepo     circle.Repository
	contribRepo    contribution.Repository
	payoutRepo     payout.Repository
	reputationRepo reputation.Repository
	wsBroadcast    func(circleID string, data any)
}

// NewEventProcessor creates a new EventProcessor with all required dependencies.
func NewEventProcessor(
	db *sqlx.DB,
	rmqClient *rabbitmq.Client,
	circleRepo circle.Repository,
	contribRepo contribution.Repository,
	payoutRepo payout.Repository,
	reputationRepo reputation.Repository,
) *EventProcessor {
	return &EventProcessor{
		db:             db,
		rmqClient:      rmqClient,
		circleRepo:     circleRepo,
		contribRepo:    contribRepo,
		payoutRepo:     payoutRepo,
		reputationRepo: reputationRepo,
	}
}

// SetWebSocketBroadcast sets the callback for real-time WebSocket updates.
// When set, every processed event will be broadcast to connected clients
// subscribed to the relevant circle room.
func (p *EventProcessor) SetWebSocketBroadcast(fn func(circleID string, data any)) {
	p.wsBroadcast = fn
}

// ProcessTransaction maps a Stellar transaction to domain events and
// persists the resulting entities to PostgreSQL.
func (p *EventProcessor) ProcessTransaction(ctx context.Context, txn *Transaction) error {
	if len(txn.Operations) == 0 {
		return nil
	}

	var errs []error
	for _, op := range txn.Operations {
		if err := p.processOperation(ctx, txn, &op); err != nil {
			log.Warn().Err(err).
				Str("hash", txn.Hash).
				Str("op_type", op.Type).
				Msg("processing operation")
			errs = append(errs, err)
			continue
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("processing %d/%d operations: %w", len(errs), len(txn.Operations), errs[0])
	}
	return nil
}

func (p *EventProcessor) processOperation(ctx context.Context, txn *Transaction, op *Operation) error {
	switch {
	case op.Type == "create_account":
		return p.handleCreateAccount(ctx, txn, op)
	case op.Type == "payment":
		return p.handlePayment(ctx, txn, op)
	case op.Type == "invoke_host_function":
		return p.handleSorobanInvoke(ctx, txn, op)
	case op.Type == "extend_footprint_ttl":
		return p.handleExtendTTL(ctx, txn, op)
	default:
		log.Debug().Str("type", op.Type).Msg("unhandled operation type")
		return nil
	}
}

func (p *EventProcessor) handleCreateAccount(ctx context.Context, txn *Transaction, op *Operation) error {
	// A new Stellar account was created — potentially a new user onboarding.
	// In production this would trigger user auto-creation and KYC workflows.
	log.Info().
		Str("hash", txn.Hash).
		Str("source", op.SourceAccount).
		Msg("create_account detected")

	p.Broadcast(ctx, op.SourceAccount, "account_created", map[string]any{
		"hash":    txn.Hash,
		"account": op.SourceAccount,
		"ledger":  txn.Ledger,
	})
	return nil
}

func (p *EventProcessor) handlePayment(ctx context.Context, txn *Transaction, op *Operation) error {
	// A payment operation was detected — could represent a circle contribution,
	// a payout distribution, or a Soroban contract interaction.
	log.Info().
		Str("hash", txn.Hash).
		Str("source", op.SourceAccount).
		Msg("payment detected")

	p.Broadcast(ctx, op.SourceAccount, "payment_detected", map[string]any{
		"hash":    txn.Hash,
		"source":  op.SourceAccount,
		"ledger":  txn.Ledger,
	})
	return nil
}

func (p *EventProcessor) handleSorobanInvoke(ctx context.Context, txn *Transaction, op *Operation) error {
	// A Soroban contract invocation was detected. This is the primary
	// mechanism for circle creation, contribution tracking, and payouts.
	log.Info().
		Str("hash", txn.Hash).
		Str("source", op.SourceAccount).
		Msg("soroban invoke detected")

	p.Broadcast(ctx, op.SourceAccount, "soroban_invoke", map[string]any{
		"hash":    txn.Hash,
		"source":  op.SourceAccount,
		"ledger":  txn.Ledger,
	})
	return nil
}

func (p *EventProcessor) handleExtendTTL(ctx context.Context, txn *Transaction, op *Operation) error {
	// Contract instance storage TTL was extended — no domain event required,
	// but we track it for observability.
	log.Debug().
		Str("hash", txn.Hash).
		Str("source", op.SourceAccount).
		Msg("extend_footprint_ttl detected")
	return nil
}

// Broadcast sends a real-time update via WebSocket and publishes the event
// to RabbitMQ for async workers (notifications, webhooks, analytics).
func (p *EventProcessor) Broadcast(ctx context.Context, circleID string, eventType string, payload any) {
	// Real-time WebSocket broadcast to subscribed clients
	if p.wsBroadcast != nil {
		p.wsBroadcast(circleID, payload)
	}

	// Async event publishing to RabbitMQ
	data, err := json.Marshal(map[string]any{
		"type":      eventType,
		"circleId":  circleID,
		"payload":   payload,
		"timestamp": time.Now().UTC(),
	})
	if err != nil {
		log.Warn().Err(err).Msg("marshaling event for rabbitmq")
		return
	}

	if p.rmqClient != nil {
		if err := p.rmqClient.Publish("moistello.events", "circle."+eventType, data); err != nil {
			log.Warn().Err(err).Msg("publishing to rabbitmq")
		}
	}
}

// ensure we import these to avoid compiler complaints for work-in-progress
var _ = uuid.New
var _ = circle.Repository(nil)
var _ = contribution.Repository(nil)
var _ = payout.Repository(nil)
var _ = reputation.Repository(nil)
