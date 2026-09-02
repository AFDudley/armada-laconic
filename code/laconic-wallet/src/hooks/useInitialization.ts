import { useCallback, useEffect, useState } from 'react';

import { IWeb3Wallet } from '@walletconnect/web3wallet';

import { createWeb3Wallet } from '../utils/wallet-connect/WalletConnectUtils';

export default function useInitialization(
  setWeb3wallet: React.Dispatch<React.SetStateAction<IWeb3Wallet | undefined>>,
) {
  const [initialized, setInitialized] = useState(false);

  const onInitialize = useCallback(async () => {
    try {
      const web3walletInstance = await createWeb3Wallet();
      setWeb3wallet(web3walletInstance);

      setInitialized(true);
    } catch (err: unknown) {
      console.error('Error for initializing', err);
    }
  }, [setWeb3wallet]);

  useEffect(() => {
    if (!initialized) {
      onInitialize();
    }
  }, [initialized, onInitialize]);

  return initialized;
}
