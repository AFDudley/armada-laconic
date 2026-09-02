import Config from 'react-native-config';

import '@walletconnect/react-native-compat';
import '@ethersproject/shims';
import { Core } from '@walletconnect/core';
import { ICore } from '@walletconnect/types';
import { Web3Wallet, IWeb3Wallet } from '@walletconnect/web3wallet';

export let core: ICore;

export async function createWeb3Wallet() {
  core = new Core({
    projectId: Config.WALLET_CONNECT_PROJECT_ID,
  });

  const web3wallet = await Web3Wallet.init({
    core,
    metadata: {
      name: 'Laconic Wallet',
      description: 'Laconic Wallet',
      url: 'https://wallet.laconic.com/',
      icons: ['https://avatars.githubusercontent.com/u/92608123'],
    },
  });

  return web3wallet;
}

export async function web3WalletPair(
  web3wallet: IWeb3Wallet,
  params: { uri: string },
) {
  if (web3wallet) {
    return await web3wallet.core.pairing.pair({ uri: params.uri });
  }
}
