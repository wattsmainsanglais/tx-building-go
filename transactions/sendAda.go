package transactions

import (
	//"bytes"
	"encoding/hex"
	"fmt"
	//"io"
	//"net/http"
	//"strings"

	"github.com/Salvionied/apollo"
	"github.com/Salvionied/apollo/serialization/AssetName"
	"github.com/Salvionied/apollo/serialization/Policy"
	"github.com/Salvionied/apollo/txBuilding/Backend/BlockFrostChainContext"
	"github.com/Salvionied/apollo/txBuilding/Backend/MaestroChainContext"
)

var (
	recipient = "addr_test1qq2mes4h8xegru89g2a957at5s5kp75e3hs8kd2axulu6ehhrwvxll95kkqesq7advhq5pfjf5tqjz7gea45wjpjet6q0h7xam"
	qty       = 200000 // 0.1 ADA in lovelace
)

// SendAda builds and signs a simple ADA payment transaction
func SendAda(mnemonic, maestroAPIKey, blockfrostProjectID string) (string, error) {
	// Use Maestro for querying UTxOs
	mfc, err := MaestroChainContext.NewMaestroChainContext(3, maestroAPIKey)
	if err != nil {
		return "", fmt.Errorf("failed to create Maestro chain context: %w", err)
	}

	bfc, err := BlockFrostChainContext.NewBlockfrostChainContext("https://cardano-preprod.blockfrost.io/api", 0, blockfrostProjectID)
	if err != nil {
		return "", fmt.Errorf("failed to create BlockFrost chain context: %w", err)
	}

	// Create empty backend and Apollo builder
	cc := apollo.NewEmptyBackend()
	apolloBE := apollo.New(&cc)

	// Load wallet from mnemonic using Apollo's method (network 3 = preprod)
	apolloBE, err = apolloBE.SetWalletFromMnemonic(mnemonic, 3)
	if err != nil {
		return "", fmt.Errorf("failed to set wallet from mnemonic: %w", err)
	}

	// Set wallet as change address
	apolloBE, err = apolloBE.SetWalletAsChangeAddress()
	if err != nil {
		return "", fmt.Errorf("failed to set change address: %w", err)
	}

	// Get wallet address and query UTxOs
	walletAddr := *apolloBE.GetWallet().GetAddress()
	fmt.Printf("Wallet Address: %s\n", walletAddr.String())

	utxos, err := mfc.Utxos(walletAddr)
	if err != nil {
		return "", fmt.Errorf("failed to get utxos: %w", err)
	}

	fmt.Printf("Found %d UTxOs\n", len(utxos))
	totalAda := int64(0)
	totalUnits := int64(0)
	policyID := "5e74a87d8109db21fe3d407950c161cd2df7975f0868e10682a3dbfe"
	assetNameHex := "7070626c323032342d73636166666f6c642d746f6b656e"
	assetNameBytes, _ := hex.DecodeString(assetNameHex)
	assetName := string(assetNameBytes)
	for i, utxo := range utxos {
		amount := utxo.Output.GetAmount()
		filteredCoin := amount.GetAssets().Filter(func(policy Policy.PolicyId, asset AssetName.AssetName, quantity int64) bool {
			return policy.Value == policyID && asset.HexString() == assetNameHex
		})
		fmt.Printf("UTxO %d: %d lovelace\n", i, amount.GetCoin())
		for policy, assets := range filteredCoin {
			for assetName, quantity := range assets {
				fmt.Printf("Policy: %s, Asset: %s, Quantity: %d\n", policy, assetName, quantity)
				if quantity > 0 {
					totalUnits += 10
				}
			}
		}

		totalAda += int64(amount.GetCoin())
		otherToken := amount.GetAssets()
		if len(otherToken) > 0 {
			fmt.Printf("Other tokens:  %v\n", otherToken)
		}
	}
	fmt.Printf("Total ADA available: %d lovelace (%.2f ADA)\n\n", totalAda, float64(totalAda)/1000000.0)

	unitsToSend := apollo.Unit{
		PolicyId: policyID,
		Name:     assetName,
		Quantity: int(totalUnits),
	}

	// Build the transaction
	apolloBE, err = apolloBE.
		AddLoadedUTxOs(utxos...).
		PayToAddressBech32(recipient, qty, unitsToSend).
		Complete()
	if err != nil {
		return "", fmt.Errorf("failed to build transaction: %w", err)
	}

	// Sign using Apollo's method (uses internal wallet)
	apolloBE = apolloBE.Sign()

	// Get the signed transaction
	tx := apolloBE.GetTx()

	txHash, err := bfc.SubmitTx(*tx)
	if err != nil {
		return "", fmt.Errorf("failed to submit transaction: %w", err)
	}

	if txHash.Payload != nil {
		fmt.Printf("Transaction complete")
	}

	// Submit via HTTP directly to BlockFrost API
	/*fmt.Println("Submitting transaction via HTTP to BlockFrost...")
	txHash, err := submitTxBlockFrost(txCbor, blockfrostProjectID)
	if err != nil {
		return "", fmt.Errorf("failed to submit transaction: %w", err)
	}
	*/

	return hex.EncodeToString(txHash.Payload), nil
}

// submitTxBlockFrost submits a transaction directly via BlockFrost HTTP API
// This bypasses the buggy BlockFrost SDK and uses raw HTTP requests
/*func submitTxBlockFrost(txCbor string, projectId string) (string, error) {
	// BlockFrost's transaction submission endpoint for Preprod
	url := "https://cardano-preprod.blockfrost.io/api/v0/tx/submit"

	// Step 1: Convert the hex-encoded CBOR string to raw bytes
	// BlockFrost expects the actual CBOR bytes, not a hex string
	txBytes, err := hex.DecodeString(txCbor)
	if err != nil {
		return "", fmt.Errorf("failed to decode CBOR hex: %w", err)
	}

	// Step 2: Create the HTTP POST request
	// bytes.NewBuffer(txBytes) creates a Reader from our byte slice for the request body
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(txBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Step 3: Set required headers
	// Content-Type tells BlockFrost we're sending CBOR-encoded data (not JSON)
	req.Header.Set("Content-Type", "application/cbor")
	// project_id is BlockFrost's authentication mechanism (your API key)
	req.Header.Set("project_id", projectId)

	// Step 4: Send the request using Go's default HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	// defer ensures the response body is closed when this function exits
	// This prevents resource leaks
	defer resp.Body.Close()

	// Step 5: Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Step 6: Check if submission was successful
	// BlockFrost returns 200 for successful submissions
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("submission failed (status %d): %s", resp.StatusCode, string(body))
	}

	// Step 7: Parse the response
	// BlockFrost returns the transaction hash as a plain string (possibly quoted)
	txHash := string(body)
	// Remove any surrounding quotes, newlines, or whitespace
	txHash = strings.Trim(txHash, "\" \n\r")

	return txHash, nil
}*/
