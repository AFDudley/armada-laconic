import { useCallback } from "react";

import { retrieveAccounts } from "../utils/accounts";
import { useNetworks } from "../context/NetworksContext";

const useAccountsData = () => {
  const { networksData } = useNetworks();

  const getAccountsData = useCallback(async (chainId: string) => {
    const targetNetwork = networksData.find(network => network.chainId === chainId);

    if (!targetNetwork) {
      return [];
    }

    const accounts = await retrieveAccounts(targetNetwork);
    return accounts || [];
  }, [networksData]);

  return { getAccountsData };
};

export default useAccountsData;
