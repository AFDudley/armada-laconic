#!/bin/bash

# Default value for IS_RELEASE
IS_RELEASE=${IS_RELEASE:-false}

# Install dependencies
echo "Installing dependencies..."
yarn

# Create the necessary directory for assets
mkdir -p android/app/src/main/assets/

# Bundle the React Native application
yarn react-native bundle \
  --platform android \
  --dev false \
  --entry-file index.js \
  --bundle-output android/app/src/main/assets/index.android.bundle \
  --assets-dest android/app/src/main/res

# Navigate to the android directory
cd android

# Run the Gradle build based on the IS_RELEASE flag
if [ "$IS_RELEASE" = "true" ]; then
  echo "Building release version..."
  ./gradlew assembleRelease
else
  echo "Building debug version..."
  ./gradlew assembleDebug
fi
