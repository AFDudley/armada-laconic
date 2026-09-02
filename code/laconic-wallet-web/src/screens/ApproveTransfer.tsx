import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Image, ScrollView, View } from 'react-native';
import {
  ActivityIndicator,
  Button,
  Text,
  Appbar,
  TextInput,
} from 'react-native-paper';
import { providers, BigNumber } from 'ethers';
import { Deferrable } from 'ethers/lib/utils';

import { useNavigation } from '@react-navigation/native';
import {
  NativeStackNavigationProp,
  NativeStackScreenProps,
} from '@react-navigation/native-stack';
import { getHeaderTitle } from '@react-navigation/elements';
import { DirectSecp256k1Wallet } from '@cosmjs/proto-signing';
import {
  calculateFee,
  GasPrice,
  MsgSendEncodeObject,
  SigningStargateClient,
} from '@cosmjs/stargate';

import { Account, StackParamsList } from '../types';
import AccountDetails from '../components/AccountDetails';
import styles from '../styles/stylesheet';
import { retrieveSingleAccount } from '../utils/accounts';
import {
  approveWalletConnectRequest,
  rejectWalletConnectRequest,
  WalletConnectRequests,
} from '../utils/wallet-connect/wallet-connect-requests';
import { useWalletConnect } from '../context/WalletConnectContext';
import DataBox from '../components/DataBox';
import { getPathKey } from '../utils/misc';
import { useNetworks } from '../context/NetworksContext';
import { COSMOS, EIP155, IS_NUMBER_REGEX } from '../utils/constants';
import TxErrorDialog from '../components/TxErrorDialog';
import { EIP155_SIGNING_METHODS } from '../utils/wallet-connect/EIP155Data';
import { COSMOS_METHODS } from '../utils/wallet-connect/COSMOSData';

export const MEMO = 'Sending signed tx from Laconic Wallet';
// Reference: https://ethereum.org/en/developers/docs/gas/#what-is-gas-limit
const ETH_MINIMUM_GAS = 21000;

type SignRequestProps = NativeStackScreenProps<
  StackParamsList,
  'ApproveTransfer'
>;

