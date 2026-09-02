import React, { useState } from "react";
import { View } from "react-native";
import { List } from "react-native-paper";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";

import { NetworkDropdownProps, NetworksDataState } from "../types";
import styles from "../styles/stylesheet";
import { useNetworks } from "../context/NetworksContext";
import { Accordion, AccordionSummary } from "@mui/material";

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
      <Accordion expanded={expanded} onClick={() => setExpanded(!expanded)}>
        <AccordionSummary expandIcon={<ExpandMoreIcon />}>
          {selectedNetwork!.networkName}
        </AccordionSummary>

        {networksData.map((networkData) => (
          <List.Item
            key={networkData.networkId}
            title={networkData.networkName}
            onPress={() => handleNetworkPress(networkData)}
          />
        ))}
      </Accordion>
    </View>
  );
};

export { NetworkDropdown };
