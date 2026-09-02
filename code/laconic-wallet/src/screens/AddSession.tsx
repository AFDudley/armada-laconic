import React, { useCallback, useEffect, useState } from 'react';
import { AppState, TouchableOpacity, View } from 'react-native';
import { Button, Text, TextInput } from 'react-native-paper';
import {
  Camera,
  useCameraDevice,
  useCameraPermission,
  useCodeScanner,
} from 'react-native-vision-camera';
import { Linking } from 'react-native';

import { useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';

import { web3WalletPair } from '../utils/wallet-connect/WalletConnectUtils';
import styles from '../styles/stylesheet';
import { StackParamsList } from '../types';
import { useWalletConnect } from '../context/WalletConnectContext';

const AddSession = () => {
  const navigation =
    useNavigation<NativeStackNavigationProp<StackParamsList>>();

  const { hasPermission, requestPermission } = useCameraPermission();
  const device = useCameraDevice('back');

  const { web3wallet } = useWalletConnect();

  const [currentWCURI, setCurrentWCURI] = useState<string>('');
  const [isActive, setIsActive] = useState(AppState.currentState === 'active');
  const [isScanning, setScanning] = useState(true);

  const codeScanner = useCodeScanner({
    codeTypes: ['qr'],
    onCodeScanned: codes => {
      if (isScanning) {
        codes.forEach(code => {
          if (code.value) {
            setCurrentWCURI(code.value);
            setScanning(false);
          }
        });
      }
    },
  });

  const linkToSettings = async () => {
    await Linking.openSettings();
  };

  const pair = useCallback(async () => {
    if (!web3wallet) {
      return;
    }

    const pairing = await web3WalletPair(web3wallet, { uri: currentWCURI });
    navigation.navigate('WalletConnect');
    return pairing;
  }, [web3wallet, currentWCURI, navigation]);

  useEffect(() => {
    const handleAppStateChange = (newState: string) => {
      setIsActive(newState === 'active');
    };

    AppState.addEventListener('change', handleAppStateChange);

    if (!hasPermission) {
      requestPermission();
    }
  }, [hasPermission, requestPermission]);

  return (
    <View style={styles.appContainer}>
      {!hasPermission || !device ? (
        <>
          <Text>
            {!hasPermission
              ? 'No Camera Permission granted'
              : 'No Camera Selected'}
          </Text>
          <TouchableOpacity onPress={linkToSettings}>
            <Text variant="titleSmall" style={[styles.hyperlink]}>
              Go to settings
            </Text>
          </TouchableOpacity>
        </>
      ) : (
        <>
          <View style={styles.cameraContainer}>
            {isActive ? (
              <Camera
                style={styles.camera}
                device={device}
                isActive={isActive}
                codeScanner={codeScanner}
                video={false}
              />
            ) : (
              <Text>No Camera Selected!</Text>
            )}
          </View>

          <View style={styles.inputContainer}>
            <Text variant="titleMedium">Enter WalletConnect URI</Text>
            <TextInput
              mode="outlined"
              onChangeText={setCurrentWCURI}
              value={currentWCURI}
              numberOfLines={4}
              multiline={true}
              style={styles.walletConnectUriText}
            />

            <View style={styles.signButton}>
              <Button mode="contained" onPress={pair}>
                Pair Session
              </Button>
            </View>
          </View>
        </>
      )}
    </View>
  );
};
export default AddSession;
