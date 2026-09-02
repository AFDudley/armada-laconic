import { useEffect, useCallback } from 'react';

import { useNetworks } from '../context/NetworksContext';
import { sendMessage } from '../utils/misc';
import useAccountsData from '../hooks/useAccountsData';
import { addAccount } from '../utils/accounts';
import { useAccounts } from '../context/AccountsContext';
import { Account, NetworksDataState } from '../types';
import { ADD_ACCOUNT_RESPONSE, REQUEST_ADD_ACCOUNT } from '../utils/constants';

const REACT_APP_ALLOWED_URLS = import.meta.env.REACT_APP_ALLOWED_URLS;

const useAddAccountEmbed = () => {
  const { networksData } = useNetworks();
  const { setAccounts, setCurrentIndex } = useAccounts();
  const { getAccountsData } = useAccountsData();

  const addAccountHandler = useCallback(async (network: NetworksDataState) => {
    const newAccount = await addAccount(network.chainId);
    if (newAccount) {
      setAccounts(prev => [...prev, newAccount]);
      setCurrentIndex(newAccount.index);
    }
  }, [setAccounts, setCurrentIndex]);

  useEffect(() => {
    const handleAddAccount = async (event: MessageEvent) => {
      if (event.data.type !== REQUEST_ADD_ACCOUNT) return;

      if (!REACT_APP_ALLOWED_URLS) {
        console.log('Unauthorized app origin:', event.origin);
        return;
      }

      const allowedUrls = REACT_APP_ALLOWED_URLS.split(',').map(url => url.trim());

      if (!allowedUrls.includes(event.origin)) {
        console.log('Unauthorized app.');
        return;
      }

      const network = networksData.find(network => network.chainId === event.data.chainId);

      await addAccountHandler(network!);
      const updatedAccounts = await getAccountsData(event.data.chainId);
      const addresses = updatedAccounts.map((account: Account) => account.address);

      sendMessage(event.source as Window, ADD_ACCOUNT_RESPONSE, addresses, event.origin);
    };

    window.addEventListener('message', handleAddAccount);

    return () => {
      window.removeEventListener('message', handleAddAccount);
    };
  }, [networksData, getAccountsData, addAccountHandler]);
};

export default useAddAccountEmbed;
