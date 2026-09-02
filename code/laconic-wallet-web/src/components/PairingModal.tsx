import React, { useEffect, useMemo, useState } from 'react';
import { Image, View, Modal, ScrollView } from 'react-native';
import { Button, Text } from 'react-native-paper';
import mergeWith from 'lodash/mergeWith';

import { buildApprovedNamespaces, getSdkError } from '@walletconnect/utils';

import { PairingModalProps } from '../types';
import styles from '../styles/stylesheet';
import { useAccounts } from '../context/AccountsContext';
import { useWalletConnect } from '../context/WalletConnectContext';
import { useNetworks } from '../context/NetworksContext';
import { getNamespaces } from '../utils/wallet-connect/helpers';

const PairingModal = ({
  visible,
  currentProposal,
  setCurrentProposal,
  setModalVisible,
  setToastVisible,
}: PairingModalProps) => {
  const { accounts } = useAccounts();
  const { selectedNetwork, networksData } = useNetworks();
  const [isLoading, setIsLoading] = useState(false);
  const [chainError, setChainError] = useState('');

  const dappName = currentProposal?.params?.proposer?.metadata.name;
  const url = currentProposal?.params?.proposer?.metadata.url;
  const icon = currentProposal?.params.proposer?.metadata.icons[0];

  const [walletConnectData, setWalletConnectData] = useState<{
    walletConnectMethods: string[];
    walletConnectEvents: string[];
    walletConnectChains: string[];
  }>({
    walletConnectMethods: [],
    walletConnectEvents: [],
    walletConnectChains: [],
  });

  const [supportedNamespaces, setSupportedNamespaces] = useState<
    Record<
      string,
      {
        chains: string[];
        methods: string[];
        events: string[];
        accounts: string[];
      }
    >
  >();

  useEffect(() => {
    if (!currentProposal) {
      return;
    }
    const { params } = currentProposal;
    const { requiredNamespaces, optionalNamespaces } = params;

    setWalletConnectData({
      walletConnectMethods: [],
      walletConnectEvents: [],
      walletConnectChains: [],
    });

    const combinedNamespaces = mergeWith(
      requiredNamespaces,
      optionalNamespaces,
      (obj, src) =>
        Array.isArray(obj) && Array.isArray(src) ? [...src, ...obj] : undefined,
    );

    Object.keys(combinedNamespaces).forEach(key => {
      const { methods, events, chains } = combinedNamespaces[key];

      setWalletConnectData(prevData => {
        return {
          walletConnectMethods: [...prevData.walletConnectMethods, ...methods],
          walletConnectEvents: [...prevData.walletConnectEvents, ...events],
          walletConnectChains: chains
            ? [...prevData.walletConnectChains, ...chains]
            : [...prevData.walletConnectChains],
        };
      });
    });
  }, [currentProposal]);

  const { setActiveSessions, web3wallet } = useWalletConnect();

  useEffect(() => {
    const getSupportedNamespaces = async () => {
      if (!currentProposal) {
        return;
      }

      const { optionalNamespaces, requiredNamespaces } = currentProposal.params;
      try {
        const nameSpaces = await getNamespaces(
          optionalNamespaces,
          requiredNamespaces,
          networksData,
          selectedNetwork!,
          accounts,
        );
        setSupportedNamespaces(nameSpaces);
      } catch (err) {
        setChainError((err as Error).message);

        const { id } = currentProposal;
        await web3wallet!.rejectSession({
          id,
          reason: getSdkError('UNSUPPORTED_CHAINS'),
        });
        setCurrentProposal(undefined);
        setWalletConnectData({
          walletConnectMethods: [],
          walletConnectEvents: [],
          walletConnectChains: [],
        });
      }
    };

    getSupportedNamespaces();
  }, [
    currentProposal,
    networksData,
    selectedNetwork,
    accounts,
    setCurrentProposal,
    setModalVisible,
    web3wallet
  ]);

  const namespaces = useMemo(() => {
    return (
      currentProposal &&
      supportedNamespaces
      &&
      buildApprovedNamespaces({
        proposal: currentProposal.params,
        supportedNamespaces,
      })
    );
  }, [currentProposal, supportedNamespaces]);

  const handleAccept = async () => {
    try {
      if (currentProposal && namespaces) {
        setIsLoading(true);
        const { id } = currentProposal;

        await web3wallet!.approveSession({
          id,
          namespaces,
        });

        const sessions = web3wallet!.getActiveSessions();
        setIsLoading(false);
        setActiveSessions(sessions);
        setModalVisible(false);
        setToastVisible(true);
        setCurrentProposal(undefined);
        setSupportedNamespaces(undefined);
        setWalletConnectData({
          walletConnectMethods: [],
          walletConnectEvents: [],
          walletConnectChains: [],
        });
      }
    } catch (error) {
      console.error('Error in approve session:', error);
      throw error;
    }
  };

  const handleClose = () => {
    setChainError('');
    setModalVisible(false);
  };

  const handleReject = async () => {
    if (currentProposal) {
      const { id } = currentProposal;
      await web3wallet!.rejectSession({
        id,
        reason: getSdkError('USER_REJECTED_METHODS'),
      });

      setModalVisible(false);
      setCurrentProposal(undefined);
      setWalletConnectData({
        walletConnectMethods: [],
        walletConnectEvents: [],
        walletConnectChains: [],
      });
    }
  };

  return (
    <View>
      {chainError !== '' ? (
        <Modal visible={visible} animationType="slide" transparent>
          <View style={styles.modalContentContainer}>
            <View style={styles.container}>
              {icon && (
                <>
                  {icon.endsWith('.svg') ? (
                    <View style={styles.dappLogo}>
                      <Text>SvgURI requstIcon</Text>
                    </View>
                  ) : (
                    <Image style={styles.dappLogo} source={{ uri: icon }} />
                  )}
                </>
              )}

              <Text variant="titleMedium">{dappName}</Text>
              <Text variant="bodyMedium">{url}</Text>
              <Text variant="titleMedium">{chainError}</Text>
            </View>
            <View style={styles.flexRow}>
              <Button mode="outlined" onPress={handleClose}>
                Close
              </Button>
            </View>
          </View>
        </Modal>
      ) : (
        <Modal visible={visible} animationType="slide" transparent>
          <View style={styles.modalOuterContainer}>
            <View style={styles.modalContentContainer}>
              <ScrollView showsVerticalScrollIndicator={true}>
                <View style={styles.container}>
                  {icon && (
                    <>
                      {icon.endsWith('.svg') ? (
                        <View style={styles.dappLogo}>
                          <Text>SvgURI requstIcon</Text>
                        </View>
                      ) : (
                        <Image style={styles.dappLogo} source={{ uri: icon }} />
                      )}
                    </>
                  )}

                  <Text variant="titleMedium">{dappName}</Text>
                  <Text variant="bodyMedium">{url}</Text>
                  <View style={styles.marginVertical8} />
                  <Text variant="titleMedium">Connect to this site?</Text>

                  {walletConnectData.walletConnectMethods.length > 0 && (
                    <View>
                      <Text variant="titleMedium">Chains:</Text>
                      {walletConnectData.walletConnectChains.map(chain => (
                        <Text style={styles.centerText} key={chain}>
                          {chain}
                        </Text>
                      ))}
                    </View>
                  )}

                  {walletConnectData.walletConnectMethods.length > 0 && (
                    <View style={styles.marginVertical8}>
                      <Text variant="titleMedium">Methods Requested:</Text>
                      {walletConnectData.walletConnectMethods.map(method => (
                        <Text style={styles.centerText} key={method}>
                          {method}
                        </Text>
                      ))}
                    </View>
                  )}

                  {walletConnectData.walletConnectEvents.length > 0 && (
                    <View style={styles.marginVertical8}>
                      <Text variant="titleMedium">Events Requested:</Text>
                      {walletConnectData.walletConnectEvents.map(event => (
                        <Text style={styles.centerText} key={event}>
                          {event}
                        </Text>
                      ))}
                    </View>
                  )}
                </View>
              </ScrollView>

              <View style={styles.flexRow}>
                <Button
                  mode="contained"
                  onPress={handleAccept}
                  loading={isLoading}
                  disabled={isLoading}>
                  {isLoading ? 'Connecting' : 'Yes'}
                </Button>
                <View style={styles.space} />
                <Button mode="outlined" onPress={handleReject}>
                  No
                </Button>
              </View>
            </View>
          </View>
        </Modal>
      )}
    </View>
  );
};

export default PairingModal;
