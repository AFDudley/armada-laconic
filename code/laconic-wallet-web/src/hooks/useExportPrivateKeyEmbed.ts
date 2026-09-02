import { useEffect } from 'react';

import { useAccounts } from '../context/AccountsContext';
import { getPathKey, sendMessage } from '../utils/misc';
import { ACCOUNT_PK_RESPONSE, REQUEST_ACCOUNT_PK } from '../utils/constants';

const useExportPKEmbed = () => {
  const {  accounts } = useAccounts();

  useEffect(() => {
    const handleMessage = async (event: MessageEvent) => {
      const { type, chainId, address } = event.data;

      if (type !== REQUEST_ACCOUNT_PK) return;

      try {
        const selectedAccount = accounts.find(account => account.address === address);
        if (!selectedAccount) {
          throw new Error("Account not found")
        }

        const pathKey = await getPathKey(chainId, selectedAccount.index);
        const privateKey = pathKey.privKey;

        sendMessage(
          event.source as Window,
          ACCOUNT_PK_RESPONSE,
          { privateKey },
          event.origin,
        );
      } catch (error) {
        console.error('Error fetching private key:', error);
      }
    };

    window.addEventListener('message', handleMessage);
    return () => {
      window.removeEventListener('message', handleMessage);
    };
  }, [accounts]);
};

export default useExportPKEmbed;
