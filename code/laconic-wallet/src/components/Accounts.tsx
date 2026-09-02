import React, { useEffect, useState } from 'react';
import { ScrollView, TouchableOpacity, View } from 'react-native';
import { Button, List, Text, useTheme } from 'react-native-paper';
import { setInternetCredentials } from 'react-native-keychain';

import { useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';

import { StackParamsList, Account } from '../types';
import { addAccount } from '../utils/accounts';
import styles from '../styles/stylesheet';
import HDPathDialog from './HDPathDialog';
import AccountDetails from './AccountDetails';
import { useAccounts } from '../context/AccountsContext';
import { useNetworks } from '../context/NetworksContext';
import ConfirmDialog from './ConfirmDialog';
import { getNamespaces } from '../utils/wallet-connect/helpers';
import ShowPKDialog from './ShowPKDialog';
import { useWalletConnect } from '../context/WalletConnectContext';

const Accounts = () => {
  const navigation =
    useNavigation<NativeStackNavigationProp<StackParamsList>>();

  const { accounts, setAccounts, setCurrentIndex, currentIndex } =
    useAccounts();
  const { networksData, selectedNetwork, setNetworksData, setSelectedNetwork } =
    useNetworks();

  const { web3wallet } = useWalletConnect();
  const [expanded, setExpanded] = useState(false);
  const [isAccountCreating, setIsAccountCreating] = useState(false);
  const [hdDialog, setHdDialog] = useState(false);
  const [pathCode, setPathCode] = useState('');
  const [deleteNetworkDialog, setDeleteNetworkDialog] =
    useState<boolean>(false);

  const theme = useTheme();

  const handlePress = () => setExpanded(!expanded);

  const hideDeleteNetworkDialog = () => setDeleteNetworkDialog(false);

  const updateAccounts = (account: Account) => {
    setAccounts([...accounts, account]);
  };

  useEffect(() => {
    const updateSessions = async () => {
      const sessions = (web3wallet && web3wallet.getActiveSessions()) || {};
      // Iterate through each session

      for (const topic in sessions) {
        const session = sessions[topic];
        const { optionalNamespaces, requiredNamespaces } = session;

        const updatedNamespaces = await getNamespaces(
          optionalNamespaces,
          requiredNamespaces,
          networksData,
          selectedNetwork!,
          accounts,
        );

        if (!updatedNamespaces) {
          return;
        }

        await web3wallet!.updateSession({
          topic,
          namespaces: updatedNamespaces,
        });
      }
    };
    // Call the updateSessions function when the 'accounts' dependency changes
    updateSessions();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [accounts, web3wallet]);

  const addAccountHandler = async () => {
    setIsAccountCreating(true);
    const newAccount = await addAccount(selectedNetwork!);
    setIsAccountCreating(false);
    if (newAccount) {
      updateAccounts(newAccount);
      setCurrentIndex(newAccount.index);
    }
  };

  const renderAccountItems = () =>
    accounts.map(account => (
      <List.Item
        key={account.index}
        title={`Account ${account.index + 1}`}
        onPress={() => {
          setCurrentIndex(account.index);
          setExpanded(false);
        }}
      />
    ));

  const handleRemove = async () => {
    const updatedNetworks = networksData.filter(
      networkData => selectedNetwork!.networkId !== networkData.networkId,
    );

    await setInternetCredentials(
      'networks',
      '_',
      JSON.stringify(updatedNetworks),
    );

    setSelectedNetwork(updatedNetworks[0]);
    setCurrentIndex(0);
    setDeleteNetworkDialog(false);
    setNetworksData(updatedNetworks);
  };

  return (
    <ScrollView>
      <View>
        <HDPathDialog
          visible={hdDialog}
          hideDialog={() => setHdDialog(false)}
          updateAccounts={updateAccounts}
          pathCode={pathCode}
        />
        <List.Accordion
          title={`Account ${currentIndex + 1}`}
          expanded={expanded}
          onPress={handlePress}>
          {renderAccountItems()}
        </List.Accordion>

        <View style={styles.addAccountButton}>
          <Button
            mode="contained"
            onPress={addAccountHandler}
            loading={isAccountCreating}
            disabled={isAccountCreating}>
            {isAccountCreating ? 'Adding' : 'Add Account'}
          </Button>
        </View>

        <View style={styles.addAccountButton}>
          <Button
            mode="contained"
            onPress={() => {
              setHdDialog(true);
              setPathCode(`m/44'/${selectedNetwork!.coinType}'/`);
            }}>
            Add Account from HD path
          </Button>
        </View>

        <AccountDetails account={accounts[currentIndex]} />

        <View style={styles.signLink}>
          <TouchableOpacity
            onPress={() => {
              navigation.navigate('SignMessage', {
                selectedNamespace: selectedNetwork!.namespace,
                selectedChainId: selectedNetwork!.chainId,
                accountInfo: accounts[currentIndex],
              });
            }}>
            <Text
              variant="titleSmall"
              style={[styles.hyperlink, { color: theme.colors.primary }]}>
              Sign Message
            </Text>
          </TouchableOpacity>
        </View>

        <View style={styles.signLink}>
          <TouchableOpacity
            onPress={() => {
              navigation.navigate('AddNetwork');
            }}>
            <Text
              variant="titleSmall"
              style={[styles.hyperlink, { color: theme.colors.primary }]}>
              Add Network
            </Text>
          </TouchableOpacity>
        </View>

        <View style={styles.signLink}>
          <TouchableOpacity
            onPress={() => {
              navigation.navigate('EditNetwork', {
                selectedNetwork: selectedNetwork!,
              });
            }}>
            <Text
              variant="titleSmall"
              style={[styles.hyperlink, { color: theme.colors.primary }]}>
              Edit Network
            </Text>
          </TouchableOpacity>
        </View>

        {!selectedNetwork!.isDefault && (
          <View style={styles.signLink}>
            <TouchableOpacity
              onPress={() => {
                setDeleteNetworkDialog(true);
              }}>
              <Text
                variant="titleSmall"
                style={[styles.hyperlink, { color: theme.colors.primary }]}>
                Delete Network
              </Text>
            </TouchableOpacity>
          </View>
        )}
        <ConfirmDialog
          title="Delete Network"
          visible={deleteNetworkDialog}
          hideDialog={hideDeleteNetworkDialog}
          onConfirm={handleRemove}
        />

        <ShowPKDialog />
      </View>
    </ScrollView>
  );
};

export default Accounts;
