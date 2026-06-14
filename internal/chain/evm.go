// EVM testnet provider — implements ChainAnchor against Sepolia (or any
// EVM-compatible RPC). Used when CHAIN_PROVIDER=sepolia. Deploys the
// ZTCVReceiptAnchor contract and submits real on-chain anchor txs.
package chain

import (
	"context"
	"crypto/ecdsa"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/gem-squared/gem2-ZTCV/internal/types"
)

// Embedded compiled ABI artifact from contracts/artifacts/.../ZTCVReceiptAnchor.json
// (Hardhat output — full artifact JSON with abi + bytecode + metadata).
//
//go:embed embedded/ztcv_receipt_anchor.json
var embeddedABIArtifact []byte

// EVMTestnetProvider is a real-slot Sepolia (or any EVM testnet) anchor.
// Submits recordVerification txs to a deployed ZTCVReceiptAnchor contract.
type EVMTestnetProvider struct {
	client          *ethclient.Client
	chainID         *big.Int
	contractAddr    common.Address
	parsedABI       abi.ABI
	privateKey      *ecdsa.PrivateKey
	explorerBaseURL string
}

// NewEVMTestnetProvider constructs an EVM-backed anchor.
//
// rpcURL          : RPC endpoint (Alchemy / public Sepolia / OmniOne portal)
// privateKeyHex   : 0x-prefixed deployer private key (Demo wallet ONLY — per W1)
// contractAddrHex : deployed ZTCVReceiptAnchor address
// explorerBaseURL : tx-explorer prefix (e.g. https://sepolia.etherscan.io/tx/)
func NewEVMTestnetProvider(rpcURL, privateKeyHex, contractAddrHex, explorerBaseURL string) (*EVMTestnetProvider, error) {
	if rpcURL == "" {
		return nil, fmt.Errorf("rpcURL required")
	}
	if contractAddrHex == "" {
		return nil, fmt.Errorf("contract address required")
	}
	if privateKeyHex == "" {
		return nil, fmt.Errorf("private key required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial RPC: %w", err)
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("chainID fetch: %w", err)
	}

	pk, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	// Hardhat artifact JSON: {"abi": [...], "bytecode": "0x...", ...}
	// Extract the abi field so we can hand it to go-ethereum's abi parser.
	var artifact struct {
		ABI json.RawMessage `json:"abi"`
	}
	if err := json.Unmarshal(embeddedABIArtifact, &artifact); err != nil {
		client.Close()
		return nil, fmt.Errorf("unmarshal artifact: %w", err)
	}
	abiSource := string(artifact.ABI)
	if strings.TrimSpace(abiSource) == "" {
		// Fallback: embedded file might itself already be a bare ABI array.
		abiSource = string(embeddedABIArtifact)
	}
	parsedABI, err := abi.JSON(strings.NewReader(abiSource))
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("parse ABI: %w", err)
	}

	return &EVMTestnetProvider{
		client:          client,
		chainID:         chainID,
		contractAddr:    common.HexToAddress(contractAddrHex),
		parsedABI:       parsedABI,
		privateKey:      pk,
		explorerBaseURL: explorerBaseURL,
	}, nil
}

// AnchorReceipt submits a recordVerification(bytes32,bytes32,bool,bytes32) tx
// and returns the tx hash + explorer URL. NEVER waits for confirmation —
// the demo path returns immediately so the frontend doesn't stall.
func (p *EVMTestnetProvider) AnchorReceipt(sessionHash, receiptHash string, isSafe bool, policyVersion string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sh := hexTo32(sessionHash)
	rh := hexTo32(receiptHash)
	pv := hexTo32(policyVersion)

	auth, err := bind.NewKeyedTransactorWithChainID(p.privateKey, p.chainID)
	if err != nil {
		return "", "", fmt.Errorf("transactor init: %w", err)
	}
	auth.Context = ctx

	contract := bind.NewBoundContract(p.contractAddr, p.parsedABI, p.client, p.client, p.client)
	tx, err := contract.Transact(auth, "recordVerification", sh, rh, isSafe, pv)
	if err != nil {
		return "", "", fmt.Errorf("send tx: %w", err)
	}

	txHash := tx.Hash().Hex()
	explorerURL := p.explorerBaseURL + txHash
	return txHash, explorerURL, nil
}

// GetReceipt is a placeholder for Phase B v1 — the demo path doesn't query
// chain reads; it just submits + returns the tx hash. Decoding the bound
// struct return into ReceiptOnChain can land later.
func (p *EVMTestnetProvider) GetReceipt(sessionHash string) (*types.ReceiptOnChain, error) {
	_ = sessionHash
	return nil, nil
}

// Close releases the RPC client.
func (p *EVMTestnetProvider) Close() {
	if p.client != nil {
		p.client.Close()
	}
}

// hexTo32 parses a 0x-prefixed hex string into a fixed [32]byte array.
// Right-pads if too short, truncates from the left if too long.
func hexTo32(s string) [32]byte {
	s = strings.TrimPrefix(s, "0x")
	b, _ := hex.DecodeString(s)
	var out [32]byte
	if len(b) > 32 {
		copy(out[:], b[len(b)-32:])
	} else {
		copy(out[32-len(b):], b)
	}
	return out
}
