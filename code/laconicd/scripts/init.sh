#!/bin/bash

if [[ -n "$CERC_SCRIPT_DEBUG" ]]; then
  set -x
fi

LACONIC_BIN=${LACONIC_BIN:-laconicd}
LACONIC_HOME="${LACONIC_HOME:-$HOME/.laconicd}"

KEYNAME="node_key"
CHAINID=${CHAINID:-"laconic_9000-1"}
MONIKER=${MONIKER:-"localtestnet"}
KEYRING=${KEYRING:-"test"}
DENOM=${DENOM:-"alnt"}
STAKING_AMOUNT=${STAKING_AMOUNT:-"1000000000000000"} # 10^15 alnt
ALLOCATION_AMOUNT=1000000000000000000000000000000    # 10^30 alnt | 10^12 lnt
LOGLEVEL=${LOGLEVEL:-"info"}

P2P_EXTERNAL_PORT=${P2P_EXTERNAL_PORT:-26656}

input_genesis_file=${GENESIS_FILE}
import_private_key=${PRIVATE_KEY}

laconicd="$LACONIC_BIN --home=$LACONIC_HOME --log_level=$LOGLEVEL"

set -e

if [ "$1" == "clean" ]; then
  echo "Clearing existing data directory in $LACONIC_HOME"
  rm -rf "$LACONIC_HOME"
else
  echo "Using existing database at $LACONIC_HOME. To replace, run '$(basename $0) clean'"
fi

# Create a new home dir if needed
if [[ ! -d "$LACONIC_HOME/data/application.db" ]]; then
  # Import the node key if passed, or generate one
  if [[ -n "$import_private_key" ]]; then
     $laconicd keys import-hex $KEYNAME 'b0b0000000000000000000000000000000000000000000000000000000000000'
  else
     printf "y\n" | $laconicd keys add $KEYNAME --keyring-backend $KEYRING --no-backup
  fi

  # Set moniker and chain-id (Moniker can be anything, chain-id must be an integer)
  $laconicd init $MONIKER --chain-id $CHAINID --default-denom $DENOM
else
  KEYNAME=$($laconicd keys list --list-names --keyring-backend $KEYRING | head -1)
fi

if [[ ! -d "$db_path" ]]; then
  $laconicd config set client chain-id $CHAINID
  $laconicd config set client keyring-backend $KEYRING

  if [[ -f ${input_genesis_file} ]]; then
    # Use provided genesis config
    cp $input_genesis_file $LACONIC_HOME/config/genesis.json
  fi

  $laconicd genesis validate

  update_comet_config() {       # comet config is not supported by confix cli
    local file="$LACONIC_HOME/config/config.toml"
    cp -f "$file" "${file}.bak" && tomlq --toml-output ".$1 = $2" "${file}.bak" > "$file"
  }

  # disable empty blocks
  update_comet_config consensus.create_empty_blocks 'false'

  # Allow requests from any origin
  update_comet_config rpc.cors_allowed_origins '["*"]'

  # Set persistent peers
  update_comet_config p2p.seeds "\"$P2P_PEERS\""
  update_comet_config p2p.persistent_peers "\"$P2P_PEERS\""
  update_comet_config p2p.external_address "\"tcp://127.0.0.1:${P2P_EXTERNAL_PORT}\""
  update_comet_config p2p.addr_book_strict false

  # Enable telemetry (prometheus metrics: http://localhost:1317/metrics?format=prometheus)
  $laconicd config set app telemetry.enable true
  $laconicd config set app telemetry.prometheus-retention-time 60
  update_comet_config instrumentation.prometheus true

  # Enable validator distsig
  $laconicd config set app distsig.enable true
  $laconicd config set app distsig.longterm-key "$KEYNAME"

  # Nitro config, WIP
  # $laconicd config set app nitro.use-distsig true # TODO
  $laconicd config set app nitro.eth-key "$KEYNAME"
  $laconicd config set app nitro.eth-url "$NITRO_ETH_URL"
  $laconicd config set app nitro.eth-start-block "$NITRO_ETH_START_BLOCK"
  $laconicd config set app nitro.eth-na-address "'$NITRO_ADJUDICATOR_ADDRESS'"
  $laconicd config set app nitro.eth-vpa-address "'$VIRTUAL_PAYMENT_APP_ADDRESS'"
  $laconicd config set app nitro.eth-ca-address "'$CONSENSUS_APP_ADDRESS'"
else
  echo "Using existing database at $LACONIC_HOME."
fi

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
# TODO new pruning config
$laconicd start \
  --server.minimum-gas-prices="1$DENOM" \
  --rpc.laddr="tcp://0.0.0.0:26657" \
  --grpc.address="127.0.0.1:9091" \
  --gql.enable --gql.playground
