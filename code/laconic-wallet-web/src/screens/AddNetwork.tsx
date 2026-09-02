import React, { useCallback, useEffect, useState } from "react";
import { useForm, Controller, useWatch, FieldErrors } from "react-hook-form";
import { TextInput, HelperText } from "react-native-paper";
import { chains } from "chain-registry";
import { useDebouncedCallback } from "use-debounce";
import { z } from "zod";

import { NativeStackNavigationProp } from "@react-navigation/native-stack";
import { useNavigation } from "@react-navigation/native";
import { zodResolver } from "@hookform/resolvers/zod";
import { Divider, Grid } from "@mui/material";
import { LoadingButton } from "@mui/lab";

import { StackParamsList } from "../types";
import { SelectNetworkType } from "../components/SelectNetworkType";
import { addNewNetwork } from "../utils/accounts";
import { useNetworks } from "../context/NetworksContext";
import {
  EIP155,
  CHAINID_DEBOUNCE_DELAY,
  EMPTY_FIELD_ERROR,
  INVALID_URL_ERROR,
  IS_NUMBER_REGEX,
} from "../utils/constants";
import ETH_CHAINS from "../assets/ethereum-chains.json";
import { Layout } from "../components/Layout";

const ethNetworkDataSchema = z.object({
  chainId: z.string().nonempty({ message: EMPTY_FIELD_ERROR }),
  networkName: z.string().nonempty({ message: EMPTY_FIELD_ERROR }),
  rpcUrl: z.string().url({ message: INVALID_URL_ERROR }),
  blockExplorerUrl: z
    .string()
    .url({ message: INVALID_URL_ERROR })
    .or(z.literal("")),
  coinType: z
    .string()
    .nonempty({ message: EMPTY_FIELD_ERROR })
    .regex(IS_NUMBER_REGEX),
  currencySymbol: z.string().nonempty({ message: EMPTY_FIELD_ERROR }),
});

const cosmosNetworkDataSchema = z.object({
  chainId: z.string().nonempty({ message: EMPTY_FIELD_ERROR }),
  networkName: z.string().nonempty({ message: EMPTY_FIELD_ERROR }),
  rpcUrl: z.string().url({ message: INVALID_URL_ERROR }),
  blockExplorerUrl: z
    .string()
    .url({ message: INVALID_URL_ERROR })
    .or(z.literal("")),
  coinType: z
    .string()
    .nonempty({ message: EMPTY_FIELD_ERROR })
    .regex(IS_NUMBER_REGEX),
  nativeDenom: z.string().nonempty({ message: EMPTY_FIELD_ERROR }),
  addressPrefix: z.string().nonempty({ message: EMPTY_FIELD_ERROR }),
  gasPrice: z
    .string()
    .nonempty({ message: EMPTY_FIELD_ERROR })
    .regex(/^\d+(\.\d+)?$/),
});

