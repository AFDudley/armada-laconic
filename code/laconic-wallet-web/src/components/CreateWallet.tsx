import React from "react";
import { View } from "react-native";
import { LoadingButton } from "@mui/lab";

import { CreateWalletProps } from "../types";

const CreateWallet = ({
  isWalletCreating,
  createWalletHandler,
}: CreateWalletProps) => {
  return (
    <View>
      <View>
        <LoadingButton
          variant="contained"
          loading={isWalletCreating}
          disabled={isWalletCreating}
          onClick={createWalletHandler}
        >
          {isWalletCreating ? "Creating" : "CREATE WALLET"}
        </LoadingButton>
      </View>
    </View>
  );
};

export default CreateWallet;
