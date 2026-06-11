// File: drivers/lightning/lnd_client.go
// Purpose: Real LND gRPC client using official lightningnetwork/lnd protobuf definitions

package lightning

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// LNDClient implements LightningDriver using real LND gRPC
type LNDClient struct {
	conn         *grpc.ClientConn
	client       lnrpc.LightningClient
	name         string
	RPCAddr      string
	tlsPath      string
	macaroonPath string
	connected    bool
}

// NodeConfig holds connection parameters for Alice or Bob
type NodeConfig struct {
	Name         string
	RPCAddr      string
	TLSCertPath  string
	MacaroonPath string
}

// NewLNDClient creates a client for Alice or Bob
func NewLNDClient(cfg NodeConfig) *LNDClient {
	return &LNDClient{
		name:         cfg.Name,
		RPCAddr:      cfg.RPCAddr,
		tlsPath:      cfg.TLSCertPath,
		macaroonPath: cfg.MacaroonPath,
		connected:    false,
	}
}

// Connect establishes gRPC with TLS + macaroon auth
func (c *LNDClient) Connect() error {
	creds, err := credentials.NewClientTLSFromFile(c.tlsPath, "")
	if err != nil {
		return fmt.Errorf("load tls cert: %w", err)
	}

	macaroon, err := loadMacaroon(c.macaroonPath)
	if err != nil {
		return fmt.Errorf("load macaroon: %w", err)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(macaroon),
	}

	conn, err := grpc.Dial(c.RPCAddr, opts...)
	if err != nil {
		return fmt.Errorf("dial lnd: %w", err)
	}

	c.conn = conn
	c.client = lnrpc.NewLightningClient(conn)
	c.connected = true
	return nil
}

// Disconnect closes connection
func (c *LNDClient) Disconnect() error {
	if c.conn != nil {
		c.conn.Close()
	}
	c.connected = false
	return nil
}

// IsConnected returns state
func (c *LNDClient) IsConnected() bool {
	return c.connected
}

// GetInfo returns real LND node identity
func (c *LNDClient) GetInfo(ctx context.Context) (*NodeInfo, error) {
	if !c.connected {
		return nil, fmt.Errorf("node %s not connected", c.name)
	}

	resp, err := c.client.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		return nil, fmt.Errorf("getinfo: %w", err)
	}

	return &NodeInfo{
		Alias:    resp.Alias,
		Pubkey:   resp.IdentityPubkey,
		Channels: int(resp.NumActiveChannels + resp.NumInactiveChannels),
		Status:   "online",
	}, nil
}

// GetWalletBalance returns real confirmed/unconfirmed balance
func (c *LNDClient) GetWalletBalance(ctx context.Context) (*WalletBalance, error) {
	if !c.connected {
		return nil, fmt.Errorf("node %s not connected", c.name)
	}

	resp, err := c.client.WalletBalance(ctx, &lnrpc.WalletBalanceRequest{})
	if err != nil {
		return nil, fmt.Errorf("walletbalance: %w", err)
	}

	return &WalletBalance{
		TotalBalance:       resp.TotalBalance,
		ConfirmedBalance:   resp.ConfirmedBalance,
		UnconfirmedBalance: resp.UnconfirmedBalance,
	}, nil
}

// ListChannels returns real open channels
func (c *LNDClient) ListChannels(ctx context.Context) ([]Channel, error) {
	if !c.connected {
		return nil, fmt.Errorf("node %s not connected", c.name)
	}

	resp, err := c.client.ListChannels(ctx, &lnrpc.ListChannelsRequest{
		ActiveOnly: false,
	})
	if err != nil {
		return nil, fmt.Errorf("listchannels: %w", err)
	}

	var channels []Channel
	for _, ch := range resp.Channels {
		channels = append(channels, Channel{
			ChannelID:     fmt.Sprintf("%d", ch.ChanId),
			Capacity:      ch.Capacity,
			LocalBalance:  ch.LocalBalance,
			RemoteBalance: ch.RemoteBalance,
			Active:        ch.Active,
			RemotePubkey:  ch.RemotePubkey,
		})
	}
	return channels, nil
}

// AddInvoice creates a real BOLT 11 invoice
func (c *LNDClient) AddInvoice(ctx context.Context, amountSats int64, memo string) (*InvoiceResult, error) {
	if !c.connected {
		return nil, fmt.Errorf("node %s not connected", c.name)
	}

	resp, err := c.client.AddInvoice(ctx, &lnrpc.Invoice{
		Value: amountSats,
		Memo:  memo,
	})
	if err != nil {
		return nil, fmt.Errorf("addinvoice: %w", err)
	}

	return &InvoiceResult{
		PaymentRequest: resp.PaymentRequest,
		PaymentHash:    hex.EncodeToString(resp.RHash),
		AmountSats:     amountSats,
	}, nil
}

// SendPayment pays a real invoice via LND
func (c *LNDClient) SendPayment(ctx context.Context, invoice string, amountSats int64) (*PaymentResult, error) {
	if !c.connected {
		return nil, fmt.Errorf("node %s not connected", c.name)
	}

	resp, err := c.client.SendPaymentSync(ctx, &lnrpc.SendRequest{
		PaymentRequest: invoice,
		Amt:            amountSats,
	})
	if err != nil {
		return nil, fmt.Errorf("sendpayment: %w", err)
	}

	status := "failed"
	if resp.PaymentError == "" {
		status = "settled"
	}

	return &PaymentResult{
		Status:      status,
		PaymentHash: hex.EncodeToString(resp.PaymentHash),
		Preimage:    hex.EncodeToString(resp.PaymentPreimage),
		FeeSats:     resp.PaymentRoute.TotalFees,
	}, nil
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

// macaroonCredential implements gRPC PerRPCCredentials
type macaroonCredential struct {
	macaroon []byte
}

func (m *macaroonCredential) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"macaroon": hex.EncodeToString(m.macaroon),
	}, nil
}

func (m *macaroonCredential) RequireTransportSecurity() bool {
	return true
}

func loadMacaroon(path string) (*macaroonCredential, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read macaroon: %w", err)
	}
	return &macaroonCredential{macaroon: data}, nil
}
