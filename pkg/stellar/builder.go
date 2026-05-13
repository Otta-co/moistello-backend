package stellar

import (
	"encoding/json"
	"fmt"
	"time"
)

// CreateAccountOp creates a Stellar create-account operation
type CreateAccountOp struct {
	Destination     string `json:"destination"`
	StartingBalance string `json:"starting_balance"`
}

// PaymentOp creates a Stellar payment operation
type PaymentOp struct {
	Destination string `json:"destination"`
	Asset       Asset  `json:"asset"`
	Amount      string `json:"amount"`
}

// SorobanInvokeOp is a Soroban contract invocation operation
type SorobanInvokeOp struct {
	ContractID string        `json:"contract_id"`
	Function   string        `json:"function"`
	Args       []SorobanArg  `json:"args"`
}

type SorobanArg struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Asset struct {
	Code   string `json:"code,omitempty"`
	Issuer string `json:"issuer,omitempty"`
}

// Transaction represents a Stellar transaction to be signed
type Transaction struct {
	SourceAccount string        `json:"source_account"`
	Fee           int64         `json:"fee"`
	Sequence      int64         `json:"sequence"`
	Operations    []interface{} `json:"operations"`
	Memo          string        `json:"memo,omitempty"`
	TimeBounds    *TimeBounds   `json:"time_bounds,omitempty"`
}

type TimeBounds struct {
	MinTime int64 `json:"min_time"`
	MaxTime int64 `json:"max_time"`
}

// TransactionBuilder constructs Stellar transactions
type TransactionBuilder struct {
	sourceAccount string
	fee           int64
	operations    []interface{}
	memo          string
	timeout       time.Duration
}

func NewTransactionBuilder(sourceAccount string) *TransactionBuilder {
	return &TransactionBuilder{
		sourceAccount: sourceAccount,
		fee:           100,
		timeout:       30 * time.Second,
	}
}

func (b *TransactionBuilder) SetFee(fee int64) *TransactionBuilder {
	b.fee = fee
	return b
}

func (b *TransactionBuilder) SetTimeout(d time.Duration) *TransactionBuilder {
	b.timeout = d
	return b
}

func (b *TransactionBuilder) SetMemo(memo string) *TransactionBuilder {
	b.memo = memo
	return b
}

func (b *TransactionBuilder) AddOperation(op interface{}) *TransactionBuilder {
	b.operations = append(b.operations, op)
	return b
}

func (b *TransactionBuilder) AddCreateAccount(destination, startingBalance string) *TransactionBuilder {
	return b.AddOperation(CreateAccountOp{Destination: destination, StartingBalance: startingBalance})
}

func (b *TransactionBuilder) AddPayment(destination string, asset Asset, amount string) *TransactionBuilder {
	return b.AddOperation(PaymentOp{Destination: destination, Asset: asset, Amount: amount})
}

func (b *TransactionBuilder) AddSorobanInvoke(contractID, function string, args []SorobanArg) *TransactionBuilder {
	return b.AddOperation(SorobanInvokeOp{ContractID: contractID, Function: function, Args: args})
}

func (b *TransactionBuilder) Build(sequence int64) *Transaction {
	now := time.Now().Unix()
	return &Transaction{
		SourceAccount: b.sourceAccount,
		Fee:           b.fee,
		Sequence:      sequence,
		Operations:    b.operations,
		Memo:          b.memo,
		TimeBounds: &TimeBounds{
			MinTime: now,
			MaxTime: now + int64(b.timeout.Seconds()) + 300,
		},
	}
}

func (t *Transaction) ToJSON() ([]byte, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("marshaling transaction: %w", err)
	}
	return data, nil
}

// ParseSequence converts a string sequence number to int64
func ParseSequence(seq string) (int64, error) {
	var result int64
	_, err := fmt.Sscanf(seq, "%d", &result)
	if err != nil {
		return 0, fmt.Errorf("parsing sequence %q: %w", seq, err)
	}
	return result, nil
}
