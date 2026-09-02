// Reference: https://github.com/lugg/react-native-config?tab=readme-ov-file#typescript-declaration-for-your-env-file
declare module 'react-native-config' {
  export interface NativeConfig {
    WALLET_CONNECT_PROJECT_ID: string;
    DEFAULT_GAS_PRICE: string;
    DEFAULT_GAS_ADJUSTMENT: string;
    LACONICD_RPC_URL: string;
  }

  export const Config: NativeConfig;
  export default Config;
}
