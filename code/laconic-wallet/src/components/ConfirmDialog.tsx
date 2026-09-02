import React from 'react';
import { Portal, Dialog, Button, Text } from 'react-native-paper';
import { ResetDialogProps } from '../types';

const ConfirmDialog = ({
  title,
  visible,
  hideDialog,
  onConfirm,
}: ResetDialogProps) => {
  return (
    <Portal>
      <Dialog visible={visible} onDismiss={hideDialog}>
        <Dialog.Title>{title}</Dialog.Title>
        <Dialog.Content>
          <Text variant="bodyMedium">Are you sure?</Text>
        </Dialog.Content>
        <Dialog.Actions>
          <Button textColor="red" onPress={onConfirm}>
            Yes
          </Button>
          <Button onPress={hideDialog}>No</Button>
        </Dialog.Actions>
      </Dialog>
    </Portal>
  );
};

export default ConfirmDialog;
