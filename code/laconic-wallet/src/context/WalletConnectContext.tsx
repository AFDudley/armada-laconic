import React, { createContext, useContext, useEffect, useState } from 'react';

import { SessionTypes } from '@walletconnect/types';
import { IWeb3Wallet } from '@walletconnect/web3wallet';

import { WalletConnectContextProps } from '../types';
import useInitialization from '../hooks/useInitialization';

const WalletConnectContext = createContext<WalletConnectContextProps>({
  activeSessions: {},
  setActiveSessions: () => {},
  web3wallet: {} as IWeb3Wallet,
  setWeb3wallet: () => {},
});

const useWalletConnect = () => {
  const walletConnectContext = useContext(WalletConnectContext);
  return walletConnectContext;
};

const WalletConnectProvider = ({ children }: { children: React.ReactNode }) => {
  const [web3wallet, setWeb3wallet] = useState<IWeb3Wallet | undefined>();

  useInitialization(setWeb3wallet);

  useEffect(() => {
    const sessions = (web3wallet && web3wallet.getActiveSessions()) || {};
    setActiveSessions(sessions);
  }, [web3wallet]);

  const [activeSessions, setActiveSessions] = useState<
    Record<string, SessionTypes.Struct>
  >({});

  return (
    <WalletConnectContext.Provider
      value={{
        activeSessions,
        setActiveSessions,
        web3wallet,
        setWeb3wallet,
      }}>
      {children}
    </WalletConnectContext.Provider>
  );
};

export { useWalletConnect, WalletConnectProvider };
