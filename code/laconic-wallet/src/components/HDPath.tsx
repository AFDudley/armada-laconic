import React, { useState } from 'react';
import { ScrollView, View, Text } from 'react-native';
import { Button, TextInput } from 'react-native-paper';

import { addAccountFromHDPath } from '../utils/accounts';
import { Account, NetworksDataState, PathState } from '../types';
import styles from '../styles/stylesheet';
import { useAccounts } from '../context/AccountsContext';

const HDPath = ({
  pathCode,
  updateAccounts,
  hideDialog,
  selectedNetwork,
}: {
  pathCode: string;
  updateAccounts: (account: Account) => void;
  hideDialog: () => void;
  selectedNetwork: NetworksDataState;
}) => {
  const { setCurrentIndex } = useAccounts();
  const [isAccountCreating, setIsAccountCreating] = useState(false);
  const [path, setPath] = useState<PathState>({
    firstNumber: '',
    secondNumber: '',
    thirdNumber: '',
  });

  const handleChange = (key: keyof PathState, value: string) => {
    if (key === 'secondNumber' && value !== '' && !['0', '1'].includes(value)) {
      return;
    }

    setPath({
      ...path,
      [key]: value.replace(/[^0-9]/g, ''),
    });
  };

  const createFromHDPathHandler = async () => {
    setIsAccountCreating(true);
    const hdPath =
      pathCode +
      `${path.firstNumber}'/${path.secondNumber}/${path.thirdNumber}`;
    try {
      const newAccount = await addAccountFromHDPath(hdPath, selectedNetwork);
      if (newAccount) {
        updateAccounts(newAccount);
        setCurrentIndex(newAccount.index);
        hideDialog();
      }
    } catch (error) {
      console.error('Error creating account:', error);
    } finally {
      setIsAccountCreating(false);
    }
  };

  return (
    <ScrollView style={styles.HDcontainer}>
      <View style={styles.HDrowContainer}>
        <Text style={styles.HDtext}>{pathCode}</Text>
        <TextInput
          keyboardType="numeric"
          mode="outlined"
          onChangeText={text => handleChange('firstNumber', text)}
          value={path.firstNumber}
          style={styles.HDtextInput}
        />
        <Text style={styles.HDtext}>'/</Text>
        <TextInput
          keyboardType="numeric"
          mode="outlined"
          onChangeText={text => handleChange('secondNumber', text)}
          value={path.secondNumber}
          style={styles.HDtextInput}
        />
        <Text style={styles.HDtext}>/</Text>
        <TextInput
          keyboardType="numeric"
          mode="outlined"
          onChangeText={text => handleChange('thirdNumber', text)}
          value={path.thirdNumber}
          style={styles.HDtextInput}
        />
      </View>
      <View style={styles.HDbuttonContainer}>
        <Button
          mode="contained"
          onPress={createFromHDPathHandler}
          loading={isAccountCreating}
          disabled={isAccountCreating}>
          {isAccountCreating ? 'Adding' : 'Add Account'}
        </Button>
      </View>
    </ScrollView>
  );
};

export default HDPath;
