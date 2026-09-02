import React, { useState } from "react";
import { Text, TextInput } from "react-native-paper";
import { Button, Divider, Stack } from "@mui/material";

import { NativeStackScreenProps } from "@react-navigation/native-stack";

import { StackParamsList } from "../types";
import { signMessage } from "../utils/sign-message";
import AccountDetails from "../components/AccountDetails";
import { Layout } from "../components/Layout";

type SignProps = NativeStackScreenProps<StackParamsList, "SignMessage">;

const SignMessage = ({ route }: SignProps) => {
  const namespace = route.params.selectedNamespace;
  const chainId = route.params.selectedChainId;
  const account = route.params.accountInfo;

  const [message, setMessage] = useState<string>("");

  const signMessageHandler = async () => {
    const signedMessage = await signMessage({
      message,
      namespace,
      chainId,
      accountId: account.index,
    });
    alert(`Signature ${signedMessage}`);
  };

  return (
    <Layout title="Sign Message">
      <Text variant="titleMedium">
        {account && `Account ${account.index + 1}`}
      </Text>
      <AccountDetails account={account} />

      <Stack spacing={4}>
        <Divider flexItem />
        <TextInput
          mode="outlined"
          placeholder="Enter your message"
          onChangeText={(text) => setMessage(text)}
          value={message}
        />

        <Button
          variant="contained"
          onClick={signMessageHandler}
          sx={{ width: "200px", px: 4, py: 1, mt: 2 }}
        >
          Sign
        </Button>
      </Stack>
    </Layout>
  );
};

export default SignMessage;
