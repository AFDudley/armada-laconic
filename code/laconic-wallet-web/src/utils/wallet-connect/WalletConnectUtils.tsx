
import '@ethersproject/shims';
import { Core } from '@walletconnect/core';
import { ICore } from '@walletconnect/types';
import { Web3Wallet, IWeb3Wallet } from '@walletconnect/web3wallet';

export let web3wallet:
  | IWeb3Wallet
  | undefined;
export let core: ICore;

export async function createWeb3Wallet() {
  core = new Core({
    projectId: import.meta.env.REACT_APP_WALLET_CONNECT_PROJECT_ID,
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
