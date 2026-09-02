import React from 'react';
import 'text-encoding-polyfill';
import { AppRegistry } from 'react-native';
import { PaperProvider } from 'react-native-paper';

import { NavigationContainer } from '@react-navigation/native';

import App from './src/App';
import { AccountsProvider } from './src/context/AccountsContext';
import { NetworksProvider } from './src/context/NetworksContext';
import { WalletConnectProvider } from './src/context/WalletConnectContext';
import { name as appName } from './app.json';

export default function Main() {
  const linking = {
    prefixes: ['https://wallet.laconic.com'],
    config: {
      screens: {
        SignRequest: {
          path: 'sign/:namespace/:chaindId/:address/:message',
        },
      },
    },
  };
  return (
    <PaperProvider theme={'light'}>
      <NetworksProvider>
        <AccountsProvider>
          <WalletConnectProvider>
            <NavigationContainer linking={linking}>
              <App />
            </NavigationContainer>
          </WalletConnectProvider>
        </AccountsProvider>
      </NetworksProvider>
    </PaperProvider>
  );
}

AppRegistry.registerComponent(appName, () => Main);
