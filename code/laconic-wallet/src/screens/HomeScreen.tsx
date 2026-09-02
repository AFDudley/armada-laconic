import React, { useCallback, useEffect, useState } from 'react';
import { View, ActivityIndicator, Image } from 'react-native';
import { Button, Text } from 'react-native-paper';

import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { useNavigation } from '@react-navigation/native';
import { getSdkError } from '@walletconnect/utils';

import { createWallet, resetWallet, retrieveAccounts } from '../utils/accounts';
import { DialogComponent } from '../components/Dialog';
import { NetworkDropdown } from '../components/NetworkDropdown';
import Accounts from '../components/Accounts';
import CreateWallet from '../components/CreateWallet';
import ConfirmDialog from '../components/ConfirmDialog';
import styles from '../styles/stylesheet';
import { useAccounts } from '../context/AccountsContext';
import { useWalletConnect } from '../context/WalletConnectContext';
import { NetworksDataState, StackParamsList } from '../types';
import { useNetworks } from '../context/NetworksContext';

const WCLogo = () => {
  return (
    <Image
      style={styles.walletConnectLogo}
      source={require('../assets/WalletConnect-Icon-Blueberry.png')}
    />
  );
};

const HomeScreen = () => {
  const { accounts, setAccounts, setCurrentIndex } = useAccounts();

  const { networksData, selectedNetwork, setSelectedNetwork, setNetworksData } =
    useNetworks();
  const { web3wallet, setActiveSessions } = useWalletConnect();

  const navigation =
    useNavigation<NativeStackNavigationProp<StackParamsList>>();
  useEffect(() => {
    if (accounts.length > 0) {
      navigation.setOptions({
        // eslint-disable-next-line react/no-unstable-nested-components
        headerRight: () => (
          <Button onPress={() => navigation.navigate('WalletConnect')}>
            {<WCLogo />}
          </Button>
        ),
      });
    } else {
      navigation.setOptions({
        headerRight: undefined,
      });
    }
  }, [navigation, accounts]);

  const [isWalletCreated, setIsWalletCreated] = useState<boolean>(false);
  const [isWalletCreating, setIsWalletCreating] = useState<boolean>(false);
  const [walletDialog, setWalletDialog] = useState<boolean>(false);
  const [resetWalletDialog, setResetWalletDialog] = useState<boolean>(false);
  const [isAccountsFetched, setIsAccountsFetched] = useState<boolean>(false);
  const [phrase, setPhrase] = useState('');

  const hideWalletDialog = () => setWalletDialog(false);
  const hideResetDialog = () => setResetWalletDialog(false);

  const fetchAccounts = useCallback(async () => {
    if (!selectedNetwork) {
      return;
    }

    const loadedAccounts = await retrieveAccounts(selectedNetwork);

    if (loadedAccounts) {
      setAccounts(loadedAccounts);
      setIsWalletCreated(true);
    }

    setIsAccountsFetched(true);
  }, [selectedNetwork, setAccounts]);

  const createWalletHandler = async () => {
    setIsWalletCreating(true);
    const mnemonic = await createWallet(networksData);
    if (mnemonic) {
      fetchAccounts();
      setWalletDialog(true);
      setPhrase(mnemonic);
      setSelectedNetwork(networksData[0]);
    }
  };

  const confirmResetWallet = useCallback(async () => {
    setIsWalletCreated(false);
    setIsWalletCreating(false);
    setAccounts([]);
    setCurrentIndex(0);
    setNetworksData([]);
    setSelectedNetwork(undefined);
    await resetWallet();
    const sessions = web3wallet!.getActiveSessions();

    Object.keys(sessions).forEach(async sessionId => {
      await web3wallet!.disconnectSession({
        topic: sessionId,
        reason: getSdkError('USER_DISCONNECTED'),
      });
    });
    setActiveSessions({});

    hideResetDialog();
  }, [
    web3wallet,
    setAccounts,
    setActiveSessions,
    setCurrentIndex,
    setNetworksData,
    setSelectedNetwork,
  ]);

  const updateNetwork = (networkData: NetworksDataState) => {
    setSelectedNetwork(networkData);
    setCurrentIndex(0);
  };

  useEffect(() => {
    fetchAccounts();
  }, [networksData, setAccounts, selectedNetwork, fetchAccounts]);

  return (
    <View style={styles.appContainer}>
      {!isAccountsFetched ? (
        <View style={styles.spinnerContainer}>
          <Text style={styles.LoadingText}>Loading...</Text>
          <ActivityIndicator size="large" color="#0000ff" />
        </View>
      ) : isWalletCreated && selectedNetwork ? (
        <>
          <NetworkDropdown updateNetwork={updateNetwork} />
          <View style={styles.accountComponent}>
            <Accounts />
          </View>
          <View style={styles.resetContainer}>
            <Button
              style={styles.resetButton}
              mode="contained"
              buttonColor="#B82B0D"
              onPress={() => {
                setResetWalletDialog(true);
              }}>
              Reset Wallet
            </Button>
          </View>
        </>
      ) : (
        <CreateWallet
          isWalletCreating={isWalletCreating}
          createWalletHandler={createWalletHandler}
        />
      )}
      <DialogComponent
        visible={walletDialog}
        hideDialog={hideWalletDialog}
        contentText={phrase}
      />
      <ConfirmDialog
        title="Reset wallet"
        visible={resetWalletDialog}
        hideDialog={hideResetDialog}
        onConfirm={confirmResetWallet}
      />
    </View>
  );
};

export default HomeScreen;
