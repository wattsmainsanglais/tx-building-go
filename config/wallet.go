package config

import (
	"encoding/hex"
	"fmt"

	"github.com/Salvionied/apollo/serialization"
	"github.com/Salvionied/apollo/serialization/Address"
	"github.com/Salvionied/apollo/serialization/Key"
	"github.com/blinklabs-io/bursa"
)

// Wallet represents the wallet configuration for transaction building
type Wallet struct {
	PKH     serialization.PubKeyHash
	Vkey    Key.VerificationKey
	Skey    Key.SigningKey
	Address Address.Address
}

var (
	myWallet = &Wallet{}
)

var mnemonic = "pigeon essay giggle armor divert edit soda asthma spider code satoshi fatal keen scissors certain outside error deposit turn glow dwarf crush kitten chimney"
var password = "" // Empty password - standard for most wallets

// LoadWalletFromMnemonic initializes the wallet from a mnemonic phrase
func LoadWalletFromMnemonic(mnemonic string, networkID int) error {
	// Safety check: only allow preprod
	if networkID == 1 {
		return fmt.Errorf("mainnet is not supported - this configuration is for preprod testing only")
	}

	rootKey, err := bursa.GetRootKeyFromMnemonic(mnemonic, password)
	if err != nil {
		return fmt.Errorf("failed to get root key from mnemonic: %w", err)
	}

	accountKey := bursa.GetAccountKey(rootKey, 0)
	paymentKey := bursa.GetPaymentKey(accountKey, 0)

	rawAddress, err := bursa.GetAddress(accountKey, "preprod", 0)
	if err != nil {
		return fmt.Errorf("cannot get Address: %w", err)
	}

	address, err := Address.DecodeAddress(rawAddress.String())
	if err != nil {
		return fmt.Errorf("failed to decode address: %w", err)
	}

	vKeyBytes, err := hex.DecodeString(bursa.GetPaymentVKey(paymentKey).CborHex)
	if err != nil {
		return fmt.Errorf("failed to decode verification key: %w", err)
	}

	sKeyBytes, err := hex.DecodeString(bursa.GetPaymentSKey(paymentKey).CborHex)
	if err != nil {
		return fmt.Errorf("failed to decode signing key: %w", err)
	}

	// Remove CBOR encoding prefix
	vKeyBytes = vKeyBytes[2:]
	sKeyBytes = sKeyBytes[2:]
	// Extract only the signing key bytes (first 64 bytes) and verification key bytes (last 32 bytes)
	sKeyBytes = append(sKeyBytes[:64], sKeyBytes[96:]...)

	myWallet = &Wallet{
		PKH:     serialization.PubKeyHash(address.PaymentPart),
		Vkey:    Key.VerificationKey{Payload: vKeyBytes},
		Skey:    Key.SigningKey{Payload: sKeyBytes},
		Address: address,
	}

	return nil
}

// GetWallet returns the configured wallet instance
func GetWallet() *Wallet {
	return myWallet
}

// GetMnemonic returns the hardcoded mnemonic for testing
func GetMnemonic() string {
	return mnemonic
}
