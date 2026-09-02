import React, { useCallback, useEffect, useMemo, useState } from "react";
import { Button, Snackbar, Surface, Text } from "react-native-paper";
import { TxBody, AuthInfo } from "cosmjs-types/cosmos/tx/v1beta1/tx";

import { SignClientTypes } from "@walletconnect/types";
import { useNavigation } from "@react-navigation/native";
import { DirectSecp256k1Wallet } from "@cosmjs/proto-signing";
import { SigningStargateClient } from "@cosmjs/stargate";
import {
  createStackNavigator,
  StackNavigationProp,
} from "@react-navigation/stack";
import { getSdkError } from "@walletconnect/utils";
import { Web3WalletTypes } from "@walletconnect/web3wallet";
import { formatJsonRpcResult } from "@json-rpc-tools/utils";

import PairingModal from "./components/PairingModal";
import { useWalletConnect } from "./context/WalletConnectContext";
import { useAccounts } from "./context/AccountsContext";
import InvalidPath from "./screens/InvalidPath";
import SignMessage from "./screens/SignMessage";
import HomeScreen from "./screens/HomeScreen";
import SignRequest from "./screens/SignRequest";
import AddSession from "./screens/AddSession";
import WalletConnect from "./screens/WalletConnect";
import ApproveTransaction from "./screens/ApproveTransaction";
import { StackParamsList } from "./types";
import { EIP155_SIGNING_METHODS } from "./utils/wallet-connect/EIP155Data";
import { getSignParamsMessage } from "./utils/wallet-connect/helpers";
import ApproveTransfer from "./screens/ApproveTransfer";
import AddNetwork from "./screens/AddNetwork";
import EditNetwork from "./screens/EditNetwork";
import { CHECK_BALANCE, COSMOS, EIP155, IS_SUFFICIENT } from "./utils/constants";
import { useNetworks } from "./context/NetworksContext";
import { NETWORK_METHODS } from "./utils/wallet-connect/common-data";
import { COSMOS_METHODS } from "./utils/wallet-connect/COSMOSData";
import styles from "./styles/stylesheet";
import { Header } from "./components/Header";
import { WalletEmbed } from "./screens/WalletEmbed";
import { AutoSignIn } from "./screens/AutoSignIn";
import { checkSufficientFunds, getPathKey, sendMessage } from "./utils/misc";
import useAccountsData from "./hooks/useAccountsData";
import { useWebViewHandler } from "./hooks/useWebViewHandler";
import SignRequestEmbed from "./screens/SignRequestEmbed";
import useAddAccountEmbed from "./hooks/useAddAccountEmbed";
import useExportPKEmbed from "./hooks/useExportPrivateKeyEmbed";
import { SignTxEmbed } from "./screens/SignTxEmbed";
import useCreateNetwork from "./hooks/useCreateNetwork";

const Stack = createStackNavigator<StackParamsList>();

