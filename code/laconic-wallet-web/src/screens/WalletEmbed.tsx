import React, { useEffect, useState, useCallback, useRef } from 'react';
import { ScrollView, View } from 'react-native';
import {
  ActivityIndicator,
  Button,
  Text,
  TextInput,
} from 'react-native-paper';
import { BigNumber } from 'ethers';

import { DirectSecp256k1Wallet } from '@cosmjs/proto-signing';
import {
  calculateFee,
  GasPrice,
  SigningStargateClient,
} from '@cosmjs/stargate';

import { retrieveSingleAccount } from '../utils/accounts';
import AccountDetails from '../components/AccountDetails';
import styles from '../styles/stylesheet';
import DataBox from '../components/DataBox';
import { checkSufficientFunds, getPathKey, sendMessage } from '../utils/misc';
import { useNetworks } from '../context/NetworksContext';
import TxErrorDialog from '../components/TxErrorDialog';
import { MEMO } from '../screens/ApproveTransfer';
import { Account, NetworksDataState } from '../types';
import useGetOrCreateAccounts from '../hooks/useGetOrCreateAccounts';
import useAccountsData from '../hooks/useAccountsData';
import { REQUEST_TX, REQUEST_WALLET_ACCOUNTS, TRANSACTION_RESPONSE, WALLET_ACCOUNTS_DATA } from '../utils/constants';

type TransactionDetails = {
  chainId: string;
  fromAddress: string;
  toAddress: string;
  amount: string;
  account: Account
  balance: string;
  requestedNetwork: NetworksDataState
};

