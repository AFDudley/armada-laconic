import React, { useCallback } from "react";
import { useForm, Controller, FieldErrors } from "react-hook-form";
import { TextInput, HelperText } from "react-native-paper";
import { z } from "zod";

import { zodResolver } from "@hookform/resolvers/zod";
import {
  NativeStackNavigationProp,
  NativeStackScreenProps,
} from "@react-navigation/native-stack";
import { useNavigation } from "@react-navigation/native";

import { setInternetCredentials } from "../utils/key-store";
import { StackParamsList } from "../types";
import { retrieveNetworksData } from "../utils/accounts";
import { useNetworks } from "../context/NetworksContext";
import {
  COSMOS,
  EIP155,
  EMPTY_FIELD_ERROR,
  INVALID_URL_ERROR,
} from "../utils/constants";
import { Divider, Grid, Typography } from "@mui/material";
import { LoadingButton } from "@mui/lab";
import { Layout } from "../components/Layout";

const ethNetworksFormSchema = z.object({
  // Adding type field for resolving typescript error
  type: z.literal(EIP155).optional(),
  networkName: z.string().nonempty({ message: EMPTY_FIELD_ERROR }),
  rpcUrl: z.string().url({ message: INVALID_URL_ERROR }),
  blockExplorerUrl: z
    .string()
    .url({ message: INVALID_URL_ERROR })
    .or(z.literal("")),
});

const cosmosNetworksFormDataSchema = z.object({
  type: z.literal(COSMOS).optional(),
  networkName: z.string().nonempty({ message: EMPTY_FIELD_ERROR }),
  rpcUrl: z.string().url({ message: INVALID_URL_ERROR }),
  blockExplorerUrl: z
    .string()
    .url({ message: INVALID_URL_ERROR })
    .or(z.literal("")),
  gasPrice: z
    .string()
    .nonempty({ message: EMPTY_FIELD_ERROR })
    .regex(/^\d+(\.\d+)?$/),
});

type EditNetworkProps = NativeStackScreenProps<StackParamsList, "EditNetwork">;

const EditNetwork = ({ route }: EditNetworkProps) => {
  const { setNetworksData } = useNetworks();
  const navigation =
    useNavigation<NativeStackNavigationProp<StackParamsList>>();

  const networkData = route.params.selectedNetwork;

  const networksFormDataSchema =
    networkData.namespace === COSMOS
      ? cosmosNetworksFormDataSchema
      : ethNetworksFormSchema;

  const {
    control,
    formState: { errors, isSubmitting },
    handleSubmit,
  } = useForm<z.infer<typeof networksFormDataSchema>>({
    mode: "onChange",
    resolver: zodResolver(networksFormDataSchema),
  });

  const submit = useCallback(
    async (data: z.infer<typeof networksFormDataSchema>) => {
      const retrievedNetworksData = await retrieveNetworksData();
      const { type, ...dataWithoutType } = data;
      const newNetworkData = { ...networkData, ...dataWithoutType };
      const index = retrievedNetworksData.findIndex(
        (network) => network.networkId === networkData.networkId,
      );

      retrievedNetworksData.splice(index, 1, newNetworkData);

      await setInternetCredentials(
        "networks",
        "_",
        JSON.stringify(retrievedNetworksData),
      );

      setNetworksData(retrievedNetworksData);

      navigation.navigate("Home");
    },
    [networkData, navigation, setNetworksData],
  );
  const isCosmos = networkData.namespace === COSMOS;

  return (
    <Layout title="Edit Network">
      <Typography fontSize="1.5rem" fontWeight={500}>
        Edit {networkData?.networkName} details
      </Typography>
      <Divider flexItem sx={{ my: 4 }} />
      <Grid container spacing={2}>
        <Grid item xs={6}>
          <Controller
            control={control}
            defaultValue={networkData.networkName}
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
            defaultValue={networkData.rpcUrl}
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

        <Grid item xs={isCosmos ? 6 : 12}>
          <Controller
            control={control}
            defaultValue={networkData.blockExplorerUrl}
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
        {isCosmos && (
          <Grid item xs={6}>
            <Controller
              control={control}
              name="gasPrice"
              defaultValue={networkData.gasPrice}
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
                          z.infer<typeof cosmosNetworksFormDataSchema>
                        >
                      ).gasPrice?.message
                    }
                  </HelperText>
                </>
              )}
            />
          </Grid>
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

export default EditNetwork;
