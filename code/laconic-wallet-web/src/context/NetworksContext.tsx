import React, { createContext, useContext, useEffect, useState } from 'react';

import { NetworksDataState } from '../types';
import { retrieveNetworksData } from '../utils/accounts';
import { DEFAULT_NETWORKS, EIP155 } from '../utils/constants';
import { setInternetCredentials } from '../utils/key-store';

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

const DEFAULT_NETWORKS_DATA = DEFAULT_NETWORKS.map((defaultNetwork, index) => (
  {
    ...defaultNetwork,
    networkId: index.toString()
  })
);

const NetworksProvider = ({ children }: { children: React.ReactNode }) => {
  const [networksData, setNetworksData] = useState<NetworksDataState[]>(DEFAULT_NETWORKS_DATA);
  const [networkType, setNetworkType] = useState<string>(EIP155);
  const [selectedNetwork, setSelectedNetwork] = useState<NetworksDataState>();

  useEffect(() => {
    const fetchData = async () => {
      let retrievedNetworks = await retrieveNetworksData();

      if (retrievedNetworks.length === 0) {
        setInternetCredentials(
          'networks',
          '_',
          JSON.stringify(DEFAULT_NETWORKS_DATA),
        );

        retrievedNetworks = DEFAULT_NETWORKS_DATA;
      }

      setNetworksData(retrievedNetworks);
      setSelectedNetwork(retrievedNetworks[0]);
    };

    fetchData();
  }, []);

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
