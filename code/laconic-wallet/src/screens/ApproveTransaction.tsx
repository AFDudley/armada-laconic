import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Image, ScrollView, View } from 'react-native';
import { Button, Text, TextInput } from 'react-native-paper';
import { SvgUri } from 'react-native-svg';
import Config from 'react-native-config';
import { MsgCreateValidator } from 'cosmjs-types/cosmos/staking/v1beta1/tx';

import {
  NativeStackNavigationProp,
  NativeStackScreenProps,
} from '@react-navigation/native-stack';
import { useNavigation } from '@react-navigation/native';
import { DirectSecp256k1Wallet, EncodeObject } from '@cosmjs/proto-signing';
import { LaconicClient } from '@cerc-io/registry-sdk';
import { GasPrice, calculateFee } from '@cosmjs/stargate';
import { formatJsonRpcError } from '@json-rpc-tools/utils';

import { useNetworks } from '../context/NetworksContext';
import { Account, StackParamsList } from '../types';
import styles from '../styles/stylesheet';
import { COSMOS, IS_NUMBER_REGEX } from '../utils/constants';
import { retrieveSingleAccount } from '../utils/accounts';
import { getPathKey } from '../utils/misc';
import {
  WalletConnectRequests,
  approveWalletConnectRequest,
  rejectWalletConnectRequest,
} from '../utils/wallet-connect/wallet-connect-requests';
import { MEMO } from './ApproveTransfer';
import TxErrorDialog from '../components/TxErrorDialog';
import AccountDetails from '../components/AccountDetails';
import { useWalletConnect } from '../context/WalletConnectContext';

type ApproveTransactionProps = NativeStackScreenProps<
  StackParamsList,
  'ApproveTransaction'
>;

