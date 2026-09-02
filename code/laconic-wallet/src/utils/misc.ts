/*  Importing this library provides react native with a secure random source.
For more information, "visit https://docs.ethers.org/v5/cookbook/react-native/#cookbook-reactnative-security" */
import 'react-native-get-random-values';

import '@ethersproject/shims';

import {
  getInternetCredentials,
  resetInternetCredentials,
  setInternetCredentials,
} from 'react-native-keychain';

import { AccountData } from '@cosmjs/amino';
import { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import { stringToPath } from '@cosmjs/crypto';
import { EIP155 } from './constants';
import { NetworksDataState } from '../types';

const getMnemonic = async (): Promise<string> => {
  const mnemonicStore = await getInternetCredentials('mnemonicServer');
  if (!mnemonicStore) {
    throw new Error('Mnemonic not found!');
  }

  const mnemonic = mnemonicStore.password;
  return mnemonic;
};

const getHDPath = (namespaceChainId: string, path: string): string => {
  const namespace = namespaceChainId.split(':')[0];
  return namespace === EIP155 ? `m/44'/60'/${path}` : `m/44'/118'/${path}`;
};

export const getDirectWallet = async (
  mnemonic: string,
  path: string,
): Promise<{ directWallet: DirectSecp256k1HdWallet; data: AccountData }> => {
  const directWallet = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic, {
    hdPaths: [stringToPath(`m/44'/118'/${path}`)],
  });
  const accountsData = await directWallet.getAccounts();
  const data = accountsData[0];

  return { directWallet, data };
};

const getPathKey = async (
  namespaceChainId: string,
  accountId: number,
): Promise<{
  path: string;
  privKey: string;
  pubKey: string;
  address: string;
}> => {
  const pathKeyStore = await getInternetCredentials(
    `accounts/${namespaceChainId}/${accountId}`,
  );

  if (!pathKeyStore) {
    throw new Error('Error while fetching counter');
  }

  const pathKeyVal = pathKeyStore.password;
  const pathkey = pathKeyVal.split(',');
  const path = pathkey[0];
  const privKey = pathkey[1];
  const pubKey = pathkey[2];
  const address = pathkey[3];

  return { path, privKey, pubKey, address };
};

const getAccountIndices = async (
  namespaceChainId: string,
): Promise<{
  accountIndices: string;
  indices: number[];
  index: number;
}> => {
  const counterStore = await getInternetCredentials(
    `accountIndices/${namespaceChainId}`,
  );

  if (!counterStore) {
    throw new Error('Error while fetching counter');
  }

  let accountIndices = counterStore.password;
  const indices = accountIndices.split(',').map(Number);
  const index = indices[indices.length - 1] + 1;

  return { accountIndices, indices, index };
};

const updateAccountIndices = async (
  namespaceChainId: string,
): Promise<{ accountIndices: string; index: number }> => {
  const accountIndicesData = await getAccountIndices(namespaceChainId);
  const accountIndices = accountIndicesData.accountIndices;
  const index = accountIndicesData.index;
  const updatedAccountIndices = `${accountIndices},${index.toString()}`;

  await resetInternetCredentials(`accountIndices/${namespaceChainId}`);
  await setInternetCredentials(
    `accountIndices/${namespaceChainId}`,
    '_',
    updatedAccountIndices,
  );

  return { accountIndices: updatedAccountIndices, index };
};

const resetKeyServers = async (namespace: string) => {
  const networksServer = await getInternetCredentials('networks');
  if (!networksServer) {
    throw new Error('Networks not found.');
  }

  const networksData: NetworksDataState[] = JSON.parse(networksServer.password);
  const filteredNetworks = networksData.filter(
    network => network.namespace === namespace,
  );

  if (filteredNetworks.length === 0) {
    throw new Error(`No networks found for namespace ${namespace}.`);
  }

  filteredNetworks.forEach(async network => {
    const { chainId } = network;
    const namespaceChainId = `${namespace}:${chainId}`;

    const idStore = await getInternetCredentials(
      `accountIndices/${namespaceChainId}`,
    );
    if (!idStore) {
      throw new Error(`Account indices not found for ${namespaceChainId}.`);
    }

    const accountIds = idStore.password;
    const ids = accountIds.split(',').map(Number);
    const latestId = Math.max(...ids);

    for (let i = 0; i <= latestId; i++) {
      await resetInternetCredentials(`accounts/${namespaceChainId}/${i}`);
    }
    await resetInternetCredentials(`addAccountCounter/${namespaceChainId}`);
    await resetInternetCredentials(`accountIndices/${namespaceChainId}`);
  });
};

export {
  getMnemonic,
  getPathKey,
  updateAccountIndices,
  getHDPath,
  resetKeyServers,
};
