# armada-metalex: plan

## Executive summary

`armada-metalex` is a new Armada repository that lets an Armada user take part in MetaLeX's
on-chain legal and capital products (agreements, fundraising rounds, grants, escrows, BORG
membership, company formation) without exposing their identity or the rest of their on-chain
activity. MetaLeX is a client of this work and will cooperate on deployment and testing.

The user acts in MetaLeX through a **clean EVM address**: a fresh secp256k1 key derived on the
user's device from their Armada root secret, one per engagement, recoverable from their mnemonic.
That address, not an Armada hub, is the party MetaLeX records, so the user holds their
agreements, allocations and grants directly. Value reaches the address only from Armada's
shielded pool (an unshield from the user's 0zk address) or from an armada-nitro hub's inventory
(a channel payout), and value leaving it is shielded back to the user's 0zk address. The address
therefore carries the engagement and nothing else: no link to the user's wallet, their other
engagements or their shielded balance.

Where MetaLeX verifies a signature instead of the transaction sender, our relayer submits the
user's signature and the clean address never needs gas. Where MetaLeX requires the sender to
act, the address is funded with gas and the needed asset through the same private routes.

Scope is Ethereum L1. The work is about **124 points** in 17 chunks. A relayed agreement
signature works end to end after about **31 points**, and the first engagement with funding and
a return path after about **53**.

## 1. Scope

**In scope**, all on Ethereum mainnet, tested on Ethereum Sepolia and on our E12 fixturenet:

- MetaLeX's shared suite at its Ethereum deployment (`docs/reference/deployments.md:17-30`):
  `CyberAgreementRegistry` (cyberSign), the cyberRAISE round and deal managers, secondary trading,
  the cap-table and issuance contracts, and LeXcheX credentials.
- The Ethereum-only `ParentCoFactory` (`deployments.md:40`) and company formation, which is paid
  in USDC on Ethereum mainnet (`docs/webapp/formation.md:35,240`).
- Products that users deploy themselves, deployed on Ethereum: MetaVesT, LeXscroW,
  RicardianTripler and BORG.

**Out of scope:** every MetaLeX element that runs on another chain. That covers ACE and the
PumpCorp stack (Base only), Solana-bridged tokens, and the shared suite's Base, Arbitrum and
zkSync Era deployments. Off-chain MetaLeX tools that involve no address or transaction (the GAIBE
assistant, cap-table planning tools) need no wrapper.

## 2. How it works

Every engagement uses the same four steps.

1. **Derive the clean address.** `engagementKey = HKDF-SHA256(ikm = rootSecret, salt =
   "armada-metalex", info = "engagement/" || uint32(index))`, reduced to a valid secp256k1 scalar.
   The domain tag keeps it disjoint from the user's 0zk keyset, which the Armada SDK derives from
   the same root secret (`armada-sdk/src/wallet/derive.ts`). Each engagement gets its own index;
   the device keeps an encrypted map from engagements to indices.
2. **Fund it, if the action needs funds or gas.** Either unshield from the user's 0zk address to
   the clean address with `@armada/sdk`, or have the armada-nitro hub pay it from inventory
   through a channel. A non-shield channel allocation to an external address is paid by
   `_transferAsset` (`ArmadaNitroAdjudicator.sol:135-136`), which sends native ETH or any ERC-20
   (`src/nitro/MultiAssetHolder.sol`), so the hub can deliver both gas and tokens.
3. **Act.** Sign the product's message with the clean key. If the contract verifies the signature
   rather than the sender, armada-relayer submits it; otherwise the clean address sends the
   transaction itself.
4. **Return value.** Whatever the address receives (vested tokens, refunds, sale proceeds) is
   shielded to the user's 0zk address, through the SDK's gasless permit shield when the token
   supports EIP-2612.

The armada-nitro hub is never a MetaLeX party. Its role is funding: it pays from standing
inventory, and the user's dealings with it are off-chain channel updates, so its payouts don't
correspond to any per-user on-chain event.

## 3. Products

"Relayed" means our relayer submits the clean address's signature and the address needs no gas.
"Sender" means the contract checks or charges `msg.sender`, so the address sends the transaction
and needs gas.

### 3.1 cyberSign: CyberAgreementRegistry

| Action | Signature | Submission |
|---|---|---|
| Sign or countersign an agreement | EIP-712 `SignatureData` (§4.1) | Relayed through `signContractFor(signer, …)` (`CyberAgreementRegistry.sol:470`) when the agreement has no finalizer |
| Void an agreement | EIP-712 `VoidSignatureData` | Relayed through `voidContractFor`, which accepts any sender and checks the party's signature unless the finalizer submits |
| Set or revoke a signing delegate | none | Sender |
| Sign in to MetaLeX's app | Off-chain sign-in message | No transaction |

`signContractFor` checks the signer's signature (`:1016-1039`) and requires the sender to be the
finalizer or the signer **only when the agreement names a finalizer** (`:533-538`). The wrapper
reads the finalizer from the registry's public `agreements(contractId)` getter (`:116`, field
`:78`) before signing:

