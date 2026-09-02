import React from 'react';
import { Button, Dialog, Portal, Text } from 'react-native-paper';

const TxErrorDialog = ({
  error,
  visible,
  hideDialog,
}: {
  error: string;
  visible: boolean;
  hideDialog: () => void;
}) => {
  return (
    <Portal>
      <Dialog visible={visible} onDismiss={hideDialog}>
        <Dialog.Title>Transaction Error</Dialog.Title>
        <Dialog.Content>
          <Text variant="bodyMedium">{error}</Text>
        </Dialog.Content>
        <Dialog.Actions>
          <Button onPress={hideDialog}>OK</Button>
        </Dialog.Actions>
      </Dialog>
    </Portal>
  );
};

export default TxErrorDialog;
