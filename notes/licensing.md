# Licensing — internal note (not published)

This note holds the licensing analysis that is deliberately kept off the rendered
site. It is not under `engineering/` and is not in the site nav, so neither
`tools/render.sh` nor `tools/render-engineering.sh` touches it. Keep licensing
discussion here, not in any page that renders to HTML.

## Shielded pool + JoinSplit circuits (T0.0 / T0.1)

Railgun's on-chain pool contracts (`Railgun-Privacy/contract`) are SPDX
`UNLICENSED`, and `Railgun-Privacy/circuits-v2` carries a `License.md` stating
*"No License is provided for any party under any circumstances."* Redeploying or
recompiling those sources is therefore not permissible. This is the reason the
pool and circuits are authored by us as our own implementation of the Railgun
*design* (spec-compatible with the note/commitment/nullifier format,
`snarkSafetyVector`, and public-signal arity pinned in A.1.1/A.1.2), using no
Railgun-licensed source.

On the published pages this is stated only as "our own implementation of the
Railgun design" with the non-licensing rationale (fee=0, our POI policy, the
settlement/deposit-payout hook, and upgrade/audit control per ADR-0002). The
licensing basis lives only here.

- `Railgun-Community/engine` and `Railgun-Community/cookbook` are MIT and may be
  used as reference (behavioral oracle, recipe patterns), but they are not the
  on-chain/circuit pieces.
- `circomlib` (used by the circuits) is GPL-3.0; a circom build that `include`s
  circomlib templates is a derivative of GPL-3.0 code. Reconcile the intended
  license of our own circuits against that before release.
- `go-nitro` / `ts-nitro` are separately licensed OSS and unaffected.

## Adapters (T5.0)

The Aave-v4 yield and Swaps adapters may reuse Railgun's RelayAdapt / Cookbook
recipe *if the license permits*; otherwise they are built. This is an open gate
(work package C). On the site it reads as "build (or adapt the recipe pattern)"
without the license-gate wording.

## DSS for the nitro custodian (ADR-0016, post-v1)

Two candidate threshold-signing libraries, both effectively unproven for our use:

- **`cerc-io/chain-signatures`** — MIT. It is Chainlink's `core/services/signatures`
  code (`ethschnorr` / `ethdss` / `secp256k1` + `SchnorrSECP256K1.sol`), extracted
  by cerc-io with one functional change (SEC1 point compression for Cosmos). Its
  parent is Chainlink's **OCR2VRF** (`smartcontractkit/chainlink-vrf`), which is
  archived and self-labeled *"in development, do not use in production"* and was
  never deployed to production. So it is MIT and EVM-verifiable-Schnorr-shaped, but
  its "do not use until audited / NOT CONSTANT TIME" banners are current, not stale.
- **`ship-armada/multi-party-schnorr`** — a fork of `silence-laboratories/multi-party-schnorr`
  under the **Silence Laboratories Non-Commercial Use License (SLL)**, which requires
  a paid commercial license for any commercial or internal-business use. This is a
  hard blocker for a commercial USDC product until relicensed, notwithstanding its
  (partial) HashCloak audit.

The DSS choice re-centers on (a) license and (b) which base is most analyzable for a
full-system audit. Both paths still require net-new EVM-verifier work and one
full-system audit; neither carries a transferable production audit.

## Shipped-Armada divergence (context)

The upstream `ship-armada/armada-poc` pool is a Railgun adaptation that imports the
UNLICENSED Railgun Solidity directly. Our decision to author our own implementation
is a divergence from that, motivated by the licensing status above plus fee/POI/hook
control.
