import { COSMOS_TESTNET_CHAINS } from './wallet-connect/COSMOSData';
import { EIP155_CHAINS } from './wallet-connect/EIP155Data';
import { NetworksFormData } from '../types';

export const EIP155 = 'eip155';
export const COSMOS = 'cosmos';

export const DEFAULT_NETWORKS: NetworksFormData[] = [
  {
    chainId: 'laconic-mainnet',
    networkName: 'laconicd mainnet',
    namespace: COSMOS,
    rpcUrl: import.meta.env.REACT_APP_LACONICD_RPC_URL,
    blockExplorerUrl: 'https://explorer.laconic.com/laconic-mainnet',
    nativeDenom: 'alnt',
    addressPrefix: 'laconic',
    coinType: '118',
    gasPrice: '0.001',
    isDefault: false,
  },
  {
    chainId: 'laconic-testnet-2',
    networkName: 'laconicd testnet-2',
    namespace: COSMOS,
    rpcUrl: 'https://laconicd-sapo.laconic.com',
    blockExplorerUrl: '',
    nativeDenom: 'alnt',
    addressPrefix: 'laconic',
    coinType: '118',
    gasPrice: '0.001',
    isDefault: false,
  },
  {
    chainId: '1',
    networkName: EIP155_CHAINS['eip155:1'].name,
    namespace: EIP155,
    rpcUrl: EIP155_CHAINS['eip155:1'].rpc,
    blockExplorerUrl: '',
    currencySymbol: 'ETH',
    coinType: '60',
    isDefault: true,
  },
  {
    chainId: 'theta-testnet-001',
    networkName: COSMOS_TESTNET_CHAINS['cosmos:theta-testnet-001'].name,
    namespace: COSMOS,
    rpcUrl: COSMOS_TESTNET_CHAINS['cosmos:theta-testnet-001'].rpc,
    blockExplorerUrl: '',
    nativeDenom: 'uatom',
    addressPrefix: 'cosmos',
    coinType: '118',
    gasPrice: '0.025',
    isDefault: true,
  },
];

export const CHAINID_DEBOUNCE_DELAY = 250;

export const EMPTY_FIELD_ERROR = 'Field cannot be empty';
export const INVALID_URL_ERROR = 'Invalid URL';

export const IS_NUMBER_REGEX = /^\d+$/;

export const IS_IMPORT_WALLET_ENABLED = false;

// iframe request types
export const REQUEST_COSMOS_ACCOUNTS = 'REQUEST_COSMOS_ACCOUNTS';
export const REQUEST_SIGN_TX = 'REQUEST_SIGN_TX';
export const REQUEST_SIGN_MESSAGE = 'REQUEST_SIGN_MESSAGE';
export const REQUEST_WALLET_ACCOUNTS = 'REQUEST_WALLET_ACCOUNTS';
export const REQUEST_CREATE_OR_GET_ACCOUNTS = 'REQUEST_CREATE_OR_GET_ACCOUNTS';
export const REQUEST_TX = 'REQUEST_TX';
export const REQUEST_ACCOUNT_PK = 'REQUEST_ACCOUNT_PK';
export const REQUEST_ADD_ACCOUNT = 'REQUEST_ADD_ACCOUNT';
export const AUTO_SIGN_IN = 'AUTO_SIGN_IN';
export const CHECK_BALANCE = 'CHECK_BALANCE';
export const REQUEST_ADD_NETWORK = 'REQUEST_ADD_NETWORK';

// iframe response types
export const COSMOS_ACCOUNTS_RESPONSE = 'COSMOS_ACCOUNTS_RESPONSE';
export const SIGN_TX_RESPONSE = 'SIGN_TX_RESPONSE';
export const SIGN_MESSAGE_RESPONSE = 'SIGN_MESSAGE_RESPONSE';
export const TRANSACTION_RESPONSE = 'TRANSACTION_RESPONSE';
export const SIGN_IN_RESPONSE = 'SIGN_IN_RESPONSE';
export const ACCOUNT_PK_RESPONSE = 'ACCOUNT_PK_RESPONSE';
export const ADD_ACCOUNT_RESPONSE = 'ADD_ACCOUNT_RESPONSE';
export const WALLET_ACCOUNTS_DATA = 'WALLET_ACCOUNTS_DATA';
export const IS_SUFFICIENT = 'IS_SUFFICIENT';
export const NETWORK_ADDED_RESPONSE = "NETWORK_ADDED_RESPONSE";
export const NETWORK_ALREADY_EXISTS_RESPONSE = "NETWORK_ALREADY_EXISTS_RESPONSE";
export const NETWORK_ADD_FAILED_RESPONSE = "NETWORK_ADD_FAILED_RESPONSE";
