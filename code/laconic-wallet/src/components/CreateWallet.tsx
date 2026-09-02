import { View } from 'react-native';
import React from 'react';
import { Button } from 'react-native-paper';

import { CreateWalletProps } from '../types';
import styles from '../styles/stylesheet';

const CreateWallet = ({
  isWalletCreating,
  createWalletHandler,
}: CreateWalletProps) => {
  return (
    <View>
      <View style={styles.createWalletContainer}>
        <Button
          mode="contained"
          loading={isWalletCreating}
          disabled={isWalletCreating}
          onPress={createWalletHandler}>
          {isWalletCreating ? 'Creating' : 'Create Wallet'}
        </Button>
      </View>
    </View>
  );
};

export default CreateWallet;
