import React from 'react';
import { Dialog, DialogTitle, DialogContent, DialogActions, Button, Typography } from '@mui/material';

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
    <Dialog open={visible} onClose={hideDialog}>
      <DialogTitle>Transaction Error</DialogTitle>
      <DialogContent>
        <Typography variant="body1">{error}</Typography>
      </DialogContent>
      <DialogActions>
        <Button onClick={hideDialog}>OK</Button>
      </DialogActions>
    </Dialog>
  );
};

export default TxErrorDialog;
