import React, { useEffect } from 'react';
import { Image, TouchableOpacity, View } from 'react-native';
import { List, Text } from 'react-native-paper';
import { SvgUri } from 'react-native-svg';

import { getSdkError } from '@walletconnect/utils';

import { useWalletConnect } from '../context/WalletConnectContext';
import styles from '../styles/stylesheet';

export default function WalletConnect() {
  const { web3wallet, activeSessions, setActiveSessions } = useWalletConnect();

  const disconnect = async (sessionId: string) => {
    await web3wallet!.disconnectSession({
      topic: sessionId,
      reason: getSdkError('USER_DISCONNECTED'),
    });
    const sessions = web3wallet?.getActiveSessions() || {};
    setActiveSessions(sessions);
    return;
  };

  useEffect(() => {
    const sessions = web3wallet?.getActiveSessions() || {};
    setActiveSessions(sessions);
  }, [web3wallet, setActiveSessions]);

  return (
    <View>
      {Object.keys(activeSessions).length > 0 ? (
        <>
          <View style={styles.sessionsContainer}>
            <Text variant="titleMedium">Active Sessions</Text>
          </View>
          <List.Section>
            {Object.entries(activeSessions).map(([sessionId, session]) => (
              <List.Item
                style={styles.sessionItem}
                key={sessionId}
                title={`${session.peer.metadata.name}`}
                descriptionNumberOfLines={7}
                description={`${sessionId} \n\n${session.peer.metadata.url}\n\n${session.peer.metadata.description}`}
                // reference: https://github.com/react-navigation/react-navigation/issues/11371#issuecomment-1546543183
                // eslint-disable-next-line react/no-unstable-nested-components
                left={() => (
                  <>
                    {session.peer.metadata.icons[0].endsWith('.svg') ? (
                      <View style={styles.dappLogo}>
                        <SvgUri
                          height="50"
                          width="50"
                          uri={session.peer.metadata.icons[0]}
                        />
                      </View>
                    ) : (
                      <Image
                        style={styles.dappLogo}
                        source={{ uri: session.peer.metadata.icons[0] }}
                      />
                    )}
                  </>
                )}
                // eslint-disable-next-line react/no-unstable-nested-components
                right={() => (
                  <TouchableOpacity
                    onPress={() => disconnect(sessionId)}
                    style={styles.disconnectSession}>
                    <List.Icon icon="close" />
                  </TouchableOpacity>
                )}
              />
            ))}
          </List.Section>
        </>
      ) : (
        <View style={styles.noActiveSessions}>
          <Text>You have no active sessions</Text>
        </View>
      )}
    </View>
  );
}
