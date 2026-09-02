# cerc-io laconic gql

> Browser : <http://localhost:9473> for gql

## Run gqlgen

On having some change in the GQL schema (for example: adding a new query) update the [schema.graphql](./cerc-io/laconicd/schema.graphql) file as required and run the following:

- Generate GQL bindings:

  ```bash
  # Install gqlgen if not present
  go get github.com/99designs/gqlgen@v0.17.22

  # Generate bindings
  cd gql
  go run github.com/99designs/gqlgen generate
  ```

- Implement the required resolver method(s)

- Regenerate GQL bindings:

  ```bash
  # Generate bindings
  go run github.com/99designs/gqlgen generate
  ```

## Start server

```shell
laconicd start --gql.playground --gql.server
```

Basic node status:

```graphql
{
  getStatus {
    version
    node {
      id
      network
      moniker
    }
    sync {
      latestBlockHeight
      catchingUp
    }
    numPeers
    peers {
      isOutbound
      remoteIp
    }
    diskUsage
  }
}
```

Full node status:

```graphql
{
  getStatus {
    version
    node {
      id
      network
      moniker
    }
    sync {
      latestBlockHash
      latestBlockTime
      latestBlockHeight
      catchingUp
    }
    validator {
      address
      votingPower
      proposerPriority
    }
    validators {
      address
      votingPower
      proposerPriority
    }
    numPeers
    peers {
      node {
        id
        network
        moniker
      }
      isOutbound
      remoteIp
    }
    diskUsage
  }
}
```

Get records by IDs.

```graphql
{
  getRecordsByIds(ids: ["bafyreigswvbm4dbpnbwkyegrcxd6kynqe4bivflo6esedgxnuofpzonwdy"]) {
    id
    names
    bondId
    createTime
    expiryTime
    owners
    attributes {
      key
      value {
        string
      }
    }
  }
}
```

Query records.

```graphql
{
  queryRecords(attributes: [{ key: "type", value: { string: "crn:bot" } }]) {
    id
    names
    bondId
    createTime
    expiryTime
    owners
    attributes {
      key
      value {
        string
      }
    }
  }
}
```

Get account details:

```graphql
{
  getAccounts(addresses: ["laconic17t5ywvqxntu0afc96tz0yxcx92ss0e2alhx2c2"]) {
    address
    pubKey
    number
    sequence
    balance {
      type
      quantity
    }
  }
}
```

Query bonds:

```graphql
{
  queryBonds(
    attributes: [
      {
        key: "owner"
        value: { string: "laconic17t5ywvqxntu0afc96tz0yxcx92ss0e2alhx2c2" }
      }
    ]
  ) {
    id
    owner
    balance {
      type
      quantity
    }
  }
}
```

Get bonds by IDs.

```graphql
{
  getBondsByIds(
    ids: [
      "1c2b677cb2a27c88cc6bf8acf675c94b69051125b40c4fd073153b10f046dd87"
      "c3f7a78c5042d2003880962ba31ff3b01fcf5942960e0bc3ca331f816346a440"
    ]
  ) {
    id
    owner
    balance {
      type
      quantity
    }
  }
}
```

Query Bonds by Owner

```graphql
{
  queryBondsByOwner(
    ownerAddresses: ["laconic17t5ywvqxntu0afc96tz0yxcx92ss0e2alhx2c2"]
  ) {
    owner
    bonds {
      id
      owner
      balance {
        type
        quantity
      }
    }
  }
}
```

Query auctions by ids

```graphql
{
  getAuctionsByIds(
    ids: ["be98f2073c246194276554eefdb4c95b682a35a0f06fbe619a6da57c10c93e90"]
  ) {
    id
    ownerAddress
    createTime
    minimumBid {
      type
      quantity
    }
    commitFee {
      type
      quantity
    }
    commitsEndTime
    revealFee {
      type
      quantity
    }
    revealsEndTime
    winnerBid {
      type
      quantity
    }
    winnerPrice {
      type
      quantity
    }
    winnerAddress
    bids {
      bidderAddress
      commitHash
      commitTime
      commitFee {
        type
        quantity
      }
      revealFee {
        type
        quantity
      }
      revealTime
      bidAmount {
        type
        quantity
      }
      status
    }
  }
}
```

LookUp Authorities

```graphql
{
  lookupAuthorities(names: []) {
    ownerAddress
    ownerAddress
    height
    bondId
    status
    expiryTime
    auction {
      id
      ownerAddress
      createTime
      minimumBid {
        type
        quantity
      }
      commitFee {
        type
        quantity
      }
      commitsEndTime
      revealFee {
        type
        quantity
      }
      revealsEndTime
      winnerBid {
        type
        quantity
      }
      winnerPrice {
        type
        quantity
      }
      winnerAddress
      bids {
        bidderAddress
        commitHash
        commitTime
        commitFee {
          type
          quantity
        }
        revealFee {
          type
          quantity
        }
        revealTime
        bidAmount {
          type
          quantity
        }
        status
      }
    }
  }
}
```

LookUp Names

```graphql
{
  lookupNames(names: ["crn://hello/test"]) {
    latest {
      id
      height
    }
    history {
      id
      height
    }
  }
}
```

Resolve Names

```graphql
{
  resolveNames(names: ["asd"]) {
    id
    names
    bondId
    createTime
    expiryTime
    owners
    attributes {
      key
      value {
        string
      }
    }
  }
}
```

Query participants:

```graphql
{
  getParticipants {
    cosmosAddress
    nitroAddress
  }
}
```
