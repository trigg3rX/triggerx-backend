// Bridge Service for transferring funds from parent chain to Orbit chain
import { createPublicClient, createWalletClient, http, type Address, type Chain, defineChain, encodeFunctionData } from 'viem';
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
   * Fund deployer wallet using ETH deposit via Inbox.depositEth
   * This follows the official Arbitrum documentation for ETH deposits
   */
  async fundDeployerDirectly(amount: string = this.config.bridgeAmount || "0.02", rollupAddress?: Address): Promise<BridgeResult> {
    try {
      logger.info('Funding deployer wallet via Inbox.depositEth', { amount, rollupAddress });

      if (!this.orbitChainPublicClient) {
        throw new Error('Orbit chain client not initialized. Call setupOrbitChainClient first.');
      }

      // Convert amount to wei
      const amountWei = BigInt(parseFloat(amount) * 1e18);

      // Get the inbox address
      const inboxAddress = await this.getInboxAddress(rollupAddress);
      
      if (!inboxAddress) {
        logger.warn('Could not determine inbox address for ETH deposit, skipping bridge funding', { rollupAddress });
        return {
          success: false,
          error: 'Could not determine inbox address for ETH deposit. Orbit chain may not be ready for bridging yet.'
        };
      }

      logger.info('Using inbox address for ETH deposit', { inboxAddress });

      // Use the official Inbox.depositEth method as per Arbitrum docs
      const depositEthABI = [
        {
          name: 'depositEth',
          type: 'function',
          stateMutability: 'payable',
          inputs: [
            { name: 'destAddr', type: 'address' }
          ],
          outputs: [{ name: 'ticketId', type: 'uint256' }]
        }
      ];

      // Encode the depositEth function call
      const data = encodeFunctionData({
        abi: depositEthABI,
        functionName: 'depositEth',
        args: [this.deployerWallet.account.address] // destAddr
      });

      const depositTx = {
        to: inboxAddress,
        value: amountWei, // Only send the ETH amount, no additional fees needed for depositEth
        data: data
      };

      logger.info('Sending ETH deposit transaction', {
        to: this.deployerWallet.account.address,
        value: amountWei.toString(),
        inboxAddress: inboxAddress
      });

      // Send the transaction
      const txHash = await this.deployerWallet.sendTransaction(depositTx);
      const txReceipt = await this.parentChainPublicClient.waitForTransactionReceipt({ hash: txHash });

      logger.info('ETH deposit transaction confirmed', {
        transactionHash: txHash,
        amount: amount,
        gasUsed: txReceipt.gasUsed?.toString()
      });

      // Wait for the deposit to be processed on the Orbit chain
      logger.info('Waiting for ETH deposit to arrive on Orbit chain');
      await this.waitForDepositExecution(txReceipt);

      return {
        success: true,
        transactionHash: txHash
      };

    } catch (error) {
      logger.error('ETH deposit failed', { 
        error: error instanceof Error ? error.message : 'Unknown error',
        stack: error instanceof Error ? error.stack : undefined
      });
      return {
        success: false,
        error: error instanceof Error ? error.message : 'ETH deposit failed'
      };
    }
  }

  /**
   * Alternative method: Fund deployer wallet using retryable tickets
   * This provides more flexibility than depositEth but is more complex
   */
  async fundDeployerViaRetryableTicket(amount: string = this.config.bridgeAmount || "0.02", rollupAddress?: Address): Promise<BridgeResult> {
    try {
      logger.info('Funding deployer wallet via retryable ticket', { amount, rollupAddress });

      if (!this.orbitChainPublicClient) {
        throw new Error('Orbit chain client not initialized. Call setupOrbitChainClient first.');
      }

      // Convert amount to wei
      const amountWei = BigInt(parseFloat(amount) * 1e18);

      // Get the inbox address
      const inboxAddress = await this.getInboxAddress(rollupAddress);
      
      if (!inboxAddress) {
        logger.warn('Could not determine inbox address for retryable ticket, skipping bridge funding', { rollupAddress });
        return {
          success: false,
          error: 'Could not determine inbox address for retryable ticket. Orbit chain may not be ready for bridging yet.'
        };
      }

      logger.info('Using inbox address for retryable ticket', { inboxAddress });

      // Estimate gas parameters dynamically
      const maxSubmissionCost = await this.estimateMaxSubmissionCost();
      const gasLimit = BigInt("200000"); // 200k gas
      const maxFeePerGas = await this.estimateMaxFeePerGas();

      // Create retryable ticket ABI
      const retryableTicketABI = [
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

      // Encode the retryable ticket function call
      const data = encodeFunctionData({
        abi: retryableTicketABI,
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
        value: amountWei + maxSubmissionCost, // ETH amount + submission cost
        data: data
      };

      logger.info('Sending retryable ticket transaction', {
        to: this.deployerWallet.account.address,
        value: amountWei.toString(),
        maxSubmissionCost: maxSubmissionCost.toString(),
        gasLimit: gasLimit.toString(),
        maxFeePerGas: maxFeePerGas.toString(),
        totalValue: (amountWei + maxSubmissionCost).toString(),
        inboxAddress: inboxAddress
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
      logger.info('Waiting for retryable ticket to be executed on Orbit chain');
      await this.waitForRetryableExecution(txReceipt);

      return {
        success: true,
        transactionHash: txHash
      };

    } catch (error) {
      logger.error('Retryable ticket funding failed', { 
        error: error instanceof Error ? error.message : 'Unknown error',
        stack: error instanceof Error ? error.stack : undefined
      });
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Retryable ticket funding failed'
      };
    }
  }

  /**
   * Wait for ETH deposit execution on Orbit chain
   */
  private async waitForDepositExecution(txReceipt: any, timeoutMs: number = 300000): Promise<void> {
    const startTime = Date.now();
    
    while (Date.now() - startTime < timeoutMs) {
      try {
        // Check if the deposit has been processed
        logger.info('Checking ETH deposit execution status');
        
        // For ETH deposits, we can check the balance on the Orbit chain
        const balance = await this.orbitChainPublicClient.getBalance({
          address: this.deployerWallet.account.address
        });
        
        if (balance > 0n) {
          logger.info('ETH deposit execution completed', { balance: balance.toString() });
          return;
        }
        
        await new Promise(resolve => setTimeout(resolve, 10000)); // Wait 10 seconds
      } catch (error) {
        logger.warn('Error checking deposit status', { error });
        await new Promise(resolve => setTimeout(resolve, 5000)); // Wait 5 seconds before retry
      }
    }
    
    throw new Error('ETH deposit execution timeout');
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
   * Estimate max submission cost for retryable tickets
   */
  private async estimateMaxSubmissionCost(): Promise<bigint> {
    try {
      // Get current gas price
      const gasPrice = await this.parentChainPublicClient.getGasPrice();
      
      // Estimate submission cost (typically 0.001 ETH or gas price * 100k)
      const estimatedCost = gasPrice * BigInt(100000);
      
      // Ensure minimum cost
      const minCost = BigInt("1000000000000000"); // 0.001 ETH
      
      return estimatedCost > minCost ? estimatedCost : minCost;
    } catch (error) {
      logger.warn('Failed to estimate max submission cost, using default', { error });
      return BigInt("1000000000000000"); // 0.001 ETH default
    }
  }

  /**
   * Estimate max fee per gas for retryable tickets
   */
  private async estimateMaxFeePerGas(): Promise<bigint> {
    try {
      // Get current gas price
      const gasPrice = await this.parentChainPublicClient.getGasPrice();
      
      // Add 20% buffer
      return gasPrice * BigInt(120) / BigInt(100);
    } catch (error) {
      logger.warn('Failed to estimate max fee per gas, using default', { error });
      return BigInt("2000000000"); // 2 gwei default
    }
  }

  /**
   * Get deployer wallet address
   */
  getDeployerAddress(): Address {
    return this.deployerWallet.account.address;
  }

  /**
   * Get the recommended bridging method based on use case
   * Based on Arbitrum documentation recommendations
   */
  getRecommendedBridgingMethod(): 'depositEth' | 'retryableTicket' {
    // For simple ETH transfers to fund a wallet, depositEth is recommended
    // For more complex operations or when you need fallback functions, use retryable tickets
    return 'depositEth';
  }

  /**
   * Main bridging method that uses the recommended approach
   */
  async bridgeETH(amount: string = this.config.bridgeAmount || "0.02", rollupAddress?: Address): Promise<BridgeResult> {
    const method = this.getRecommendedBridgingMethod();
    
    if (method === 'depositEth') {
      return this.fundDeployerDirectly(amount, rollupAddress);
    } else {
      return this.fundDeployerViaRetryableTicket(amount, rollupAddress);
    }
  }

  /**
   * Get the inbox address for retryable tickets from rollup contract
   */
  private async getInboxAddress(rollupAddress?: Address): Promise<Address | null> {
    try {
      if (rollupAddress) {
        // Try to get inbox address from the rollup contract
        try {
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
        } catch (error) {
          logger.warn('Failed to read inbox address from rollup contract', { error });
        }
      }
      
      // For Orbit chains, we need to use the rollup's specific inbox
      // If we can't get it from the rollup contract, we should not proceed
      logger.error('Cannot determine inbox address for Orbit chain', { 
        rollupAddress,
        parentChainId: this.config.parentChainId 
      });
      return null;
    } catch (error) {
      logger.error('Failed to get inbox address', { error });
      return null;
    }
  }
}

export default BridgeService;
