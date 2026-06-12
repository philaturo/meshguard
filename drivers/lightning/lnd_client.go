// File: drivers/lightning/lnd_client.go
// Purpose: Real LND gRPC client using raw grpc.Invoke with JSON payloads
// Connects to: interfaces.go (LightningDriver), api/handlers.go
// Note: Uses google.golang.org/grpc directly, no protobuf generated code
// LND gRPC accepts JSON-encoded messages on the wire for simple types

package lightning

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// LNDClient implements LightningDriver using raw gRPC
type LNDClient struct {
	conn         *grpc.ClientConn
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

// invoke performs a raw gRPC call with JSON request/response
func (c *LNDClient) invoke(ctx context.Context, method string, req, resp interface{}) error {
	if !c.connected {
		return fmt.Errorf("node %s not connected", c.name)
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	err = c.conn.Invoke(ctx, "/lnrpc.Lightning/"+method, json.RawMessage(reqBytes), resp)
	if err != nil {
		return fmt.Errorf("invoke %s: %w", method, err)
	}
	return nil
}

// GetInfo returns real LND node identity
func (c *LNDClient) GetInfo(ctx context.Context) (*NodeInfo, error) {
	var resp struct {
		Alias                 string `json:"alias"`
		IdentityPubkey        string `json:"identity_pubkey"`
		NumActiveChannels     uint32 `json:"num_active_channels"`
		NumInactiveChannels   uint32 `json:"num_inactive_channels"`
	}
	if err := c.invoke(ctx, "GetInfo", struct{}{}, &resp); err != nil {
		return nil, err
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
	var resp struct {
		TotalBalance       int64 `json:"total_balance"`
		ConfirmedBalance   int64 `json:"confirmed_balance"`
		UnconfirmedBalance int64 `json:"unconfirmed_balance"`
	}
	if err := c.invoke(ctx, "WalletBalance", struct{}{}, &resp); err != nil {
		return nil, err
	}

	return &WalletBalance{
		TotalBalance:       resp.TotalBalance,
		ConfirmedBalance:   resp.ConfirmedBalance,
		UnconfirmedBalance: resp.UnconfirmedBalance,
	}, nil
}

// ListChannels returns real open channels
func (c *LNDClient) ListChannels(ctx context.Context) ([]Channel, error) {
	var resp struct {
		Channels []struct {
			ChanId        uint64 `json:"chan_id,string"`
			Capacity      int64  `json:"capacity,string"`
			LocalBalance  int64  `json:"local_balance,string"`
			RemoteBalance int64  `json:"remote_balance,string"`
			Active        bool   `json:"active"`
			RemotePubkey  string `json:"remote_pubkey"`
		} `json:"channels"`
	}

	req := struct {
		ActiveOnly bool `json:"active_only"`
	}{ActiveOnly: false}

	if err := c.invoke(ctx, "ListChannels", req, &resp); err != nil {
		return nil, err
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
	var resp struct {
		PaymentRequest string `json:"payment_request"`
		RHash          []byte `json:"r_hash"`
	}

	req := struct {
		Value int64  `json:"value,string"`
		Memo  string `json:"memo"`
	}{
		Value: amountSats,
		Memo:  memo,
	}

	if err := c.invoke(ctx, "AddInvoice", req, &resp); err != nil {
		return nil, err
	}

	return &InvoiceResult{
		PaymentRequest: resp.PaymentRequest,
		PaymentHash:    hex.EncodeToString(resp.RHash),
		AmountSats:     amountSats,
	}, nil
}

// SendPayment pays a real invoice via LND
func (c *LNDClient) SendPayment(ctx context.Context, invoice string, amountSats int64) (*PaymentResult, error) {
	var resp struct {
		PaymentError    string `json:"payment_error"`
		PaymentHash     []byte `json:"payment_hash"`
		PaymentPreimage []byte `json:"payment_preimage"`
		PaymentRoute    struct {
			TotalFees int64 `json:"total_fees,string"`
		} `json:"payment_route"`
	}

	req := struct {
		PaymentRequest string `json:"payment_request"`
		Amt            int64  `json:"amt,string"`
	}{
		PaymentRequest: invoice,
		Amt:            amountSats,
	}

	if err := c.invoke(ctx, "SendPaymentSync", req, &resp); err != nil {
		return nil, err
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
