/**
 * @desc Refference list of eip155 chains
 * @url https://chainlist.org
 */

// Taken from https://github.com/WalletConnect/web-examples/blob/main/advanced/wallets/react-wallet-v2/src/data/EIP155Data.ts

/**
 * Types
 */
export type TEIP155Chain = keyof typeof EIP155_CHAINS;

export type EIP155Chain = {
  chainId: number;
  name: string;
  logo: string;
  rgb: string;
  rpc: string;
  namespace: string;
  smartAccountEnabled?: boolean;
};

/**
 * Chains
 */
export const EIP155_CHAINS: Record<string, EIP155Chain> = {
  'eip155:1': {
    chainId: 1,
    name: 'Ethereum',
    logo: '/chain-logos/eip155-1.png',
    rgb: '99, 125, 234',
    rpc: 'https://cloudflare-eth.com/',
    namespace: 'eip155',
  },
};

/**
 * Methods
 */
export const EIP155_SIGNING_METHODS = {
  PERSONAL_SIGN: 'personal_sign',
  ETH_SEND_TRANSACTION: 'eth_sendTransaction',
};