- Standalone cyberSign agreements, created through `createStandaloneContractAndSignFor`, always
  have no finalizer (`:346-387`) and are always relayed.
- Agreements created by MetaLeX's contracts name that contract as finalizer: LeXcheX
  (`creds/lexchexMinter.sol:182,257`), deals (`storage/DealManagerStorage.sol:154`), rounds
  (`storage/RoundManagerStorage.sol:249`), secondary trades (`storage/SecondaryTradeStorage.sol:398`)
  and the SegCo and board-consent agreements (`ParentCoFactory.sol:512,539`;
  `MetaDAOFactory.sol:447,474`). Those signatures travel inside the product's own transaction.

### 3.2 cyberRAISE: rounds and deals

| Action | Submission | Funding and return |
|---|---|---|
| Submit an expression of interest (`submitEOI`), with the registry signature and payment | Sender; payment is taken from the caller | Fund the ticket amount and gas; the allocated LET lands at the clean address |
| Sign and pay a deal (`signDealAndPay`) | Sender | Same |
| Recall an expired expression of interest (`recallEOI`) | Sender | Shield the refund back |
| Secondary trades | Sender | Same pattern as primary deals |

A round can credential a qualifying investor with LeXcheX at allocation, binding the credential
to the clean address (§3.4). The issuer-side officer actions (`createRound`, `allocate`, `reject`)
are sender transactions, with the officer's escrowed signature checked under each RoundManager's
own EIP-712 domain, read from the instance at runtime.

### 3.3 Cap table and company operations

Founder and officer actions are all sender transactions: setting up or forming a company
(formation costs 1,000 USDC on Ethereum mainnet), publishing officers, seating a board, and the
IssuanceManager's `onlyOwner` certificate and scrip functions. Holders transfer, scripify and
de-scripify LETs as senders, and shield any proceeds back. Approving security-class terms is an
off-chain signature. Board consents produce a cyberSign agreement (§3.1).

### 3.4 LeXcheX credentials

The user verifies their wallets with off-chain signatures and signs the LeXcheX agreement, whose
finalizer is the LeXcheX minter, so that signature rides in the minter's transaction. The
credential is soulbound to the address that earns it and records investor type, jurisdiction and
KYC facts. An address holding a credential is therefore a durable **investor persona** that the
user creates on purpose, kept separate from one-off engagement addresses, and the wrapper warns
the user before minting.

### 3.5 MetaVesT grants

The grant's award agreement is signed through cyberSign. Every grantee action on the grant itself
is sender-gated (`onlyGrantee`, `BaseAllocation.sol:306`; `TokenOptionAllocation.sol:134`;
`RestrictedTokenAllocation.sol:152`): withdrawing vested tokens, exercising options (strike paid
in USDC from the clean address), claiming repurchase proceeds, confirming milestones and
transferring rights. The address is funded with gas, and with the strike for an exercise, and
received tokens are shielded back.

### 3.6 LeXscroW

| Action | Submission |
|---|---|
| `depositTokensWithPermit` | Relayed: the clean address signs an EIP-2612 permit and the relayer submits the deposit |
| `depositTokens` | Sender |
| `execute()` | Relayed: unpermissioned, anyone can call it |
| `updateBuyer`, `updateSeller`, `electToTerminate` | Sender |

Escrow outputs land at the party addresses and are shielded back.

### 3.7 RicardianTripler

