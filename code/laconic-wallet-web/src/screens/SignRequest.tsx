import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Image, ScrollView, View } from 'react-native';
import { ActivityIndicator, Button, Text, Appbar } from 'react-native-paper';

import { useNavigation } from '@react-navigation/native';
import {
  NativeStackNavigationProp,
  NativeStackScreenProps,
} from '@react-navigation/native-stack';
import { getHeaderTitle } from '@react-navigation/elements';

import { Account, StackParamsList } from '../types';
import AccountDetails from '../components/AccountDetails';
import styles from '../styles/stylesheet';
import { signMessage } from '../utils/sign-message';
import { retrieveSingleAccount } from '../utils/accounts';
import {
  approveWalletConnectRequest,
  rejectWalletConnectRequest,
  WalletConnectRequests,
} from '../utils/wallet-connect/wallet-connect-requests';
import { useWalletConnect } from '../context/WalletConnectContext';
import { EIP155_SIGNING_METHODS } from '../utils/wallet-connect/EIP155Data';
import { useNetworks } from '../context/NetworksContext';
import { COSMOS_METHODS } from '../utils/wallet-connect/COSMOSData';

type SignRequestProps = NativeStackScreenProps<StackParamsList, 'SignRequest'>;

const SignRequest = ({ route }: SignRequestProps) => {
  const { networksData } = useNetworks();
  const {web3wallet} = useWalletConnect();

  const requestSession = route.params.requestSessionData;
  const requestName = requestSession?.peer?.metadata?.name;
  const requestIcon = requestSession?.peer?.metadata?.icons[0];
  const requestURL = requestSession?.peer?.metadata?.url;

  const [account, setAccount] = useState<Account>();
  const [message, setMessage] = useState<string>('');
  const [namespace, setNamespace] = useState<string>('');
  const [chainId, setChainId] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);
  const [isApproving, setIsApproving] = useState(false);
  const [isRejecting, setIsRejecting] = useState(false);

  const navigation =
    useNavigation<NativeStackNavigationProp<StackParamsList>>();

  const isCosmosSignDirect = useMemo(() => {
    const requestParams = route.params.requestEvent;

    if (!requestParams?.id) {
      return false;
    }

    return requestParams.params.request.method === 'cosmos_signDirect';
  }, [route.params]);

  const isEthSendTransaction = useMemo(() => {
    const requestParams = route.params.requestEvent;

    if (!requestParams?.id) {
      return false;
    }

    return (
      requestParams.params.request.method ===
      EIP155_SIGNING_METHODS.ETH_SEND_TRANSACTION
    );
  }, [route.params]);

  const retrieveData = useCallback(
    async (
      requestNamespace: string,
      requestChainId: string,
      requestAddress: string,
      requestMessage: string,
    ) => {
      const requestAccount = await retrieveSingleAccount(
        requestNamespace,
        requestChainId,
        requestAddress,
      );
      if (!requestAccount) {
        navigation.navigate('InvalidPath');
        return;
      }

      setAccount(requestAccount);
      setMessage(decodeURIComponent(requestMessage));
      setNamespace(requestNamespace);
      setChainId(requestChainId);
      setIsLoading(false);
    },
    [navigation],
  );

  const sanitizePath = useCallback(
    (path: string) => {
      const regex = /^\/sign\/(eip155|cosmos)\/(.+)\/(.+)\/(.+)$/;
      const match = path.match(regex);
      if (match) {
        const [, pathNamespace, pathChainId, pathAddress, pathMessage] = match;
        return {
          namespace: pathNamespace,
          chainId: pathChainId,
          address: pathAddress,
          message: pathMessage,
        };
      } else {
        navigation.navigate('InvalidPath');
      }
      return null;
    },
    [navigation],
  );

  useEffect(() => {
    if (route.path) {
      const sanitizedRoute = sanitizePath(route.path);
      sanitizedRoute &&
        retrieveData(
          sanitizedRoute.namespace,
          sanitizedRoute.chainId,
          sanitizedRoute.address,
          sanitizedRoute.message,
        );
      return;
    }
    const requestEvent = route.params.requestEvent;
    const requestChainId = requestEvent?.params.chainId;

    const requestedChain = networksData.find(
      networkData => networkData.chainId === requestChainId?.split(':')[1],
    );

    retrieveData(
      requestedChain!.namespace,
      requestedChain!.chainId,
      route.params.address,
      route.params.message,
    );
  }, [retrieveData, sanitizePath, route, networksData]);

  const handleWalletConnectRequest = async () => {
    const { requestEvent } = route.params || {};

    if (!account) {
      throw new Error('account not found');
    }

    if (!requestEvent) {
      throw new Error('Request event not found');
    }

    let options: WalletConnectRequests;

    switch (requestEvent.params.request.method) {
      case COSMOS_METHODS.COSMOS_SIGN_DIRECT:
        options = {
          type: 'cosmos_signDirect',
          message,
        };
        break;
      case COSMOS_METHODS.COSMOS_SIGN_AMINO:
        options = {
          type: 'cosmos_signAmino',
          message,
        };
        break;
      case EIP155_SIGNING_METHODS.PERSONAL_SIGN:
        options = {
          type: 'personal_sign',
          message,
        };
        break;

      default:
        throw new Error('Invalid Method');
    }

    const response = await approveWalletConnectRequest(
      requestEvent,
      account,
      namespace,
      chainId,
      options,
    );

    const { topic } = requestEvent;
    await web3wallet!.respondSessionRequest({ topic, response });
  };

  const handleIntent = async () => {
    if (!account) {
      throw new Error('Account is not valid');
    }
    if (message) {
      const signedMessage = await signMessage({
        message,
        namespace,
        chainId,
        accountId: account.index,
      });
      
      // Send the result back to Android and close dialog
      if (window.Android?.onSignatureComplete) {
        window.Android.onSignatureComplete(signedMessage || "");
      } else {
        alert(`Signature: ${signedMessage}`);
      }
    }
  };

  const signMessageHandler = async () => {
    setIsApproving(true);
    if (route.params.requestEvent) {
      await handleWalletConnectRequest();
    } else {
      await handleIntent();
    }

    setIsApproving(false);
    navigation.navigate('Home');
  };

  const rejectRequestHandler = async () => {
    setIsRejecting(true);
    if (route.params?.requestEvent) {
      const response = rejectWalletConnectRequest(route.params?.requestEvent);
      const { topic } = route.params?.requestEvent;
      await web3wallet!.respondSessionRequest({
        topic,
        response,
      });
    }

    setIsRejecting(false);
    if (window.Android?.onSignatureCancelled) {
      window.Android.onSignatureCancelled();
    } else {
      navigation.navigate('Home');
    }
  };

  useEffect(() => {
    navigation.setOptions({
      // eslint-disable-next-line react/no-unstable-nested-components
      header: ({ options, back }) => {
        const title = getHeaderTitle(options, 'Sign Request');

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
                <>
                  {requestIcon.endsWith('.svg') ? (
                    <View style={styles.dappLogo}>
                      <Text>SvgURI requstIcon</Text>
                    </View>
                  ) : (
                    <Image
                      style={styles.dappLogo}
                      source={{ uri: requestIcon }}
                    />
                  )}
                </>
              )}
              <Text>{requestName}</Text>
              <Text variant="bodyMedium">{requestURL}</Text>
            </View>
            <AccountDetails account={account} />
            {isCosmosSignDirect || isEthSendTransaction ? (
              <View style={styles.requestDirectMessage}>
                <ScrollView nestedScrollEnabled>
                  <Text variant="bodyLarge">{message}</Text>
                </ScrollView>
              </View>
            ) : (
              <View style={styles.requestMessage}>
                <Text variant="bodyLarge">{message}</Text>
              </View>
            )}
          </ScrollView>
          <View style={styles.buttonContainer}>
            <Button
              mode="contained"
              onPress={signMessageHandler}
              loading={isApproving}
              disabled={isApproving}>
              Yes
            </Button>
            <Button
              mode="contained"
              onPress={rejectRequestHandler}
              loading={isRejecting}
              buttonColor="#B82B0D">
              No
            </Button>
          </View>
        </>
      )}
    </>
  );
};

export default SignRequest;
