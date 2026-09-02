import React from 'react';
import { Portal, Dialog } from 'react-native-paper';

import { HDPathDialogProps } from '../types';
import HDPath from './HDPath';
import { useNetworks } from '../context/NetworksContext';

const HDPathDialog = ({
  visible,
  hideDialog,
  updateAccounts,
  pathCode,
}: HDPathDialogProps) => {
  const { selectedNetwork } = useNetworks();

  return (
    <Portal>
      <Dialog visible={visible} onDismiss={hideDialog}>
        <Dialog.Title>Add account from HD path</Dialog.Title>
        <Dialog.Content>
          <HDPath
            selectedNetwork={selectedNetwork!}
            pathCode={pathCode}
            updateAccounts={updateAccounts}
            hideDialog={hideDialog}
          />
        </Dialog.Content>
      </Dialog>
    </Portal>
  );
};

export default HDPathDialog;
