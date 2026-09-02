import React, { useEffect, useState, useCallback } from 'react';
import { ScrollView, View } from 'react-native';
import {
  Button,
  Text,
} from 'react-native-paper';
import JSONbig from 'json-bigint';
import { AuthInfo, SignDoc } from "cosmjs-types/cosmos/tx/v1beta1/tx";

import { DirectSecp256k1Wallet, Algo, TxBodyEncodeObject, decodeOptionalPubkey } from '@cosmjs/proto-signing';
import { SigningStargateClient } from '@cosmjs/stargate';
import { toHex } from '@cosmjs/encoding';

import { getCosmosAccountByHDPath, retrieveAccounts, retrieveSingleAccount } from '../utils/accounts';
import AccountDetails from '../components/AccountDetails';
import styles from '../styles/stylesheet';
import { getMnemonic, getPathKey, sendMessage } from '../utils/misc';
import { useNetworks } from '../context/NetworksContext';
import TxErrorDialog from '../components/TxErrorDialog';
import { Account, NetworksDataState } from '../types';
import { REQUEST_SIGN_TX, REQUEST_COSMOS_ACCOUNTS, COSMOS_ACCOUNTS_RESPONSE, SIGN_TX_RESPONSE } from '../utils/constants';

// Type Definitions
interface GetAccountsRequestData {
  chainId: string,
}

interface SignTxRequestData {
  address: string;
  signDoc: SignDoc;
  txBody: TxBodyEncodeObject;
}

type IncomingMessageData = SignTxRequestData | GetAccountsRequestData;

interface IncomingMessageEventData {
  id: string;
  type: typeof REQUEST_SIGN_TX | typeof REQUEST_COSMOS_ACCOUNTS;
  data: IncomingMessageData;
}

type TransactionDetails = {
  source: MessageEventSource;
  origin: string;
  signerAddress: string;
  chainId: string;
  account: Account;
  requestedNetwork: NetworksDataState;
  balance: string;
  signDoc: SignDoc;
  txBody: TxBodyEncodeObject;
};

interface GetAccountsResponse {
  accounts: Array<{
    algo: Algo;
    address: string;
    pubkey: string;
  }>;
}

const REACT_APP_ALLOWED_URLS = import.meta.env.REACT_APP_ALLOWED_URLS;

