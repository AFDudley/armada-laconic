import { TouchableOpacity, View } from 'react-native';
import React, { useState } from 'react';
import { Button, Text, Portal, Dialog, useTheme } from 'react-native-paper';

import styles from '../styles/stylesheet';
import { getPathKey } from '../utils/misc';
import { useNetworks } from '../context/NetworksContext';
import { useAccounts } from '../context/AccountsContext';

const ShowPKDialog = () => {
  const { currentIndex } = useAccounts();
  const { selectedNetwork } = useNetworks();

  const [privateKey, setprivateKey] = useState<string>();
  const [showPKDialog, setShowPKDialog] = useState<boolean>(false);

  const theme = useTheme();

  const handleShowPrivateKey = async () => {
    const pathKey = await getPathKey(
      `${selectedNetwork!.namespace}:${selectedNetwork!.chainId}`,
      currentIndex,
    );

    setprivateKey(pathKey.privKey);
  };

  const hideShowPKDialog = () => {
    setShowPKDialog(false);
    setprivateKey(undefined);
  };

  return (
    <>
      <View style={styles.signLink}>
        <TouchableOpacity
          onPress={() => {
            setShowPKDialog(true);
          }}>
          <Text
            variant="titleSmall"
            style={[styles.hyperlink, { color: theme.colors.primary }]}>
            Show Private Key
          </Text>
        </TouchableOpacity>
      </View>
      <View>
        <Portal>
          <Dialog visible={showPKDialog} onDismiss={hideShowPKDialog}>
            <Dialog.Title>
              {!privateKey ? (
                <Text>Show Private Key?</Text>
              ) : (
                <Text>Private Key</Text>
              )}
            </Dialog.Title>
            <Dialog.Content>
              {privateKey && (
                <View style={[styles.dataBox, styles.dataBoxContainer]}>
                  <Text
                    selectable
                    variant="bodyMedium"
                    style={styles.dataBoxData}>
                    {privateKey}
                  </Text>
                </View>
              )}
              <View>
                <Text variant="bodyMedium" style={styles.dialogWarning}>
                  <Text style={[styles.highlight, styles.dialogWarning]}>
                    Warning:
                  </Text>
                  Never disclose this key. Anyone with your private keys can
                  steal any assets held in your account.
                </Text>
              </View>
            </Dialog.Content>
            <Dialog.Actions>
              {!privateKey ? (
                <>
                  <Button onPress={handleShowPrivateKey} textColor="red">
                    Yes
                  </Button>
                  <Button onPress={hideShowPKDialog}>No</Button>
                </>
              ) : (
                <Button onPress={hideShowPKDialog}>Ok</Button>
              )}
            </Dialog.Actions>
          </Dialog>
        </Portal>
      </View>
    </>
  );
};

export default ShowPKDialog;