Proposing and confirming a DoubleToken LeXscroW agreement are sender transactions.
`adoptBORGParticipationAgreementWithSignature` accepts a signature from anyone and is relayed.
That signature is over the **raw digest** `keccak256(abi.encode(details))`, checked with
`ecrecover` or ERC-1271 (`RicardianTriplerBORGParticipation.sol:144-148`), not an EIP-712 or
personal-sign message, so the wrapper's signer supports raw-digest signing.

### 3.8 BORG

A BORG member is an owner of a Gnosis Safe that runs MetaLeX's guard. The member co-signs Safe
transactions with EIP-712 `SafeTx` signatures; the relayer submits `execTransaction`, and the guard
enforces the BORG's whitelist. Self-removal through `ejectImplant` is a sender transaction. Joining
starts with the BORG Participation Agreement (§3.7). Membership is long-lived, so a BORG address is
a durable persona, like a LeXcheX one.

## 4. Signing

### 4.1 CyberAgreementRegistry

- Domain: `EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)`
  with name `CyberAgreementRegistry`, version `1`, the chain ID fixed at initialization, and
  verifying contract `0xa9E808B8eCBB60Bb19abF026B5b863215BC4c134`.
- Type: `SignatureData(bytes32 contractId,address signer,string legalContractUri,string[]
  globalFields,string[] partyFields,string[] globalValues,string[] partyValues)`, with string
  arrays hashed element by element (`_hashStringArray`, `:1068-1076`).
- Confirmed against the live Ethereum deployment: `DOMAIN_SEPARATOR()` returns
  `0x2449d9cb31bee23a0bc07db884e744da4fdb3d6614e0811a30fb96d3b874a4f4` and
  `SIGNATUREDATA_TYPEHASH()` returns
  `0xe37d17c3ab7740aee31093101f9d27d139a5c3b35324b266efe5d085d6486f05`, both equal to the values
  computed from source.
- `VoidSignatureData(bytes32 contractId,address party)` uses the same domain.

### 4.2 Other signatures

| Signature | Format |
|---|---|
| RoundManager / DealManager officer signatures | EIP-712 under each instance's own domain, read at runtime |
| RicardianTripler adoption | Raw `keccak256(abi.encode(details))` digest |
| BORG co-signing | Safe EIP-712 `SafeTx` under the Safe's domain |
| LeXscroW permit deposit | EIP-2612 permit for the token |
| Return to the pool | `@armada/sdk` gasless shield intent (domain `ArmadaGaslessShield`, version `1`, `armada-sdk/src/tx/gasless-shield.ts`) |

## 5. Privacy

The clean address and everything it does are public by design: it is the user's MetaLeX
identity for that engagement. The plan keeps it from linking to anything else.

| Possible link | How the plan prevents it |
|---|---|
| Where the address's funds come from | Its only inbound transfers are unshields from the pool and channel payouts from hub inventory; every gate asserts this |
| Gas | Relayed actions need none. Gas for sender actions arrives through the same two routes |
| Amount and timing of an unshield | Fund with rounded, slightly larger amounts, shield the change back, and fund ahead of the action where the product allows |
| Where its value goes | Back to the user's 0zk address through a shield, never to another address of theirs |
| Reuse across engagements | One address per engagement; durable personas (LeXcheX, BORG, a founder identity) only when the user creates one |
| Network metadata | The address's reads and writes go through Armada's relayer and its read API, never the user's own node |

Some products carry identity by nature. A LeXcheX credential records KYC facts, a company
founder is a named party, and formation's fixed 1,000 USDC fee is a recognizable amount. For
these, the guarantee is that the address's funding came from the pool, not that the person
behind it is unknown, and the wrapper tells the user so.

## 6. Repository

`armada-metalex` follows `armada-sdk`'s conventions (`SPEC.md`, `CLAUDE.md`, tsup, vitest,
`ABOUTME:` headers, subpath exports, `@armada/sdk` pinned by git SHA) and runs under exophial like
armada-nitro.

