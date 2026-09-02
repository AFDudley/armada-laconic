import React, { useCallback, useState } from "react";
import { View } from "react-native";
import { Text, TextInput } from "react-native-paper";

import { useNavigation } from "@react-navigation/native";
import { NativeStackNavigationProp } from "@react-navigation/native-stack";
import { Box, Button } from "@mui/material";

import { web3WalletPair } from "../utils/wallet-connect/WalletConnectUtils";
import styles from "../styles/stylesheet";
import { StackParamsList } from "../types";
import { useWalletConnect } from "../context/WalletConnectContext";
import { Layout } from "../components/Layout";

const AddSession = () => {
  const navigation =
    useNavigation<NativeStackNavigationProp<StackParamsList>>();

  const [currentWCURI, setCurrentWCURI] = useState<string>("");

  const { web3wallet } = useWalletConnect();

  const pair = useCallback(async () => {
    if (!web3wallet) {
      return;
    }

    const pairing = await web3WalletPair(web3wallet, { uri: currentWCURI });
    navigation.navigate("WalletConnect");
    return pairing;
  }, [currentWCURI, navigation, web3wallet]);

  return (
    <Layout title="Add Session">
      <View style={styles.inputContainer}>
        <Text variant="titleMedium">Enter WalletConnect URI</Text>
        <TextInput
          mode="outlined"
          onChangeText={setCurrentWCURI}
          value={currentWCURI}
          numberOfLines={4}
          multiline={true}
          style={styles.walletConnectUriText}
        />

        <Box sx={{ mt: 2 }}>
          <Button variant="contained" onClick={pair}>
            Pair Session
          </Button>
        </Box>
      </View>
    </Layout>
  );
};
export default AddSession;
