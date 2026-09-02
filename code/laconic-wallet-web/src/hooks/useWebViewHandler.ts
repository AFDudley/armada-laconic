import { useEffect, useCallback } from 'react';
import { useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';

import { useAccounts } from '../context/AccountsContext';
import { useNetworks } from '../context/NetworksContext';
import { StackParamsList } from '../types';
import useGetOrCreateAccounts from './useGetOrCreateAccounts';

export const useWebViewHandler = () => {
  // Navigation and context hooks
  const navigation = useNavigation<NativeStackNavigationProp<StackParamsList>>();
  const { selectedNetwork } = useNetworks();
  const { accounts, currentIndex } = useAccounts();

  // Initialize accounts
  useGetOrCreateAccounts();

  // Core navigation handler
  const navigateToSignRequest = useCallback((message: string) => {
    try {
      // Validation checks
      if (!selectedNetwork?.namespace || !selectedNetwork?.chainId) {
        window.Android?.onSignatureError?.('Invalid network configuration');
        return;
      }

      if (!accounts?.length) {
        window.Android?.onSignatureError?.('No accounts available');
        return;
      }

      const currentAccount = accounts[currentIndex];
      if (!currentAccount) {
        window.Android?.onSignatureError?.('Current account not found');
        return;
      }

      // Create the path and validate with regex
      const path = `/sign/${selectedNetwork.namespace}/${selectedNetwork.chainId}/${currentAccount.address}/${encodeURIComponent(message)}`;
      const pathRegex = /^\/sign\/(eip155|cosmos)\/(.+)\/(.+)\/(.+)$/;
      const match = path.match(pathRegex);

      if (!match) {
        window.Android?.onSignatureError?.('Invalid signing path');
        return;
      }

      const [, pathNamespace, pathChainId, pathAddress, pathMessage] = match;

      // Reset navigation stack and navigate to sign request
      navigation.reset({
        index: 0,
        routes: [
          {
            name: 'SignRequest',
            path,
            params: {
              namespace: pathNamespace,
              chainId: pathChainId,
              address: pathAddress,
              message: decodeURIComponent(pathMessage),
              accountInfo: currentAccount,
            },
          },
        ],
      });
    } catch (error) {
      window.Android?.onSignatureError?.(`Navigation error: ${error}`);
    }
  }, [selectedNetwork, accounts, currentIndex, navigation]);

  useEffect(() => {
    // Assign the function to the window object
    window.receiveSignRequestFromAndroid = navigateToSignRequest;

    return () => {
      window.receiveSignRequestFromAndroid = undefined;
    };
  }, [navigateToSignRequest]); // Only the function reference as dependency
};
