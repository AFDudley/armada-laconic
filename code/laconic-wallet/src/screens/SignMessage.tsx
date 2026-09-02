import React, { useState } from 'react';
import { ScrollView, View, Alert } from 'react-native';
import { Button, Text, TextInput } from 'react-native-paper';

import { NativeStackScreenProps } from '@react-navigation/native-stack';

import { StackParamsList } from '../types';
import styles from '../styles/stylesheet';
import { signMessage } from '../utils/sign-message';
import AccountDetails from '../components/AccountDetails';

type SignProps = NativeStackScreenProps<StackParamsList, 'SignMessage'>;

const SignMessage = ({ route }: SignProps) => {
  const namespace = route.params.selectedNamespace;
  const chainId = route.params.selectedChainId;
  const account = route.params.accountInfo;

  const [message, setMessage] = useState<string>('');

  const signMessageHandler = async () => {
    const signedMessage = await signMessage({
      message,
      namespace,
      chainId,
      accountId: account.index,
    });
    Alert.alert('Signature', signedMessage);
  };

  return (
    <ScrollView style={styles.signPage}>
      <View style={styles.accountInfo}>
        <View>
          <Text variant="titleMedium">
            {account && `Account ${account.index + 1}`}
          </Text>
        </View>
        <View style={styles.accountContainer}>
          <AccountDetails account={account} />
        </View>
      </View>

      <TextInput
        mode="outlined"
        placeholder="Enter your message"
        onChangeText={text => setMessage(text)}
        value={message}
      />

      <View style={styles.signButton}>
        <Button mode="contained" onPress={signMessageHandler}>
          Sign
        </Button>
      </View>
    </ScrollView>
  );
};

export default SignMessage;
