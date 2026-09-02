import React from 'react';
import { View, Text } from 'react-native';
import { Button } from 'react-native-paper';

import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { useNavigation } from '@react-navigation/native';

import { StackParamsList } from '../types';
import styles from '../styles/stylesheet';

const InvalidPath = () => {
  const navigation =
    useNavigation<NativeStackNavigationProp<StackParamsList>>();
  return (
    <View style={styles.badRequestContainer}>
      <Text style={styles.invalidMessageText}>
        The signature request was invalid.
      </Text>
      <Button
        mode="contained"
        onPress={() => {
          navigation.navigate('Home');
        }}>
        Home
      </Button>
    </View>
  );
};

export default InvalidPath;
