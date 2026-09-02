import React, { useEffect, useState } from "react";
import styles from "../styles/stylesheet";
import {
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  TextField,
  Grid,
  Button,
  Typography,
} from "@mui/material";

const ImportWalletDialog = ({
  visible,
  hideDialog,
  importWalletHandler,
}: {
  visible: boolean;
  hideDialog: () => void;
  importWalletHandler: (recoveryPhrase: string) => Promise<void>;
}) => {
  const [words, setWords] = useState(Array(12).fill(""));

  const handleWordChange = (index: number, value: string) => {
    const newWords = [...words];
    newWords[index] = value;
    setWords(newWords);
  };

  const handlePaste = (event: React.ClipboardEvent<HTMLDivElement>) => {
    const pastedText = event.clipboardData.getData("Text");
    const splitWords = pastedText.trim().split(/\s+/);

    if (splitWords.length === 12) {
      setWords(splitWords);
      event.preventDefault();
    }
  };

  useEffect(() => {
    setWords(Array(12).fill(""));
  }, [visible]);

  return (
    <Dialog open={visible} onClose={hideDialog}>
      <DialogTitle style={styles.mnemonicTitle}>
        Import your wallet from your mnemonic
      </DialogTitle>
      <DialogContent style={styles.mnemonicContainer}>
        <Typography component="div">
          (You can paste your entire mnemonic into the first textbox)
        </Typography>
        <Grid container spacing={2}>
          {words.map((word, index) => (
            <Grid item xs={6} sm={4} key={index}>
              <TextField
                value={word}
                onChange={(e) => handleWordChange(index, e.target.value)}
                onPaste={index === 0 ? handlePaste : undefined}
                placeholder={`Word ${index + 1}`}
                fullWidth
                margin="normal"
              />
            </Grid>
          ))}
        </Grid>
      </DialogContent>
      <DialogActions style={styles.mnemonicButtonRow}>
        <Button
          onClick={() => importWalletHandler(words.join(" "))}
          style={styles.mnemonicButton}
        >
          Import Wallet
        </Button>
        <Button onClick={hideDialog} style={styles.mnemonicButton}>
          Cancel
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default ImportWalletDialog;
