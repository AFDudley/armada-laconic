/*  Importing this library provides react native with a secure random source.
For more information, "visit https://docs.ethers.org/v5/cookbook/react-native/#cookbook-reactnative-security" */
import 'react-native-get-random-values';

import '@ethersproject/shims';

import { Wallet } from 'ethers';
import { SignDoc } from 'cosmjs-types/cosmos/tx/v1beta1/tx';

import { SignMessageParams } from '../types';
import { getDirectWallet, getMnemonic, getPathKey } from './misc';
import { getCosmosAccounts } from './accounts';
import { COSMOS, EIP155 } from './constants';

const signMessage = async ({
  message,
  namespace,
  chainId,
  accountId,
}: SignMessageParams): Promise<string | undefined> => {
  const path = await getPathKey(`${namespace}:${chainId}`, accountId);

  switch (namespace) {
    case EIP155:
      return await signEthMessage(message, accountId, chainId);
    case COSMOS:
      return await signCosmosMessage(message, path.path);
    default:
      throw new Error('Invalid wallet type');
  }
};

const signEthMessage = async (
  message: string,
  accountId: number,
  chainId: string,
): Promise<string | undefined> => {
  try {
    const privKey = (await getPathKey(`${EIP155}:${chainId}`, accountId))
      .privKey;
    const wallet = new Wallet(privKey);
    const signature = await wallet.signMessage(message);

    return signature;
  } catch (error) {
    console.error('Error signing Ethereum message:', error);
    throw error;
  }
};

const signCosmosMessage = async (
  message: string,
  path: string,
): Promise<string | undefined> => {
  try {
    const mnemonic = await getMnemonic();
    const cosmosAccount = await getCosmosAccounts(mnemonic, path);
    const address = cosmosAccount.data.address;
    const cosmosSignature = await cosmosAccount.cosmosWallet.signAmino(
      address,
      {
        chain_id: '',
        account_number: '0',
        sequence: '0',
        fee: {
          gas: '0',
          amount: [],
        },
        msgs: [
          {
            type: 'sign/MsgSignData',
            value: {
              signer: address,
              data: btoa(message),
            },
          },
        ],
        memo: '',
      },
    );

    return cosmosSignature.signature.signature;
  } catch (error) {
    console.error('Error signing Cosmos message:', error);
    throw error;
  }
};

const signDirectMessage = async (
  namespaceChainId: string,
  accountId: number,
  signDoc: SignDoc,
): Promise<string | undefined> => {
  try {
    const path = (await getPathKey(namespaceChainId, accountId)).path;
    const mnemonic = await getMnemonic();
    const { directWallet, data } = await getDirectWallet(mnemonic, path);

    const directSignature = await directWallet.signDirect(
      data.address,
      signDoc,
    );

    return directSignature.signature.signature;
  } catch (error) {
    console.error('Error signing Cosmos message:', error);
    throw error;
  }
};

export { signMessage, signEthMessage, signCosmosMessage, signDirectMessage };
