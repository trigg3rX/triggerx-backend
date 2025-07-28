# 🚀 TriggerX Imuachain Integration

This document provides complete instructions for registering as a validator on Imuachain using the TriggerX CLI.

## 📋 Overview

TriggerX now supports full integration with Imuachain, allowing operators to:
- Register as operators on Imuachain
- Deposit and delegate tokens for staking
- Opt-in to AVS (Actively Validated Services)
- Complete TriggerX AVS registration on Ethereum
- Associate operators with EVM stakers for post-bootstrap scenarios

## 🛠️ Prerequisites

### Required Software
- `imuad` - Imua chain binary
- `cast` - Foundry's cast tool for Ethereum interactions
- `jq` - JSON processor (optional, CLI provides fallbacks)

### Required Tokens
- **ETH on Sepolia** - For gas fees on Ethereum testnet
- **exoETH on Sepolia** - For staking (get from faucet)
- **IMUA tokens** - For gas on Imuachain (automatically funded via CLI)

## ⚙️ Environment Setup

Update your `.env` file with these additional variables:

```bash
# Imuachain Configuration
OPERATOR_NAME=YourValidatorName
IMUA_HOME_DIR=/home/user/.imuad
IMUA_ACCOUNT_KEY_NAME=YourKeyName
IMUA_CHAIN_ID=imuachaintestnet_233-9
IMUA_COS_GRPC_URL=https://api-cosmos-grpc.exocore-restaking.com:443

# Optional Configuration
DEPOSIT_AMOUNT=1000
TOKEN_ADDRESS=0x83E6850591425e3C1E263c054f4466838B9Bd9e4
IMUA_ETH_RPC_URL=https://api-eth.exocore-restaking.com
IMUA_FAUCET_URL=https://241009-faucet.exocore-restaking.com/
```

## 🎯 Quick Start

For a fully automated registration process:

```bash
# First, setup Imuachain keys (creates validator key if needed)
./triggerx setup-imua-keys

# Fund your account with IMUA tokens (for gas fees)
./triggerx fund-imua-account

# Then run the complete registration process
./triggerx complete-imua-registration
```

## 📚 Individual Commands

Run each step individually:

```bash
# 0. Setup Imuachain keys (optional - auto-runs when needed)
./triggerx setup-imua-keys

# 1. Fund account with IMUA tokens for gas fees
./triggerx fund-imua-account

# 2. Check your IMUA token balance
./triggerx check-imua-balance

# 3. Register on Imuachain
./triggerx register-imua-operator

./triggerx get-imeth-tokens

# 4. Deposit and delegate tokens
./triggerx deposit-and-delegate

# 5. Opt-in to AVS
./triggerx opt-in-to-avs

# 6. Complete TriggerX registration
./triggerx complete-registration

# 7. Associate operator (post-bootstrap only)
./triggerx associate-operator
```

## 🔧 Key Management

### Setting Up Validator Keys

Before registering as an operator, you need to create a validator key:

```bash
./triggerx setup-imua-keys
```

This command will:
- Check if `imuad` is installed and available
- Initialize the Imua home directory if it doesn't exist
- Create the validator key if it's missing
- Display the validator address for funding

### Manual Key Creation

If you prefer to create keys manually:

```bash
# Initialize Imua directory
imuad init validator --home ~/.imuad

# Create validator key
imuad --home ~/.imuad keys add validator

# List all keys
imuad --home ~/.imuad keys list
```

## 💰 Account Funding

### Automatic Funding (Recommended)

Use the CLI to automatically fund your account:

```bash
# Fund your validator account with IMUA tokens
./triggerx fund-imua-account

# Check your current balance
./triggerx check-imua-balance
```

The funding command will:
- Get your validator address automatically
- Request IMUA tokens from the testnet faucet
- Wait for transaction confirmation
- Verify your updated balance

### Manual Funding

If you prefer to fund manually:

```bash
# Get your validator address
VALIDATOR_ADDRESS=$(imuad --home $IMUA_HOME_DIR keys show -a $IMUA_ACCOUNT_KEY_NAME)

# Request tokens from faucet
curl -X POST https://241009-faucet.exocore-restaking.com/ \
  -H "Content-Type: application/json" \
  -d "{\"address\": \"$VALIDATOR_ADDRESS\"}"

# Check balance
imuad query bank balances $VALIDATOR_ADDRESS \
  --node $IMUA_COS_GRPC_URL --output json
```

## 🔍 Verification

Check your registration status:

```bash
# Check operator info
imuad --home $IMUA_HOME_DIR query operator get-operator-info \
    $(imuad --home $IMUA_HOME_DIR keys show -a $IMUA_ACCOUNT_KEY_NAME) \
    --node $IMUA_COS_GRPC_URL --output json

# Check consensus key
imuad --home $IMUA_HOME_DIR query operator get-operator-cons-key \
    $(imuad --home $IMUA_HOME_DIR keys show -a $IMUA_ACCOUNT_KEY_NAME) \
    $IMUA_CHAIN_ID --node $IMUA_COS_GRPC_URL --output json
```

## 🚨 Troubleshooting

Common issues and solutions:

1. **"validator.info: key not found"** - Run `./triggerx setup-imua-keys` to create the missing validator key
2. **"account not found" or "insufficient balance"** - Run `./triggerx fund-imua-account` to get IMUA tokens
3. **"Operator not registered with chain"** - Run `./triggerx register-operator-with-chain` first
4. **"BLS_PRIVATE_KEY not found"** - Run `./triggerx generate-keys` first
5. **"imuad binary not found"** - Install the imuad binary and ensure it's in your PATH

### Detailed Error Solutions

#### Account Funding Issues
If you see errors like "account im1... not found: key not found", this means:
- Your validator account has no IMUA tokens for gas fees
- **Solution**: Run `./triggerx fund-imua-account`

#### Validator Key Issues
If you see errors like "failed to get operator address: exit status 1", this usually means:
- The validator key doesn't exist in the keyring
- The IMUA_ACCOUNT_KEY_NAME doesn't match an existing key
- The home directory isn't properly initialized

**Solution**: Run `./triggerx setup-imua-keys` to automatically fix these issues.

#### Faucet Issues
If the faucet request fails:
- Check your internet connection
- Verify the faucet URL in your environment variables
- Try again after a few minutes (rate limiting)
- Check the transaction hash on the Imua explorer

### Balance Checking
Always check your balance before operations:

```bash
# Quick balance check
./triggerx check-imua-balance

# Detailed balance query
imuad query bank balances $(imuad --home $IMUA_HOME_DIR keys show -a $IMUA_ACCOUNT_KEY_NAME) \
  --node $IMUA_COS_GRPC_URL --output json
```

## 📊 Requirements

- **Minimum Self-Delegation:** 1,000 USD in token value
- **Validator Set:** Top 50 validators by total stake
- **Hardware:** 4+ CPU cores, 8GB+ RAM, 100GB+ SSD
- **Gas Fees:** ~1 IMUA token for multiple transactions

## 🔐 Security

- Never share private keys
- Use secure secret management
- Keep backups in safe locations
- Monitor validator performance regularly

---

For detailed documentation, see the full README and Imua documentation. 