const App = (): React.JSX.Element => {
  const navigation = useNavigation<StackNavigationProp<StackParamsList>>();

  const { web3wallet, setActiveSessions } = useWalletConnect();
  const { accounts, setCurrentIndex } = useAccounts();
  const { networksData, selectedNetwork, setSelectedNetwork } = useNetworks();
  const { getAccountsData } = useAccountsData();

  const [modalVisible, setModalVisible] = useState(false);
  const [toastVisible, setToastVisible] = useState(false);
  const [currentProposal, setCurrentProposal] = useState<
    SignClientTypes.EventArguments["session_proposal"] | undefined
  >();

  const onSessionProposal = useCallback(
    async (proposal: SignClientTypes.EventArguments["session_proposal"]) => {
      if (!accounts.length || !accounts.length) {
        const { id } = proposal;
        await web3wallet!.rejectSession({
          id,
          reason: getSdkError("UNSUPPORTED_ACCOUNTS"),
        });
        return;
      }
      setModalVisible(true);
      setCurrentProposal(proposal);
    },
    [accounts, web3wallet],
  );

  const onSessionRequest = useCallback(
    async (requestEvent: Web3WalletTypes.SessionRequest) => {
      const { topic, params, id } = requestEvent;
      const { request } = params;

      const requestSessionData =
        web3wallet!.engine.signClient.session.get(topic);
      switch (request.method) {
        case NETWORK_METHODS.GET_NETWORKS:
          const currentNetworkId = networksData.find(
            (networkData) =>
              networkData.networkId === selectedNetwork!.networkId,
          )?.networkId;

          const networkNamesData = networksData.map((networkData) => {
            return {
              id: networkData.networkId,
              name: networkData.networkName,
            };
          });

          const formattedResponse = formatJsonRpcResult(id, {
            currentNetworkId,
            networkNamesData,
          });

          await web3wallet!.respondSessionRequest({
            topic,
            response: formattedResponse,
          });
          break;

        case NETWORK_METHODS.CHANGE_NETWORK:
          const networkNameData = request.params[0];
          const network = networksData.find(
            (networkData) => networkData.networkId === networkNameData.id,
          );
          setCurrentIndex(0);
          setSelectedNetwork(network);

          const response = formatJsonRpcResult(id, {
            response: "true",
          });

          await web3wallet!.respondSessionRequest({
            topic,
            response: response,
          });
          break;

        case EIP155_SIGNING_METHODS.ETH_SEND_TRANSACTION:
          navigation.navigate("ApproveTransfer", {
            transaction: request.params[0],
            requestEvent,
            requestSessionData,
          });
          break;

        case EIP155_SIGNING_METHODS.PERSONAL_SIGN:
          navigation.navigate("SignRequest", {
            namespace: EIP155,
            address: request.params[1],
            message: getSignParamsMessage(request.params),
            requestEvent,
            requestSessionData,
          });
          break;

        case COSMOS_METHODS.COSMOS_SIGN_DIRECT:
          const message = {
            txbody: TxBody.toJSON(
              TxBody.decode(
                Uint8Array.from(
                  Buffer.from(request.params.signDoc.bodyBytes, "hex"),
                ),
              ),
            ),
            authInfo: AuthInfo.toJSON(
              AuthInfo.decode(
                Uint8Array.from(
                  Buffer.from(request.params.signDoc.authInfoBytes, "hex"),
                ),
              ),
            ),
          };
          navigation.navigate("SignRequest", {
            namespace: COSMOS,
            address: request.params.signerAddress,
            message: JSON.stringify(message, undefined, 2),
            requestEvent,
            requestSessionData,
          });
          break;

        case COSMOS_METHODS.COSMOS_SIGN_AMINO:
          navigation.navigate("SignRequest", {
            namespace: COSMOS,
            address: request.params.signerAddress,
            message: request.params.signDoc.memo,
            requestEvent,
            requestSessionData,
          });
          break;

        case COSMOS_METHODS.COSMOS_SEND_TOKENS:
          navigation.navigate("ApproveTransfer", {
            transaction: request.params[0],
            requestEvent,
            requestSessionData,
          });
          break;

        case COSMOS_METHODS.COSMOS_SEND_TRANSACTION:
          const { transactionMessage, signer } = request.params;
          navigation.navigate("ApproveTransaction", {
            transactionMessage,
            signer,
            requestEvent,
            requestSessionData,
          });
          break;

        default:
          throw new Error("Invalid method");
      }
    },
    [
      navigation,
      networksData,
      setSelectedNetwork,
      setCurrentIndex,
      selectedNetwork,
      web3wallet,
    ],
  );

  const onSessionDelete = useCallback(() => {
    const sessions = web3wallet!.getActiveSessions();
    setActiveSessions(sessions);
  }, [setActiveSessions, web3wallet]);

  useEffect(() => {
    web3wallet?.on("session_proposal", onSessionProposal);
    web3wallet?.on("session_request", onSessionRequest);
    web3wallet?.on("session_delete", onSessionDelete);
    return () => {
      web3wallet?.off("session_proposal", onSessionProposal);
      web3wallet?.off("session_request", onSessionRequest);
      web3wallet?.off("session_delete", onSessionDelete);
    };
  });

  useEffect(() => {
    const handleCheckBalance = async (event: MessageEvent) => {
      if (event.data.type !== CHECK_BALANCE) return;

      const { chainId, amount } = event.data;
      const network = networksData.find(net => net.chainId === chainId);

      if (!network) {
        console.error('Network not found');
        throw new Error('Requested network not supported.');
      }

      if (network.namespace !== COSMOS) {
        throw new Error('Unsupported network');
      }

      const accounts = await getAccountsData(chainId);
      const account = accounts[0];

      if (!account) {
        throw new Error(`No accounts in network ${chainId}`);
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

      const areFundsSufficient = checkSufficientFunds(amount, balance.amount);

      sendMessage(event.source as Window, IS_SUFFICIENT, areFundsSufficient, event.origin);
    };

    window.addEventListener('message', handleCheckBalance);

    return () => {
      window.removeEventListener('message', handleCheckBalance);
    };
  }, [networksData, getAccountsData]);

  const showWalletConnect = useMemo(() => accounts.length > 0, [accounts]);

  useWebViewHandler();
  useAddAccountEmbed();
  useExportPKEmbed();
  useCreateNetwork();

  return (
    <Surface style={styles.appSurface}>
      <Stack.Navigator
        screenOptions={{
          headerBackTitleVisible: true,
        }}
      >
        <Stack.Screen
          name="Home"
          component={HomeScreen}
          options={{
            // eslint-disable-next-line react/no-unstable-nested-components
            header: () => (
              <Header title="Wallet" showWalletConnect={showWalletConnect} />
            ),
          }}
        />
        <Stack.Screen
          name="SignMessage"
          component={SignMessage}
          options={{
            // eslint-disable-next-line react/no-unstable-nested-components
            header: () => <Header title="Wallet" />,
          }}
        />
        <Stack.Screen
          name="SignRequest"
          component={SignRequest}
          options={{
            // eslint-disable-next-line react/no-unstable-nested-components
            header: () => <Header title="Wallet" />,
          }}
        />
        <Stack.Screen
          name="InvalidPath"
          component={InvalidPath}
          options={{
            // eslint-disable-next-line react/no-unstable-nested-components
            headerTitle: () => <Text variant="titleLarge">Bad Request</Text>,
          }}
        />
        <Stack.Screen
          name="AddSession"
          component={AddSession}
          options={{
            header: () => <Header title="Wallet" />,
          }}
        />
        <Stack.Screen
          name="WalletConnect"
          component={WalletConnect}
          options={{
            // eslint-disable-next-line react/no-unstable-nested-components
            headerTitle: () => <Text variant="titleLarge">WalletConnect</Text>,
            // eslint-disable-next-line react/no-unstable-nested-components
            headerRight: () => (
              <Button
                onPress={() => {
                  navigation.navigate("AddSession");
                }}
              >
                {<Text>PAIR</Text>}
              </Button>
            ),
          }}
        />
        <Stack.Screen
          name="ApproveTransfer"
          component={ApproveTransfer}
          options={{
            header: () => <Header title="Wallet" />,
          }}
        />
        <Stack.Screen
          name="AddNetwork"
          component={AddNetwork}
          options={{
            header: () => <Header title="Wallet" />,
          }}
        />
        <Stack.Screen
          name="EditNetwork"
          component={EditNetwork}
          options={{
            header: () => <Header title="Wallet" />,
          }}
        />
        <Stack.Screen
          name="ApproveTransaction"
          component={ApproveTransaction}
          options={{
            header: () => <Header title="Wallet" />,
          }}
        />
        <Stack.Screen
          name="wallet-embed"
          component={WalletEmbed}
          options={{
            header: () => <></>,
          }}
        />
        <Stack.Screen
          name="sign-tx-request-embed"
          component={SignTxEmbed}
          options={{
            header: () => <></>,
          }}
        />
        <Stack.Screen
          name="auto-sign-in"
          component={AutoSignIn}
          options={{
            header: () => <></>,
          }}
        />
        <Stack.Screen
          name="sign-message-request-embed"
          component={SignRequestEmbed}
          options={{
            header: () => <Header title="Wallet" />,
          }}
        />
      </Stack.Navigator>
      <PairingModal
        visible={modalVisible}
        setModalVisible={setModalVisible}
        currentProposal={currentProposal}
        setCurrentProposal={setCurrentProposal}
        setToastVisible={setToastVisible}
      />
      <Snackbar
        visible={toastVisible}
        onDismiss={() => setToastVisible(false)}
        duration={3000}
      >
        Session approved
      </Snackbar>
    </Surface>
  );
};

export default App;
