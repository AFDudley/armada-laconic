FROM golang:1.23.6-bookworm AS builder

# Set working directory for the build
WORKDIR /go/src/git.vdb.to/cerc-io/laconicd

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Add source files
COPY . .

# Make the binary
RUN make build

# Final image
# DEV golang image
FROM golang:1.23.6-bookworm AS base

# Install ca-certificates, jq, curl, bash, and other necessary packages
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    jq curl netcat-openbsd bash bc yq \
    && rm -rf /var/lib/apt/lists/*

# FROM base

# # Copy over binary from the builder
# COPY --from=builder /go/src/git.vdb.to/cerc-io/laconicd/build/laconicd /usr/bin/laconicd

# #  Copy over init script from builder
# COPY --from=builder /go/src/git.vdb.to/cerc-io/laconicd/scripts/init.sh scripts/init.sh

# WORKDIR /

# Run laconicd by default
ENTRYPOINT ["laconicd"]
