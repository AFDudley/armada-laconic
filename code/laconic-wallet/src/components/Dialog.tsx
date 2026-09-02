import React, { useState } from 'react';
import { View } from 'react-native';
import { Button, Dialog, Portal, Snackbar, Text } from 'react-native-paper';

import Clipboard from '@react-native-clipboard/clipboard';

import styles from '../styles/stylesheet';
import GridView from './Grid';
import { CustomDialogProps } from '../types';

const DialogComponent = ({
  visible,
  hideDialog,
  contentText,
}: CustomDialogProps) => {
  const [toastVisible, setToastVisible] = useState(false);

  const words = contentText.split(' ');

  const handleCopy = () => {
    Clipboard.setString(contentText);
    setToastVisible(true);
  };

  return (
    <>
      <Portal>
        <Dialog visible={visible} onDismiss={hideDialog}>
          <Dialog.Content>
            <Text variant="titleLarge">Mnemonic</Text>
            <View style={styles.dialogTitle}>
              <Text variant="titleMedium">
                Your mnemonic provides full access to your wallet and funds.
                Make sure to note it down.{' '}
              </Text>
              <Text variant="titleMedium" style={styles.dialogWarning}>
                Do not share your mnemonic with anyone
              </Text>
              <GridView words={words} />
            </View>
          </Dialog.Content>
          <Dialog.Actions>
            <Button onPress={handleCopy}>Copy</Button>
            <Button onPress={hideDialog}>Done</Button>
          </Dialog.Actions>
        </Dialog>
      </Portal>
      <Snackbar
        visible={toastVisible}
        onDismiss={() => setToastVisible(false)}
        duration={1000}>
        Mnemonic copied to clipboard
      </Snackbar>
    </>
  );
};

export { DialogComponent };
