import React from 'react';
import { View, Text } from 'react-native';

import styles from '../styles/stylesheet';

const DataBox = ({ label, data }: { label: string; data: string }) => {
  return (
    <View style={styles.dataBoxContainer}>
      <Text style={styles.dataBoxLabel}>{label}</Text>
      <View style={styles.dataBox}>
        <Text style={styles.dataBoxData}>{data}</Text>
      </View>
    </View>
  );
};

export default DataBox;
