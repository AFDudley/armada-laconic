import React from "react";
import {
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Button,
  Typography,
} from "@mui/material";

import styles from "../styles/stylesheet";
import GridView from "./Grid";
import { CustomDialogProps } from "../types";

const DialogComponent = ({
  visible,
  hideDialog,
  contentText,
}: CustomDialogProps) => {
  const words = contentText.split(" ");

  const copyMnemonic = () => {
    navigator.clipboard.writeText(contentText);
  };

  return (
    <Dialog open={visible} onClose={hideDialog}>
      <DialogTitle style={{...styles.mnemonicTitle}}>Mnemonic</DialogTitle>
      <DialogContent style={{...styles.mnemonicContainer}}>
        <Typography component="div">
          Your mnemonic provides full access to your wallet and funds.<br/>
          Make sure to note it down.
        </Typography>
        <Typography component="div" style={{ ...styles.mnemonicDialogWarning }}>
          Do not share your mnemonic with anyone
        </Typography>
        <GridView words={words} />
      </DialogContent>
      <DialogActions style={{...styles.mnemonicButtonRow}}>
        <Button style={{...styles.mnemonicButton}} onClick={copyMnemonic}>Copy</Button>
        <Button style={{...styles.mnemonicButton}} onClick={hideDialog}>Done</Button>
      </DialogActions>
    </Dialog>
  );
};

export { DialogComponent };