```
armada-metalex/
  SPEC.md  CLAUDE.md  package.json  .exophial/config.yaml
  src/
    keys/          engagement-key derivation; viem signer (EIP-712, personal-sign, raw digest)
    engagements/   encrypted engagement registry, stored through @armada/sdk storage
    funding/       fund from pool (unshield), fund from nitro (channel payout), return to pool
    eip712/        typed-data and digest builders for every signature in section 4
    clients/       cyber-sign, cyber-raise, issuance, company, lexchex, metavest,
                   lexscrow, ricardian, borg
    relay/         client for armada-relayer's MetaLeX endpoint
    config/        MetaLeX Ethereum mainnet and Sepolia addresses
  test/vectors/    derivation and digest vectors
  ui/              embeddable React components (wagmi, viem), built last
```

Changes in existing repositories:

- **armada-relayer** (`relayer-v2/actor`): a `/relay/metalex` endpoint with an allow-list of
  MetaLeX selectors (`signContractFor`, `voidContractFor`, LeXscroW `execute` and
  `depositTokensWithPermit`, RicardianTripler adoption, Safe `execTransaction`), modeled on the
  existing selector list (`src/relay/selectors.ts`).
- **armada-nitro**: a MetaLeX deploy module in the fixturenet stack (`stack/`, `env/reconciler`)
  and the per-product gates under `env/gates/`.
- **armada-interface**: optionally mounts the wrapper's components.
- **armada-sdk**: no change.

## 7. Testing

- **Unit:** derivation vectors (deterministic, domain-separated, disjoint from the 0zk keyset,
  recoverable from the mnemonic); digest vectors for every signature in section 4, with
  `SignatureData` checked against the live separator and typehash; a check that funding only
  ever targets derived addresses and returns only ever target the user's 0zk address.
- **Integration on E12:** MetaLeX's contracts deployed by the fixturenet stack, with their
  addresses in its manifest. Each client runs end to end: derive, fund, act, return. The first
  integration run also checks a relayed agreement signature against the registry's own digest.
- **Gates,** one per product in armada-nitro's `env/gates` style: `metalex-cybersign` (relayed,
  zero gas at the clean address), `metalex-cyberraise`, `metalex-metavest`, `metalex-lexscrow`,
  `metalex-ricardian`, `metalex-lexchex` and `metalex-borg`.
- **Privacy assertion,** in every gate: every inbound transfer to the clean address comes from the
  pool, the relayer or the hub.

## 8. Work breakdown

Chunks are sized for exophial workers (at most 12 points, one point being about 10k tokens of
peak context).

| # | Chunk | Depends on | Points |
|---|---|---|---|
| C1 | Repository bootstrap: build, test and exophial config; MetaLeX addresses; `@armada/sdk` dependency | — | 4 |
| C2 | Engagement-key derivation and signer, with derivation vectors | C1 | 7 |
| C3 | Signature builders for section 4, with digest vectors | C1 | 6 |
| C4 | Encrypted engagement registry | C1 | 4 |
| C5 | armada-relayer MetaLeX endpoint and allow-list | — | 8 |
| C6 | cyberSign client and the first relayed signature, with its gate | C2, C3, C5 | 6 |
| C7 | Funding from the pool and return to the pool, with the privacy assertion | C2 | 9 |
| C8 | Funding from armada-nitro channel payouts | C7 | 7 |
| C9 | MetaLeX deploy module for the E12 fixturenet | C1 | 10 |
| C10 | cyberRAISE client and gate | C3, C7, C9 | 9 |
| C11 | MetaVesT client and gate | C7, C9 | 7 |
| C12 | LeXscroW and RicardianTripler clients and gates | C3, C5, C7, C9 | 9 |
| C13 | LeXcheX client and the persona model, with its gate | C3, C9 | 7 |
| C14 | BORG client and gate | C3, C5, C9 | 8 |
| C15 | Cap-table and company-officer clients, including secondary trades, with a gate | C7, C9 | 9 |
| C16 | Embeddable UI components | C6, C10 | 10 |
| C17 | SPEC and documentation, cross-repository cleanup | all | 4 |

**Total: 124 points.**

- First relayed signature end to end: C1, C2, C3, C5, C6, about **31 points**.
- First engagement with funding and a return path: C1, C2, C3, C5, C7, C9, C10, about
  **53 points**.
- After C1, chunks C2 to C5 and C9 can run in parallel. The product clients C10 to C15 are
  independent of each other once C3, C7 and C9 have landed. C16 and C17 come last.
