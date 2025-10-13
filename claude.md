# Claude Context: Cardano Transaction Builder

## Project Overview
A working Cardano transaction builder in Go using the Apollo library. Successfully creates, signs, and submits ADA payment transactions on the Cardano Preprod testnet.

## Tech Stack
- **Language**: Go 1.24.5
- **Main Library**: [Apollo](https://github.com/Salvionied/apollo) v1.3.0 - Pure Go Cardano transaction builder (relatively new, limited documentation)
- **Chain Provider**: Maestro (for UTxO queries) + BlockFrost (for submission)
- **Network**: Cardano Preprod testnet only (mainnet explicitly blocked)

## Status: ✅ WORKING

Successfully submitted first transaction on 2025-10-13:
**TX ID**: `efb7104944800c3f98879cb8d31ba45586927692517ea637b2b4a5b4d1bebf15`

### What's Working
1. **Wallet Management** (`config/wallet.go`)
   - Apollo's built-in `SetWalletFromMnemonic()` handles key derivation
   - Network ID: 3 (preprod)
   - No password required (empty string is standard)
   - Address derivation matches Eternl wallet

2. **Transaction Building** (`transactions/sendAda.go`)
   - ✅ Connects to Maestro API for UTxO queries (network ID: 3)
   - ✅ Creates Apollo transaction builder with empty backend
   - ✅ Loads wallet from mnemonic using Apollo's native method
   - ✅ Sets wallet as change address
   - ✅ Queries UTxOs from Maestro
   - ✅ Builds transaction with auto fee calculation
   - ✅ Signs transaction using Apollo's `.Sign()` method
   - ✅ Submits via direct HTTP to BlockFrost API

3. **HTTP Transaction Submission** (`submitTxBlockFrost()`)
   - Bypasses buggy SDK implementations
   - Direct HTTP POST to BlockFrost
   - Sends raw CBOR bytes (not hex string)
   - Proper headers: `Content-Type: application/cbor`, `project_id` authentication
   - Returns transaction hash on success

### Key Learnings & Solutions

1. **Network ID Confusion**
   - Maestro uses different network IDs than other providers
   - Maestro: 0=mainnet, 1=testnet, 2=preview, 3=preprod
   - Must use network ID 3 for preprod with Maestro

2. **Wallet Derivation**
   - Initially tried manual key extraction from Bursa - caused invalid witness errors
   - **Solution**: Use Apollo's `SetWalletFromMnemonic()` method
   - Apollo handles all key derivation correctly internally
   - Must use empty password ("") - standard for most wallets

3. **SDK Bugs Encountered**
   - Maestro Go SDK's `SubmitTx()` returns cryptic pointer errors
   - BlockFrost SDK has caching issues (tries to read non-existent tmp files)
   - **Solution**: Bypass SDKs and use direct HTTP POST for submission

4. **Transaction Submission via HTTP**
   - BlockFrost expects raw CBOR bytes, not hex strings or JSON
   - Must use `Content-Type: application/cbor` header
   - Authentication via `project_id` header (not Authorization)
   - Convert hex CBOR to bytes before sending: `hex.DecodeString()`

### Configuration
- **Mnemonic**: Loaded from `config/wallet.go:25`
- **Recipient**: `addr_test1qq2mes4h8xegru89g2a957at5s5kp75e3hs8kd2axulu6ehhrwvxll95kkqesq7advhq5pfjf5tqjz7gea45wjpjet6q0h7xam`
- **Amount**: 100,000 lovelace (0.1 ADA)
- **Network**: Preprod (ID: 3 for Maestro)
- **Maestro API Key**: `khQrk6nqDgRVfnI8Y4W18173oSzlAEPp`
- **BlockFrost Project ID**: `preprodSAO4rpor12EVJ7r5jKAOgNBaISPQPaqQ`

## Apollo Library Notes

### Working Pattern (from Apollo docs)
```go
// Create chain context (for UTxO queries)
bfc, _ := MaestroChainContext.NewMaestroChainContext(3, apiKey)

// Create empty backend and builder
cc := apollo.NewEmptyBackend()
apollob := apollo.New(&cc)

// Load wallet from mnemonic
apollob, _ = apollob.SetWalletFromMnemonic(mnemonic, 3)
apollob, _ = apollob.SetWalletAsChangeAddress()

// Get UTxOs from chain provider
utxos, _ := bfc.Utxos(*apollob.GetWallet().GetAddress())

// Build and sign transaction
apollob, _ = apollob.
    AddLoadedUTxOs(utxos...).
    PayToAddressBech32(recipient, amount).
    Complete()
apollob = apollob.Sign()  // Uses internal wallet

// Get transaction
tx := apollob.GetTx()
```

### Backend Types
- `NewEmptyBackend()` - For building transactions (doesn't need chain access)
- Chain contexts (Maestro/BlockFrost) - For querying chain state
- Keep them separate - use empty backend for Apollo, chain context for queries

### Documentation Challenges
- Apollo is relatively new and not well documented
- Examples in README sometimes outdated
- Must read source code to understand some patterns
- Community support available on Discord

## File Structure
```
tx-building-go/
├── config/
│   └── wallet.go          # Wallet init (legacy, not used anymore)
├── transactions/
│   └── sendAda.go         # Transaction building and HTTP submission
├── tmp/                   # BlockFrost cache dir (auto-created)
├── main.go                # Entry point
├── go.mod                 # Dependencies
├── claude.md              # This file
├── mnemonic.txt           # Mnemonic file (should be in .gitignore)
└── newandamiotest.json    # Unknown JSON file
```

## Next Steps / Improvements
- ✅ ~~Test transaction submission~~ **DONE - TX submitted successfully**
- Move secrets to environment variables (mnemonic, API keys)
- Create CLI interface for dynamic inputs (recipient, amount)
- Add support for token transfers
- Add metadata to transactions
- Implement multi-input transactions
- Add transaction status checking
- Refactor to remove unused config/wallet.go code

## Safety Features
- Only preprod network ID (3) used
- Mnemonic and API keys need to be moved to env vars before GitHub commit

## Dependencies
```
github.com/Salvionied/apollo v1.3.0
github.com/blinklabs-io/bursa v0.11.1
github.com/blinklabs-io/gouroboros v0.129.0
```
