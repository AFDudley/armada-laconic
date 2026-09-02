// Taken from https://github.com/WalletConnect/web-examples/blob/main/advanced/wallets/react-wallet-v2/src/data/COSMOSData.ts

/**
 * Types
 */
export type TCosmosChain = keyof typeof COSMOS_TESTNET_CHAINS;

/**
 * Chains
 */

// Added for pay.laconic.com
export const COSMOS_TESTNET_CHAINS: Record<
  string,
  {
    chainId: string;
    name: string;
    rpc: string;
    namespace: string;
  }
> = {
  'cosmos:theta-testnet-001': {
    chainId: 'theta-testnet-001',
    name: 'Cosmos Hub Testnet',
    rpc: 'https://rpc-t.cosmos.nodestake.top',
    namespace: 'cosmos',
  },
};

/**
 * Methods
 */
export const COSMOS_SIGNING_METHODS = {
  COSMOS_SIGN_DIRECT: 'cosmos_signDirect',
  COSMOS_SIGN_AMINO: 'cosmos_signAmino',
};

export const COSMOS_METHODS = {
  ...COSMOS_SIGNING_METHODS,
  COSMOS_SEND_TOKENS: 'cosmos_sendTokens', // Added for pay.laconic.com
  COSMOS_SEND_TRANSACTION: 'cosmos_sendTransaction', // Added for testnet onboarding app
};