const AddNetwork = () => {
  const navigation =
    useNavigation<NativeStackNavigationProp<StackParamsList>>();

  const { setNetworksData } = useNetworks();

  const [namespace, setNamespace] = useState<string>(EIP155);

  const networksFormDataSchema =
    namespace === EIP155 ? ethNetworkDataSchema : cosmosNetworkDataSchema;

  const {
    control,
    formState: { errors, isSubmitting },
    handleSubmit,
    setValue,
    reset,
  } = useForm<z.infer<typeof networksFormDataSchema>>({
    mode: "onChange",
    resolver: zodResolver(networksFormDataSchema),
  });

  const watchChainId = useWatch({
    control,
    name: "chainId",
  });

  const updateNetworkType = (newNetworkType: string) => {
    setNamespace(newNetworkType);
  };

  const fetchChainDetails = useDebouncedCallback((chainId: string) => {
    if (namespace === EIP155) {
      const ethChainDetails = ETH_CHAINS.find(
        (chain) => chain.chainId === Number(chainId),
      );
      if (!ethChainDetails) {
        return;
      }
      setValue("networkName", ethChainDetails.name);
      setValue("rpcUrl", ethChainDetails.rpc[0]);
      setValue("blockExplorerUrl", ethChainDetails.explorers?.[0].url || "");
      setValue("coinType", String(ethChainDetails.slip44 ?? "60"));
      setValue("currencySymbol", ethChainDetails.nativeCurrency.symbol);
      return;
    }
    const cosmosChainDetails = chains.find(
      ({ chain_id }) => chain_id === chainId,
    );
    if (!cosmosChainDetails) {
      return;
    }
    setValue("networkName", cosmosChainDetails.pretty_name);
    setValue("rpcUrl", cosmosChainDetails.apis?.rpc?.[0]?.address || "");
    setValue("blockExplorerUrl", cosmosChainDetails.explorers?.[0].url || "");
    setValue("addressPrefix", cosmosChainDetails.bech32_prefix);
    setValue("coinType", String(cosmosChainDetails.slip44 ?? "118"));
    setValue("nativeDenom", cosmosChainDetails.fees?.fee_tokens[0].denom || "");
    setValue(
      "gasPrice",
      String(
        cosmosChainDetails.fees?.fee_tokens[0].average_gas_price ||
          String(import.meta.env.REACT_APP_DEFAULT_GAS_PRICE),
      ),
    );
  }, CHAINID_DEBOUNCE_DELAY);

  const submit = useCallback(
    async (data: z.infer<typeof networksFormDataSchema>) => {
      const newNetworkData = {
        ...data,
        namespace,
        isDefault: false,
      };

      const retrievedNetworksData = await addNewNetwork(newNetworkData);

      setNetworksData(retrievedNetworksData);

      navigation.navigate("Home");
    },
    [navigation, namespace, setNetworksData],
  );

  useEffect(() => {
    fetchChainDetails(watchChainId);
  }, [watchChainId, fetchChainDetails]);

  useEffect(() => {
    reset();
  }, [namespace, reset]);

  return (
    <Layout title="Add Network">
      <SelectNetworkType updateNetworkType={updateNetworkType} />
      <Divider flexItem sx={{ my: 4 }} />

      <Grid container spacing={2} sx={{ px: 1 }}>
        <Grid item xs={6}>
          <Controller
            control={control}
            name="chainId"
            defaultValue=""
            render={({ field: { onChange, onBlur, value } }) => (
              <>
                <TextInput
                  mode="outlined"
                  value={value}
                  label="Chain ID"
                  onBlur={onBlur}
                  onChangeText={(textValue) => onChange(textValue)}
                />
                <HelperText type="error">{errors.chainId?.message}</HelperText>
              </>
            )}
          />
        </Grid>

        <Grid item xs={6}>
          <Controller
            control={control}
            defaultValue=""
            name="networkName"
            render={({ field: { onChange, onBlur, value } }) => (
              <>
                <TextInput
                  mode="outlined"
                  label="Network Name"
                  value={value}
                  onBlur={onBlur}
                  onChangeText={(textValue) => onChange(textValue)}
                />
                <HelperText type="error">
                  {errors.networkName?.message}
                </HelperText>
              </>
            )}
          />
        </Grid>

        <Grid item xs={6}>
          <Controller
            control={control}
            name="rpcUrl"
            defaultValue=""
            render={({ field: { onChange, onBlur, value } }) => (
              <>
                <TextInput
                  mode="outlined"
                  label="New RPC URL"
                  onBlur={onBlur}
                  value={value}
                  onChangeText={(textValue) => onChange(textValue)}
                />
                <HelperText type="error">{errors.rpcUrl?.message}</HelperText>
              </>
            )}
          />
        </Grid>

        <Grid item xs={6}>
          <Controller
            control={control}
            defaultValue=""
            name="blockExplorerUrl"
            render={({ field: { onChange, onBlur, value } }) => (
              <>
                <TextInput
                  mode="outlined"
                  value={value}
                  label="Block Explorer URL  (Optional)"
                  onBlur={onBlur}
                  onChangeText={(textValue) => onChange(textValue)}
                />
                <HelperText type="error">
                  {errors.blockExplorerUrl?.message}
                </HelperText>
              </>
            )}
          />
        </Grid>
        <Grid item xs={namespace === EIP155 ? 12 : 6}>
          <Controller
            control={control}
            name="coinType"
            defaultValue=""
            render={({ field: { onChange, onBlur, value } }) => (
              <>
                <TextInput
                  mode="outlined"
                  value={value}
                  label="Coin Type"
                  onBlur={onBlur}
                  onChangeText={onChange}
                />
                <HelperText type="error">{errors.coinType?.message}</HelperText>
              </>
            )}
          />
        </Grid>
        {namespace === EIP155 ? (
          <Grid item xs={12}>
            <Controller
              control={control}
              name="currencySymbol"
              defaultValue=""
              render={({ field: { onChange, onBlur, value } }) => (
                <>
                  <TextInput
                    mode="outlined"
                    value={value}
                    label="Currency Symbol"
                    onBlur={onBlur}
                    onChangeText={(textValue) => onChange(textValue)}
                  />
                  <HelperText type="error">
                    {
                      (
                        errors as FieldErrors<
                          z.infer<typeof ethNetworkDataSchema>
                        >
                      ).currencySymbol?.message
                    }
                  </HelperText>
                </>
              )}
            />
          </Grid>
        ) : (
          <>
            <Grid item xs={6}>
              <Controller
                control={control}
                name="nativeDenom"
                defaultValue=""
                render={({ field: { onChange, onBlur, value } }) => (
                  <>
                    <TextInput
                      mode="outlined"
                      value={value}
                      label="Native Denom"
                      onBlur={onBlur}
                      onChangeText={(textValue) => onChange(textValue)}
                    />
                    <HelperText type="error">
                      {
                        (
                          errors as FieldErrors<
                            z.infer<typeof cosmosNetworkDataSchema>
                          >
                        ).nativeDenom?.message
                      }
                    </HelperText>
                  </>
                )}
              />
            </Grid>
            <Grid item xs={6}>
              <Controller
                control={control}
                name="addressPrefix"
                defaultValue=""
                render={({ field: { onChange, onBlur, value } }) => (
                  <>
                    <TextInput
                      mode="outlined"
                      value={value}
                      label="Address Prefix"
                      onBlur={onBlur}
                      onChangeText={(textValue) => onChange(textValue)}
                    />
                    <HelperText type="error">
                      {
                        (
                          errors as FieldErrors<
                            z.infer<typeof cosmosNetworkDataSchema>
                          >
                        ).addressPrefix?.message
                      }
                    </HelperText>
                  </>
                )}
              />
            </Grid>
            <Grid item xs={6}>
              <Controller
                control={control}
                name="gasPrice"
                defaultValue=""
                render={({ field: { onChange, onBlur, value } }) => (
                  <>
                    <TextInput
                      mode="outlined"
                      value={value}
                      label="Gas Price"
                      onBlur={onBlur}
                      onChangeText={onChange}
                    />
                    <HelperText type="error">
                      {
                        (
                          errors as FieldErrors<
                            z.infer<typeof cosmosNetworkDataSchema>
                          >
                        ).gasPrice?.message
                      }
                    </HelperText>
                  </>
                )}
              />
            </Grid>
          </>
        )}
      </Grid>
      <LoadingButton
        variant="contained"
        loading={isSubmitting}
        disabled={isSubmitting}
        onClick={handleSubmit(submit)}
        sx={{ minWidth: "200px", px: 4, py: 1, mt: 2 }}
      >
        {isSubmitting ? "Adding" : "Submit"}
      </LoadingButton>
    </Layout>
  );
};

export default AddNetwork;
