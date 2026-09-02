import { PopulatedTransaction } from 'ethers';

import { SignClientTypes, SessionTypes } from '@walletconnect/types';
import { IWeb3Wallet, Web3WalletTypes } from '@walletconnect/web3wallet';
import { EncodeObject } from '@cosmjs/proto-signing';

export type StackParamsList = {
  Home: undefined;
  SignMessage: {
    selectedNamespace: string;
    selectedChainId: string;
    accountInfo: Account;
  };
  SignRequest: {
    namespace: string;
    chainId?: string;
    address: string;
    message: string;
    accountInfo?: Account;
    requestEvent?: Web3WalletTypes.SessionRequest;
    requestSessionData?: SessionTypes.Struct;
  };
  ApproveTransfer: {
    transaction: PopulatedTransaction;
    requestEvent: Web3WalletTypes.SessionRequest;
    requestSessionData: SessionTypes.Struct;
  };
  InvalidPath: undefined;
  WalletConnect: undefined;
  AddSession: undefined;
  AddNetwork: undefined;
  EditNetwork: {
    selectedNetwork: NetworksDataState;
  };
  ApproveTransaction: {
    transactionMessage: EncodeObject;
    signer: string;
    requestEvent: Web3WalletTypes.SessionRequest;
    requestSessionData: SessionTypes.Struct;
  };
  "wallet-embed": undefined;
  "auto-sign-in": undefined;
  "sign-message-request-embed": undefined;
  "sign-tx-request-embed": undefined;
};

export type Account = {
  index: number;
  pubKey: string;
  address: string;
  hdPath: string;
};

export type NetworkDropdownProps = {
  updateNetwork: (networksData: NetworksDataState) => void;
};

export type NetworksFormData = {
  networkName: string;
  rpcUrl: string;
  chainId: string;
  currencySymbol?: string;
  blockExplorerUrl?: string;
  namespace: string;
  nativeDenom?: string;
  addressPrefix?: string;
  coinType: string;
  gasPrice?: string;
  isDefault: boolean;
};

export interface NetworksDataState extends NetworksFormData {
  networkId: string;
}

export type SignMessageParams = {
  message: string;
  namespace: string;
  chainId: string;
  accountId: number;
};

export type CreateWalletProps = {
  isWalletCreating: boolean;
  createWalletHandler: () => Promise<void>;
};

export type ResetDialogProps = {
  title: string;
  visible: boolean;
  hideDialog: () => void;
  onConfirm: () => void;
};

export type HDPathDialogProps = {
  pathCode: string;
  visible: boolean;
  hideDialog: () => void;
  updateAccounts: (account: Account) => void;
};

export type CustomDialogProps = {
  visible: boolean;
  hideDialog: () => void;
  contentText: string;
  titleText?: string;
};

export type GridViewProps = {
  words: string[];
};

export type PathState = {
  firstNumber: string;
  secondNumber: string;
  thirdNumber: string;
};

export interface PairingModalProps {
  visible: boolean;
  setModalVisible: (arg1: boolean) => void;
  currentProposal:
    | SignClientTypes.EventArguments['session_proposal']
    | undefined;
  setCurrentProposal: (
    arg1:
      | SignClientTypes.EventArguments['session_proposal']
      | undefined,
  ) => void;
  setToastVisible: (arg1: boolean) => void;
}

export interface WalletConnectContextProps {
  activeSessions: Record<string, SessionTypes.Struct>;
  setActiveSessions: (
    activeSessions: Record<string, SessionTypes.Struct>,
  ) => void;
  web3wallet: IWeb3Wallet | undefined;
  setWeb3wallet: React.Dispatch<React.SetStateAction<IWeb3Wallet | undefined>>;
}
