// File: drivers/bitcoin/rpc_client.go
// Purpose: Bitcoin Core RPC client — live connection to regtest/mainnet node

package bitcoin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// RPCClient connects to Bitcoin Core via JSON-RPC
type RPCClient struct {
	url      string
	user     string
	password string
	client   *http.Client
}

// NewRPCClient creates a client for the given endpoint
func NewRPCClient(host, user, password string) *RPCClient {
	return &RPCClient{
		url:      fmt.Sprintf("http://%s", host),
		user:     user,
		password: password,
		client:   &http.Client{},
	}
}

// RPCRequest is the standard Bitcoin Core JSON-RPC envelope
type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// RPCResponse wraps all Bitcoin Core responses
type RPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *RPCError       `json:"error"`
	ID     int             `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// call executes a raw RPC method and returns the result
func (c *RPCClient) call(method string, params interface{}) (json.RawMessage, error) {
	reqBody := RPCRequest{
		JSONRPC: "1.0",
		ID:      1,
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.user, c.password)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var rpcResp RPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

// BlockchainInfo mirrors getblockchaininfo response
type BlockchainInfo struct {
	Chain        string  `json:"chain"`
	Blocks       int64   `json:"blocks"`
	Headers      int64   `json:"headers"`
	BestBlockHash string `json:"bestblockhash"`
	Difficulty   float64 `json:"difficulty"`
	VerificationProgress float64 `json:"verificationprogress"`
}

// GetBlockchainInfo returns current chain state
func (c *RPCClient) GetBlockchainInfo() (*BlockchainInfo, error) {
	result, err := c.call("getblockchaininfo", nil)
	if err != nil {
		return nil, fmt.Errorf("getblockchaininfo: %w", err)
	}

	var info BlockchainInfo
	if err := json.Unmarshal(result, &info); err != nil {
		return nil, fmt.Errorf("parse blockchaininfo: %w", err)
	}
	return &info, nil
}

// MempoolInfo mirrors getmempoolinfo response
type MempoolInfo struct {
	Size  int64 `json:"size"`
	Bytes int64 `json:"bytes"`
	Usage int64 `json:"usage"`
}

// GetMempoolInfo returns current mempool state
func (c *RPCClient) GetMempoolInfo() (*MempoolInfo, error) {
	result, err := c.call("getmempoolinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("getmempoolinfo: %w", err)
	}

	var info MempoolInfo
	if err := json.Unmarshal(result, &info); err != nil {
		return nil, fmt.Errorf("parse mempoolinfo: %w", err)
	}
	return &info, nil
}

// GetBlockCount returns current block height
func (c *RPCClient) GetBlockCount() (int64, error) {
	result, err := c.call("getblockcount", nil)
	if err != nil {
		return 0, err
	}

	var count int64
	if err := json.Unmarshal(result, &count); err != nil {
		return 0, fmt.Errorf("parse blockcount: %w", err)
	}
	return count, nil
}

// HealthCheck verifies Bitcoin Core is reachable
func (c *RPCClient) HealthCheck() error {
	_, err := c.GetBlockCount()
	return err
}
