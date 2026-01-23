# EigenCloud Integration

[EigenCloud](https://www.eigenlayer.xyz/) is a restaking protocol on Ethereum that enables staked ETH to secure multiple services simultaneously. TriggerX operates as an **AVS (Actively Validated Service)** on EigenCloud.

## How TriggerX Uses EigenCloud

- **Keeper Registration**: Keepers stake ETH through EigenCloud contracts to participate in task execution
- **Economic Security**: Staked assets provide cryptoeconomic guarantees for honest behavior
- **Slashing**: Malicious or faulty behavior results in slashing of staked assets
- **Reward Distribution**: Execution rewards distributed through EigenCloud contracts

## Othentic Framework

TriggerX uses [Othentic](https://othentic.xyz/) as the consensus layer - a modular AVS framework built on EigenCloud.

### Othentic Services

The `othentic/` directory contains Docker configuration for:

- **Aggregator**: Collects executed tasks and coordinates consensus
- **Performer**: Executes tasks and submits proofs
- **Attester**: Validates peer executions with BFT consensus

### Configuration

Othentic services are configured via environment variables:

```bash
# Contract addresses
AVS_GOVERNANCE_ADDRESS=0x...
ATTESTATION_CENTER_ADDRESS=0x...

# Network
L1_RPC=...  # Ethereum L1 RPC
L2_RPC=...  # L2 chain RPC

# P2P Bootstrap
OTHENTIC_BOOTSTRAP_ID=...
OTHENTIC_BOOTSTRAP_SEED=...
```

## Resources

- [EigenCloud Documentation](https://docs.eigenlayer.xyz/)
- [EigenCloud AVS Guide](https://docs.eigenlayer.xyz/eigenlayer/avs-guides/avs-developer-guide)
- [Othentic Documentation](https://docs.othentic.xyz/)
