// Taken from https://medium.com/walletconnect/how-to-build-a-wallet-in-react-native-with-the-web3wallet-sdk-b6f57bf02f9a
import { BigNumber, Wallet, providers } from 'ethers';

import { formatJsonRpcError, formatJsonRpcResult } from '@json-rpc-tools/utils';
import { SignClientTypes } from '@walletconnect/types';
import { getSdkError } from '@walletconnect/utils';
import {
  SigningStargateClient,
  StdFee,
  MsgSendEncodeObject,
} from '@cosmjs/stargate';
import { EncodeObject } from '@cosmjs/proto-signing';
import { LaconicClient } from '@cerc-io/registry-sdk';

import { EIP155_SIGNING_METHODS } from './EIP155Data';
import { signDirectMessage, signEthMessage } from '../sign-message';
import { Account } from '../../types';
import { getMnemonic, getPathKey } from '../misc';
import { getCosmosAccounts } from '../accounts';
import { COSMOS_METHODS } from './COSMOSData';

interface EthSendTransaction {
  type: 'eth_sendTransaction';
  provider: providers.JsonRpcProvider;
  ethGasLimit: BigNumber;
  ethGasPrice: string | null;
  maxPriorityFeePerGas: BigNumber | null;
  maxFeePerGas: BigNumber | null;
}

interface SignMessage {
  message: string;
}
interface EthPersonalSign extends SignMessage {
  type: 'personal_sign';
}

interface CosmosSignDirect extends SignMessage {
  type: 'cosmos_signDirect';
}

interface CosmosSignAmino extends SignMessage {
  type: 'cosmos_signAmino';
}

interface CosmosSendTokens {
  type: 'cosmos_sendTokens';
  signingStargateClient: SigningStargateClient;
  cosmosFee: StdFee;
  sendMsg: MsgSendEncodeObject;
  memo: string;
}

interface CosmosSendTransaction {
  type: 'cosmos_sendTransaction';
  LaconicClient: LaconicClient;
  cosmosFee: StdFee;
  txMsg: EncodeObject;
}

export type WalletConnectRequests =
  | EthSendTransaction
  | EthPersonalSign
  | CosmosSignDirect
  | CosmosSignAmino
  | CosmosSendTokens
  | CosmosSendTransaction;

export async function approveWalletConnectRequest(
  requestEvent: SignClientTypes.EventArguments['session_request'],
  account: Account,
  namespace: string,
  chainId: string,
  options: WalletConnectRequests,
) {
  const { params, id } = requestEvent;
  const { request } = params;

  const path = (await getPathKey(`${namespace}:${chainId}`, account.index))
    .path;
  const mnemonic = await getMnemonic();
  const cosmosAccount = await getCosmosAccounts(mnemonic, path);
  const address = account.address;

  switch (request.method) {
    case EIP155_SIGNING_METHODS.ETH_SEND_TRANSACTION:
      if (!(options.type === 'eth_sendTransaction')) {
        throw new Error('Incorrect parameters passed');
      }

      const privKey = (
        await getPathKey(`${namespace}:${chainId}`, account.index)
      ).privKey;
      const wallet = new Wallet(privKey);
      const sendTransaction = request.params[0];
      const updatedTransaction =
        options.maxFeePerGas && options.maxPriorityFeePerGas
          ? {
              ...sendTransaction,
              gasLimit: options.ethGasLimit,
              maxFeePerGas: options.maxFeePerGas,
              maxPriorityFeePerGas: options.maxPriorityFeePerGas,
              type: 2,
            }
          : {
              ...sendTransaction,
              gasLimit: options.ethGasLimit,
              gasPrice: options.ethGasPrice,
              type: 0,
            };

      const connectedWallet = wallet.connect(options.provider);

      const hash = await connectedWallet.sendTransaction(updatedTransaction);
      const receipt = typeof hash === 'string' ? hash : hash?.hash;
      return formatJsonRpcResult(id, {
        signature: receipt,
      });

    case EIP155_SIGNING_METHODS.PERSONAL_SIGN:
      if (!(options.type === 'personal_sign')) {
        throw new Error('Incorrect parameters passed');
      }

      const ethSignature = await signEthMessage(
        options.message,
        account.index,
        chainId,
      );
      return formatJsonRpcResult(id, ethSignature);

    case COSMOS_METHODS.COSMOS_SIGN_DIRECT:
      // Reference: https://github.com/confio/cosmjs-types/blob/66e52711914fccd2a9d1a03e392d3628fdf499e2/src/cosmos/tx/v1beta1/tx.ts#L51
      // According above doc, in the signDoc interface 'bodyBytes' and 'authInfoBytes' have Uint8Array type
      if (!(options.type === 'cosmos_signDirect')) {
        throw new Error('Incorrect parameters passed');
      }

      const bodyBytesArray = Uint8Array.from(
        Buffer.from(request.params.signDoc.bodyBytes, 'hex'),
      );
      const authInfoBytesArray = Uint8Array.from(
        Buffer.from(request.params.signDoc.authInfoBytes, 'hex'),
      );

      const cosmosDirectSignature = await signDirectMessage(
        `${namespace}:${chainId}`,
        account.index,
        {
          ...request.params.signDoc,
          bodyBytes: bodyBytesArray,
          authInfoBytes: authInfoBytesArray,
        },
      );

      return formatJsonRpcResult(id, {
        signature: cosmosDirectSignature,
      });

    case COSMOS_METHODS.COSMOS_SIGN_AMINO:
      if (!(options.type === 'cosmos_signAmino')) {
        throw new Error('Incorrect parameters passed');
      }

      const cosmosAminoSignature = await cosmosAccount.cosmosWallet.signAmino(
        address,
        request.params.signDoc,
      );

      if (!cosmosAminoSignature) {
        throw new Error('Error signing message');
      }

      return formatJsonRpcResult(id, {
        signature: cosmosAminoSignature.signature.signature,
      });

    case COSMOS_METHODS.COSMOS_SEND_TOKENS:
      if (!(options.type === 'cosmos_sendTokens')) {
        throw new Error('Incorrect parameters passed');
      }

      const result = await options.signingStargateClient.signAndBroadcast(
        address,
        [options.sendMsg],
        options.cosmosFee,
        options.memo,
      );

      return formatJsonRpcResult(id, {
        signature: result.transactionHash,
      });

    case COSMOS_METHODS.COSMOS_SEND_TRANSACTION:
      if (!(options.type === 'cosmos_sendTransaction')) {
        throw new Error('Incorrect parameters passed');
      }
      const resultFromTx = await options.LaconicClient.signAndBroadcast(
        address,
        [options.txMsg],
        options.cosmosFee,
      );

      return formatJsonRpcResult(id, {
        code: resultFromTx.code,
      });

    default:
      throw new Error(getSdkError('INVALID_METHOD').message);
  }
}

export function rejectWalletConnectRequest(
  request: SignClientTypes.EventArguments['session_request'],
) {
  const { id } = request;

  return formatJsonRpcError(id, getSdkError('USER_REJECTED_METHODS').message);
}
