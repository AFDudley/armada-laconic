import Config from 'react-native-config';
import { COSMOS_TESTNET_CHAINS } from './wallet-connect/COSMOSData';
import { EIP155_CHAINS } from './wallet-connect/EIP155Data';

export const EIP155 = 'eip155';
export const COSMOS = 'cosmos';
export const DEFAULT_NETWORKS = [
  {
    chainId: 'laconic_9000-1',
    networkName: 'laconicd',
    namespace: COSMOS,
    rpcUrl: Config.LACONICD_RPC_URL,
    blockExplorerUrl: '',
    nativeDenom: 'alnt',
    addressPrefix: 'laconic',
    coinType: '118',
    gasPrice: '1',
    isDefault: true,
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