export const WalletEmbed = () => {
  const [isTxRequested, setIsTxRequested] = useState<boolean>(false);
  const [transactionDetails, setTransactionDetails] = useState<TransactionDetails | null>(null);
  const [fees, setFees] = useState<string>('');
  const [gasLimit, setGasLimit] = useState<string>('');
  const [isTxLoading, setIsTxLoading] = useState(false);
  const [txError, setTxError] = useState<string | null>(null);
  const txEventRef = useRef<MessageEvent | null>(null);

  const { networksData } = useNetworks();
  const { getAccountsData } = useAccountsData();

  useEffect(() => {
    const handleGetAccounts = async (event: MessageEvent) => {
      if (event.data.type !== REQUEST_WALLET_ACCOUNTS) return;

      const accountsData = await getAccountsData(event.data.chainId);

      if (accountsData.length === 0) {
        sendMessage(event.source as Window, 'ERROR', 'Wallet accounts not found', event.origin);
        return;
      }

      sendMessage(
        event.source as Window,
        WALLET_ACCOUNTS_DATA,
        accountsData.map(account => account.address),
        event.origin
      );
    };

    window.addEventListener('message', handleGetAccounts);

    return () => {
      window.removeEventListener('message', handleGetAccounts);
    };
  }, [getAccountsData]);

  // Custom hook for adding listener to get accounts data
  useGetOrCreateAccounts();

  const handleTxRequested = useCallback(
    async (event: MessageEvent) => {
      try {
        if (event.data.type !== REQUEST_TX) return;

        txEventRef.current = event;

        const { chainId, fromAddress, toAddress, amount } = event.data;
        const network = networksData.find(net => net.chainId === chainId);

        if (!network) {
          console.error('Network not found');
          throw new Error('Requested network not supported.');
        }

        const account = await retrieveSingleAccount(network.namespace, network.chainId, fromAddress);
        if (!account) {
          throw new Error('Account not found for the requested address.');
        }

        const cosmosPrivKey = (
          await getPathKey(`${network.namespace}:${chainId}`, account.index)
        ).privKey;

        const sender = await DirectSecp256k1Wallet.fromKey(
          Buffer.from(cosmosPrivKey.split('0x')[1], 'hex'),
          network.addressPrefix
        );

        const client = await SigningStargateClient.connectWithSigner(network.rpcUrl!, sender);

        const balance = await client.getBalance(
          account.address,
          network.nativeDenom!.toLowerCase()
        );

        const sendMsg = {
          typeUrl: '/cosmos.bank.v1beta1.MsgSend',
          value: {
            fromAddress: fromAddress,
            toAddress: toAddress,
            amount: [
              {
                amount: String(amount),
                denom: network.nativeDenom!,
              },
            ],
          },
        };

        setTransactionDetails({
          chainId,
          fromAddress,
          toAddress,
          amount,
          account,
          balance: balance.amount,
          requestedNetwork: network,
        });

        if (!checkSufficientFunds(amount, balance.amount)) {
          throw new Error('Insufficient funds');
        }

        const gasEstimation = await client.simulate(fromAddress, [sendMsg], MEMO);
        const gasLimit = String(
          Math.round(gasEstimation * Number(import.meta.env.REACT_APP_GAS_ADJUSTMENT))
        );
        setGasLimit(gasLimit);

        const gasPrice = GasPrice.fromString(`${network.gasPrice}${network.nativeDenom}`);
        const cosmosFees = calculateFee(Number(gasLimit), gasPrice);
        setFees(cosmosFees.amount[0].amount);

        setIsTxRequested(true);
      } catch (error) {
        if (!(error instanceof Error)) {
          throw error;
        }
        setTxError(error.message);
      }
    }, [networksData]);

  useEffect(() => {
    window.addEventListener('message', handleTxRequested);
    return () => window.removeEventListener('message', handleTxRequested);
  }, [handleTxRequested]);

  const acceptRequestHandler = async () => {
    try {
      setIsTxLoading(true);
      if (!transactionDetails) {
        throw new Error('Tx details not set');
      }
      const balanceBigNum = BigNumber.from(transactionDetails.balance);
      const amountBigNum = BigNumber.from(String(transactionDetails.amount));
      if (amountBigNum.gte(balanceBigNum)) {
        throw new Error('Insufficient funds');
      }

      const cosmosPrivKey = (
        await getPathKey(`${transactionDetails.requestedNetwork.namespace}:${transactionDetails.chainId}`, transactionDetails.account.index)
      ).privKey;

      const sender = await DirectSecp256k1Wallet.fromKey(
        Buffer.from(cosmosPrivKey.split('0x')[1], 'hex'),
        transactionDetails.requestedNetwork.addressPrefix
      );

      const client = await SigningStargateClient.connectWithSigner(
        transactionDetails.requestedNetwork.rpcUrl!,
        sender
      );

      const fee = calculateFee(
        Number(gasLimit),
        GasPrice.fromString(`${transactionDetails.requestedNetwork.gasPrice}${transactionDetails.requestedNetwork.nativeDenom}`)
      );

      const txResult = await client.sendTokens(
        transactionDetails.fromAddress,
        transactionDetails.toAddress,
        [{ amount: String(transactionDetails.amount), denom: transactionDetails.requestedNetwork.nativeDenom! }],
        fee
      );

      const event = txEventRef.current;
      if (event?.source) {
        sendMessage(event.source as Window, TRANSACTION_RESPONSE, txResult.transactionHash, event.origin);
      } else {
        console.error('No event source available to send message');
      }
    } catch (error) {
      if (!(error instanceof Error)) {
        throw error;
      }
      setTxError(error.message);
    } finally {
      setIsTxLoading(false);
    }
  };

  const rejectRequestHandler = () => {
    const event = txEventRef.current;

    setIsTxRequested(false);
    setTransactionDetails(null);
    if (event?.source) {
      sendMessage(event.source as Window, TRANSACTION_RESPONSE, null, event.origin);
    } else {
      console.error('No event source available to send message');
    }
  };

  return (
    <>
      {isTxRequested && transactionDetails ? (
        <>
          <ScrollView contentContainerStyle={styles.appContainer}>
            <View style={styles.dataBoxContainer}>
              <Text style={styles.dataBoxLabel}>From</Text>
              <View style={styles.dataBox}>
                <AccountDetails account={transactionDetails.account} />
              </View>
            </View>
            <DataBox
              label={`Balance (${transactionDetails.requestedNetwork.nativeDenom})`}
              data={
                transactionDetails.balance === '' ||
                  transactionDetails.balance === undefined
                  ? 'Loading balance...'
                  : `${transactionDetails.balance}`
              }
            />
            <View style={styles.approveTransfer}>
              <DataBox label="To" data={transactionDetails.toAddress} />
              <DataBox
                label={`Amount (${transactionDetails.requestedNetwork.nativeDenom})`}
                data={transactionDetails.amount}
              />
              <TextInput
                mode="outlined"
                label="Fee"
                value={fees}
                onChangeText={setFees}
                style={styles.transactionFeesInput}
              />
              <TextInput
                mode="outlined"
                label="Gas Limit"
                value={gasLimit}
                onChangeText={value =>
                  /^\d+$/.test(value) ? setGasLimit(value) : null
                }
              />
            </View>
          </ScrollView>
          <View style={styles.buttonContainer}>
            <Button
              mode="contained"
              onPress={acceptRequestHandler}
              loading={isTxLoading}
              disabled={!transactionDetails.balance || !fees || isTxLoading}
            >
              {isTxLoading ? 'Processing' : 'Yes'}
            </Button>
            <Button
              mode="contained"
              onPress={rejectRequestHandler}
              buttonColor="#B82B0D"
              disabled={isTxLoading}
            >
              No
            </Button>
          </View>
        </>
      ) : (
        <View style={styles.spinnerContainer}>
          <View style={{ marginTop: 50 }}></View>
          <ActivityIndicator size="large" color="#0000ff" />
        </View>
      )}
      <TxErrorDialog
        error={txError!}
        visible={!!txError}
        hideDialog={() => {
          setTxError(null)
          if (window.parent) {
            sendMessage(window.parent, TRANSACTION_RESPONSE, null, '*');
            sendMessage(window.parent, 'closeIframe', null, '*');
          }
        }}
      />
    </>
  );
};