const ApproveTransfer = ({ route }: SignRequestProps) => {
  const { networksData } = useNetworks();
  const { web3wallet } = useWalletConnect();

  const requestSession = route.params.requestSessionData;
  const requestName = requestSession.peer.metadata.name;
  const requestIcon = requestSession.peer.metadata.icons[0];
  const requestURL = requestSession.peer.metadata.url;
  const transaction = route.params.transaction;
  const requestEvent = route.params.requestEvent;
  const chainId = requestEvent.params.chainId;
  const requestMethod = requestEvent.params.request.method;

  const [account, setAccount] = useState<Account>();
  const [isLoading, setIsLoading] = useState(true);
  const [balance, setBalance] = useState<string>('');
  const [isTxLoading, setIsTxLoading] = useState(false);
  const [cosmosStargateClient, setCosmosStargateClient] =
    useState<SigningStargateClient>();
  const [fees, setFees] = useState<string>('');
  const [cosmosGasLimit, setCosmosGasLimit] = useState<string>('');
  const [txError, setTxError] = useState<string>();
  const [isTxErrorDialogOpen, setIsTxErrorDialogOpen] = useState(false);
  const [ethGasPrice, setEthGasPrice] = useState<BigNumber | null>();
  const [ethGasLimit, setEthGasLimit] = useState<BigNumber>();
  const [ethMaxFee, setEthMaxFee] = useState<BigNumber | null>();
  const [ethMaxPriorityFee, setEthMaxPriorityFee] =
    useState<BigNumber | null>();

  const isSufficientFunds = useMemo(() => {
    if (!transaction.value) {
      return;
    }

    if (!balance) {
      return;
    }

    const amountBigNum = BigNumber.from(String(transaction.value));
    const balanceBigNum = BigNumber.from(balance);

    if (amountBigNum.gte(balanceBigNum)) {
      return false;
    } else {
      return true;
    }
  }, [balance, transaction]);

  const requestedNetwork = networksData.find(
    networkData =>
      `${networkData.namespace}:${networkData.chainId}` === chainId,
  );
  const namespace = requestedNetwork!.namespace;

  const sendMsg: MsgSendEncodeObject = useMemo(() => {
    return {
      typeUrl: '/cosmos.bank.v1beta1.MsgSend',
      value: {
        fromAddress: transaction.from,
        toAddress: transaction.to,
        amount: [
          {
            amount: String(transaction.value),
            denom: requestedNetwork!.nativeDenom!,
          },
        ],
      },
    };
  }, [requestedNetwork, transaction]);

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
        const client = await SigningStargateClient.connectWithSigner(
          requestedNetwork?.rpcUrl!,
          sender,
        );

        setCosmosStargateClient(client);
      } catch (error) {
        if (!(error instanceof Error)) {
          throw error;
        }

        setTxError(error.message);
        setIsTxErrorDialogOpen(true);
      }
    };

    setClient();
  }, [account, requestedNetwork, chainId, namespace]);

  const provider = useMemo(() => {
    if (namespace === EIP155) {
      if (!requestedNetwork) {
        throw new Error('Requested chain not supported');
      }
      try {
        const ethProvider = new providers.JsonRpcProvider(
          requestedNetwork.rpcUrl,
        );

        return ethProvider;
      } catch (error) {
        if (!(error instanceof Error)) {
          throw error;
        }

        setTxError(error.message);
        setIsTxErrorDialogOpen(true);
      }
    }
  }, [requestedNetwork, namespace]);

  const navigation =
    useNavigation<NativeStackNavigationProp<StackParamsList>>();

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
    // Set loading to false when gas values for requested chain are fetched
    // If requested chain is EVM compatible, the cosmos gas values will be undefined and vice-versa, hence the condition checks only one of them at the same time
    if (
      // If requested chain is EVM compatible, set loading to false when ethMaxFee and ethPriorityFee have been populated
      (ethMaxFee !== undefined && ethMaxPriorityFee !== undefined) ||
      // Or if requested chain is a cosmos chain, set loading to false when cosmosGasLimit has been populated
      !cosmosGasLimit
    ) {
      setIsLoading(false);
    }
  }, [ethMaxFee, ethMaxPriorityFee, cosmosGasLimit]);

  useEffect(() => {
    if (namespace === EIP155) {
      const ethFees = BigNumber.from(ethGasLimit ?? 0)
        .mul(BigNumber.from(ethMaxFee ?? ethGasPrice ?? 0))
        .toString();
      setFees(ethFees);
    } else {
      const gasPrice = GasPrice.fromString(
        requestedNetwork?.gasPrice! + requestedNetwork?.nativeDenom,
      );

      if (!cosmosGasLimit) {
        return;
      }

      const cosmosFees = calculateFee(Number(cosmosGasLimit), gasPrice);

      setFees(cosmosFees.amount[0].amount);
    }
  }, [
    transaction,
    namespace,
    ethGasLimit,
    ethGasPrice,
    cosmosGasLimit,
    requestedNetwork,
    ethMaxFee,
  ]);
  useEffect(() => {
    retrieveData(transaction.from!);
  }, [retrieveData, transaction]);

  const isEIP1559 = useMemo(() => {
    if (cosmosGasLimit) {
      return;
    }
    if (ethMaxFee !== null && ethMaxPriorityFee !== null) {
      return true;
    }
    return false;
  }, [cosmosGasLimit, ethMaxFee, ethMaxPriorityFee]);

  const acceptRequestHandler = async () => {
    setIsTxLoading(true);
    try {
      if (!account) {
        throw new Error('account not found');
      }

      if (ethGasLimit && ethGasLimit.lt(ETH_MINIMUM_GAS)) {
        throw new Error(`Atleast ${ETH_MINIMUM_GAS} gas limit is required`);
      }

      if (ethMaxFee && ethMaxPriorityFee && ethMaxFee.lte(ethMaxPriorityFee)) {
        throw new Error(
          `Max fee per gas (${ethMaxFee.toNumber()}) cannot be lower than or equal to max priority fee per gas (${ethMaxPriorityFee.toNumber()})`,
        );
      }

      let options: WalletConnectRequests;

      switch (requestMethod) {
        case EIP155_SIGNING_METHODS.ETH_SEND_TRANSACTION:
          if (
            ethMaxFee === undefined ||
            ethMaxPriorityFee === undefined ||
            ethGasPrice === undefined
          ) {
            throw new Error('Gas values not found');
          }

          options = {
            type: 'eth_sendTransaction',
            provider: provider!,
            ethGasLimit: BigNumber.from(ethGasLimit),
            ethGasPrice: ethGasPrice ? ethGasPrice.toHexString() : null,
            maxFeePerGas: ethMaxFee,
            maxPriorityFeePerGas: ethMaxPriorityFee,
          };
          break;
        case COSMOS_METHODS.COSMOS_SEND_TOKENS:
          if (!cosmosStargateClient) {
            throw new Error('Cosmos stargate client not found');
          }

          options = {
            type: 'cosmos_sendTokens',
            signingStargateClient: cosmosStargateClient,
            // StdFee object
            cosmosFee: {
              // This amount is total fees required for transaction
              amount: [
                {
                  amount: fees,
                  denom: requestedNetwork!.nativeDenom!,
                },
              ],
              gas: cosmosGasLimit,
            },
            sendMsg,
            memo: MEMO,
          };

          break;

        default:
          throw new Error('Invalid method');
      }

      const response = await approveWalletConnectRequest(
        requestEvent,
        account,
        namespace,
        requestedNetwork!.chainId,
        options,
      );

      const { topic } = requestEvent;
      await web3wallet!.respondSessionRequest({ topic, response });
      navigation.navigate('Home');
    } catch (error) {
      if (!(error instanceof Error)) {
        throw error;
      }

      setTxError(error.message);
      setIsTxErrorDialogOpen(true);
    }
    setIsTxLoading(false);
  };

  const rejectRequestHandler = async () => {
    const response = rejectWalletConnectRequest(requestEvent);
    const { topic } = requestEvent;
    await web3wallet!.respondSessionRequest({
      topic,
      response,
    });

    navigation.navigate('Home');
  };

  useEffect(() => {
    const getAccountBalance = async () => {
      try {
        if (!account) {
          return;
        }
        if (namespace === EIP155) {
          if (!provider) {
            return;
          }
          const fetchedBalance = await provider.getBalance(account.address);
          setBalance(fetchedBalance ? fetchedBalance.toString() : '0');
        } else {
          const cosmosBalance = await cosmosStargateClient?.getBalance(
            account.address,
            requestedNetwork!.nativeDenom!.toLowerCase(),
          );

          setBalance(cosmosBalance?.amount!);
        }
      } catch (error) {
        if (!(error instanceof Error)) {
          throw error;
        }

        setTxError(error.message);
        setIsTxErrorDialogOpen(true);
      }
    };

    getAccountBalance();
  }, [account, provider, namespace, cosmosStargateClient, requestedNetwork]);

  useEffect(() => {
    navigation.setOptions({
      // eslint-disable-next-line react/no-unstable-nested-components
      header: ({ options, back }) => {
        const title = getHeaderTitle(options, 'Approve Transaction');

        return (
          <Appbar.Header>
            {back && (
              <Appbar.BackAction
                onPress={async () => {
                  await rejectRequestHandler();
                  navigation.navigate('Home');
                }}
              />
            )}
            <Appbar.Content title={title} />
          </Appbar.Header>
        );
      },
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [navigation, route.name]);

  useEffect(() => {
    const getEthGas = async () => {
      try {
        if (!isSufficientFunds || !provider) {
          return;
        }

        const data = await provider.getFeeData();

        setEthMaxFee(data.maxFeePerGas);
        setEthMaxPriorityFee(data.maxPriorityFeePerGas);
        setEthGasPrice(data.gasPrice);

        if (transaction.gasLimit) {
          setEthGasLimit(BigNumber.from(transaction.gasLimit));
        } else {
          const transactionObject: Deferrable<providers.TransactionRequest> = {
            from: transaction.from!,
            to: transaction.to!,
            data: transaction.data!,
            value: transaction.value!,
            maxFeePerGas: data.maxFeePerGas ?? undefined,
            maxPriorityFeePerGas: data.maxPriorityFeePerGas ?? undefined,
            gasPrice: data.maxFeePerGas
              ? undefined
              : data.gasPrice ?? undefined,
          };
          const gasLimit = await provider.estimateGas(transactionObject);
          setEthGasLimit(gasLimit);
        }
      } catch (error) {
        if (!(error instanceof Error)) {
          throw error;
        }

        setTxError(error.message);
        setIsTxErrorDialogOpen(true);
      }
    };
    getEthGas();
  }, [provider, transaction, isSufficientFunds]);

  useEffect(() => {
    const getCosmosGas = async () => {
      try {
        if (!cosmosStargateClient) {
          return;
        }
        if (!isSufficientFunds) {
          return;
        }

        const gasEstimation = await cosmosStargateClient.simulate(
          transaction.from!,
          [sendMsg],
          MEMO,
        );

        setCosmosGasLimit(
          String(
            Math.round(gasEstimation * Number(import.meta.env.REACT_APP_GAS_ADJUSTMENT)),
          ),
        );
      } catch (error) {
        if (!(error instanceof Error)) {
          throw error;
        }

        setTxError(error.message);
        setIsTxErrorDialogOpen(true);
      }
    };
    getCosmosGas();
  }, [cosmosStargateClient, isSufficientFunds, sendMsg, transaction]);

  useEffect(() => {
    if (balance && !isSufficientFunds) {
      setTxError('Insufficient funds');
      setIsTxErrorDialogOpen(true);
    }
  }, [isSufficientFunds, balance]);

  return (
    <>
      {isLoading ? (
        <View style={styles.spinnerContainer}>
          <ActivityIndicator size="large" color="#0000ff" />
        </View>
      ) : (
        <>
          <ScrollView contentContainerStyle={styles.appContainer}>
            <View style={styles.dappDetails}>
              {requestIcon && (
                <Image
                  style={styles.dappLogo}
                  source={requestIcon ? { uri: requestIcon } : undefined}
                />
              )}
              <Text>{requestName}</Text>
              <Text variant="bodyMedium">{requestURL}</Text>
            </View>
            <View style={styles.dataBoxContainer}>
              <Text style={styles.dataBoxLabel}>From</Text>
              <View style={styles.dataBox}>
                <AccountDetails account={account} />
              </View>
            </View>
            <DataBox
              label={`Balance (${
                namespace === EIP155 ? 'wei' : requestedNetwork!.nativeDenom
              })`}
              data={
                balance === '' || balance === undefined
                  ? 'Loading balance...'
                  : `${balance}`
              }
            />
            {transaction && (
              <View style={styles.approveTransfer}>
                <DataBox label="To" data={transaction.to!} />
                <DataBox
                  label={`Amount (${
                    namespace === EIP155 ? 'wei' : requestedNetwork!.nativeDenom
                  })`}
                  data={BigNumber.from(
                    transaction.value?.toString(),
                  ).toString()}
                />

                {namespace === EIP155 ? (
                  <>
                    {isEIP1559 === false ? (
                      <>
                        <Text style={styles.dataBoxLabel}>
                          {'Gas Price (wei)'}
                        </Text>
                        <TextInput
                          mode="outlined"
                          value={ethGasPrice?.toNumber().toString()}
                          onChangeText={value =>
                            setEthGasPrice(BigNumber.from(value))
                          }
                          style={styles.transactionFeesInput}
                        />
                      </>
                    ) : (
                      <>
                        <Text style={styles.dataBoxLabel}>
                          Max Fee Per Gas (wei)
                        </Text>
                        <TextInput
                          mode="outlined"
                          value={ethMaxFee?.toNumber().toString()}
                          onChangeText={value => {
                            if (IS_NUMBER_REGEX.test(value)) {
                              setEthMaxFee(BigNumber.from(value));
                            }
                          }}
                          style={styles.transactionFeesInput}
                        />
                        <Text style={styles.dataBoxLabel}>
                          Max Priority Fee Per Gas (wei)
                        </Text>
                        <TextInput
                          mode="outlined"
                          value={ethMaxPriorityFee?.toNumber().toString()}
                          onChangeText={value => {
                            if (IS_NUMBER_REGEX.test(value)) {
                              setEthMaxPriorityFee(BigNumber.from(value));
                            }
                          }}
                          style={styles.transactionFeesInput}
                        />
                      </>
                    )}
                    <Text style={styles.dataBoxLabel}>Gas Limit</Text>
                    <TextInput
                      mode="outlined"
                      value={ethGasLimit?.toNumber().toString()}
                      onChangeText={value => {
                        if (IS_NUMBER_REGEX.test(value)) {
                          setEthGasLimit(BigNumber.from(value));
                        }
                      }}
                      style={styles.transactionFeesInput}
                    />
                    <DataBox
                      label={`${
                        isEIP1559 === true ? 'Max Fee' : 'Gas Fee'
                      } (wei)`}
                      data={fees!}
                    />
                    <DataBox label="Data" data={transaction.data!} />
                  </>
                ) : (
                  <>
                    <Text style={styles.dataBoxLabel}>{`Fee (${
                      requestedNetwork!.nativeDenom
                    })`}</Text>
                    <TextInput
                      mode="outlined"
                      value={fees}
                      onChangeText={value => setFees(value)}
                      style={styles.transactionFeesInput}
                    />
                    <Text style={styles.dataBoxLabel}>Gas Limit</Text>
                    <TextInput
                      mode="outlined"
                      value={cosmosGasLimit}
                      onChangeText={value => {
                        if (IS_NUMBER_REGEX.test(value)) {
                          setCosmosGasLimit(value);
                        }
                      }}
                    />
                  </>
                )}
              </View>
            )}
          </ScrollView>
          <View style={styles.buttonContainer}>
            <Button
              mode="contained"
              onPress={acceptRequestHandler}
              loading={isTxLoading}
              disabled={!balance || !fees}>
              {isTxLoading ? 'Processing' : 'Yes'}
            </Button>
            <Button
              mode="contained"
              onPress={rejectRequestHandler}
              buttonColor="#B82B0D">
              No
            </Button>
          </View>
        </>
      )}
      <TxErrorDialog
        error={txError!}
        visible={isTxErrorDialogOpen}
        hideDialog={() => {
          setIsTxErrorDialogOpen(false);
          if (!isSufficientFunds || !balance || !fees) {
            rejectRequestHandler();
            navigation.navigate('Home');
          }
        }}
      />
    </>
  );
};

export default ApproveTransfer;
