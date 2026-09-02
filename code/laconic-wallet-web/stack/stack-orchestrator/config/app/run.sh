#!/bin/bash

set -e
if [ -n "$CERC_SCRIPT_DEBUG" ]; then
  set -x
fi

echo "Using the following env variables:"
echo "WALLET_CONNECT_ID: ${WALLET_CONNECT_ID}"
echo "CERC_DEFAULT_GAS_PRICE: ${CERC_DEFAULT_GAS_PRICE}"
echo "CERC_GAS_ADJUSTMENT: ${CERC_GAS_ADJUSTMENT}"
echo "CERC_LACONICD_RPC_URL: ${CERC_LACONICD_RPC_URL}"
echo "CERC_ALLOWED_URLS: ${CERC_ALLOWED_URLS}"

# Build with required env
export REACT_APP_WALLET_CONNECT_PROJECT_ID=$WALLET_CONNECT_ID
export REACT_APP_DEFAULT_GAS_PRICE=$CERC_DEFAULT_GAS_PRICE
export REACT_APP_GAS_ADJUSTMENT=$CERC_GAS_ADJUSTMENT
export REACT_APP_LACONICD_RPC_URL=$CERC_LACONICD_RPC_URL
export REACT_APP_ALLOWED_URLS=$CERC_ALLOWED_URLS

# Set env variables in build
yarn set-env

# Define the directory and file path
FILE_PATH="/app/build/.well-known/walletconnect.txt"

# Create the directory if it doesn't exist
mkdir -p "$(dirname "$FILE_PATH")"
# Write verification code to the file
echo "$WALLET_CONNECT_VERIFY_CODE" > "$FILE_PATH"

# Serve build dir with explicit config
serve -s -l 80 -c /app/serve.json /app/build
