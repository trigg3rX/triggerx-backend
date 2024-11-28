package errors

import "errors"

// Keystore errors
var (
	ErrInvalidCurve 			= errors.New("invalid cryptographic curve")
	ErrInvalidPassword 			= errors.New("invalid password")
	ErrKeyFileExists 			= errors.New("keystore file already exists")
	ErrKeystoreNotFound     	= errors.New("keystore file not found")
	ErrInvalidBLSKeyFile    	= errors.New("invalid bls key file, missing public key")
	ErrKeystoreDecryption   	= errors.New("failed to decrypt keystore")
	ErrKeystoreEncryption   	= errors.New("failed to encrypt keystore")
	ErrDirectoryCreation    	= errors.New("failed to create directory")
	ErrKeystoreListing      	= errors.New("failed to list keystore files")
)

// BLS errors
var (
	ErrDeserializeSignature 	= errors.New("failed to deserialize signature")
	ErrDeserializePublicKey    	= errors.New("failed to deserialize public key")
	ErrPublicKeyNotInGroup     	= errors.New("public key not in group")
	ErrAggregateSignatures     	= errors.New("failed to aggregate signatures")
	ErrAggregatePublicKeys     	= errors.New("failed to aggregate public keys")
	ErrInvalidPrivateKey       	= errors.New("invalid private key")
	ErrSignatureVerification   	= errors.New("signature verification failed")
)

// Poseidon errors
var (
	ErrPoseidonHash            	= errors.New("poseidon hash failed")
	ErrInvalidSignature        	= errors.New("invalid signature")
	ErrChainIDRetrieval        	= errors.New("failed to retrieve chain ID")
	ErrPrivateKeyConversion    	= errors.New("failed to convert private key from hex")
)