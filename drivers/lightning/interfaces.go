// File: drivers/lightning/interfaces.go
// Purpose: Clean LightningDriver interface

package lightning

import (
	"context"
)

// LightningDriver is the abstraction all MeshGuard components use
type LightningDriver interface {
	GetInfo(ctx context.Context) (*NodeInfo, error)
	GetWalletBalance(ctx context.Context) (*WalletBalance, error)
	ListChannels(ctx context.Context) ([]Channel, error)
	AddInvoice(ctx context.Context, amountSats int64, memo string) (*InvoiceResult, error)
	SendPayment(ctx context.Context, invoice string, amountSats int64) (*PaymentResult, error)
	Connect() error
	Disconnect() error
	IsConnected() bool
}

// InvoiceResult holds created invoice data
type InvoiceResult struct {
	PaymentRequest string `json:"payment_request"`
	PaymentHash    string `json:"payment_hash"`
	AmountSats     int64  `json:"amount_sats"`
}

// NodeInfo holds identity data
type NodeInfo struct {
	Alias    string `json:"alias"`
	Pubkey   string `json:"pubkey"`
	Channels int    `json:"channels"`
	Status   string `json:"status"`
}

// WalletBalance holds satoshi balances
type WalletBalance struct {
	TotalBalance       int64 `json:"total_balance"`
	ConfirmedBalance   int64 `json:"confirmed_balance"`
	UnconfirmedBalance int64 `json:"unconfirmed_balance"`
}

// Channel represents a Lightning channel
type Channel struct {
	ChannelID     string `json:"channel_id"`
	Capacity      int64  `json:"capacity"`
	LocalBalance  int64  `json:"local_balance"`
	RemoteBalance int64  `json:"remote_balance"`
	Active        bool   `json:"active"`
	RemotePubkey  string `json:"remote_pubkey"`
}

// PaymentResult holds payment outcome
type PaymentResult struct {
	Status      string `json:"status"`
	PaymentHash string `json:"payment_hash"`
	Preimage    string `json:"preimage,omitempty"`
	FeeSats     int64  `json:"fee_sats,omitempty"`
}

// Ensure LNDClient implements LightningDriver
var _ LightningDriver = (*LNDClient)(nil)
