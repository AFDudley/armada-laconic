import React from 'react';
import { View } from 'react-native';
import { Text } from 'react-native-paper';

import styles from '../styles/stylesheet';
import { GridViewProps } from '../types';

const GridView = ({ words }: GridViewProps) => {
  return (
    <View style={styles.gridContainer}>
      {words.map((word, index) => (
        <View key={index} style={styles.gridItem}>
          <Text>{index + 1}. </Text>
          <Text variant="bodySmall">{word}</Text>
        </View>
      ))}
    </View>
  );
};

export default GridView;
