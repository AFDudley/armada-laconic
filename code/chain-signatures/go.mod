module git.vdb.to/cerc-io/chain-signatures

go 1.23.4

require (
	github.com/btcsuite/btcd/btcec/v2 v2.3.4
	github.com/ethereum/go-ethereum v1.14.12
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.10.0
	go.dedis.ch/fixbuf v1.0.3
	go.dedis.ch/kyber/v3 v3.1.0
	golang.org/x/crypto v0.32.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/holiman/uint256 v1.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.dedis.ch/protobuf v1.0.11 // indirect
	golang.org/x/sys v0.29.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace go.dedis.ch/kyber/v3 => github.com/cerc-io/kyber/v3 v3.0.0-20250728035006-f80208a7f291 // branch dev-3.x
