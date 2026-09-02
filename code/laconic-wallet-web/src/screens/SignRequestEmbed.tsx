import React, { useCallback, useEffect, useState } from 'react';
import { ScrollView, View } from 'react-native';
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
import { getCosmosAccountByHDPath, retrieveSingleAccount } from '../utils/accounts';
import { getMnemonic, getPathKey, sendMessage } from '../utils/misc';
import { COSMOS, REQUEST_SIGN_MESSAGE, SIGN_MESSAGE_RESPONSE } from '../utils/constants';
import { useNetworks } from '../context/NetworksContext';

const REACT_APP_ALLOWED_URLS = import.meta.env.REACT_APP_ALLOWED_URLS;

type SignRequestProps = NativeStackScreenProps<StackParamsList, 'sign-message-request-embed'>;

const SignRequestEmbed = ({ route }: SignRequestProps) => {
  const [displayAccount, setDisplayAccount] = useState<Account>();
  const [message, setMessage] = useState<string>('');
  const [chainId, setChainId] = useState<string>('');
  const [namespace, setNamespace] = useState<string>('');
  const [signDoc, setSignDoc] = useState<any>(null);
  const [signerAddress, setSignerAddress] = useState<string>('');
  const [origin, setOrigin] = useState<string>('');
  const [sourceWindow, setSourceWindow] = useState<Window | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isApproving, setIsApproving] = useState(false);

  const { networksData } = useNetworks();
  const navigation =
    useNavigation<NativeStackNavigationProp<StackParamsList>>();

  const signMessageHandler = async () => {
    if (!signDoc || !signerAddress || !sourceWindow) return;

    setIsApproving(true);
    try {
      if (namespace !== COSMOS) {
        // TODO: Support ETH namespace
        throw new Error(`namespace ${namespace} is not supported`)
      }

      const requestAccount = await retrieveSingleAccount(namespace, chainId, signerAddress);
      const path = (await getPathKey(`${namespace}:${chainId}`, requestAccount!.index)).path;
      const mnemonic = await getMnemonic();

      const requestedNetworkData = networksData.find(networkData => networkData.chainId === chainId)
      if (!requestedNetworkData) {
        throw new Error("Requested network not found")
      }

      const cosmosAccount = await getCosmosAccountByHDPath(mnemonic, path, requestedNetworkData?.addressPrefix);

      const cosmosAminoSignature = await cosmosAccount.cosmosWallet.signAmino(
        signerAddress,
        signDoc,
      );

      const signature = cosmosAminoSignature.signature.signature;

      sendMessage(
        sourceWindow,
        SIGN_MESSAGE_RESPONSE,
        { signature },
        origin,
      );

      navigation.navigate('Home');
    } catch (err) {
      console.error('Signing failed:', err);
      sendMessage(
        sourceWindow!,
        SIGN_MESSAGE_RESPONSE,
        { error: err },
        origin,
      );
    } finally {
      setIsApproving(false);
    }
  };

  const rejectRequestHandler = useCallback(async () => {
    if (sourceWindow && origin) {
      sendMessage(
        sourceWindow,
        SIGN_MESSAGE_RESPONSE,
        { error: 'User rejected the request' },
        origin,
      );
    }
  }, [sourceWindow, origin]);

  useEffect(() => {
    const handleCosmosSignMessage = async (event: MessageEvent) => {
      if (event.data.type !== REQUEST_SIGN_MESSAGE) return;


      if (!REACT_APP_ALLOWED_URLS) {
        console.log('Allowed URLs are not set');
        return;
      }

      const allowedUrls = REACT_APP_ALLOWED_URLS.split(',').map(url => url.trim());

      if (!allowedUrls.includes(event.origin)) {
        console.log('Unauthorized app.');
        return;
      }

      try {
        const { signerAddress, signDoc } = event.data.params;

        const receivedNamespace = event.data.chainId.split(':')[0]
        const receivedChainId = event.data.chainId.split(':')[1]

        if (receivedNamespace !== COSMOS) {
          // TODO: Support ETH namespace
          throw new Error(`namespace ${receivedNamespace} is not supported`)
        }

        setSignerAddress(signerAddress);
        setSignDoc(signDoc);
        setMessage(signDoc.memo || '');
        setOrigin(event.origin);
        setSourceWindow(event.source as Window);
        setNamespace(receivedNamespace);
        setChainId(receivedChainId);

        const requestAccount = await retrieveSingleAccount(
          receivedNamespace,
          receivedChainId,
          signerAddress,
        );

        setDisplayAccount(requestAccount);
        setIsLoading(false);
      } catch (err) {
        console.error('Error preparing sign request:', err);
        setIsLoading(false);
      }
    };

    window.addEventListener('message', handleCosmosSignMessage);
    return () => window.removeEventListener('message', handleCosmosSignMessage);
  }, []);

  useEffect(() => {
    navigation.setOptions({
      // eslint-disable-next-line react/no-unstable-nested-components
      header: ({ options, back }) => {
        const title = getHeaderTitle(options, 'Sign Message');

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
  }, [navigation, rejectRequestHandler]);

  return (
    <>
      {isLoading ? (
        <View style={styles.spinnerContainer}>
          <ActivityIndicator size="large" color="#0000ff" />
        </View>
      ) : (
        <>
          <ScrollView contentContainerStyle={styles.appContainer}>
            <AccountDetails account={displayAccount} />
            <View style={styles.requestMessage}>
              <Text variant="bodyLarge">{message}</Text>
            </View>
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
              buttonColor="#B82B0D">
              No
            </Button>
          </View>
        </>
      )}
    </>
  );
};

export default SignRequestEmbed;
