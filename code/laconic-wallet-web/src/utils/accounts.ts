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
} from './key-store';
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
  recoveryPhrase?: string,
): Promise<string> => {
  const mnemonic = recoveryPhrase ? recoveryPhrase : utils.entropyToMnemonic(utils.randomBytes(16));

  const hdNode = HDNode.fromMnemonic(mnemonic);
  await setInternetCredentials('mnemonicServer', 'mnemonic', mnemonic);

  await createWalletFromMnemonic(networksData, hdNode, mnemonic);

  return mnemonic;
};

const createWalletFromMnemonic = async (
  networksData: NetworksDataState[],
  hdNode: HDNode,
  mnemonic: string
): Promise<void> => {
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
          await getCosmosAccountByHDPath(mnemonic, hdPath, network.addressPrefix)
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
};

const addAccount = async (
  chainId: string,
): Promise<Account | undefined> => {
  try {
    let selectedNetworkAccount
    const networksData = await retrieveNetworksData();

    // Add account to all networks and return account for selected network
    for (const network of networksData) {
      const namespaceChainId = `${network.namespace}:${network.chainId}`;
      const id = await getNextAccountId(namespaceChainId);
      const hdPath = getHDPath(namespaceChainId, `0'/0/${id}`);
      const account = await addAccountFromHDPath(hdPath, network);
      await updateAccountCounter(namespaceChainId, id);

      if (network.chainId === chainId) {
        selectedNetworkAccount = account;
      }
    }

    return selectedNetworkAccount;
  } catch (error) {
    console.error('Error creating account:', error);
  }
};

const addAccountsForNetwork = async (
  network: NetworksDataState,
  numberOfAccounts: number,
): Promise<void> => {
  try {
    const namespaceChainId = `${network.namespace}:${network.chainId}`;

    for (let i = 0; i < numberOfAccounts; i++) {
      const id = await getNextAccountId(namespaceChainId);
      const hdPath = getHDPath(namespaceChainId, `0'/0/${id}`);
      await addAccountFromHDPath(hdPath, network);
      await updateAccountCounter(namespaceChainId, id);
    }
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

const addNewNetwork = async (
  newNetworkData: NetworksFormData
): Promise<NetworksDataState[]> => {
  const mnemonicServer = await getInternetCredentials("mnemonicServer");
  const mnemonic = mnemonicServer;

  if (!mnemonic) {
    throw new Error("Mnemonic not found");
  }

  const hdNode = HDNode.fromMnemonic(mnemonic);

  const hdPath = `m/44'/${newNetworkData.coinType}'/0'/0/0`;
  const node = hdNode.derivePath(hdPath);
  let address;

  switch (newNetworkData.namespace) {
    case EIP155:
      address = node.address;
      break;

    case COSMOS:
      address = (
        await getCosmosAccountByHDPath(
          mnemonic,
          hdPath,
          newNetworkData.addressPrefix,
        )
      ).data.address;
      break;

    default:
      throw new Error("Unsupported namespace");
  }

  const accountInfo = `${hdPath},${node.privateKey},${node.publicKey},${address}`;

  await Promise.all([
    setInternetCredentials(
      `accounts/${newNetworkData.namespace}:${newNetworkData.chainId}/0`,
      "_",
      accountInfo,
    ),
    setInternetCredentials(
      `addAccountCounter/${newNetworkData.namespace}:${newNetworkData.chainId}`,
      "_",
      "1",
    ),
    setInternetCredentials(
      `accountIndices/${newNetworkData.namespace}:${newNetworkData.chainId}`,
      "_",
      "0",
    ),
  ]);

  const retrievedNetworksData = await storeNetworkData(newNetworkData);

  // Get number of accounts in first network
  const nextAccountId = await getNextAccountId(
    `${retrievedNetworksData[0].namespace}:${retrievedNetworksData[0].chainId}`,
  );

  const selectedNetwork = retrievedNetworksData.find(
    (network) => network.chainId === newNetworkData.chainId,
  );

  await addAccountsForNetwork(selectedNetwork!, nextAccountId - 1);

  return retrievedNetworksData;
}

const storeNetworkData = async (
  networkData: NetworksFormData,
): Promise<NetworksDataState[]> => {
  const networks = await getInternetCredentials('networks');
  let retrievedNetworks = [];
  if (networks) {
    retrievedNetworks = JSON.parse(networks!);
  }
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

  if (!networks) {
    return [];
  }

  const parsedNetworks: NetworksDataState[] = JSON.parse(networks);

  return parsedNetworks;
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
  const accountIndices = accountIndicesServer;
  if (!accountIndices) {
    return;
  }

  const loadedAccounts = await retrieveAccountsForNetwork(
    `${currentNetworkData.namespace}:${currentNetworkData.chainId}`,
    accountIndices,
  )

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
  const accountIndices = accountIndicesServer;

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

  const mnemonic = mnemonicStore;
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
        await getCosmosAccountByHDPath(mnemonic, hdPath, networkData.addressPrefix)
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

  const accountCounter = idStore;
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

const getCosmosAccountByHDPath = async (
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

const checkNetworkForChainID = async (
  chainId: string,
): Promise<boolean> => {
  const networks = await getInternetCredentials('networks');

  if (!networks) {
    return false;
  }

  const networksData: NetworksFormData[] = JSON.parse(networks);

  return networksData.some((network) => network.chainId === chainId);
}

const isWalletCreated = async (
): Promise<boolean> => {
  const mnemonicServer = await getInternetCredentials("mnemonicServer");
  const mnemonic = mnemonicServer;

  return mnemonic !== null;
};

export {
  createWallet,
  addAccount,
  addAccountFromHDPath,
  addAccountsForNetwork,
  storeNetworkData,
  retrieveNetworksData,
  retrieveAccounts,
  retrieveSingleAccount,
  resetWallet,
  accountInfoFromHDPath,
  getNextAccountId,
  updateAccountCounter,
  getCosmosAccountByHDPath,
  addNewNetwork,
  checkNetworkForChainID,
  isWalletCreated
};
