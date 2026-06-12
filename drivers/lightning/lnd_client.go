// File: drivers/lightning/lnd_client.go
// Purpose: Real LND REST API client with lncli fallback for LND 0.21.99-beta
// Connects to: interfaces.go (LightningDriver), api/handlers.go
// Note: Uses LND REST API with HTTP + macaroon auth, no gRPC/protobuf

package lightning

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// LNDClient implements LightningDriver using LND REST API
type LNDClient struct {
	client       *http.Client
	name         string
	RestAddr     string
	tlsPath      string
	macaroonPath string
	connected    bool
}

// NodeConfig holds connection parameters for Alice or Bob
type NodeConfig struct {
	Name         string
	RPCAddr      string  // Kept for backward compatibility
	RestAddr     string  // REST API addr, e.g. "127.0.0.1:8080"
	TLSCertPath  string
	MacaroonPath string
}

// NewLNDClient creates a client for Alice or Bob
func NewLNDClient(cfg NodeConfig) *LNDClient {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}

	return &LNDClient{
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
			Timeout: 30 * time.Second,
		},
		name:         cfg.Name,
		RestAddr:     cfg.RestAddr,
		tlsPath:      cfg.TLSCertPath,
		macaroonPath: cfg.MacaroonPath,
		connected:    false,
	}
}

// Connect marks as ready (actual auth happens per-request)
func (c *LNDClient) Connect() error {
	c.connected = true
	return nil
}

// Disconnect marks as offline
func (c *LNDClient) Disconnect() error {
	c.connected = false
	return nil
}

// IsConnected returns state
func (c *LNDClient) IsConnected() bool {
	return c.connected
}

