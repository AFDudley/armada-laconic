/*  Importing this library provides react native with a secure random source.
For more information, "visit https://docs.ethers.org/v5/cookbook/react-native/#cookbook-reactnative-security" */
import 'react-native-get-random-values';

import '@ethersproject/shims';

import { utils } from 'ethers';
import { HDNode } from 'ethers/lib/utils';
import {
  setInternetCredentials,
  resetInternetCredentials,
  getInternetCredentials,
} from 'react-native-keychain';

import { Secp256k1HdWallet } from '@cosmjs/amino';
import { AccountData } from '@cosmjs/proto-signing';
import { stringToPath } from '@cosmjs/crypto';

import { Account, NetworksDataState, NetworksFormData } from '../types';
import {
  getHDPath,
  getPathKey,
  resetKeyServers,
  updateAccountIndices,
} from './misc';
import { COSMOS, EIP155 } from './constants';

const createWallet = async (
  networksData: NetworksDataState[],
): Promise<string> => {
  const mnemonic = utils.entropyToMnemonic(utils.randomBytes(16));
  await setInternetCredentials('mnemonicServer', 'mnemonic', mnemonic);

  const hdNode = HDNode.fromMnemonic(mnemonic);

  for (const network of networksData) {
    const hdPath = `m/44'/${network.coinType}'/0'/0/0`;
    const node = hdNode.derivePath(hdPath);
    let address;

    switch (network.namespace) {
      case EIP155:
        address = node.address;
        break;

      case COSMOS:
        address = (
          await getCosmosAccounts(mnemonic, hdPath, network.addressPrefix)
        ).data.address;
        break;

      default:
        throw new Error('Unsupported namespace');
    }

    const accountInfo = `${hdPath},${node.privateKey},${node.publicKey},${address}`;

    await Promise.all([
      setInternetCredentials(
        `accounts/${network.namespace}:${network.chainId}/0`,
        '_',
        accountInfo,
      ),
      setInternetCredentials(
        `addAccountCounter/${network.namespace}:${network.chainId}`,
        '_',
        '1',
      ),
      setInternetCredentials(
        `accountIndices/${network.namespace}:${network.chainId}`,
        '_',
        '0',
      ),
    ]);
  }

  return mnemonic;
};

const addAccount = async (
  networkData: NetworksDataState,
): Promise<Account | undefined> => {
  try {
    const namespaceChainId = `${networkData.namespace}:${networkData.chainId}`;
    const id = await getNextAccountId(namespaceChainId);
    const hdPath = getHDPath(namespaceChainId, `0'/0/${id}`);
    const accounts = await addAccountFromHDPath(hdPath, networkData);
    await updateAccountCounter(namespaceChainId, id);
    return accounts;
  } catch (error) {
    console.error('Error creating account:', error);
  }
};

const addAccountFromHDPath = async (
  hdPath: string,
  networkData: NetworksDataState,
): Promise<Account | undefined> => {
  try {
    const account = await accountInfoFromHDPath(hdPath, networkData);
    if (!account) {
      throw new Error('Error while creating account');
    }

    const { privKey, pubKey, address } = account;

    const namespaceChainId = `${networkData.namespace}:${networkData.chainId}`;

    const index = (await updateAccountIndices(namespaceChainId)).index;

    await Promise.all([
      setInternetCredentials(
        `accounts/${namespaceChainId}/${index}`,
        '_',
        `${hdPath},${privKey},${pubKey},${address}`,
      ),
    ]);

    return { index, pubKey, address, hdPath };
  } catch (error) {
    console.error(error);
  }
};

const storeNetworkData = async (
  networkData: NetworksFormData,
): Promise<NetworksDataState[]> => {
  const networks = await getInternetCredentials('networks');
  const retrievedNetworks =
    networks && networks.password ? JSON.parse(networks.password) : [];
  let networkId = 0;
  if (retrievedNetworks.length > 0) {
    networkId = retrievedNetworks[retrievedNetworks.length - 1].networkId + 1;
  }

  const updatedNetworks: NetworksDataState[] = [
    ...retrievedNetworks,
    {
      ...networkData,
      networkId: String(networkId),
    },
  ];
  await setInternetCredentials(
    'networks',
    '_',
    JSON.stringify(updatedNetworks),
  );
  return updatedNetworks;
};

const retrieveNetworksData = async (): Promise<NetworksDataState[]> => {
  const networks = await getInternetCredentials('networks');
  const retrievedNetworks: NetworksDataState[] =
    networks && networks.password ? JSON.parse(networks.password) : [];

  return retrievedNetworks;
};

