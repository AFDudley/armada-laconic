import React, { useState } from 'react';
import { View } from 'react-native';
import { List } from 'react-native-paper';

import { NetworkDropdownProps, NetworksDataState } from '../types';
import styles from '../styles/stylesheet';
import { useNetworks } from '../context/NetworksContext';

const NetworkDropdown = ({ updateNetwork }: NetworkDropdownProps) => {
  const { networksData, selectedNetwork, setSelectedNetwork } = useNetworks();

  const [expanded, setExpanded] = useState<boolean>(false);

  const handleNetworkPress = (networkData: NetworksDataState) => {
    updateNetwork(networkData);
    setSelectedNetwork(networkData);
    setExpanded(false);
  };

  return (
    <View style={styles.networkDropdown}>
      <List.Accordion
        title={selectedNetwork!.networkName}
        expanded={expanded}
        onPress={() => setExpanded(!expanded)}>
        {networksData.map(networkData => (
          <List.Item
            key={networkData.networkId}
            title={networkData.networkName}
            onPress={() => handleNetworkPress(networkData)}
          />
        ))}
      </List.Accordion>
    </View>
  );
};

export { NetworkDropdown };
