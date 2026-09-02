import React, { useCallback, useEffect, useState } from "react";
import { View, ActivityIndicator } from "react-native";
import { Text } from "react-native-paper";
import { Button, Divider, Portal, Snackbar } from "@mui/material";

import { getSdkError } from "@walletconnect/utils";

import { createWallet, resetWallet, retrieveAccounts } from "../utils/accounts";
import { NetworkDropdown } from "../components/NetworkDropdown";
import Accounts from "../components/Accounts";
import CreateWallet from "../components/CreateWallet";
import ConfirmDialog from "../components/ConfirmDialog";
import styles from "../styles/stylesheet";
import { useAccounts } from "../context/AccountsContext";
import { useWalletConnect } from "../context/WalletConnectContext";
import { NetworksDataState } from "../types";
import { useNetworks } from "../context/NetworksContext";
import ImportWalletDialog from "../components/ImportWalletDialog";
import { MnemonicDialog } from "../components/MnemonicDialog";
import { Container } from "../components/Container";
import { IS_IMPORT_WALLET_ENABLED } from "../utils/constants";

const HomeScreen = () => {
  const { setAccounts, setCurrentIndex } = useAccounts();

  const { networksData, selectedNetwork, setSelectedNetwork, setNetworksData } =
    useNetworks();
  const { setActiveSessions, web3wallet } = useWalletConnect();

  const [isWalletCreated, setIsWalletCreated] = useState<boolean>(false);
  const [isWalletCreating, setIsWalletCreating] = useState<boolean>(false);
  const [importWalletDialog, setImportWalletDialog] = useState<boolean>(false);
  const [mnemonicDialog, setMnemonicDialog] = useState<boolean>(false);
  const [resetWalletDialog, setResetWalletDialog] = useState<boolean>(false);
  const [toastVisible, setToastVisible] = useState(false);
  const [invalidMnemonicError, setInvalidMnemonicError] = useState("");
  const [isAccountsFetched, setIsAccountsFetched] = useState<boolean>(true);
  const [phrase, setPhrase] = useState("");

  const hideMnemonicDialog = () => setMnemonicDialog(false);
  const hideResetDialog = () => setResetWalletDialog(false);

  const fetchAccounts = useCallback(async () => {
    if (!selectedNetwork) {
      return;
    }

    const loadedAccounts = await retrieveAccounts(selectedNetwork);

    if (loadedAccounts) {
      setAccounts(loadedAccounts);
      setIsWalletCreated(true);
    }

    setIsAccountsFetched(true);
  }, [selectedNetwork, setAccounts]);

  const createWalletHandler = async () => {
    setIsWalletCreating(true);
    const mnemonic = await createWallet(networksData);
    if (mnemonic) {
      fetchAccounts();
      setMnemonicDialog(true);
      setPhrase(mnemonic);
      setSelectedNetwork(networksData[0]);
    }
  };

  const importWalletHandler = async (recoveryPhrase: string) => {
    try {
      const mnemonic = await createWallet(networksData, recoveryPhrase);
      if (mnemonic) {
        fetchAccounts();
        setPhrase(mnemonic);
        setSelectedNetwork(networksData[0]);
        setImportWalletDialog(false);
      }
    } catch (error: any) {
      setInvalidMnemonicError((error.message as string).toUpperCase());
      setToastVisible(true);
    }
  };

  const confirmResetWallet = useCallback(async () => {
    setIsWalletCreated(false);
    setIsWalletCreating(false);
    setAccounts([]);
    setCurrentIndex(0);
    setNetworksData([]);
    setSelectedNetwork(undefined);
    await resetWallet();
    const sessions = web3wallet!.getActiveSessions();

    Object.keys(sessions).forEach(async (sessionId) => {
      await web3wallet!.disconnectSession({
        topic: sessionId,
        reason: getSdkError("USER_DISCONNECTED"),
      });
    });
    setActiveSessions({});

    hideResetDialog();
  }, [
    setAccounts,
    setActiveSessions,
    setCurrentIndex,
    setNetworksData,
    setSelectedNetwork,
    web3wallet,
  ]);

  const updateNetwork = (networkData: NetworksDataState) => {
    setSelectedNetwork(networkData);
    setCurrentIndex(0);
  };

  useEffect(() => {
    fetchAccounts();
  }, [networksData, setAccounts, selectedNetwork, fetchAccounts]);

  return (
    <View style={styles.appContainer}>
      <Container>
        {!isAccountsFetched ? (
          <View style={styles.spinnerContainer}>
            <Text style={styles.LoadingText}>Loading...</Text>
            <ActivityIndicator size="large" color="#0000ff" />
          </View>
        ) : isWalletCreated && selectedNetwork ? (
          <>
            <NetworkDropdown updateNetwork={updateNetwork} />
            <View style={styles.accountComponent}>
              <Accounts />
            </View>
            <Divider flexItem sx={{ my: 4 }} />
            <Button
              variant="contained"
              onClick={() => {
                setResetWalletDialog(true);
              }}
              color="error"
            >
              Reset Wallet
            </Button>
          </>
        ) : (
          <>
            <CreateWallet
              isWalletCreating={isWalletCreating}
              createWalletHandler={createWalletHandler}
            />
            {IS_IMPORT_WALLET_ENABLED && (
              <View style={styles.createWalletContainer}>
                <Button
                  variant="contained"
                  onClick={() => setImportWalletDialog(true)}
                >
                  Import Wallet
                </Button>
              </View>
            )}
          </>
        )}
        <ImportWalletDialog
          visible={importWalletDialog}
          hideDialog={() => setImportWalletDialog(false)}
          importWalletHandler={importWalletHandler}
        />
        <MnemonicDialog
          visible={mnemonicDialog}
          hideDialog={hideMnemonicDialog}
          contentText={phrase}
        />
        <ConfirmDialog
          title="Reset wallet"
          visible={resetWalletDialog}
          hideDialog={hideResetDialog}
          onConfirm={confirmResetWallet}
        />
      </Container>
      <Portal>
        <Snackbar
          open={toastVisible}
          autoHideDuration={3000}
          message={invalidMnemonicError}
          onClose={() => setToastVisible(false)}
          anchorOrigin={{ horizontal: "center", vertical: "bottom" }}
          ContentProps={{ style: { backgroundColor: "red", color: "white" } }}
        />
      </Portal>
    </View>
  );
};

export default HomeScreen;
