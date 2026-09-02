import React, { useCallback } from 'react';
import { ScrollView, View } from 'react-native';
import { useForm, Controller, FieldErrors } from 'react-hook-form';
import { TextInput, Button, HelperText, Text } from 'react-native-paper';
import { setInternetCredentials } from 'react-native-keychain';
import { z } from 'zod';

import { zodResolver } from '@hookform/resolvers/zod';
import {
  NativeStackNavigationProp,
  NativeStackScreenProps,
} from '@react-navigation/native-stack';
import { useNavigation } from '@react-navigation/native';

import { StackParamsList } from '../types';
import styles from '../styles/stylesheet';
import { retrieveNetworksData } from '../utils/accounts';
import { useNetworks } from '../context/NetworksContext';
import {
  COSMOS,
  EIP155,
  EMPTY_FIELD_ERROR,
  INVALID_URL_ERROR,
} from '../utils/constants';

const ethNetworksFormSchema = z.object({
  // Adding type field for resolving typescript error
  type: z.literal(EIP155).optional(),
  networkName: z.string().nonempty({ message: EMPTY_FIELD_ERROR }),
  rpcUrl: z.string().url({ message: INVALID_URL_ERROR }),
  blockExplorerUrl: z
    .string()
    .url({ message: INVALID_URL_ERROR })
    .or(z.literal('')),
});

const cosmosNetworksFormDataSchema = z.object({
  type: z.literal(COSMOS).optional(),
  networkName: z.string().nonempty({ message: EMPTY_FIELD_ERROR }),
  rpcUrl: z.string().url({ message: INVALID_URL_ERROR }),
  blockExplorerUrl: z
    .string()
    .url({ message: INVALID_URL_ERROR })
    .or(z.literal('')),
  gasPrice: z
    .string()
    .nonempty({ message: EMPTY_FIELD_ERROR })
    .regex(/^\d+(\.\d+)?$/),
});

type EditNetworkProps = NativeStackScreenProps<StackParamsList, 'EditNetwork'>;

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
    mode: 'onChange',
    resolver: zodResolver(networksFormDataSchema),
  });

  const submit = useCallback(
    async (data: z.infer<typeof networksFormDataSchema>) => {
      const retrievedNetworksData = await retrieveNetworksData();
      const { type, ...dataWithoutType } = data;
      const newNetworkData = { ...networkData, ...dataWithoutType };
      const index = retrievedNetworksData.findIndex(
        network => network.networkId === networkData.networkId,
      );

      retrievedNetworksData.splice(index, 1, newNetworkData);

      await setInternetCredentials(
        'networks',
        '_',
        JSON.stringify(retrievedNetworksData),
      );

      setNetworksData(retrievedNetworksData);

      navigation.navigate('Laconic');
    },
    [networkData, navigation, setNetworksData],
  );

  return (
    <ScrollView contentContainerStyle={styles.signPage}>
      <View>
        <Text style={styles.subHeading}>
          Edit {networkData?.networkName} details
        </Text>
      </View>
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
              onChangeText={textValue => onChange(textValue)}
            />
            <HelperText type="error">{errors.networkName?.message}</HelperText>
          </>
        )}
      />
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
              onChangeText={textValue => onChange(textValue)}
            />
            <HelperText type="error">{errors.rpcUrl?.message}</HelperText>
          </>
        )}
      />

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
              onChangeText={textValue => onChange(textValue)}
            />
            <HelperText type="error">
              {errors.blockExplorerUrl?.message}
            </HelperText>
          </>
        )}
      />
      {networkData.namespace === COSMOS && (
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
      )}
      <Button
        mode="contained"
        loading={isSubmitting}
        disabled={isSubmitting}
        onPress={handleSubmit(submit)}>
        {isSubmitting ? 'Adding' : 'Submit'}
      </Button>
    </ScrollView>
  );
};

export default EditNetwork;
