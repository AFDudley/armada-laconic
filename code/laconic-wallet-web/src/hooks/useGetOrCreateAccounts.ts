import { useEffect, useCallback } from "react";

import { createWallet, isWalletCreated } from "../utils/accounts";
import { sendMessage } from "../utils/misc";
import useAccountsData from "./useAccountsData";
import { useNetworks } from "../context/NetworksContext";
import { useAccounts } from "../context/AccountsContext";
import { REQUEST_CREATE_OR_GET_ACCOUNTS, WALLET_ACCOUNTS_DATA } from "../utils/constants";

const REACT_APP_ALLOWED_URLS = import.meta.env.REACT_APP_ALLOWED_URLS;

const useGetOrCreateAccounts = () => {
  const { networksData } = useNetworks();
  const { getAccountsData } = useAccountsData();
  const { setAccounts } = useAccounts();

  // Wrap the function in useCallback to prevent recreation on each render
  const getOrCreateAccountsForChain = useCallback(async (chainId: string) => {
    const isCreated = await isWalletCreated();

    if (!isCreated) {
      console.log("Accounts not found, creating wallet...");
      await createWallet(networksData);
    }

    const accountsData = await getAccountsData(chainId);

    // Update the AccountsContext with the new accounts
    setAccounts(accountsData);

    return accountsData;
  }, [networksData, getAccountsData, setAccounts]);

  useEffect(() => {
    const handleCreateAccounts = async (event: MessageEvent) => {
      if (event.data.type !== REQUEST_CREATE_OR_GET_ACCOUNTS) return;

      if (!REACT_APP_ALLOWED_URLS) {
        console.log('Allowed URLs are not set');
        return;
      }

      const allowedUrls = REACT_APP_ALLOWED_URLS.split(',').map(url => url.trim());

      if (!allowedUrls.includes(event.origin)) {
        console.log('Unauthorized app.');
        return;
      }

      const accountsData = await getOrCreateAccountsForChain(event.data.chainId);
      const accountsAddressList = accountsData.map(account => account.address);
      console.log('Sending WALLET_ACCOUNTS_DATA accounts:', accountsAddressList);

      sendMessage(
        event.source as Window, WALLET_ACCOUNTS_DATA,
        accountsAddressList,
        event.origin
      );
    };

    const autoCreateAccounts = async () => {
      const defaultChainId = networksData[0]?.chainId;

      if (!defaultChainId) {
        console.log('useGetOrCreateAccounts: No default chainId found');
        return;
      }
      const accounts = await getOrCreateAccountsForChain(defaultChainId);

      // Only notify Android when we actually have accounts
      if (accounts.length > 0 && window.Android?.onAccountsReady) {
        window.Android.onAccountsReady();
      } else {
        console.log('No accounts created or Android bridge not available');
      }
    };

    window.addEventListener('message', handleCreateAccounts);

    const isAndroidWebView = !!(window.Android);

    if (isAndroidWebView) {
      autoCreateAccounts();
    }

    return () => {
      window.removeEventListener('message', handleCreateAccounts);
    };
  }, [networksData, getAccountsData, getOrCreateAccountsForChain]);
};

export default useGetOrCreateAccounts;