// restCall performs an authenticated HTTP request to LND REST API
func (c *LNDClient) restCall(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	if !c.connected {
		return nil, fmt.Errorf("node %s not connected", c.name)
	}

	macaroonHex, err := loadMacaroonHex(c.macaroonPath)
	if err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := fmt.Sprintf("https://%s%s", c.RestAddr, path)
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Grpc-Metadata-macaroon", macaroonHex)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetInfo returns LND node identity
func (c *LNDClient) GetInfo(ctx context.Context) (*NodeInfo, error) {
	data, err := c.restCall(ctx, "GET", "/v1/getinfo", nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Alias             string `json:"alias"`
		IdentityPubkey    string `json:"identity_pubkey"`
		NumActiveChannels int    `json:"num_active_channels"`
		NumPeers          int    `json:"num_peers"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &NodeInfo{
		Alias:    resp.Alias,
		Pubkey:   resp.IdentityPubkey,
		Channels: resp.NumActiveChannels,
		Status:   "online",
	}, nil
}

// GetWalletBalance returns confirmed/unconfirmed balance
func (c *LNDClient) GetWalletBalance(ctx context.Context) (*WalletBalance, error) {
	data, err := c.restCall(ctx, "GET", "/v1/balance/blockchain", nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		TotalBalance       int64 `json:"total_balance,string"`
		ConfirmedBalance   int64 `json:"confirmed_balance,string"`
		UnconfirmedBalance int64 `json:"unconfirmed_balance,string"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		var resp2 struct {
			TotalBalance       int64 `json:"total_balance"`
			ConfirmedBalance   int64 `json:"confirmed_balance"`
			UnconfirmedBalance int64 `json:"unconfirmed_balance"`
		}
		if err2 := json.Unmarshal(data, &resp2); err2 != nil {
			return nil, fmt.Errorf("unmarshal balance: %w", err)
		}
		resp.TotalBalance = resp2.TotalBalance
		resp.ConfirmedBalance = resp2.ConfirmedBalance
		resp.UnconfirmedBalance = resp2.UnconfirmedBalance
	}

	return &WalletBalance{
		TotalBalance:       resp.TotalBalance,
		ConfirmedBalance:   resp.ConfirmedBalance,
		UnconfirmedBalance: resp.UnconfirmedBalance,
	}, nil
}

// ListChannels returns open channels
func (c *LNDClient) ListChannels(ctx context.Context) ([]Channel, error) {
	data, err := c.restCall(ctx, "GET", "/v1/channels", nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Channels []struct {
			ChanId        string `json:"chan_id"`
			Capacity      int64  `json:"capacity,string"`
			LocalBalance  int64  `json:"local_balance,string"`
			RemoteBalance int64  `json:"remote_balance,string"`
			Active        bool   `json:"active"`
			RemotePubkey  string `json:"remote_pubkey"`
		} `json:"channels"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal channels: %w", err)
	}

	var channels []Channel
	for _, ch := range resp.Channels {
		channels = append(channels, Channel{
			ChannelID:     ch.ChanId,
			Capacity:      ch.Capacity,
			LocalBalance:  ch.LocalBalance,
			RemoteBalance: ch.RemoteBalance,
			Active:        ch.Active,
			RemotePubkey:  ch.RemotePubkey,
		})
	}
	return channels, nil
}

// AddInvoice creates a BOLT 11 invoice
func (c *LNDClient) AddInvoice(ctx context.Context, amountSats int64, memo string) (*InvoiceResult, error) {
	reqBody := map[string]interface{}{
		"value": amountSats,
		"memo":  memo,
	}

	data, err := c.restCall(ctx, "POST", "/v1/invoices", reqBody)
	if err != nil {
		return nil, err
	}

	var resp struct {
		PaymentRequest string `json:"payment_request"`
		RHash          string `json:"r_hash"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal invoice: %w", err)
	}

	return &InvoiceResult{
		PaymentRequest: resp.PaymentRequest,
		PaymentHash:    resp.RHash,
		AmountSats:     amountSats,
	}, nil
}

// SendPayment pays an invoice via the Lightning Network
func (c *LNDClient) SendPayment(ctx context.Context, invoice string, amountSats int64) (*PaymentResult, error) {
	log.Printf("[PAY] Starting payment: node=%s, amount=%d", c.name, amountSats)

	// 1. Try lncli first (isolated context)
	result, cliErr := c.sendPaymentViaCLI(invoice)
	if cliErr == nil {
		log.Printf("[PAY] lncli SUCCESS: hash=%s", result.PaymentHash)
		return result, nil
	}
	log.Printf("[PAY] lncli FAILED: %v", cliErr)

	// 2. REST fallback
	log.Printf("[PAY] Falling back to REST...")
	reqBody := map[string]interface{}{
		"payment_request": invoice,
		"fee_limit":       map[string]interface{}{"fixed": 1000},
	}

	data, err := c.restCall(ctx, "POST", "/v1/payments", reqBody)
	if err != nil {
		data, err = c.restCall(ctx, "POST", "/v1/channels/transactions", reqBody)
		if err != nil {
			return nil, fmt.Errorf("all methods failed (lncli: %v, rest: %w)", cliErr, err)
		}
	}

	// Parse REST response
	var resp struct {
		Status          string `json:"status"`
		PaymentError    string `json:"payment_error"`
		PaymentHash     string `json:"payment_hash"`
		PaymentPreimage string `json:"payment_preimage"`
		TotalFees       int64  `json:"total_fees,string"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal rest payment: %w", err)
	}

	status := "settled"
	if resp.Status == "FAILED" || resp.PaymentError != "" {
		status = "failed"
	}

	return &PaymentResult{
		Status:      status,
		PaymentHash: resp.PaymentHash,
		Preimage:    resp.PaymentPreimage,
		FeeSats:     resp.TotalFees,
	}, nil
}

// sendPaymentViaCLI uses lncli shell command to send payment
func (c *LNDClient) sendPaymentViaCLI(invoice string) (*PaymentResult, error) {
	lncliPath := "/home/aturo/go/bin/lncli"
	if _, err := os.Stat(lncliPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("lncli binary not found at %s", lncliPath)
	}

	var rpcServer, tlsCertPath, macaroonPath string
	switch c.name {
	case "Alice":
		rpcServer = "127.0.0.1:10009"
		tlsCertPath = c.tlsPath
		macaroonPath = c.macaroonPath
	case "Bob":
		rpcServer = "127.0.0.1:10010"
		tlsCertPath = c.tlsPath
		macaroonPath = c.macaroonPath
	default:
		return nil, fmt.Errorf("unknown node: %s", c.name)
	}

	// CRITICAL: Create fresh timeout so HTTP handler cancellation doesn't kill lncli
	cmdCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, lncliPath,
		"--network=regtest",
		"--rpcserver="+rpcServer,
		"--tlscertpath="+tlsCertPath,
		"--macaroonpath="+macaroonPath,
		"sendpayment",
		"--pay_req="+invoice,
		"--force",
	)

	// Explicit environment setup
	homeDir, _ := os.UserHomeDir()
	cmd.Env = append(os.Environ(), "HOME="+homeDir)
	cmd.Dir = "/" // Avoid weird working directory issues

	log.Printf("[LNCLI] Executing: %s %s", lncliPath, strings.Join(cmd.Args[1:], " "))

	output, err := cmd.CombinedOutput()
	outStr := string(output)

	if err != nil {
		log.Printf("[LNCLI] EXEC ERROR: %v | Output: %s", err, outStr)
		return nil, fmt.Errorf("lncli exec failed: %v", err)
	}

	log.Printf("[LNCLI] RAW OUTPUT: %s", outStr)

	if !strings.Contains(outStr, "SUCCEEDED") {
		return nil, fmt.Errorf("lncli did not succeed: %s", outStr)
	}

	// Parse hash & preimage from text output
	hash := "unknown"
	preimage := "unknown"
	for _, line := range strings.Split(outStr, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Payment hash:") {
			hash = strings.TrimSpace(strings.TrimPrefix(line, "Payment hash:"))
		}
		if strings.Contains(line, "preimage:") {
			parts := strings.Split(line, "preimage:")
			if len(parts) > 1 {
				preimage = strings.TrimSpace(strings.Split(parts[1], ",")[0])
			}
		}
	}

	return &PaymentResult{
		Status:      "settled",
		PaymentHash: hash,
		Preimage:    preimage,
	}, nil
}

// loadMacaroonHex reads a binary macaroon file and returns hex-encoded string
func loadMacaroonHex(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read macaroon: %w", err)
	}
	return hex.EncodeToString(data), nil
}
