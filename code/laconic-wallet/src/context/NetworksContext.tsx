import React, { createContext, useContext, useEffect, useState } from 'react';

import { NetworksDataState } from '../types';
import { retrieveNetworksData, storeNetworkData } from '../utils/accounts';
import { DEFAULT_NETWORKS, EIP155 } from '../utils/constants';

const NetworksContext = createContext<{
  networksData: NetworksDataState[];
  setNetworksData: React.Dispatch<React.SetStateAction<NetworksDataState[]>>;
  networkType: string;
  setNetworkType: (networkType: string) => void;
  selectedNetwork?: NetworksDataState;
  setSelectedNetwork: React.Dispatch<
    React.SetStateAction<NetworksDataState | undefined>
  >;
}>({
  networksData: [],
  setNetworksData: () => {},
  networkType: '',
  setNetworkType: () => {},
  selectedNetwork: {} as NetworksDataState,
  setSelectedNetwork: () => {},
});

const useNetworks = () => {
  const networksContext = useContext(NetworksContext);
  return networksContext;
};

const NetworksProvider = ({ children }: { children: any }) => {
  const [networksData, setNetworksData] = useState<NetworksDataState[]>([]);
  const [networkType, setNetworkType] = useState<string>(EIP155);
  const [selectedNetwork, setSelectedNetwork] = useState<NetworksDataState>();

  useEffect(() => {
    const fetchData = async () => {
      const retrievedNetworks = await retrieveNetworksData();
      if (retrievedNetworks.length === 0) {
        for (const defaultNetwork of DEFAULT_NETWORKS) {
          await storeNetworkData(defaultNetwork);
        }
      }
      const retrievedNewNetworks = await retrieveNetworksData();
      setNetworksData(retrievedNewNetworks);
      setSelectedNetwork(retrievedNewNetworks[0]);
    };

    if (networksData.length === 0) {
      fetchData();
    }
  }, [networksData]);

  useEffect(() => {
    setSelectedNetwork(prevSelectedNetwork => {
      return networksData.find(
        networkData => networkData.networkId === prevSelectedNetwork?.networkId,
      );
    });
  }, [networksData]);

  return (
    <NetworksContext.Provider
      value={{
        networksData,
        setNetworksData,
        networkType,
        setNetworkType,
        selectedNetwork,
        setSelectedNetwork,
      }}>
      {children}
    </NetworksContext.Provider>
  );
};

export { useNetworks, NetworksProvider };