export const SignTxEmbed = () => {
  const [isTxApprovalVisible, setIsTxApprovalVisible] = useState<boolean>(false);
  const [transactionDetails, setTransactionDetails] = useState<TransactionDetails | null>(null);
  const [isTxLoading, setIsTxLoading] = useState(false);
  const [txError, setTxError] = useState<string | null>(null);

  const { networksData } = useNetworks();

  // Message Handlers

  const handleGetCosmosAccountsRequest = useCallback(async (event: MessageEvent<IncomingMessageEventData>) => {
    const { data } = event.data;
    const source = event.source as Window;
    const origin = event.origin;
    const requestData = data as GetAccountsRequestData;
    const mnemonic = await getMnemonic();

    try {
      const requestedNetworkData = networksData.find(networkData => networkData.chainId === requestData.chainId)

      if(!requestedNetworkData) {
        throw new Error("Network data not found")
      }

      const allAccounts = await retrieveAccounts(requestedNetworkData);

      if (!allAccounts || allAccounts.length === 0) {
        throw new Error("Accounts not found for network")
      }

      const responseAccounts = await Promise.all(
        allAccounts.map(async (acc) => {
          const cosmosAccount = (await getCosmosAccountByHDPath(mnemonic, acc.hdPath, requestedNetworkData.addressPrefix)).data;
          return {
            ...cosmosAccount,
            pubkey: toHex(cosmosAccount.pubkey),
          };
        })
      );

      const response: GetAccountsResponse = { accounts: responseAccounts };
      sendMessage(source, COSMOS_ACCOUNTS_RESPONSE, {data: response}, origin);
    } catch (error: unknown) {
      console.error(`Error handling ${REQUEST_COSMOS_ACCOUNTS}:`, error);
      const errorMsg = error instanceof Error ? error.message : String(error);

      sendMessage(source, COSMOS_ACCOUNTS_RESPONSE, { error: `Failed to get accounts: ${errorMsg}` }, origin);
    }
  }, [networksData]);

  const handleSignTxRequest = useCallback(async (event: MessageEvent<IncomingMessageEventData>) => {
    const { data } = event.data;
    const source = event.source as Window;
    const origin = event.origin;
    const requestData = data as SignTxRequestData;

    setIsTxApprovalVisible(false);
    setTransactionDetails(null);
    setTxError(null);

    try {
      const { address: signerAddress, signDoc, txBody } = requestData;

      const network = networksData.find(net => net.chainId === signDoc.chainId);
      if (!network) throw new Error(`Network with chainId "${signDoc.chainId}" not supported.`);

      const account = await retrieveSingleAccount(network.namespace, network.chainId, signerAddress);
      if (!account) throw new Error(`Account not found for address "${signerAddress}" on chain "${signDoc.chainId}".`);

      // Balance Check
      // Use a temporary read-only client for balance
      const { privKey } = await getPathKey(`${network.namespace}:${network.chainId}`, account.index)

      const tempWallet = await DirectSecp256k1Wallet.fromKey(
        new Uint8Array(Buffer.from(privKey.replace(/^0x/, ''), 'hex')),
        network.addressPrefix
      );

      const client = await SigningStargateClient.connectWithSigner(network.rpcUrl!, tempWallet);
      const balance = await client.getBalance(account.address, network.nativeDenom!);
      client.disconnect();

      if (!balance || balance.amount === "0") {
        throw new Error(`${account.address} does not have any balance`)
      }

      setTransactionDetails({
        source: source,
        origin: origin,
        signerAddress,
        chainId: signDoc.chainId,
        account,
        requestedNetwork: network,
        balance: balance.amount,
        signDoc,
        txBody
      });

      setIsTxApprovalVisible(true);

    } catch (error: unknown) {
      console.error(`Error handling ${REQUEST_SIGN_TX}:`, error);
      const errorMsg = error instanceof Error ? error.message : String(error);

      sendMessage(source, SIGN_TX_RESPONSE, { error: `Failed to prepare transaction: ${errorMsg}` }, origin);
      setTxError(errorMsg);
    }
  }, [networksData]);

  const handleIncomingMessage = useCallback((event: MessageEvent) => {
    if (!event.data || typeof event.data !== 'object' || !event.data.type || !event.source || event.source === window) {
      return; // Basic validation
    }

    if (!REACT_APP_ALLOWED_URLS) {
      console.log('Allowed URLs are not set');
      return;
    }

    const allowedUrls = REACT_APP_ALLOWED_URLS.split(',').map(url => url.trim());

    if (!allowedUrls.includes(event.origin)) {
      console.log('Unauthorized app.');
      return;
    }

    const messageData = event.data as IncomingMessageEventData;

    switch (messageData.type) {
      case REQUEST_COSMOS_ACCOUNTS:
        handleGetCosmosAccountsRequest(event as MessageEvent<IncomingMessageEventData>);
        break;
      case REQUEST_SIGN_TX:
        handleSignTxRequest(event as MessageEvent<IncomingMessageEventData>);
        break;
      default:
        console.warn(`Received unknown message type: ${messageData.type}`);
    }
  }, [handleGetCosmosAccountsRequest, handleSignTxRequest]);

  useEffect(() => {
    window.addEventListener('message', handleIncomingMessage);
    return () => {
      window.removeEventListener('message', handleIncomingMessage);
    };
  }, [handleIncomingMessage]);

  // Action Handlers

  const acceptRequestHandler = async () => {
    if (!transactionDetails) {
      setTxError("Transaction details are missing.");
      return;
    }

    setIsTxLoading(true);
    setTxError(null);

    const { source, origin, requestedNetwork, chainId, account, signerAddress, signDoc } = transactionDetails;

    try {
      const { privKey } = await getPathKey(`${requestedNetwork.namespace}:${chainId}`, account.index);

      const privateKeyBytes = Buffer.from(privKey.replace(/^0x/, ''), 'hex');
      const wallet = await DirectSecp256k1Wallet.fromKey(new Uint8Array(privateKeyBytes), requestedNetwork.addressPrefix); // Wrap in Uint8Array

      // Perform the actual signing
      const signResponse = await wallet.signDirect(signerAddress, signDoc);

      sendMessage(source as Window, SIGN_TX_RESPONSE, {data: signResponse}, origin);

      setIsTxApprovalVisible(false);
      setTransactionDetails(null);

    } catch (error: unknown) {
      console.error("Error during signDirect:", error);
      const errorMsg = error instanceof Error ? error.message : String(error);

      setTxError(errorMsg);
      sendMessage(source as Window, SIGN_TX_RESPONSE, { error: `Failed to sign transaction: ${errorMsg}` }, origin);
    } finally {
      setIsTxLoading(false);
    }
  };

  const rejectRequestHandler = () => {
    if (!transactionDetails) return;
    const { source, origin } = transactionDetails;

    sendMessage(source as Window, SIGN_TX_RESPONSE, { error: "User rejected the signature request." }, origin);
    setIsTxApprovalVisible(false);
    setTransactionDetails(null);
    setTxError(null);
  };

  const decodedAuth = React.useMemo(() => {
    if (!transactionDetails) {
      return
    }

    const info = AuthInfo.decode(transactionDetails.signDoc.authInfoBytes);
    return {
      ...info,
      signerInfos: info.signerInfos.map((signerInfo) => ({
        ...signerInfo,
        publicKey: decodeOptionalPubkey(signerInfo.publicKey),
      })),
    };
  }, [transactionDetails]);

  const formattedTxBody = React.useMemo(
    () => {
      if (!transactionDetails) {
        return
      }

      return JSONbig.stringify(transactionDetails.txBody, null, 2)
    },
    [transactionDetails]
  );

  const formattedAuthInfo = React.useMemo(
    () => JSONbig.stringify(decodedAuth, null, 2),
    [decodedAuth]
  );

  const formattedSignDoc = React.useMemo(
    () =>
    {
      if (!transactionDetails) {
        return
      }

      return JSONbig.stringify(
        {
          ...transactionDetails.signDoc,
          bodyBytes: toHex(transactionDetails.signDoc.bodyBytes),
          authInfoBytes: toHex(transactionDetails.signDoc.authInfoBytes),
        },
        null,
        2
      )
    },
    [transactionDetails]
  );

  return (
    <>
      {isTxApprovalVisible && transactionDetails ? (
        <>
          <ScrollView contentContainerStyle={styles.appContainer}>
            <View style={{ marginBottom: 16 }}>
              <Text variant='titleLarge'>Account</Text>
              <View style={styles.dataBox}>
                <AccountDetails account={transactionDetails.account} />
              </View>
            </View>

            <View style={{ marginBottom: 16 }}>
              <Text variant='titleLarge'>{`Balance (${transactionDetails.requestedNetwork.nativeDenom})`}</Text>
              <View style={styles.dataBox}>
                <Text variant='bodyLarge'>
                  {transactionDetails.balance}
                </Text>
              </View>
            </View>

            <View style={{ marginBottom: 16 }}>
              <Text variant='titleLarge'>Transaction Body</Text>
              <ScrollView style={styles.messageBody} contentContainerStyle={{ padding: 10 }}>
                <Text style={styles.codeText}>
                  {formattedTxBody}
                </Text>
              </ScrollView>
            </View>

            <View style={{ marginBottom: 16 }}>
              <Text variant='titleLarge'>Auth Info</Text>
              <ScrollView style={styles.messageBody} contentContainerStyle={{ padding: 10 }}>
                <Text style={styles.codeText}>
                  {formattedAuthInfo}
                </Text>
              </ScrollView>
            </View>

            <View style={{ marginBottom: 16 }}>
              <Text variant='titleLarge'>Transaction Data To Be Signed</Text>
              <ScrollView style={styles.messageBody} contentContainerStyle={{ padding: 10 }}>
                <Text style={styles.codeText}>
                  {formattedSignDoc}
                </Text>
              </ScrollView>
            </View>
          </ScrollView>

          <View style={styles.buttonContainer}>
            <Button
              mode="contained"
              onPress={acceptRequestHandler}
              loading={isTxLoading}
              disabled={isTxLoading}
              style={{marginTop: 10}}
            >
              {isTxLoading ? 'Processing...' : 'Approve'}
            </Button>
            <Button
              mode="contained"
              onPress={rejectRequestHandler}
              buttonColor="#B82B0D"
              disabled={isTxLoading}
              style={{ marginTop: 10 }}
            >
              Reject
            </Button>
          </View>
        </>
      ) : (
        <View style={styles.spinnerContainer}>
          <Text style={{ marginTop: 50, textAlign: 'center' }}>Waiting for request...</Text>
        </View>
      )}
      <TxErrorDialog
        error={txError!}
        visible={!!txError}
        hideDialog={() => setTxError(null)}
      />
    </>
  );
};
