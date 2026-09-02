import React, { useCallback, useEffect, useState } from 'react';
import { ScrollView } from 'react-native';
import { useForm, Controller, useWatch, FieldErrors } from 'react-hook-form';
import { TextInput, Button, HelperText } from 'react-native-paper';
import {
  getInternetCredentials,
  setInternetCredentials,
} from 'react-native-keychain';
import { HDNode } from 'ethers/lib/utils';
import { chains } from 'chain-registry';
import { useDebouncedCallback } from 'use-debounce';
import { z } from 'zod';

import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { useNavigation } from '@react-navigation/native';
import { zodResolver } from '@hookform/resolvers/zod';

import styles from '../styles/stylesheet';
import { StackParamsList } from '../types';
import { SelectNetworkType } from '../components/SelectNetworkType';
import { storeNetworkData } from '../utils/accounts';
import { useNetworks } from '../context/NetworksContext';
import {
  COSMOS,
  EIP155,
  CHAINID_DEBOUNCE_DELAY,
  EMPTY_FIELD_ERROR,
  INVALID_URL_ERROR,
  IS_NUMBER_REGEX,
} from '../utils/constants';
import { getCosmosAccounts } from '../utils/accounts';
import ETH_CHAINS from '../assets/ethereum-chains.json';

const ethNetworkDataSchema = z.object({
  chainId: z.string().nonempty({ message: EMPTY_FIELD_ERROR }),
  networkName: z.string().nonempty({ message: EMPTY_FIELD_ERROR }),
  rpcUrl: z.string().url({ message: INVALID_URL_ERROR }),
  blockExplorerUrl: z
    .string()
    .url({ message: INVALID_URL_ERROR })
    .or(z.literal('')),
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
    .or(z.literal('')),
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
    mode: 'onChange',
    resolver: zodResolver(networksFormDataSchema),
  });

  const watchChainId = useWatch({
    control,
    name: 'chainId',
  });

  const updateNetworkType = (newNetworkType: string) => {
    setNamespace(newNetworkType);
  };

  const fetchChainDetails = useDebouncedCallback((chainId: string) => {
    if (namespace === EIP155) {
      const ethChainDetails = ETH_CHAINS.find(
        chain => chain.chainId === Number(chainId),
      );
      if (!ethChainDetails) {
        return;
      }
      setValue('networkName', ethChainDetails.name);
      setValue('rpcUrl', ethChainDetails.rpc[0]);
      setValue('blockExplorerUrl', ethChainDetails.explorers?.[0].url || '');
      setValue('coinType', String(ethChainDetails.slip44 ?? '60'));
      setValue('currencySymbol', ethChainDetails.nativeCurrency.symbol);
      return;
    }
    const cosmosChainDetails = chains.find(
      ({ chain_id }) => chain_id === chainId,
    );
    if (!cosmosChainDetails) {
      return;
    }
    setValue('networkName', cosmosChainDetails.pretty_name);
    setValue('rpcUrl', cosmosChainDetails.apis?.rpc?.[0]?.address || '');
    setValue('blockExplorerUrl', cosmosChainDetails.explorers?.[0].url || '');
    setValue('addressPrefix', cosmosChainDetails.bech32_prefix);
    setValue('coinType', String(cosmosChainDetails.slip44 ?? '118'));
    setValue('nativeDenom', cosmosChainDetails.fees?.fee_tokens[0].denom || '');
    setValue(
      'gasPrice',
      String(
        cosmosChainDetails.fees?.fee_tokens[0].average_gas_price ||
          String(process.env.DEFAULT_GAS_PRICE),
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

      const mnemonicServer = await getInternetCredentials('mnemonicServer');
      const mnemonic = mnemonicServer && mnemonicServer.password;

      if (!mnemonic) {
        throw new Error('Mnemonic not found');
      }

      const hdNode = HDNode.fromMnemonic(mnemonic);

      const hdPath = `m/44'/${newNetworkData.coinType}'/0'/0/0`;
      const node = hdNode.derivePath(hdPath);
      let address;

      switch (newNetworkData.namespace) {
        case EIP155:
          address = node.address;
          break;

        case COSMOS:
          address = (
            await getCosmosAccounts(
              mnemonic,
              hdPath,
              (newNetworkData as z.infer<typeof cosmosNetworkDataSchema>)
                .addressPrefix,
            )
          ).data.address;
          break;

        default:
          throw new Error('Unsupported namespace');
      }

      const accountInfo = `${hdPath},${node.privateKey},${node.publicKey},${address}`;

      const retrievedNetworksData = await storeNetworkData(newNetworkData);
      setNetworksData(retrievedNetworksData);

      await Promise.all([
        setInternetCredentials(
          `accounts/${newNetworkData.namespace}:${newNetworkData.chainId}/0`,
          '_',
          accountInfo,
        ),
        setInternetCredentials(
          `addAccountCounter/${newNetworkData.namespace}:${newNetworkData.chainId}`,
          '_',
          '1',
        ),
        setInternetCredentials(
          `accountIndices/${newNetworkData.namespace}:${newNetworkData.chainId}`,
          '_',
          '0',
        ),
      ]);

      navigation.navigate('Laconic');
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
    <ScrollView contentContainerStyle={styles.signPage}>
      <SelectNetworkType updateNetworkType={updateNetworkType} />

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
              onChangeText={textValue => onChange(textValue)}
            />
            <HelperText type="error">{errors.chainId?.message}</HelperText>
          </>
        )}
      />
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
              onChangeText={textValue => onChange(textValue)}
            />
            <HelperText type="error">{errors.networkName?.message}</HelperText>
          </>
        )}
      />
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
              onChangeText={textValue => onChange(textValue)}
            />
            <HelperText type="error">{errors.rpcUrl?.message}</HelperText>
          </>
        )}
      />

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
              onChangeText={textValue => onChange(textValue)}
            />
            <HelperText type="error">
              {errors.blockExplorerUrl?.message}
            </HelperText>
          </>
        )}
      />
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
      {namespace === EIP155 ? (
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
                onChangeText={textValue => onChange(textValue)}
              />
              <HelperText type="error">
                {
                  (errors as FieldErrors<z.infer<typeof ethNetworkDataSchema>>)
                    .currencySymbol?.message
                }
              </HelperText>
            </>
          )}
        />
      ) : (
        <>
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
                  onChangeText={textValue => onChange(textValue)}
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
                  onChangeText={textValue => onChange(textValue)}
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
        </>
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

export default AddNetwork;