export const retrieveAccountsForNetwork = async (
  namespaceChainId: string,
  accountsIndices: string,
): Promise<Account[]> => {
  const accountsIndexArray = accountsIndices.split(',');

  const loadedAccounts = await Promise.all(
    accountsIndexArray.map(async i => {
      const { address, path, pubKey } = await getPathKey(
        namespaceChainId,
        Number(i),
      );

      const account: Account = {
        index: Number(i),
        pubKey,
        address,
        hdPath: path,
      };
      return account;
    }),
  );

  return loadedAccounts;
};

const retrieveAccounts = async (
  currentNetworkData: NetworksDataState,
): Promise<Account[] | undefined> => {
  const accountIndicesServer = await getInternetCredentials(
    `accountIndices/${currentNetworkData.namespace}:${currentNetworkData.chainId}`,
  );
  const accountIndices = accountIndicesServer && accountIndicesServer.password;
  const loadedAccounts =
    accountIndices !== false
      ? await retrieveAccountsForNetwork(
          `${currentNetworkData.namespace}:${currentNetworkData.chainId}`,
          accountIndices,
        )
      : undefined;

  return loadedAccounts;
};

const retrieveSingleAccount = async (
  namespace: string,
  chainId: string,
  address: string,
) => {
  let loadedAccounts;

  const accountIndicesServer = await getInternetCredentials(
    `accountIndices/${namespace}:${chainId}`,
  );
  const accountIndices = accountIndicesServer && accountIndicesServer.password;

  if (!accountIndices) {
    throw new Error('Indices for given chain not found');
  }

  loadedAccounts = await retrieveAccountsForNetwork(
    `${namespace}:${chainId}`,
    accountIndices,
  );

  if (!loadedAccounts) {
    throw new Error('Accounts for given chain not found');
  }

  return loadedAccounts.find(account => account.address === address);
};

const resetWallet = async () => {
  try {
    await Promise.all([
      resetInternetCredentials('mnemonicServer'),
      resetKeyServers(EIP155),
      resetKeyServers(COSMOS),
      setInternetCredentials('networks', '_', JSON.stringify([])),
    ]);
  } catch (error) {
    console.error('Error resetting wallet:', error);
    throw error;
  }
};

const accountInfoFromHDPath = async (
  hdPath: string,
  networkData: NetworksDataState,
): Promise<
  { privKey: string; pubKey: string; address: string } | undefined
> => {
  const mnemonicStore = await getInternetCredentials('mnemonicServer');
  if (!mnemonicStore) {
    throw new Error('Mnemonic not found!');
  }

  const mnemonic = mnemonicStore.password;
  const hdNode = HDNode.fromMnemonic(mnemonic);
  const node = hdNode.derivePath(hdPath);

  const privKey = node.privateKey;
  const pubKey = node.publicKey;

  let address: string;

  switch (networkData.namespace) {
    case EIP155:
      address = node.address;
      break;
    case COSMOS:
      address = (
        await getCosmosAccounts(mnemonic, hdPath, networkData.addressPrefix)
      ).data.address;
      break;
    default:
      throw new Error('Invalid wallet type');
  }
  return { privKey, pubKey, address };
};

const getNextAccountId = async (namespaceChainId: string): Promise<number> => {
  const idStore = await getInternetCredentials(
    `addAccountCounter/${namespaceChainId}`,
  );
  if (!idStore) {
    throw new Error('Account id not found');
  }

  const accountCounter = idStore.password;
  const nextCounter = Number(accountCounter);
  return nextCounter;
};

const updateAccountCounter = async (
  namespaceChainId: string,
  id: number,
): Promise<void> => {
  const idStore = await getInternetCredentials(
    `addAccountCounter/${namespaceChainId}`,
  );
  if (!idStore) {
    throw new Error('Account id not found');
  }

  const updatedCounter = String(id + 1);
  await resetInternetCredentials(`addAccountCounter/${namespaceChainId}`);
  await setInternetCredentials(
    `addAccountCounter/${namespaceChainId}`,
    '_',
    updatedCounter,
  );
};

const getCosmosAccounts = async (
  mnemonic: string,
  path: string,
  prefix: string = COSMOS,
): Promise<{ cosmosWallet: Secp256k1HdWallet; data: AccountData }> => {
  const cosmosWallet = await Secp256k1HdWallet.fromMnemonic(mnemonic, {
    hdPaths: [stringToPath(path)],
    prefix,
  });

  const accountsData = await cosmosWallet.getAccounts();
  const data = accountsData[0];

  return { cosmosWallet, data };
};

export {
  createWallet,
  addAccount,
  addAccountFromHDPath,
  storeNetworkData,
  retrieveNetworksData,
  retrieveAccounts,
  retrieveSingleAccount,
  resetWallet,
  accountInfoFromHDPath,
  getNextAccountId,
  updateAccountCounter,
  getCosmosAccounts,
};
