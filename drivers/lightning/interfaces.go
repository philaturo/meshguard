// File: drivers/lightning/interfaces.go
// Purpose: Clean LightningDriver interface

package lightning

import "context"

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

// Ensure LNDClient implements LightningDriver
var _ LightningDriver = (*LNDClient)(nil)
