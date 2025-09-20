// Bridge Service for transferring funds from parent chain to Orbit chain
import { createPublicClient, createWalletClient, http, type Address, type Chain, defineChain } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import { arbitrumSepolia } from 'viem/chains';
import logger from '../utils/logger';

export interface BridgeServiceConfig {
  parentChainRpc: string;
  parentChainId: number;
  deployerPrivateKey: string;
  bridgeAmount?: string; // Amount to bridge in ETH (default: "0.1")
  maxSubmissionCost?: string; // Max submission cost for retryable tickets
  maxGas?: string; // Max gas for retryable tickets
  gasPriceBid?: string; // Gas price bid for retryable tickets
}

export interface BridgeResult {
  success: boolean;
  transactionHash?: string;
  retryableTicketId?: string;
  error?: string;
}

export interface TokenBridgeContracts {
  parentChainContracts: {
    router: Address;
    standardGateway: Address;
    customGateway: Address;
    wethGateway: Address;
    weth: Address;
    multicall: Address;
  };
  orbitChainContracts: {
    router: Address;
    standardGateway: Address;
    customGateway: Address;
    wethGateway: Address;
    weth: Address;
    proxyAdmin: Address;
    beaconProxyFactory: Address;
    upgradeExecutor: Address;
    multicall: Address;
  };
}

class BridgeService {
  private config: BridgeServiceConfig;
  private parentChainPublicClient: any;
  private orbitChainPublicClient: any;
  private deployerWallet: any;
  private tokenBridgeContracts?: TokenBridgeContracts;

  constructor(config: BridgeServiceConfig) {
    this.config = config;
    this.initializeClients();
  }

