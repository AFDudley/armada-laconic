#!/usr/bin/env bash

set -e

# Enter the proto files dir
cd proto

echo "Generating gogo proto code"
proto_dirs=$(find . -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    # Check if the go_package in the file is pointing to laconicd
    if grep -q "option go_package.*laconicd" "$file"; then
      buf generate --template buf.gen.gogo.yaml "$file"
    fi
  done
done

echo "Generating pulsar proto code"
buf generate --template buf.gen.pulsar.yaml

# Go back to root dir
cd ..

# Copy over the generated files and cleanup
cp -r git.vdb.to/cerc-io/laconicd/* ./
rm -rf git.vdb.to

go mod tidy
