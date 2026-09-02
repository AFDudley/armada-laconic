import React, { useState } from "react";
import { TouchableOpacity, View } from "react-native";
import { Button, Link, Typography } from "@mui/material";
import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";

import styles from "../styles/stylesheet";
import { getPathKey } from "../utils/misc";
import { useNetworks } from "../context/NetworksContext";
import { useAccounts } from "../context/AccountsContext";

const ShowPKDialog = () => {
  const { currentIndex } = useAccounts();
  const { selectedNetwork } = useNetworks();

  const [privateKey, setPrivateKey] = useState<string>();
  const [showPKDialog, setShowPKDialog] = useState<boolean>(false);

  const handleShowPrivateKey = async () => {
    const pathKey = await getPathKey(
      `${selectedNetwork!.namespace}:${selectedNetwork!.chainId}`,
      currentIndex,
    );

    setPrivateKey(pathKey.privKey);
  };

  const hideShowPKDialog = () => {
    setShowPKDialog(false);
    setPrivateKey(undefined);
  };

  return (
    <>
      <View style={styles.signLink}>
        <TouchableOpacity
          onPress={() => {
            setShowPKDialog(true);
          }}
        >
          <Link>Show Private Key</Link>
        </TouchableOpacity>
      </View>
      <View>
        <Dialog open={showPKDialog} onClose={hideShowPKDialog}>
          <DialogTitle>
            {!privateKey ? (
              <Typography>Show Private Key?</Typography>
            ) : (
              <Typography>Private Key</Typography>
            )}
          </DialogTitle>
          <DialogContent>
            {privateKey && (
              <View style={[styles.dataBox, styles.dataBoxContainer]}>
                <Typography
                  component="pre"
                  variant="body1"
                  style={styles.dataBoxData}
                  sx={{
                    wordWrap: "break-word",
                    whiteSpace: "initial"
                  }}
                >
                  {privateKey}
                </Typography>
              </View>
            )}
            <View>
              <Typography variant="body1" style={styles.dialogWarning}>
                Warning: Never disclose this key. Anyone with your private keys can steal
                any assets held in your account.
              </Typography>
            </View>
          </DialogContent>
          <DialogActions>
            {!privateKey ? (
              <>
                <Button style={{...styles.buttonRed, ...styles.button}} onClick={handleShowPrivateKey}>Yes</Button>
                <Button style={{...styles.buttonBlue, ...styles.button}} onClick={hideShowPKDialog}>No</Button>
              </>
            ) : (
              <Button style={{...styles.buttonBlue, ...styles.button}} onClick={hideShowPKDialog}>Ok</Button>
            )}
          </DialogActions>
        </Dialog>
      </View>
    </>
  );
};

export default ShowPKDialog;
