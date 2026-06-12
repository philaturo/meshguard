// File: drivers/bitcoin/rpc_client.go
// Purpose: Bitcoin Core RPC adapter for blockchain and mempool queries
// Connects to: api/handlers.go (GET /api/bitcoin/status)
// Used by: dashboard BitcoinCore component

package bitcoin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// RPCClient connects to Bitcoin Core JSON-RPC
type RPCClient struct {
	host string
	user string
	pass string
}

// NewRPCClient creates a client with credentials
func NewRPCClient(host, user, pass string) *RPCClient {
	return &RPCClient{
		host: host,
		user: user,
		pass: pass,
	}
}

// rpcCall performs a generic JSON-RPC request
func (c *RPCClient) rpcCall(method string, params []interface{}) (map[string]interface{}, error) {
	reqBody := map[string]interface{}{
		"jsonrpc": "1.0",
		"id":      "meshguard",
		"method":  method,
		"params":  params,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "http://"+c.host, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.user, c.pass)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var result struct {
		Result map[string]interface{} `json:"result"`
		Error  map[string]interface{} `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("rpc error: %v", result.Error)
	}

	return result.Result, nil
}

// GetBlockchainInfo returns chain state
func (c *RPCClient) GetBlockchainInfo() (*BlockchainInfo, error) {
	result, err := c.rpcCall("getblockchaininfo", []interface{}{})
	if err != nil {
		return nil, err
	}

	info := &BlockchainInfo{}
	if chain, ok := result["chain"].(string); ok {
		info.Chain = chain
	}
	if blocks, ok := result["blocks"].(float64); ok {
		info.Blocks = int64(blocks)
	}
	if headers, ok := result["headers"].(float64); ok {
		info.Headers = int64(headers)
	}
	if hash, ok := result["bestblockhash"].(string); ok {
		info.BestBlockHash = hash
	}

	return info, nil
}

// GetMempoolInfo returns mempool statistics
func (c *RPCClient) GetMempoolInfo() (*MempoolInfo, error) {
	result, err := c.rpcCall("getmempoolinfo", []interface{}{})
	if err != nil {
		return nil, err
	}

	info := &MempoolInfo{}
	if size, ok := result["size"].(float64); ok {
		info.Size = int64(size)
	}
	if bytes, ok := result["bytes"].(float64); ok {
		info.Bytes = int64(bytes)
	}

	return info, nil
}

// BlockchainInfo holds chain state
type BlockchainInfo struct {
	Chain         string `json:"chain"`
	Blocks        int64  `json:"blocks"`
	Headers       int64  `json:"headers"`
	BestBlockHash string `json:"bestblockhash"`
}

// MempoolInfo holds mempool statistics
type MempoolInfo struct {
	Size  int64 `json:"size"`
	Bytes int64 `json:"bytes"`
}
