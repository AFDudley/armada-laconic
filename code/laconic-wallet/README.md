# laconic-wallet

## Install

- Install [Node](https://nodejs.org/en/download/package-manager/)

- Install [version 17](https://download.oracle.com/java/17/latest/jdk-17_linux-x64_bin.deb) of the Java SE Development Kit (JDK) (Clicking on the link will download the .deb file)

  - Install using the `.deb` file
    ```bash
    sudo dpkg -i jdk-17_linux-x64_bin.deb
    ```

  - Check Oracle JDK docs for installation in other platforms: https://docs.oracle.com/en/java/javase/17/install/overview-jdk-installation.html

- Install [Android Studio](https://developer.android.com/studio/index.html)

- In the installation wizard check the following items:

  - Android SDK

  - Android SDK Platform

  - Android Virtual Device

- Install Android SDK

  - Open Android Studio -> Configure -> SDK Manager -> SDK Platform Tab.

  - Check the box next to "Show Package Details" in the bottom right corner. Look for and expand the Android 13 (Tiramisu) entry, then make sure the following items are checked:

    - Android SDK Platform 33

    - Intel x86 Atom_64 System Image or Google APIs Intel x86 Atom System Image

  - Select SDK Tools and check the box next to "Show Package Details"

  - Select Android SDK Build-Tools 33.0.0

  - Click Apply

- Configure the `ANDROID_HOME` environment variable

- Add the following lines to your `$HOME/.bash_profile` or `$HOME/.bashrc` (if you are using zsh then `~/.zprofile` or `~/.zshrc`) config file:

  ```bash
  export ANDROID_HOME=$HOME/Android/Sdk
  export PATH=$PATH:$ANDROID_HOME/emulator
  export PATH=$PATH:$ANDROID_HOME/platform-tools
  ```

- Type `source $HOME/.bash_profile` for bash or `source $HOME/.zprofile` to load the config into your current shell. Verify that `ANDROID_HOME` has been set by running

  ```bash
  echo $ANDROID_HOME
  ```

- Check that the appropriate directories have been added to your path by running

  ```bash
  echo $PATH
  ```

## Setup for laconic-wallet

1. Clone the repository

   ```
   git clone git@git.vdb.to:cerc-io/laconic-wallet.git
   ```

2. Enter the project directory

   ```
   cd laconic-wallet
   ```

3. Install dependencies

   ```
   yarn
   ```

4. Setup .env
   - Copy and update [`.env`](./.env)

     ```
     cp .env.example .env
     ```

   - In the `.env` file add your WalletConnect project id. You can generate your own ProjectId at https://cloud.walletconnect.com

     ```
     WALLET_CONNECT_PROJECT_ID=39bc93c...
     ```

5. Add SDK directory to project
   - Inside the [`android`](./android/) directory, create a file `local.properties` and add your Android SDK path
     ```
     sdk.dir = /home/USERNAME/Android/Sdk
     ```
     Where `USERNAME` is your linux username

6. Set up the Android device

   - For a physical device, refer to the [React Native documentation for running on a physical device](https://reactnative.dev/docs/running-on-device)

   - For a virtual device, continue with the steps

7. Start the application

   ```
   yarn start
   ```

8. Press `a` to run the application on android

## Flow for the app

- User scans QR Code on dApp from wallet to connect

- After clicking on pair button, dApp emits an event 'session_proposal'

- Wallet listens to this event and opens a modal to either accept or reject the proposal

  - Modal shows information about methods, chains and events the dApp is requesting for

    - This information is taken from [namespaces](https://docs.walletconnect.com/advanced/glossary#namespaces) object received from dApp

- On accepting, wallet sends the [namespaces](https://docs.walletconnect.com/advanced/glossary#namespaces) object with information about the accounts that are present in the wallet in response

- Once the [session](https://docs.walletconnect.com/advanced/glossary#session) is established, it is shown on the active sessions page

- The wallet can be connected to multiple dApps

## signature-requester-app

### Setup

1. Clone the repository

   ```
   git clone git@git.vdb.to:cerc-io/signature-requester-app.git
   ```

2. Enter the project directory

   ```
   cd signature-requester-app
   ```

3. Install dependencies
    ```
    yarn
    ```
4. Start the application

   ```
   yarn start
   ```
5. Press `a` to run the application on android

You should see both the apps running on your emulator or physical device

### Run

1. Open the app

2. Add details about chain and address you wish to sign the message with

    - You can see the URL based on the data you entered

3. Click on `Open Link` button to open Laconic Wallet and sign message

    - If this does not work, follow steps below

      - Open Laconic Wallet app settings

      - Open `Open by default` section

      - Click on `Add link`

      - You should see `wallet.laconic.com` link

      - Once selected, signature-requester-app should work as intended

## Troubleshooting

- To clean the build

  ```
  cd android

  ./gradlew clean
  ```
