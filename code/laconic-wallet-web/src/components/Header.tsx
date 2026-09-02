import {
  Button,
  Divider,
  Link,
  Stack,
  SvgIcon,
  Typography,
} from "@mui/material";
import React from "react";
import { useNavigation } from "@react-navigation/native";
import { Image } from "react-native";
import { NativeStackNavigationProp } from "@react-navigation/native-stack";
import { StackParamsList } from "../types";
import styles from "../styles/stylesheet";

const WCLogo = () => {
  return (
    <Image
      style={styles.walletConnectLogo}
      source={require("../assets/WalletConnect-Icon-Blueberry.png")}
    />
  );
};

export const Header: React.FC<{
  title: string;
  showWalletConnect?: boolean;
}> = ({ title, showWalletConnect }) => {
  const navigation =
    useNavigation<NativeStackNavigationProp<StackParamsList>>();
  return (
    <Stack
      direction="row"
      sx={{
        backgroundColor: "background.paper",
        pl: 2,
        alignItems: "center",
        py: 1,
        justifyContent: "space-between",
      }}
      spacing={1}
    >
      <Stack direction="row">
        <Button
          component={Link}
          onClick={() => {
            navigation.navigate("Home");
          }}
        >
          <SvgIcon sx={{ height: 20, width: 100 }}>
            <svg
              width="115"
              height="20"
              viewBox="0 0 115 20"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                fillRule="evenodd"
                clipRule="evenodd"
                d="M3.37388 10.5194C5.70149 8.19185 7.14225 4.97748 7.1416 1.42853C7.14246 0.94681 7.11586 0.470456 7.063 0L-0.000488281 0.000643078L-0.000273922 13.5723C-0.000917354 15.2174 0.62632 16.863 1.88091 18.1175C3.1356 19.3721 4.78235 20.0001 6.42772 19.9993L6.42729 19.9997L19.9995 20L19.999 12.9355C19.5296 12.8838 19.0532 12.857 18.5704 12.8569C15.0224 12.8574 11.8079 14.298 9.48026 16.6255C7.78654 18.2768 5.07093 18.2771 3.39812 16.6043C1.72638 14.9325 1.72562 12.2161 3.37388 10.5194ZM18.5344 1.46863C16.5837 -0.481929 13.4146 -0.48268 11.4633 1.46863C9.512 3.41984 9.51276 6.58895 11.4633 8.53941C13.415 10.491 16.5831 10.4907 18.5344 8.53941C20.4857 6.5882 20.4861 3.42016 18.5344 1.46863Z"
                fill="#FBFBFB"
              />
              <path
                d="M31.4741 18.5838H39.2552V16.3302H34.075V1.41351H31.4741V18.5838Z"
                fill="#FBFBFB"
              />
              <path
                d="M49.8108 1.41351H45.4976L40.9893 18.5838H43.6769L44.8039 14.2913H50.3744L51.5014 18.5838H54.3191L49.8108 1.41351ZM45.3458 12.145L47.6 3.2593H47.6866L49.8541 12.145H45.3458Z"
                fill="#FBFBFB"
              />
              <path
                d="M62.9292 8.06885H65.9636C65.9636 3.17534 64.3813 1.07196 60.6967 1.07196C56.8169 1.07196 55.1479 3.73341 55.1479 9.97909C55.1479 16.2462 56.8169 18.9291 60.6967 18.9291C64.3813 18.9291 65.9636 16.8901 65.9853 12.1468H62.9508C62.9292 15.8599 62.474 16.7828 60.6967 16.7828C58.6593 16.7828 58.1607 15.4307 58.1824 9.97909C58.1824 4.54896 58.6809 3.19678 60.6967 3.21823C62.474 3.21823 62.9292 4.18413 62.9292 8.06885Z"
                fill="#FBFBFB"
              />
              <path
                d="M73.7781 1.07209C77.7229 1.09364 79.4135 3.77643 79.4135 10.0007C79.4135 16.2249 77.7229 18.9078 73.7781 18.9292C69.8117 18.9507 68.1211 16.2678 68.1211 10.0007C68.1211 3.73354 69.8117 1.05064 73.7781 1.07209ZM71.1555 10.0007C71.1555 15.4308 71.6757 16.783 73.7781 16.783C75.8589 16.783 76.3791 15.4308 76.3791 10.0007C76.3791 4.54909 75.8589 3.19691 73.7781 3.21847C71.6757 3.23992 71.1555 4.59209 71.1555 10.0007Z"
                fill="#FBFBFB"
              />
              <path
                d="M85.0819 18.5624L82.481 18.5838V1.41351H87.0544L91.3243 15.4073H91.3676V1.41351H93.968V18.5838H89.677L85.1254 3.51689H85.0819V18.5624Z"
                fill="#FBFBFB"
              />
              <path
                d="M100.468 1.41351H97.8677V18.5838H100.468V1.41351Z"
                fill="#FBFBFB"
              />
              <path
                d="M111.139 8.06885H114.174C114.174 3.17534 112.591 1.07196 108.906 1.07196C105.028 1.07196 103.358 3.73341 103.358 9.97909C103.358 16.2462 105.028 18.9291 108.906 18.9291C112.591 18.9291 114.174 16.8901 114.195 12.1468H111.161C111.139 15.8599 110.684 16.7828 108.906 16.7828C106.869 16.7828 106.371 15.4307 106.393 9.97909C106.393 4.54896 106.891 3.19678 108.906 3.21823C110.684 3.21823 111.139 4.18413 111.139 8.06885Z"
                fill="#FBFBFB"
              />
            </svg>
          </SvgIcon>
        </Button>
        <Divider
          flexItem
          orientation="vertical"
          color="#FBFBFB"
          sx={{ height: "20px", width: "1px", alignSelf: "center" }}
        />
        <Typography fontSize="1.25rem" sx={{ paddingLeft: 1 }}>
          {title}
        </Typography>
      </Stack>

      {showWalletConnect && (
        <Button onClick={() => navigation.navigate("WalletConnect")}>
          {<WCLogo />}
        </Button>
      )}
    </Stack>
  );
};
