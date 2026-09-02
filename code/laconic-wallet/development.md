# Development

## WalletConnect details

- Docs - https://docs.walletconnect.com/api/sign/overview

- Doc for terminologies - https://docs.walletconnect.com/advanced/glossary

- Function for creating web3 wallet

  ```js
  export let core: ICore;

  export async function createWeb3Wallet() {
    const core = new Core({
      projectId: Config.WALLET_CONNECT_PROJECT_ID,
    });

    web3wallet = await Web3Wallet.init({
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
  ```

- Hook used for intializing web3 wallet

  ```js
  export default function useInitialization(setWeb3wallet) {
    const [initialized, setInitialized] = useState(false);

    const onInitialize = useCallback(async () => {
      try {
        const web3walletInstance = await createWeb3Wallet();
        setWeb3wallet(web3walletInstance);

        setInitialized(true);
      } catch (err: unknown) {
        console.error('Error for initializing', err);
      }
    }, [setWeb3wallet]);

    useEffect(() => {
      if (!initialized) {
        onInitialize();
      }
    }, [initialized, onInitialize]);

    return initialized;
  }
  ```

- Once user clicks on pair, this function is triggered

  ```js
  export async function web3WalletPair(params: { uri: string }) {
    return await web3wallet.core.pairing.pair({ uri: params.uri });
  }
  ```

- In a useEffect, we keep on listening to events that are emitted by dapp and do actions based on it

  ```js
  useEffect(() => {
    web3wallet?.on('session_proposal', onSessionProposal);
    web3wallet?.on('session_request', onSessionRequest);
    web3wallet?.on('session_delete', onSessionDelete);
    return () => {
      web3wallet?.off('session_proposal', onSessionProposal);
      web3wallet?.off('session_request', onSessionRequest);
      web3wallet?.off('session_delete', onSessionDelete);
    }
  })
  ```
- Signing messages

  - Cosmos methods info

    - The signDoc format for signAmino and signDirect is different

      - Reference - https://docs.leapwallet.io/cosmos/for-dapps-connect-to-leap/api-reference

    - For signDirect, the message is protobuf encoded , hence to sign it , we have to first convert it to uint8Array

      ```js
      const bodyBytesArray = Uint8Array.from(
        Buffer.from(request.params.signDoc.bodyBytes, 'hex'),
      );
      const authInfoBytesArray = Uint8Array.from(
        Buffer.from(request.params.signDoc.authInfoBytes, 'hex'),
      );
      ```

    - This will give us correct signature

## patch-package

- <https://www.npmjs.com/package/patch-package>

- The `pbkdf2` function that `ethers.js` uses is implemented in javascript, so it is slow

- There we must patch the ethers package to use the `pbkdf2Sync` function of [react-native-quick-crypto](https://www.npmjs.com/package/react-native-quick-crypto) to improve performance

- The post-install script is used to achieve this task

  ```
  "postinstall": "patch-package"
  ```

  - This looks for `patches` folder in root of the project and looks for any patch files present and applies those patches

- Reference - <https://github.com/ethers-io/ethers.js/issues/2250#issuecomment-1195416579>

- Time (in ms) taken before patching

  ```
  LOG  Creating HD node, 53272.81766900001
  ```

- Time (in ms) taken after patching

  ```
  LOG  Creating HD node, 653.8724599992856
  ```

## Data structure of keystore

  ```json
  {
    // Accounts data -> hdpath, privateKey, publicKey, address
    "accounts/eip155:1/0":{
      "username": "",
      "password": "m/44'/60'/0'/0/0,0x0654623fe7a0e3d74f518e22818c1cfd58517e80a232df5d6d20a3afd99fd3bb,0x02cced178c903835bb29e337fa227ba0204a3285eb35797b766ed975b478b4beb6,0xe5fA0Dd92659e287e5e3Fa582Ee20fcf74fb1116"
    },
    "accounts/cosmos:cosmoshub-4/0":{
      "username": "",
      "password": "m/44'/118'/0'/0/0,0xbb9ccc3029178a61ba847c22108318ba119220451c99e6358efd7b85d7d49ed5,0x03a8002f56f99126f930ca3ac1963731e3f08df30822031db7c1dd50851bdfcc3c,cosmos1s5a5ls3d6xmks5fc8tst9x7tx2jv8mfjmn0aq8"
    },

    // Networks Data
    "networks":{
      "username":"",
      "password":[
        {
          "networkId":"1",
          "namespace": "eip155",
          "chainId": "1",
          "networkName": "Ethereum",
          "rpcUrl": "https://cloudflare-eth.com/",
          "currencySymbol": "ETH",
          "coinType": "60"
        },
        {
          "networkId":"2",
          "namespace": "cosmos",
          "chainId": "theta-testnet-001",
          "networkName": "Cosmos Hub Testnet",
          "rpcUrl": "https://rpc-t.cosmos.nodestake.top",
          "nativeDenom": "ATOM",
          "addressPrefix": "cosmos",
          "coinType": "118"
        }
      ]
    },

  // Stores count of total accounts for specific chain
    "addAccountCounter/eip155:1":{
      "username": "",
      "password": "3"
    },
    "addAccountCounter/cosmos:cosmoshub-4":{
      "username": "",
      "password": "3"
    },

    // Stores ids of accounts for specific chain
    "accountIndices/eip155:1":{
      "username": "",
      "password": "0,1,2"
    },
    "accountIndices/cosmos:cosmoshub-4":{
      "username": "",
      "password": "0,1,2"
    }
  }
  ```