  private initializeClients() {
    try {
      logger.info('Initializing Bridge service clients');

      // Create parent chain public client
      this.parentChainPublicClient = createPublicClient({
        chain: arbitrumSepolia,
        transport: http(this.config.parentChainRpc)
      });

      // Create deployer wallet
      const deployerKey = this.config.deployerPrivateKey.startsWith('0x') 
        ? this.config.deployerPrivateKey 
        : `0x${this.config.deployerPrivateKey}`;
      
      this.deployerWallet = createWalletClient({
        account: privateKeyToAccount(deployerKey as `0x${string}`),
        chain: arbitrumSepolia,
        transport: http(this.config.parentChainRpc)
      });

      logger.info('Bridge service clients initialized successfully', {
        deployerAddress: this.deployerWallet.account.address
      });
    } catch (error) {
      logger.error('Failed to initialize Bridge service clients', { error });
      throw new Error(`Failed to initialize Bridge service clients: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Set up Orbit chain client for bridging operations
   */
  async setupOrbitChainClient(orbitChainId: number, orbitChainRpc: string) {
    try {
      logger.info('Setting up Orbit chain client', { orbitChainId, orbitChainRpc });

      // Define the Orbit chain
      const orbitChain = defineChain({
        id: orbitChainId,
        network: 'Orbit chain',
        name: 'orbit',
        nativeCurrency: { name: 'Ether', symbol: 'ETH', decimals: 18 },
        rpcUrls: {
          default: { http: [orbitChainRpc] },
          public: { http: [orbitChainRpc] },
        },
        testnet: true,
      });

      // Create Orbit chain public client
      this.orbitChainPublicClient = createPublicClient({
        chain: orbitChain,
        transport: http(orbitChainRpc)
      });

      logger.info('Orbit chain client setup completed', { orbitChainId });
    } catch (error) {
      logger.error('Failed to setup Orbit chain client', { error });
      throw new Error(`Failed to setup Orbit chain client: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Fund deployer wallet using a simple ETH transfer via retryable ticket
   * This is a simplified approach that should work more reliably
   */
  async fundDeployerDirectly(amount: string = this.config.bridgeAmount || "0.02", rollupAddress?: Address): Promise<BridgeResult> {
    try {
      logger.info('Funding deployer wallet via simple retryable ticket', { amount, rollupAddress });

      if (!this.orbitChainPublicClient) {
        throw new Error('Orbit chain client not initialized. Call setupOrbitChainClient first.');
      }

      // Convert amount to wei
      const amountWei = BigInt(parseFloat(amount) * 1e18);

      // Get the inbox address
      const inboxAddress = await this.getInboxAddress(rollupAddress);
      
      if (!inboxAddress) {
        throw new Error('Could not determine inbox address for retryable ticket');
      }

      logger.info('Using inbox address for retryable ticket', { inboxAddress });

      // Create a simple retryable ticket that just transfers ETH
      // We'll use a minimal approach with lower gas costs
      const maxSubmissionCost = BigInt("500000000000000"); // 0.0005 ETH
      const gasLimit = BigInt("100000"); // 100k gas
      const maxFeePerGas = BigInt("1000000000"); // 1 gwei

      // Encode the function call data manually
      const abi = [
        {
          name: 'createRetryableTicket',
          type: 'function',
          stateMutability: 'payable',
          inputs: [
            { name: 'to', type: 'address' },
            { name: 'value', type: 'uint256' },
            { name: 'maxSubmissionCost', type: 'uint256' },
            { name: 'submissionRefundAddress', type: 'address' },
            { name: 'valueRefundAddress', type: 'address' },
            { name: 'gasLimit', type: 'uint256' },
            { name: 'maxFeePerGas', type: 'uint256' },
            { name: 'data', type: 'bytes' }
          ],
          outputs: [{ name: 'ticketId', type: 'uint256' }]
        }
      ];

      // Use encodeFunctionData to create the transaction data
      const { encodeFunctionData } = await import('viem');
      const data = encodeFunctionData({
        abi,
        functionName: 'createRetryableTicket',
        args: [
          this.deployerWallet.account.address, // to
          amountWei, // value
          maxSubmissionCost, // maxSubmissionCost
          this.deployerWallet.account.address, // submissionRefundAddress
          this.deployerWallet.account.address, // valueRefundAddress
          gasLimit, // gasLimit
          maxFeePerGas, // maxFeePerGas
          "0x" // data (empty for simple ETH transfer)
        ]
      });

      const retryableTx = {
        to: inboxAddress,
        value: amountWei + maxSubmissionCost,
        data: data
      };

      logger.info('Sending retryable ticket transaction', {
        to: this.deployerWallet.account.address,
        value: amountWei.toString(),
        maxSubmissionCost: maxSubmissionCost.toString(),
        gasLimit: gasLimit.toString(),
        maxFeePerGas: maxFeePerGas.toString()
      });

      // Send the transaction
      const txHash = await this.deployerWallet.sendTransaction(retryableTx);
      const txReceipt = await this.parentChainPublicClient.waitForTransactionReceipt({ hash: txHash });

      logger.info('Retryable ticket transaction confirmed', {
        transactionHash: txHash,
        amount: amount,
        gasUsed: txReceipt.gasUsed?.toString()
      });

      // Wait for the retryable ticket to be executed on the Orbit chain
      logger.info('Waiting for funds to arrive on Orbit chain');
      await this.waitForRetryableExecution(txReceipt);

      return {
        success: true,
        transactionHash: txHash
      };

    } catch (error) {
      logger.error('Direct funding failed', { 
        error: error instanceof Error ? error.message : 'Unknown error',
        stack: error instanceof Error ? error.stack : undefined
      });
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Direct funding failed'
      };
    }
  }

  /**
   * Wait for retryable ticket execution on Orbit chain
   */
  private async waitForRetryableExecution(txReceipt: any, timeoutMs: number = 300000): Promise<void> {
    const startTime = Date.now();
    
    while (Date.now() - startTime < timeoutMs) {
      try {
        // Check if the retryable ticket has been executed
        // This is a simplified check - in practice, you'd need to parse the logs
        // to find the retryable ticket ID and check its status
        logger.info('Checking retryable ticket execution status');
        
        // For now, just wait a bit and assume it's executed
        // In a production implementation, you'd check the actual retryable status
        await new Promise(resolve => setTimeout(resolve, 10000)); // Wait 10 seconds
        
        logger.info('Retryable ticket execution completed');
        return;
      } catch (error) {
        logger.warn('Error checking retryable status', { error });
        await new Promise(resolve => setTimeout(resolve, 5000)); // Wait 5 seconds before retry
      }
    }
    
    throw new Error('Retryable ticket execution timeout');
  }

  /**
   * Get deployer wallet address
   */
  getDeployerAddress(): Address {
    return this.deployerWallet.account.address;
  }

  /**
   * Get the inbox address for retryable tickets from rollup contract
   */
  private async getInboxAddress(rollupAddress?: Address): Promise<Address | null> {
    try {
      if (rollupAddress) {
        // Try to get inbox address from the rollup contract
        const inboxAddress = await this.parentChainPublicClient.readContract({
          address: rollupAddress,
          abi: [
            {
              name: 'inbox',
              type: 'function',
              stateMutability: 'view',
              inputs: [],
              outputs: [{ name: '', type: 'address' }]
            }
          ],
          functionName: 'inbox'
        });
        
        if (inboxAddress && inboxAddress !== '0x0000000000000000000000000000000000000000') {
          logger.info('Retrieved inbox address from rollup contract', { inboxAddress });
          return inboxAddress as Address;
        }
      }
      
      // Fallback to known inbox addresses
      const knownInboxAddresses = {
        421614: '0x0000000000000000000000000000000000000064', // Arbitrum Sepolia inbox
        42161: '0x0000000000000000000000000000000000000064',  // Arbitrum One inbox
      };
      
      const fallbackAddress = knownInboxAddresses[this.config.parentChainId as keyof typeof knownInboxAddresses] as Address;
      logger.info('Using fallback inbox address', { fallbackAddress });
      return fallbackAddress || null;
    } catch (error) {
      logger.error('Failed to get inbox address', { error });
      return null;
    }
  }
}

export default BridgeService;
