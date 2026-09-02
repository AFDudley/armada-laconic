// Taken from https://medium.com/walletconnect/how-to-build-a-wallet-in-react-native-with-the-web3wallet-sdk-b6f57bf02f9a

import { utils } from 'ethers';

import { ProposalTypes } from '@walletconnect/types';

import { Account, NetworksDataState } from '../../types';
import { EIP155_SIGNING_METHODS } from './EIP155Data';
import { mergeWith } from 'lodash';
import { retrieveAccounts } from '../accounts';
import { COSMOS, EIP155 } from '../constants';
import { NETWORK_METHODS } from './common-data';
import { COSMOS_METHODS } from './COSMOSData';

/**
 * Converts hex to utf8 string if it is valid bytes
 */
export function convertHexToUtf8(value: string) {
  if (utils.isHexString(value)) {
    return utils.toUtf8String(value);
  }

  return value;
}

/**
 * Gets message from various signing request methods by filtering out
 * a value that is not an address (thus is a message).
 * If it is a hex string, it gets converted to utf8 string
 */
export function getSignParamsMessage(params: string[]) {
  const message = params.filter(p => !utils.isAddress(p))[0];

  return convertHexToUtf8(message);
}

export const getNamespaces = async (
  optionalNamespaces: ProposalTypes.OptionalNamespaces,
  requiredNamespaces: ProposalTypes.RequiredNamespaces,
  networksData: NetworksDataState[],
  selectedNetwork: NetworksDataState,
  accounts: Account[],
) => {
  const namespaceChainId = `${selectedNetwork.namespace}:${selectedNetwork.chainId}`;

  const combinedNamespaces = mergeWith(
    requiredNamespaces,
    optionalNamespaces,
    (obj, src) =>
      Array.isArray(obj) && Array.isArray(src) ? [...src, ...obj] : undefined,
  );

  const walletConnectChains: string[] = [];

  Object.keys(combinedNamespaces).forEach(key => {
    const { chains } = combinedNamespaces[key];

    chains && walletConnectChains.push(...chains);
  });

  // If combinedNamespaces is not empty, send back namespaces object based on requested chains
  // Else send back namespaces object using currently selected network
  if (Object.keys(combinedNamespaces).length > 0) {
    if (!(walletConnectChains.length > 0)) {
      return;
    }
    // Check for unsupported chains
    const networkChains = networksData.map(
      network => `${network.namespace}:${network.chainId}`,
    );
    if (!walletConnectChains.every(chain => networkChains.includes(chain))) {
      const unsupportedChains = walletConnectChains.filter(
        chain => !networkChains.includes(chain),
      );
      throw new Error(`Unsupported chains : ${unsupportedChains.join(',')}`);
    }

    // Get required networks
    const requiredNetworks = networksData.filter(network =>
      walletConnectChains.includes(`${network.namespace}:${network.chainId}`),
    );
    // Get accounts for required networks
    const requiredAddressesPromise = requiredNetworks.map(
      async requiredNetwork => {
        const retrievedAccounts = await retrieveAccounts(requiredNetwork);

        if (!retrievedAccounts) {
          throw new Error('Accounts for given network not found');
        }

        const addresses = retrievedAccounts.map(
          retrieveAccount =>
            `${requiredNetwork.namespace}:${requiredNetwork.chainId}:${retrieveAccount.address}`,
        );

        return addresses;
      },
    );

    const requiredAddressesArray = await Promise.all(requiredAddressesPromise);
    const requiredAddresses = requiredAddressesArray.flat();

    // construct namespace object
    const newNamespaces = {
      eip155: {
        chains: walletConnectChains.filter(chain => chain.includes(EIP155)),
        // TODO: Debug optional namespace methods and events being required for approval
        methods: [
          ...Object.values(EIP155_SIGNING_METHODS),
          ...Object.values(NETWORK_METHODS),
          ...(optionalNamespaces.eip155?.methods ?? []),
          ...(requiredNamespaces.eip155?.methods ?? []),
        ],
        events: [
          ...(optionalNamespaces.eip155?.events ?? []),
          ...(requiredNamespaces.eip155?.events ?? []),
        ],
        accounts: requiredAddresses.filter(account => account.includes(EIP155)),
      },
      cosmos: {
        chains: walletConnectChains.filter(chain => chain.includes(COSMOS)),
        methods: [
          ...Object.values(COSMOS_METHODS),
          ...Object.values(NETWORK_METHODS),
          ...(optionalNamespaces.cosmos?.methods ?? []),
          ...(requiredNamespaces.cosmos?.methods ?? []),
        ],
        events: [
          ...(optionalNamespaces.cosmos?.events ?? []),
          ...(requiredNamespaces.cosmos?.events ?? []),
        ],
        accounts: requiredAddresses.filter(account => account.includes(COSMOS)),
      },
    };

    return newNamespaces;
  } else {
    switch (selectedNetwork.namespace) {
      case EIP155:
        return {
          eip155: {
            chains: [namespaceChainId],
            // TODO: Debug optional namespace methods and events being required for approval
            methods: [
              ...Object.values(EIP155_SIGNING_METHODS),
              ...Object.values(NETWORK_METHODS),
              ...(optionalNamespaces.eip155?.methods ?? []),
              ...(requiredNamespaces.eip155?.methods ?? []),
            ],
            events: [
              ...(optionalNamespaces.eip155?.events ?? []),
              ...(requiredNamespaces.eip155?.events ?? []),
            ],
            accounts: accounts.map(ethAccount => {
              return `${namespaceChainId}:${ethAccount.address}`;
            }),
          },
          cosmos: {
            chains: [],
            methods: [],
            events: [],
            accounts: [],
          },
        };
      case COSMOS:
        return {
          cosmos: {
            chains: [namespaceChainId],
            methods: [
              ...Object.values(COSMOS_METHODS),
              ...Object.values(NETWORK_METHODS),
              ...(optionalNamespaces.cosmos?.methods ?? []),
              ...(requiredNamespaces.cosmos?.methods ?? []),
            ],
            events: [
              ...(optionalNamespaces.cosmos?.events ?? []),
              ...(requiredNamespaces.cosmos?.events ?? []),
            ],
            accounts: accounts.map(cosmosAccount => {
              return `${namespaceChainId}:${cosmosAccount.address}`;
            }),
          },
          eip155: {
            chains: [],
            methods: [],
            events: [],
            accounts: [],
          },
        };
      default:
        break;
    }
  }
};
