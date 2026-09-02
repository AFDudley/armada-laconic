import React from 'react';
import { Dialog, DialogTitle, DialogContent, DialogActions, Button, Typography } from '@mui/material';
import { ResetDialogProps } from '../types';
import styles from "../styles/stylesheet";

const ConfirmDialog = ({
  title,
  visible,
  hideDialog,
  onConfirm,
}: ResetDialogProps) => {
  return (
    <Dialog open={visible} onClose={hideDialog}>
      <DialogTitle style={{...styles.resetDialogTitle}} >{title}</DialogTitle>
      <DialogContent style={{...styles.resetDialogContent}}>
        <Typography>Are you sure?</Typography>
      </DialogContent>
      <DialogActions style={{...styles.resetDialogActionRow}}>
        <Button style={{...styles.buttonRed, ...styles.button}} color="error" onClick={onConfirm}>
          Yes
        </Button>
        <Button style={{...styles.buttonBlue, ...styles.button}} onClick={hideDialog}>No</Button>
      </DialogActions>
    </Dialog>
  );
};

export default ConfirmDialog;
