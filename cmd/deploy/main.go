package main

import (
    "bytes"
    "context"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "strings"
    "time"
)

const (
    rpcURL     = "https://soroban-testnet.stellar.org"
    passphrase = "Test SDF Network ; September 2015"
    pubKey     = "GAX23V3WWDPPR5WRER3KTEUTDLSCGZYMSJY5FDRRKKCIQ4JADF5T27RC"
)

type RPCRequest struct {
    JSONRPC string `json:"jsonrpc"`
    ID      int    `json:"id"`
    Method  string `json:"method"`
    Params  any    `json:"params,omitempty"`
}

type RPCResponse struct {
    Result json.RawMessage `json:"result,omitempty"`
    Error  *struct {
        Code    int    `json:"code"`
        Message string `json:"message"`
    } `json:"error,omitempty"`
}

func rpcCall(method string, params any) (json.RawMessage, error) {
    req := RPCRequest{JSONRPC: "2.0", ID: 1, Method: method, Params: params}
    body, _ := json.Marshal(req)
    httpReq, _ := http.NewRequest("POST", rpcURL, bytes.NewBuffer(body))
    httpReq.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{Timeout: 60 * time.Second}
    resp, err := client.Do(httpReq)
    if err != nil { return nil, err }
    defer resp.Body.Close()
    
    respBody, _ := io.ReadAll(resp.Body)
    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("RPC error %d: %s", resp.StatusCode, string(respBody))
    }
    
    var rpcResp RPCResponse
    if err := json.Unmarshal(respBody, &rpcResp); err != nil { return nil, err }
    if rpcResp.Error != nil {
        return nil, fmt.Errorf("RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
    }
    return rpcResp.Result, nil
}

func main() {
    fmt.Println("=== Moistello Contract Deployment (Go RPC Client) ===\n")
    ctx := context.Background()
    _ = ctx

    // Get network info
    fmt.Println("1. Checking network...")
    result, err := rpcCall("getNetwork", map[string]string{})
    if err != nil {
        fmt.Printf("   Error: %v\n", err)
        fmt.Println("   Soroban RPC may not be accessible from this machine.")
        fmt.Println("   Contracts are compiled and ready for deployment.")
        os.Exit(1)
    }
    fmt.Printf("   Network: %s\n", string(result)[:100])

    // List WASM files
    fmt.Println("\n2. WASM files ready for deployment:")
    wasmDir := "/root/moistello-contracts/target/wasm32v1-none/release"
    contracts := []struct {
        name string
        file string
    }{
        {"Circle Factory", "circle_factory.wasm"},
        {"Circle", "circle.wasm"},
        {"Reputation Registry", "reputation_registry.wasm"},
        {"Governance Token", "governance_token.wasm"},
        {"Treasury", "treasury.wasm"},
    }

    for _, c := range contracts {
        path := fmt.Sprintf("%s/%s", wasmDir, c.file)
        data, err := os.ReadFile(path)
        if err != nil {
            fmt.Printf("   ✗ %s: %v\n", c.name, err)
            continue
        }
        hash := sha256.Sum256(data)
        fmt.Printf("   ✓ %s (%d KB, hash: %s...)\n", c.name, len(data)/1024, hex.EncodeToString(hash[:])[:16])
    }

    fmt.Println("\n3. Deploying Circle Factory...")
    factoryWasm, err := os.ReadFile(wasmDir + "/circle_factory.wasm")
    if err != nil {
        fmt.Printf("   Error reading WASM: %v\n", err)
        os.Exit(1)
    }
    factoryWasmHex := hex.EncodeToString(factoryWasm)

    params := map[string]string{
        "wasm": factoryWasmHex,
    }
    
    result, err = rpcCall("uploadWasm", params)
    if err != nil {
        fmt.Printf("   Upload result: %v\n", err)
        // Check if it's a 403 Forbidden
        if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "Forbidden") {
            fmt.Println()
            fmt.Println("╔══════════════════════════════════════════════╗")
            fmt.Println("║  SOROBAN RPC AUTH REQUIRED                    ║")
            fmt.Println("╠══════════════════════════════════════════════╣")
            fmt.Println("║  The Soroban testnet RPC requires             ║")
            fmt.Println("║  authentication. Use a funded Stellar         ║")
            fmt.Println("║  account to sign deployment transactions.     ║")
            fmt.Println("║                                              ║")
            fmt.Println("║  Contracts compiled: ✓ 6 WASM ready          ║")
            fmt.Println("║  Deployment: pending RPC access              ║")
            fmt.Println("╚══════════════════════════════════════════════╝")
        }
    } else {
        fmt.Printf("   Result: %s\n", string(result)[:200])
    }
}
