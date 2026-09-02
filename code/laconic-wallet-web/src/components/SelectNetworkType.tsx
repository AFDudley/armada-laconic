import React, { useState } from 'react';
import { View } from 'react-native';
import { Text, List } from 'react-native-paper';

import styles from '../styles/stylesheet';
import { COSMOS, EIP155 } from '../utils/constants';

const SelectNetworkType = ({
  updateNetworkType,
}: {
  updateNetworkType: (networkType: string) => void;
}) => {
  const [expanded, setExpanded] = useState<boolean>(false);
  const [selectedNetwork, setSelectedNetwork] = useState<string>('ETH');

  const networks = ['ETH', 'COSMOS'];

  const handleNetworkPress = (network: string) => {
    setSelectedNetwork(network);
    updateNetworkType(network === 'ETH' ? EIP155 : COSMOS);
    setExpanded(false);
  };

  return (
    <View style={styles.networkDropdown}>
      <Text style={styles.selectNetworkText}>Select Network Type</Text>
      <List.Accordion
        title={selectedNetwork}
        expanded={expanded}
        onPress={() => setExpanded(!expanded)}>
        {networks.map(network => (
          <List.Item
            key={network}
            title={network}
            onPress={() => handleNetworkPress(network)}
          />
        ))}
      </List.Accordion>
    </View>
  );
};

export { SelectNetworkType };
