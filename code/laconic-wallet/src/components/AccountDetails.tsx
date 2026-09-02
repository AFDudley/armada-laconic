import React from 'react';
import { View } from 'react-native';
import { Text } from 'react-native-paper';

import { Account } from '../types';
import styles from '../styles/stylesheet';

interface AccountDetailsProps {
  account: Account | undefined;
}

const AccountDetails: React.FC<AccountDetailsProps> = ({ account }) => {
  return (
    <View style={styles.accountContainer}>
      <Text variant="bodyLarge" selectable={true}>
        <Text style={styles.highlight}>Address: </Text>
        {account?.address}
      </Text>
      <Text variant="bodyLarge" selectable={true}>
        <Text style={styles.highlight}>Public Key: </Text>
        {account?.pubKey}
      </Text>
      <Text variant="bodyLarge">
        <Text style={styles.highlight}>HD Path: </Text>
        {account?.hdPath}
      </Text>
    </View>
  );
};

export default AccountDetails;
