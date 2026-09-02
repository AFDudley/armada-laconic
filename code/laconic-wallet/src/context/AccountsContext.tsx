import React, { createContext, useContext, useState } from 'react';

import { Account } from '../types';

const AccountsContext = createContext<{
  accounts: Account[];
  setAccounts: (account: Account[]) => void;
  currentIndex: number;
  setCurrentIndex: (index: number) => void;
}>({
  accounts: [],
  setAccounts: () => {},
  currentIndex: 0,
  setCurrentIndex: () => {},
});

const useAccounts = () => {
  const accountsContext = useContext(AccountsContext);
  return accountsContext;
};

const AccountsProvider = ({ children }: { children: any }) => {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [currentIndex, setCurrentIndex] = useState<number>(0);

  return (
    <AccountsContext.Provider
      value={{
        accounts,
        setAccounts,
        currentIndex,
        setCurrentIndex,
      }}>
      {children}
    </AccountsContext.Provider>
  );
};

export { useAccounts, AccountsProvider };
