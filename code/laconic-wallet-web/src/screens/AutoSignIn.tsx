import React, { useEffect } from 'react';

import { useNetworks } from '../context/NetworksContext';
import { signMessage } from '../utils/sign-message';
import { AUTO_SIGN_IN, EIP155, SIGN_IN_RESPONSE } from '../utils/constants';
import { sendMessage } from '../utils/misc';
import useAccountsData from '../hooks/useAccountsData';
import useGetOrCreateAccounts from '../hooks/useGetOrCreateAccounts';

const REACT_APP_ALLOWED_URLS = import.meta.env.REACT_APP_ALLOWED_URLS;

export const AutoSignIn = () => {
  const { networksData } = useNetworks();

  const { getAccountsData } = useAccountsData();

  useEffect(() => {
    const handleSignIn = async (event: MessageEvent) => {
      if (event.data.type !== AUTO_SIGN_IN) return;

      if (!REACT_APP_ALLOWED_URLS) {
        console.log('Allowed URLs are not set');
        return;
      }

      const allowedUrls = REACT_APP_ALLOWED_URLS.split(',').map(url => url.trim());

      if (!allowedUrls.includes(event.origin)) {
        console.log('Unauthorized app.');
        return;
      }

      const accountsData = await getAccountsData(event.data.chainId);

      if (!accountsData.length) {
        return
      }

      const signature = await signMessage({ message: event.data.message, accountId: accountsData[0].index, chainId: event.data.chainId, namespace: EIP155 })

      sendMessage(event.source as Window, SIGN_IN_RESPONSE, { message: event.data.message, signature }, event.origin);
    };

    window.addEventListener('message', handleSignIn);

    return () => {
      window.removeEventListener('message', handleSignIn);
    };
  }, [networksData, getAccountsData]);

  // Custom hook for adding listener to get accounts data
  useGetOrCreateAccounts();

  return (
    <>
    </>
  )
};
