import React from 'react';
import { Dialog, DialogTitle, DialogContent } from '@mui/material';

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
    <Dialog open={visible} onClose={hideDialog}>
      <DialogTitle>Add account from HD path</DialogTitle>
      <DialogContent>
        <HDPath
          selectedNetwork={selectedNetwork!}
          pathCode={pathCode}
          updateAccounts={updateAccounts}
          hideDialog={hideDialog}
        />
      </DialogContent>
    </Dialog>
  );
};

export default HDPathDialog;
