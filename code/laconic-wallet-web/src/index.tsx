import React from "react";
import ReactDOM from "react-dom/client";
import {
  PaperProvider,
  MD3DarkTheme as DefaultTheme,
} from "react-native-paper";
import { NavigationContainer, DarkTheme } from "@react-navigation/native";
import { Platform } from "react-native";
import { Buffer } from "buffer";

import "./index.css";
import App from "./App";
import { AccountsProvider } from "./context/AccountsContext";
import { NetworksProvider } from "./context/NetworksContext";
import reportWebVitals from "./reportWebVitals";
import { WalletConnectProvider } from "./context/WalletConnectContext";
import { createTheme, ThemeProvider } from "@mui/material";

globalThis.Buffer = Buffer;

const linking = {
  prefixes: ["https://wallet.laconic.com"],
};

const theme = {
  ...DefaultTheme,
  dark: true,
};

const navigationTheme: typeof DarkTheme = {
  ...DarkTheme,
  colors: {
    ...DarkTheme.colors,
    primary: "#0000F4",
    background: "#0F0F0F",
    card: "#18181A",
  },
};

const muiTheme = createTheme({
  components: {
    MuiAccordion: {
      defaultProps: {
        sx: {
          border: "1px solid #48474F",
          borderBottomRightRadius: 3,
          borderBottomLeftRadius: 3,
        },
      },
    },
    MuiButton: {
      defaultProps: {
        color: "primary",
        sx: {
          fontFamily: `DM Mono, monospace`,
          fontWeight: 400,
        },
      },
    },
    MuiLink: {
      defaultProps: {
        color: "text.primary",
        fontSize: "14px",
      },
    },
    MuiTypography: {
      defaultProps: {
        color: "text.primary",
        fontWeight: 400,
      },
    },
    MuiPaper: {
      defaultProps: {
        sx: {
          backgroundImage: "none",
        },
      },
    },
  },
  palette: {
    mode: "dark",
    primary: {
      main: "#0000F4",
    },
    secondary: {
      main: "#A2A2FF",
    },
    error: {
      main: "#B20710",
    },
    background: {
      default: "#0F0F0F",
      paper: "#18181A",
    },
    text: {
      primary: "#FBFBFB",
    },
    info: {
      main: "#FBFBFB",
    },
  },
});

const root = ReactDOM.createRoot(
  document.getElementById("root") as HTMLElement,
);
root.render(
  <div id="app">
    <PaperProvider theme={theme}>
      <NetworksProvider>
        <AccountsProvider>
          <WalletConnectProvider>
            <NavigationContainer
              linking={linking}
              documentTitle={{
                formatter: () => `Laconic | Wallet`,
              }}
              theme={navigationTheme}
            >
              <React.Fragment>
                {Platform.OS === "web" ? (
                  <style type="text/css">{`
                  @font-face {
                    font-family: 'MaterialCommunityIcons';
                    src: url(${require("react-native-vector-icons/Fonts/MaterialCommunityIcons.ttf")}) format('truetype');
                  }
                `}</style>
                ) : null}
                <ThemeProvider theme={muiTheme}>
                  <App />
                </ThemeProvider>
              </React.Fragment>
            </NavigationContainer>
          </WalletConnectProvider>
        </AccountsProvider>
      </NetworksProvider>
    </PaperProvider>
  </div>,
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
