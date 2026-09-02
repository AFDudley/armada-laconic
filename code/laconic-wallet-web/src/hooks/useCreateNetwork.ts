import { useEffect, useCallback } from "react";

import { addNewNetwork, createWallet, checkNetworkForChainID, isWalletCreated } from "../utils/accounts";
import { useNetworks } from "../context/NetworksContext";
import { NETWORK_ADDED_RESPONSE, NETWORK_ADD_FAILED_RESPONSE, NETWORK_ALREADY_EXISTS_RESPONSE, REQUEST_ADD_NETWORK } from "../utils/constants";
import { NetworksFormData } from "../types";
import { sendMessage } from "../utils/misc";

const REACT_APP_ALLOWED_URLS = import.meta.env.REACT_APP_ALLOWED_URLS;

const useCreateNetwork = () => {
  const { networksData, setNetworksData } = useNetworks();

  const getOrCreateNetwork = useCallback(
    async (chainId: string, networkData: NetworksFormData, sourceOrigin?: string) => {
      if (!sourceOrigin) {
        return;
      }

      try {
        const isCreated = await isWalletCreated();

        if (!isCreated) {
          console.log("Wallet not found, creating wallet...");
          await createWallet(networksData);
        }

        const isNetworkPresent = await checkNetworkForChainID(chainId);

        if (!isNetworkPresent) {
          console.log("ChainId not found. Adding network");

          const resolvedNetworkData = {
            chainId,
            namespace: networkData.namespace,
            networkName: networkData.networkName,
            rpcUrl: networkData.rpcUrl,
            blockExplorerUrl: networkData.blockExplorerUrl || "",
            addressPrefix: networkData.addressPrefix || "",
            coinType: networkData.coinType,
            nativeDenom: networkData.nativeDenom || "",
            gasPrice: networkData.gasPrice || String(import.meta.env.REACT_APP_DEFAULT_GAS_PRICE),
            currencySymbol: networkData.currencySymbol || "",
            isDefault: false,
          };

          const retrievedNetworksData = await addNewNetwork(resolvedNetworkData);
          setNetworksData(retrievedNetworksData);

          sendMessage(window.parent, NETWORK_ADDED_RESPONSE, {
            type: NETWORK_ADDED_RESPONSE,
            chainId
          }, sourceOrigin);
        } else {
          sendMessage(window.parent, NETWORK_ALREADY_EXISTS_RESPONSE, {
            type: NETWORK_ALREADY_EXISTS_RESPONSE,
            chainId
          }, sourceOrigin);
        }
      } catch (error) {
        console.error("Error in getOrCreateNetwork:", error);
        sendMessage(window.parent, NETWORK_ADD_FAILED_RESPONSE, {
          type: NETWORK_ADD_FAILED_RESPONSE,
          message: error instanceof Error ? error.message : "Unknown error"
        }, sourceOrigin);
      }
    },
    [networksData, setNetworksData]
  );

  useEffect(() => {
    const handleCreateNetwork = async (event: MessageEvent) => {
      if (event.data.type !== REQUEST_ADD_NETWORK) return;

      if (!REACT_APP_ALLOWED_URLS) {
        console.log("Allowed URLs are not set");
        return;
      }

      const allowedUrls = REACT_APP_ALLOWED_URLS.split(",").map(url => url.trim());

      if (!allowedUrls.includes(event.origin)) {
        console.log("Unauthorized app.");
        return;
      }

      const { chainId, networkData } = event.data;

      await getOrCreateNetwork(chainId, networkData, event.origin);
    };

    window.addEventListener("message", handleCreateNetwork);
    return () => {
      window.removeEventListener("message", handleCreateNetwork);
    };
  }, [getOrCreateNetwork]);
};

export default useCreateNetwork;
