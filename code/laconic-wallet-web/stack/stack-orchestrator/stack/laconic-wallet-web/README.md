# laconic-wallet-web

Instructions for running the `laconic-wallet-web` using [laconic-so](https://git.vdb.to/cerc-io/stack-orchestrator)

## Setup

* Clone the stack repo:

  ```bash
  laconic-so fetch-stack git.vdb.to/LaconicNetwork/laconic-wallet-web
  ```

* Build the container image:

  ```bash
  laconic-so --stack ~/cerc/laconic-wallet-web/stack/stack-orchestrator/stack/laconic-wallet-web build-containers
  ```

  This should create the `cerc/laconic-wallet-web` image locally

## Create a deployment

* Create a spec file for the deployment:

  ```bash
  laconic-so --stack ~/cerc/laconic-wallet-web/stack/stack-orchestrator/stack/laconic-wallet-web deploy init --output laconic-wallet-web-spec.yml
  ```

* Edit `network` in the spec file to map container ports to host ports as required:

  ```bash
  network:
    ports:
      laconic-wallet-web:
        - '3000:80'
  ```

* Create a deployment from the spec file:

  ```bash
  laconic-so --stack ~/cerc/laconic-wallet-web/stack/stack-orchestrator/stack/laconic-wallet-web deploy create --spec-file laconic-wallet-web-spec.yml --deployment-dir laconic-wallet-web-deployment
  ```

## Configuration

* Inside the `laconic-wallet-web-deployment` deployment directory, open `config.env` file and set following env variables:

  ```bash
  # WalletConnect project ID, same should be used in the laconic-wallet
  WALLET_CONNECT_ID=

  # Allowed urls is a comma separated list of allowed urls
  CERC_ALLOWED_URLS=

  # Optional

  # WalletConnect code for hostname verification
  WALLET_CONNECT_VERIFY_CODE=

  # Default gas price for txs (default: 0.025)
  CERC_DEFAULT_GAS_PRICE=

  # Gas adjustment (default: 2)
  # Reference: https://github.com/cosmos/cosmos-sdk/issues/16020
  CERC_GAS_ADJUSTMENT=

  # RPC endpoint of laconicd node (default: https://laconicd.laconic.com)
  CERC_LACONICD_RPC_URL=https://laconicd-mainnet-1.laconic.com
  ```

## Start the deployment

```bash
laconic-so deployment --dir laconic-wallet-web-deployment start
```

Open the wallet app in a browser at <http://localhost:3000>

## Clean up

* Stop the deployment:

  ```bash
  laconic-so deployment --dir laconic-wallet-web-deployment stop
  ```
