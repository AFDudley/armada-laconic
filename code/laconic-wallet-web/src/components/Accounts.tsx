import React, { useEffect, useState } from "react";
import { TouchableOpacity, View } from "react-native";
import { List } from "react-native-paper";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";

import { useNavigation } from "@react-navigation/native";
import { NativeStackNavigationProp } from "@react-navigation/native-stack";

import { StackParamsList, Account } from "../types";
import { addAccount } from "../utils/accounts";
import HDPathDialog from "./HDPathDialog";
import AccountDetails from "./AccountDetails";
import { useAccounts } from "../context/AccountsContext";
import { useWalletConnect } from "../context/WalletConnectContext";
import { useNetworks } from "../context/NetworksContext";
import ConfirmDialog from "./ConfirmDialog";
import { getNamespaces } from "../utils/wallet-connect/helpers";
import ShowPKDialog from "./ShowPKDialog";
import { setInternetCredentials } from "../utils/key-store";
import {
  Accordion,
  AccordionSummary,
  Button,
  Link,
  Stack,
} from "@mui/material";
import { LoadingButton } from "@mui/lab";

const Accounts = () => {
  const navigation =
    useNavigation<NativeStackNavigationProp<StackParamsList>>();

  const { accounts, setAccounts, setCurrentIndex, currentIndex } =
    useAccounts();
  const { networksData, selectedNetwork, setNetworksData, setSelectedNetwork } =
    useNetworks();

  const { web3wallet } = useWalletConnect();
  const [expanded, setExpanded] = useState(false);
  const [isAccountCreating, setIsAccountCreating] = useState(false);
  const [hdDialog, setHdDialog] = useState(false);
  const [pathCode, setPathCode] = useState("");
  const [deleteNetworkDialog, setDeleteNetworkDialog] =
    useState<boolean>(false);

  const handlePress = () => setExpanded(!expanded);

  const hideDeleteNetworkDialog = () => setDeleteNetworkDialog(false);

  const updateAccounts = (account: Account) => {
    setAccounts([...accounts, account]);
  };

  useEffect(() => {
    const updateSessions = async () => {
      const sessions = (web3wallet && web3wallet.getActiveSessions()) || {};
      // Iterate through each session

      for (const topic in sessions) {
        const session = sessions[topic];
        const { optionalNamespaces, requiredNamespaces } = session;

        const updatedNamespaces = await getNamespaces(
          optionalNamespaces,
          requiredNamespaces,
          networksData,
          selectedNetwork!,
          accounts,
        );

        if (!updatedNamespaces) {
          return;
        }

        await web3wallet!.updateSession({
          topic,
          namespaces: updatedNamespaces,
        });
      }
    };
    // Call the updateSessions function when the 'accounts' dependency changes
    updateSessions();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [accounts, web3wallet]);

  const addAccountHandler = async () => {
    setIsAccountCreating(true);
    const newAccount = await addAccount(selectedNetwork!.chainId);
    setIsAccountCreating(false);
    if (newAccount) {
      updateAccounts(newAccount);
      setCurrentIndex(newAccount.index);
    }
  };

  const renderAccountItems = () =>
    accounts.map((account) => (
      <List.Item
        key={account.index}
        title={`Account ${account.index + 1}`}
        onPress={() => {
          setCurrentIndex(account.index);
          setExpanded(false);
        }}
      />
    ));

  const handleRemove = async () => {
    const updatedNetworks = networksData.filter(
      (networkData) => selectedNetwork!.networkId !== networkData.networkId,
    );

    await setInternetCredentials(
      "networks",
      "_",
      JSON.stringify(updatedNetworks),
    );

    setSelectedNetwork(updatedNetworks[0]);
    setCurrentIndex(0);
    setDeleteNetworkDialog(false);
    setNetworksData(updatedNetworks);
  };

  return (
    <View>
      <View>
        <HDPathDialog
          visible={hdDialog}
          hideDialog={() => setHdDialog(false)}
          updateAccounts={updateAccounts}
          pathCode={pathCode}
        />
        <Accordion expanded={expanded} onClick={handlePress}>
          <AccordionSummary
            expandIcon={<ExpandMoreIcon />}
          >{`Account ${currentIndex + 1}`}</AccordionSummary>
          {renderAccountItems()}
        </Accordion>

        <Stack direction="row" spacing={3} sx={{ mt: 4 }}>
          <LoadingButton
            variant="contained"
            onClick={addAccountHandler}
            loading={isAccountCreating}
            disabled={isAccountCreating}
          >
            {isAccountCreating ? "Adding" : "Add Account"}
          </LoadingButton>

          <Button
            variant="contained"
            onClick={() => {
              setHdDialog(true);
              setPathCode(`m/44'/${selectedNetwork!.coinType}'/`);
            }}
          >
            Add Account from HD path
          </Button>
        </Stack>

        <AccountDetails account={accounts[currentIndex]} />
        <Stack direction="row" spacing={4}>
          <TouchableOpacity
            onPress={() => {
              navigation.navigate("SignMessage", {
                selectedNamespace: selectedNetwork!.namespace,
                selectedChainId: selectedNetwork!.chainId,
                accountInfo: accounts[currentIndex],
              });
            }}
          >
            <Link>Sign Message</Link>
          </TouchableOpacity>

          <TouchableOpacity
            onPress={() => {
              navigation.navigate("AddNetwork");
            }}
          >
            <Link>Add Network</Link>
          </TouchableOpacity>

          <TouchableOpacity
            onPress={() => {
              navigation.navigate("EditNetwork", {
                selectedNetwork: selectedNetwork!,
              });
            }}
          >
            <Link>Edit Network</Link>
          </TouchableOpacity>

          {!selectedNetwork!.isDefault && (
            <TouchableOpacity
              onPress={() => {
                setDeleteNetworkDialog(true);
              }}
            >
              <Link>Delete Network</Link>
            </TouchableOpacity>
          )}
          <ConfirmDialog
            title="Delete Network"
            visible={deleteNetworkDialog}
            hideDialog={hideDeleteNetworkDialog}
            onConfirm={handleRemove}
          />
          <ShowPKDialog />
        </Stack>
      </View>
    </View>
  );
};

export default Accounts;