const ApproveTransaction = ({ route }: ApproveTransactionProps) => {
  const { networksData } = useNetworks();

  const requestSession = route.params.requestSessionData;
  const requestName = requestSession.peer.metadata.name;
  const requestIcon = requestSession.peer.metadata.icons[0];
  const requestURL = requestSession.peer.metadata.url;
  const signer = route.params.signer;
  const requestEvent = route.params.requestEvent;
  const chainId = requestEvent.params.chainId;
  const requestEventId = requestEvent.id;
  const topic = requestEvent.topic;

  const { web3wallet } = useWalletConnect();

  const [account, setAccount] = useState<Account>();
  const [cosmosStargateClient, setCosmosStargateClient] =
    useState<LaconicClient>();
  const [cosmosGasLimit, setCosmosGasLimit] = useState<string>();
  const [fees, setFees] = useState<string>();
  const [txError, setTxError] = useState<string>();
  const [isTxErrorDialogOpen, setIsTxErrorDialogOpen] = useState(false);
  const [isRequestAccepted, setIsRequestAccepted] = useState(false);

  const navigation =
    useNavigation<NativeStackNavigationProp<StackParamsList>>();

  const requestedNetwork = networksData.find(
    networkData =>
      `${networkData.namespace}:${networkData.chainId}` === chainId,
  );
  const namespace = requestedNetwork!.namespace;

  const transactionMessage = useMemo((): EncodeObject => {
    const inputTxMsg = route.params.transactionMessage;

    // If it's a MsgCreateValidator, decode the tx msg value using MsgCreateValidator type
    if (inputTxMsg.typeUrl.includes('MsgCreateValidator')) {
      return {
        typeUrl: inputTxMsg.typeUrl,
        value: MsgCreateValidator.fromJSON(inputTxMsg.value),
      };
    }

    return inputTxMsg;
  }, [route.params.transactionMessage]);

  useEffect(() => {
    if (namespace !== COSMOS) {
      return;
    }

    const setClient = async () => {
      if (!account) {
        return;
      }

      const cosmosPrivKey = (
        await getPathKey(
          `${requestedNetwork?.namespace}:${requestedNetwork?.chainId}`,
          account.index,
        )
      ).privKey;

      const sender = await DirectSecp256k1Wallet.fromKey(
        Buffer.from(cosmosPrivKey.split('0x')[1], 'hex'),
        requestedNetwork?.addressPrefix,
      );

      try {
        const client = await LaconicClient.connectWithSigner(
          requestedNetwork?.rpcUrl!,
          sender,
        );
        setCosmosStargateClient(client);
      } catch (error: any) {
        setTxError(error.message);
        setIsTxErrorDialogOpen(true);
        const response = formatJsonRpcError(requestEventId, error.message);
        await web3wallet!.respondSessionRequest({ topic, response });
      }
    };

    setClient();
  }, [
    account,
    requestedNetwork,
    chainId,
    namespace,
    requestEventId,
    topic,
    web3wallet,
  ]);

  const retrieveData = useCallback(
    async (requestAddress: string) => {
      const requestAccount = await retrieveSingleAccount(
        requestedNetwork!.namespace,
        requestedNetwork!.chainId,
        requestAddress,
      );
      if (!requestAccount) {
        navigation.navigate('InvalidPath');
        return;
      }

      setAccount(requestAccount);
    },
    [navigation, requestedNetwork],
  );

  useEffect(() => {
    retrieveData(signer);
  }, [retrieveData, signer]);

  useEffect(() => {
    const getCosmosGas = async () => {
      try {
        if (!cosmosStargateClient) {
          return;
        }
        const gasEstimation = await cosmosStargateClient!.simulate(
          signer,
          [transactionMessage],
          MEMO,
        );

        setCosmosGasLimit(
          String(
            Math.round(gasEstimation * Number(Config.DEFAULT_GAS_ADJUSTMENT)),
          ),
        );
      } catch (error: any) {
        setTxError(error.message);
        setIsTxErrorDialogOpen(true);
        const response = formatJsonRpcError(requestEventId, error.message);
        await web3wallet!.respondSessionRequest({ topic, response });
      }
    };
    getCosmosGas();
  }, [
    cosmosStargateClient,
    transactionMessage,
    requestEventId,
    topic,
    web3wallet,
    signer,
  ]);

  useEffect(() => {
    const gasPrice = GasPrice.fromString(
      requestedNetwork?.gasPrice! + requestedNetwork?.nativeDenom,
    );

    if (!cosmosGasLimit) {
      return;
    }

    const cosmosFees = calculateFee(Number(cosmosGasLimit), gasPrice);

    setFees(cosmosFees.amount[0].amount);
  }, [namespace, cosmosGasLimit, requestedNetwork]);

  const acceptRequestHandler = async () => {
    try {
      setIsRequestAccepted(true);
      if (!account) {
        throw new Error('account not found');
      }

      let options: WalletConnectRequests;

      if (!cosmosStargateClient) {
        throw new Error('Cosmos stargate client not found');
      }

      options = {
        type: 'cosmos_sendTransaction',
        LaconicClient: cosmosStargateClient,
        // StdFee object
        cosmosFee: {
          // This amount is total fees required for transaction
          amount: [
            {
              amount: fees!,
              denom: requestedNetwork!.nativeDenom!,
            },
          ],
          gas: cosmosGasLimit!,
        },
        txMsg: transactionMessage,
      };

      const response = await approveWalletConnectRequest(
        requestEvent,
        account,
        namespace,
        requestedNetwork!.chainId,
        options,
      );

      await web3wallet!.respondSessionRequest({ topic, response });
      setIsRequestAccepted(false);
      navigation.navigate('Laconic');
    } catch (error: any) {
      setTxError(error.message);
      setIsTxErrorDialogOpen(true);
      const response = formatJsonRpcError(requestEventId, error.message);
      await web3wallet!.respondSessionRequest({ topic, response });
    }
  };

  const rejectRequestHandler = async () => {
    const response = rejectWalletConnectRequest(requestEvent);
    await web3wallet!.respondSessionRequest({
      topic,
      response,
    });

    navigation.navigate('Laconic');
  };

  const replacer = (key: string, value: any): any => {
    if (value instanceof Uint8Array) {
      return Buffer.from(value).toString('hex');
    }
    return value;
  };

  return (
    <>
      <ScrollView contentContainerStyle={styles.approveTransaction}>
        <View style={styles.dappDetails}>
          {requestIcon && (
            <>
              {requestIcon.endsWith('.svg') ? (
                <View style={styles.dappLogo}>
                  <SvgUri height="50" width="50" uri={requestIcon} />
                </View>
              ) : (
                <Image style={styles.dappLogo} source={{ uri: requestIcon }} />
              )}
            </>
          )}
          <Text>{requestName}</Text>
          <Text variant="bodySmall">{requestURL}</Text>
        </View>
        <AccountDetails account={account} />
        <Text variant="bodyLarge" style={styles.transactionLabel}>
          Message:
        </Text>
        <View style={styles.messageBody}>
          <Text variant="bodyLarge">
            {JSON.stringify(transactionMessage, replacer, 2)}
          </Text>
        </View>
        <>
          <Text variant="bodyLarge" style={styles.transactionLabel}>
            Gas Limit:
          </Text>
          <TextInput
            mode="outlined"
            style={styles.transactionFeesInput}
            value={cosmosGasLimit}
            onChangeText={value => {
              if (IS_NUMBER_REGEX.test(value)) {
                setCosmosGasLimit(value);
              }
            }}
          />
          <View style={styles.buttonContainer}>
            <Button
              mode="contained"
              onPress={acceptRequestHandler}
              loading={isRequestAccepted}
              disabled={isRequestAccepted}>
              Yes
            </Button>
            <Button
              mode="contained"
              onPress={rejectRequestHandler}
              buttonColor="#B82B0D">
              No
            </Button>
          </View>
        </>
      </ScrollView>
      <TxErrorDialog
        error={txError!}
        visible={isTxErrorDialogOpen}
        hideDialog={async () => {
          setIsTxErrorDialogOpen(false);
          navigation.navigate('Laconic');
        }}
      />
    </>
  );
};

export default ApproveTransaction;
