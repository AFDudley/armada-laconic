# `L2BridgeApp` zkVM proof verifier (2026-11-01)

The specification of the proof verifier of `L2BridgeApp`, and the operator's choice of proof system. Agent
`L2BridgeVerifierSpec` (Opus) did the research. The coordinator checked finding 1 and the `conclude` finding (§9.3,
Q6) against the code. The coordinator revised §9 after the operator's rulings.

## Rulings

1. **The verifier is where the risk is.** Operator, 2026-11-01: "it's where I see nearly all of the risk."
2. **Use RISC Zero alone.** Operator, 2026-11-01: "it seems like just supporting risc zero makes more sense." The
   reasons are in §9.1. 2-of-2 multi-proof is rejected (§9.2).
3. **A proof market is a reason for RISC Zero.** Operator, 2026-11-01: "the other reason to use risc zero is they at
   least have a marketplace for proofs, even if that's still pretty dicey it's better than nothing." Finding: Succinct
   also runs a prover market. Boundless differs because a request can go on chain (§9.1, reason 5).

## Design under review

- An OP Stack L2 uses an armada-nitro channel, the bridge channel, as its canonical bridge.
- The bridge channel has two participants. One is the custodian, a FROST committee behind `HubCustodian` (EIP-1271).
  The other is a proxy slot that anyone can use.
- The channel's app is `L2BridgeApp`, a subclass of `ZkForceMoveApp`. It supports a candidate state when three
  conditions hold:
  1. The state chains from the prior state's L2 output root.
  2. A zkVM proof shows the L2 state transition, including the L1 deposits that the transition consumed.
  3. One participant signed the state.
- The proof gives validity. ForceMove only selects the newest proven state. This is decision 18 of
  `labbook/2026-10-18-contract-owned-notes-and-zk-channel-apps.md`.
- Users have three exit paths:
  1. A proven withdrawal against a proven root.
  2. An instant swap with the custodian, locked by a hashlock.
  3. ForceMove with the user's latest co-signed channel state. This path uses no zkVM.
- The freeze rule: if two valid proofs give different roots for one L2 height, the bridge stops exit path 1. Users
  then exit by path 3.
- The target is L2BEAT Stage 2, with immutable contracts and no admin.

Sections 1 to 9 answer nine questions: the options, immutability, the public inputs, the L1 head check, program
identity and reproducibility, the soundness history, the freeze rule, the trusted primitive in Lean, and the
recommendation.

## 0. Sources and citation keys

Every claim has a citation or an `[INFERENCE]` mark. A citation is a file and line at a pinned commit, a URL with a
read date, or a measurement listed below.

| Key | Source | Pin |
|---|---|---|
| `nitro` | armada-nitro | `295b6c2eb4471162ab32ccf2bddff9f61265d54d` |
| `poc` | armada-nitro `lib/armada-poc` | `db15af172df1723e2140442de8d961068ed8e617` |
| `laconic` | armada-laconic | `6ca0b9bfd0fc215719074ead8622735edd8b005c` |
| `lean` | armada-lean, branch `integrated` | `a44db5f3815e71c05bf6acdeb9804e6dcf2e636e` |
| `lean-app` | armada-lean, branch `draft/app-property` | `48d483e91af9610e6eeada22e211e09ad3dd946e` |
| `specs` | `ethereum-optimism/specs` | `808223dd701a447f79b3249df4ef04523fba13e1` (2026-10-29) |
| `ops` | `succinctlabs/op-succinct` | `1e01cf0504c9306331641e37599cebe4e46fa635` (2026-10-28) |
| `sp1c` | `succinctlabs/sp1-contracts` | `d3629729c3216eb51bd4859d027a8eb729399fa4` (2026-04-28) |
| `r0e` | `risc0/risc0-ethereum` | `1d19771dc82531dced48fe7ff2e27bf6ad38bd59` (2026-10-22) |
| `kailua` | `boundless-xyz/kailua` | `beeeb1391ff570051242b6aa54240e6b19889be6` (2026-09-17) |
| `web` | Web pages named by URL | read 2026-11-01 |
| `ghsa` | GitHub security advisories, REST API `repos/<owner>/<repo>/security-advisories` | read 2026-11-01 |
| `chain` | Ethereum mainnet reads with `cast`, RPC `ethereum-rpc.publicnode.com` | block 26320993, 2026-11-01 |
| `gas` | `forge test` (forge 1.8.3) on the `sp1c` and `r0e` pins | run 2026-11-01 |

The research used local clones of the upstream repositories at these pins. The clones are not part of this repository.

## 0a. Findings that change the design

These four findings come from the sections below. They are listed first because they change the design.

1. **The starting root is free unless it is anchored.** `LedgerForceMoveApp` checks only that the candidate chains
   from the prior state (`nitro` `src/apps/LedgerForceMoveApp.sol:29-33`). It does not check the prior state. In the
   bridge channel, anyone can sign through the proxy slot (design). So an attacker can sign a prior state that carries a
   made-up L2 root. A zkVM proof from a made-up root is a valid proof. See §3.
2. **Both candidate STF programs use the same code base.** OP Succinct pins `kona-client/v1.8.0` (`ops`
   `Cargo.toml:84-91`). Kailua pins `kona-client/v1.6.0` (`kailua` `Cargo.toml:104-107`). A bug in kona breaks both
   zkVM paths at the same time. Two zkVMs do not give diversity at the program layer. See §1.4 and §6.3.
3. **An L1-head check in `stateIsSupported` makes support depend on time.** `blockhash` forgets a block after 256
   blocks. EIP-2935 forgets it after 8191 blocks. A state that is supported now can be unsupported later. The Nitro
   Lean model states that `Rules` "has no access to chain state" (`lean-app` `ArmadaLean/Nitro/Channel.lean:130-136`).
   See §4 and §8.
4. **The freeze needs storage writes.** `stateIsSupported` is `view` (`nitro` `src/apps/ZkForceMoveApp.sol:54-58`). A
   `view` function cannot set a freeze flag. So the freeze rule needs a non-`view` entry point and a stored record of
   proven roots. See §3.3 and §7.

---

## 1. Options

### 1.1 Summary table

| | SP1 + Groth16 | SP1 + PLONK | RISC Zero + Groth16 | Multi-proof (SP1 PLONK and RISC Zero Groth16) |
|---|---|---|---|---|
| Base verifier (mainnet) | `SP1Verifier` v6.1.0 Groth16, `0xb69f2584CBcFf99a58C4e7002E8b89Af54a6f4e2` | `SP1Verifier` v6.1.0 PLONK, `0xc3c6dDDAc8829b233Dc6536Ec024775a57b0AF2A` | `RiscZeroGroth16Verifier` 3.0.0, `0x2a098988600d87650Fb061FfAff08B97149Fa84D` | both contracts in columns 2 and 3 |
| Owner on the base verifier | none | none | none | none |
| Vendor router | `SP1VerifierGateway` (Groth16), owner is a 2-of-3 Safe | `SP1VerifierGateway` (PLONK), same owner | `RiscZeroVerifierRouter`, owner is a `TimelockController`, delay 259200 s | both routers |
| Vendor emergency stop | gateway `freezeRoute` | gateway `freezeRoute` | `RiscZeroVerifierEmergencyStop` per verifier, guardian or circuit breaker | both |
| Upgrade path on the base verifier | none | none | none | none |
| Trusted setup, L2BEAT rating | SP1 Hypercube v6.1.0 Groth16 setup, **red** | Aztec Ignition, **green** | Risc0 Groth16, **yellow** | minimum rule gives yellow `[INFERENCE]` |
| Gas of one verification (measured) | 234,718 | 291,639 | 254,073 | 2-of-2: about 545,712 |
| Verify call is `view` | yes | yes | yes | yes |

Citations for each row:

- Base verifier addresses: `sp1c` `contracts/deployments/1.json:19-20`; `r0e` `contracts/deployment.toml:68-70`. L2BEAT
  lists the same three addresses as verified deployments (`web` https://l2beat.com/zk-catalog/sp1hypercube.md,
  https://l2beat.com/zk-catalog/risc0.md). `chain`: `VERSION()` returns `"v6.1.0"` at `0xb69f…` and `0xc3c6…`, and
  `"3.0.0"` at `0x2a09…`.
- No owner on the base verifiers: the SP1 base verifier has no access control (`sp1c`
  `contracts/src/v6.1.0/SP1VerifierGroth16.sol:10-81`). The RISC Zero base verifier has none
  (`r0e` `contracts/src/groth16/RiscZeroGroth16Verifier.sol:52-183`). `chain`: `owner()` reverts on `0xb69f…` and on
  `0x2a09…`. RISC Zero states that the base verifier "has no state, it cannot be upgraded and it cannot be shutdown, even
  in the event of vulnerabilities" (`r0e` `contracts/version-management-design.md:98-101`).
- SP1 gateway: `Ownable`, `addRoute` and `freezeRoute` are `onlyOwner`, a frozen route reverts (`sp1c`
  `contracts/src/SP1VerifierGateway.sol:13-66`). Gateway addresses: `sp1c` `contracts/deployments/1.json:3-4`. `chain`:
  both gateways have owner `0xCafEf00d348Adbd57c37d1B77e0619C6244C6878`. That address is a Safe with threshold 2 and
  3 owners.
- RISC Zero router: `Ownable2Step`, `addVerifier` and `removeVerifier` are `onlyOwner`, a removed selector can never
  return (`r0e` `contracts/src/RiscZeroVerifierRouter.sol:24-102`). Router address and timelock: `r0e`
  `contracts/deployment.toml:12-22`. `chain`: the router owner is `0x0b144E07A0826182B6b59788c34b32Bfa86Fb711`, and
  `getMinDelay()` returns 259200.
- RISC Zero emergency stop: the guardian calls `estop()`. Anyone calls `estop(Receipt)` with a valid proof of an
  all-zero claim digest. A stopped proxy never restarts (`r0e` `contracts/src/RiscZeroVerifierEmergencyStop.sol:36-59`).
  `chain`: the guardian of the 3.0.0 and 2.2.0 stop proxies is `0xF616A4f81857CFEe54A4A049Ec187172574bd412`. That
  address has no code, so it is an EOA. `deployment.toml:18` names it `admin`.
- Trusted setups: `web` https://l2beat.com/zk-catalog/sp1hypercube.md ("SP1 Hypercube v6.1.0 Groth16 circuit-specific
  setup, risk red"; "Aztec Ignition, risk green"); https://l2beat.com/zk-catalog/risc0.md ("Risc0 Groth16, risk
  yellow", 55 phase-1 contributions, 238 phase-2 participants). The framework's minimum rule is for recursion
  (`web` https://forum.l2beat.com/t/the-trusted-setups-framework-for-zk-catalog/381, posted 2025-07-11). How L2BEAT
  rates two parallel systems is not stated there, so the multi-proof rating is `[INFERENCE]`.
- Gas: `gas`. The figures are the `forge test` gas of one call. They exclude the 21,000 base cost and calldata.
  SP1 Groth16: `SP1VerifierGroth16V6Test.test_VerifyProof_V6` = 234,718. SP1 PLONK:
  `SP1VerifierPlonkV6Test.test_VerifyProof_V6` = 291,639. RISC Zero: `testVerifyKnownGoodImageIdAndJournal` = 254,073.
  The SP1 test fixtures import the v6.0.0 contracts (`sp1c` `contracts/test/SP1VerifierGroth16V6.sol:5`). The v6.1.0
  contracts have the same code with different constants, so the gas is the same `[INFERENCE]`. Succinct states "~270k"
  for Groth16 and "~300k" for PLONK (`web` https://docs.succinct.xyz/docs/sp1/generating-proofs/proof-types).
  The 2-of-2 figure is the sum 291,639 + 254,073 = 545,712, without the extra hashing `[INFERENCE]`.
- `view`: `SP1Verifier.verifyProof … external view` (`sp1c` `contracts/src/v6.1.0/SP1VerifierGroth16.sol:51-55`,
  `SP1VerifierPlonk.sol:51-55`). `RiscZeroGroth16Verifier.verify … external view` (`r0e`
  `contracts/src/groth16/RiscZeroGroth16Verifier.sol:145`). Routers and stop proxies are also `view` (`sp1c`
  `SP1VerifierGateway.sol:20-24`; `r0e` `RiscZeroVerifierRouter.sol:97`; `RiscZeroVerifierEmergencyStop.sol:62`). The
  SP1 and RISC Zero verifiers call `this.verifyProof(…)`. An external self-call from a `view` function is a STATICCALL,
  so it is allowed `[INFERENCE]`.

### 1.2 SP1 options in detail

- The SP1 verifier checks a 4-byte selector, `exitCode == 0`, and `vkRoot == VK_ROOT()`. Then it verifies the SNARK over
  five public inputs: `programVKey`, the public-values digest, `exitCode`, `vkRoot`, and `nonce` (`sp1c`
  `contracts/src/v6.1.0/SP1VerifierGroth16.sol:51-81`).
- The public-values digest is `sha256(publicValues)` with the top 3 bits cleared (`SP1VerifierGroth16.sol:43-45`).
- SP1 wraps with gnark Groth16 or gnark PLONK over BN254 (`web`
  https://docs.succinct.xyz/docs/sp1/security/security-model).
- Groth16 setup, red: phase 1 is the first 54 contributions of Perpetual Powers of Tau. Phase 2 had 18 participants
  from 8 organizations, with no public call (`web` https://l2beat.com/zk-catalog/sp1hypercube.md). L2BEAT's 2026 Stage 1
  rule 1 forbids red setups (`web` https://forum.l2beat.com/t/new-stage-1-requirements-for-l2-proving-systems/413,
  posted 2026-02-16). So **SP1 Groth16 fails the Stage 1 proving requirements**. Stage 2 includes the Stage 1
  requirements `[INFERENCE]`.
- A Succinct page states that the Groth16 keys use "the Aztec Ignition ceremony + entropy contributions from members of
  the Succinct team" (`web` https://docs.succinct.xyz/docs/sp1/generating-proofs/proof-types). L2BEAT states Perpetual
  Powers of Tau for v6.1.0. The two sources disagree. This document follows L2BEAT, because L2BEAT gives the transcript
  and the key-regeneration steps.
- OP's `ZKDisputeGame` uses SP1 PLONK as its first backend (`specs`
  `specs/fault-proof/stage-one/zk/zk-dispute-game.md:105`). OP Succinct defaults to the Succinct PLONK gateway (`ops`
  `book/fault_proofs/deploy.md:51`).

### 1.3 RISC Zero option in detail

- The verifier checks a 4-byte selector. Then it verifies Groth16 over five public inputs: two halves of
  `CONTROL_ROOT`, two halves of the claim digest, and `BN254_CONTROL_ID` (`r0e`
  `contracts/src/groth16/RiscZeroGroth16Verifier.sol:155-183`).
- `verify(seal, imageId, journalDigest)` builds the claim with `ReceiptClaimLib.ok`. That claim has exit code
  `Halted(0)`, a zero input digest, and a zero assumptions digest (`r0e` `contracts/src/IRiscZeroVerifier.sol:74-82`).
- The wrapper circuit is circom and snarkjs Groth16. Its public control root lets RISC Zero change the zkVM without a
  new ceremony (`web` https://l2beat.com/zk-catalog/risc0.md, "Final wrap").
- Kailua reads the verifier address as `RISC_ZERO_VERIFIER` and calls `verify` (`kailua`
  `crates/contracts/foundry/src/KailuaVerifier.sol:39-42,337-367`). The Kailua book uses the RISC Zero managed router
  `0x8EAB…D319` (`kailua` `book/src/setup.md:52`). The book also says an application can deploy its own verifier
  (`kailua` `book/src/verifier.md:22-30`).

### 1.4 Multi-proof options

The candidate pair is SP1 PLONK (with the OP Succinct programs) and RISC Zero Groth16 (with the Kailua FPVM). SP1
Groth16 is excluded because its setup is red (§1.2).

| Rule | Safety needs | Liveness needs | Freeze |
|---|---|---|---|
| 2-of-2 | a bug in **both** systems for one false root | **both** systems prove each root | conflict between two 2-of-2 roots `[INFERENCE]` |
| 1-of-2 plus freeze on disagreement | a bug in **one** system, plus no honest counter-proof in time | **one** system proves each root | the other system proves a different root at the same height |

- 2-of-2 changes the safety hypothesis from "A is sound" to "A is sound or B is sound" at the zkVM layer `[INFERENCE]`.
- 2-of-2 does **not** change the program layer. Both STF programs use kona (finding 2 above). A kona bug gives the same
  wrong root in both systems. The two systems then agree, and no conflict appears `[INFERENCE]`.
- 1-of-2 keeps the safety of the weaker system, until an honest prover posts a conflicting proof. §7 shows that this
  trigger is not complete.
- OP `ZKDisputeGame` is not an option for this bridge as a whole. Its verifier address and `absolutePrestate` are set by
  OPCM governance, and a Guardian can pause, blacklist and retire games (`specs`
  `specs/fault-proof/stage-one/zk/zk-dispute-game.md:210-211,365`; `zk-interface.md:95-107`). Only its verifier
  interface and its public-value layout are reusable.

### 1.5 Fit with the existing armada-nitro code

- `ZkForceMoveApp.stateIsSupported` calls `StrictTurnTaking.requireValidTurnTaking` (`nitro`
  `src/apps/ZkForceMoveApp.sol:61`). Strict turn taking fixes which participant signs each turn number. The bridge rule
  is "one participant signed" (design). So `L2BridgeApp` must not call `super.stateIsSupported` `[INFERENCE]`.
- `ZkForceMoveApp` verifies with the pool's `Snark` library, a Groth16 BN254 check with a key passed in (`poc`
  `contracts/privacy-pool/types/SnarkTypes.sol:25-49`). It cannot verify PLONK. In principle it can verify the SP1 and
  RISC Zero Groth16 wraps, because each is Groth16 over BN254 with 5 public inputs `[INFERENCE]`. That path would copy
  the vendor checks (selector, `vkRoot`, exit code, digest) into new code. L2BEAT has not verified such code. This
  document recommends a call to the vendor's base verifier contract instead. This departs from N4's goal of "one Groth16
  checker in the stack" (`nitro` `docs/tasks/N4-zk-forcemove-app.md:23-25`).
- armada-nitro uses `pragma solidity 0.8.17` (`nitro` `src/apps/ZkForceMoveApp.sol:2`). `ISP1Verifier` uses
  `^0.8.20` (`sp1c` `contracts/src/v6.1.0/SP1VerifierGroth16.sol:2`). `L2BridgeApp` must declare its own interface for
  the external call `[INFERENCE]`.

---

## 2. Immutability

### 2.1 How to pin one exact verifier version

1. Call the **base verifier** address directly. Do not call `SP1VerifierGateway`, `RiscZeroVerifierRouter`, or a
   RISC Zero stop proxy. RISC Zero names "use `RiscZeroGroth16Verifier` directly" as the way to pin one immutable
   version (`r0e` `contracts/version-management-design.md:98-101`).
2. Store the address as an `immutable` in `L2BridgeApp` `[INFERENCE]`.
3. In the constructor, read and check the version fields of the target `[INFERENCE]`:
   - SP1: `VERIFIER_HASH()` equals `0x5a093a2fcb46394f5cadfe55c44d4d572fad9cec7aeb38026b0278322ef07fac` (PLONK v6.1.0)
     and `VK_ROOT()` equals `0x002f850ee998974d6cc00e50cd0814b098c05bfade466d28573240d057f25352` (`sp1c`
     `contracts/src/v6.1.0/SP1VerifierPlonk.sol:32-38`).
   - RISC Zero: `SELECTOR()` equals `0x73c457ba` (`r0e` `contracts/deployment.toml:68-70`).
4. Before deployment, check the deployed bytecode against a local build from the pinned source. L2BEAT already did this
   for the three base verifiers ("Verification: successful", `web` L2BEAT pages in §1.1). The operator should repeat
   it (§5).
5. Store the program keys as immutables: the SP1 aggregation `programVKey`, the range key commitment, and the Kailua
   `FPVM_IMAGE_ID` (§3) `[INFERENCE]`.

The selector check alone is not enough. RISC Zero states that a selector "is not intended to be collision resistant"
(`r0e` `contracts/src/groth16/RiscZeroGroth16Verifier.sol:81-85`). The pinned address is the real binding.

### 2.2 What happens when the vendor announces a soundness bug

- The vendor stops its own routing. Succinct froze the v5.0.0 and v6.0.0 routes on both gateways (`chain`:
  `routes(0xa4594c59)`, `routes(0xd4e8ecd2)`, `routes(0x0e78f4db)`, `routes(0xbb1a6f29)` all return `frozen = true`).
  RISC Zero stopped the 2.0.0-rc.3 and 2.1.0 proxies (`r0e` `contracts/deployment.toml:45-57`; `chain`: `paused()` is
  `true` on `0x68dC…40E3` and `0xDa8f…B2b1`).
- These actions do not reach a pinned base verifier. The base verifier keeps accepting proofs, because it has no owner
  and no pause (§1.1).
- `L2BridgeApp` has no admin. So it cannot move to the fixed version. It keeps trusting the broken version.
- The remaining defences are in this order:
  1. The second system in a 2-of-2 design (§1.4). The operator rejected 2-of-2 (§9.2), so this defence is absent.
  2. An on-chain bug trigger that freezes the bridge (§7). Then users exit through Nitro channels by ForceMove (design).
  3. Optionally, a "vendor can stop, never start" input: `L2BridgeApp` freezes when the vendor freezes the pinned route
     or stops the pinned proxy. Both reads are `view` (`SP1VerifierGateway.routes`, `RiscZeroVerifierEmergencyStop.paused`).
     This adds vendor trust only in the stop direction. A vendor can then cause a freeze, but it cannot cause a false
     root `[INFERENCE]`. This is an operator decision (§9.3, Q3).
- Without option 3, the gap is the time between the announcement and a public exploit. The circuit breaker and the
  conflict trigger need an exploit or a conflicting proof. An announcement alone does not trigger them `[INFERENCE]`.
- §6 lists on-chain-relevant soundness advisories. SP1 had four from 2024-10 to 2026-04 (v1.1, < 4.0, < 5.0, v6.0.x).
  RISC Zero had three from 2024-09 to 2025-08 (< 1.1.1, 2.0.x, 2.1.x). So a pinned version will probably see a
  disclosed bug during the bridge's life `[INFERENCE]`.

---

## 3. Public inputs

### 3.1 SP1 path: OP Succinct aggregation program

The aggregation program commits `AggregationOutputs` (`ops` `utils/client/src/types.rs:17-25`;
`programs/aggregation/src/main.rs:88-99`). The contract re-creates the same tuple (`ops` `contracts/src/lib/Types.sol:9-17`)
and calls `verifyProof(aggregationVkey, abi.encode(publicValues), proof)` (`ops`
`contracts/src/validity/OPSuccinctL2OutputOracle.sol:388-403`).

| Input | Meaning | Source in OP Succinct | Binding in `L2BridgeApp` | Free for an attacker? |
|---|---|---|---|---|
| `programVKey` (= `aggregationVkey`) | identifies the aggregation program | contract config (`OPSuccinctL2OutputOracle.sol:403`) | immutable constant | no, if pinned |
| `l1Head` | latest L1 block hash the proof read | `historicBlockHashes[_l1BlockNumber]` (`:388-394`) | a stored L1 checkpoint (§4) | **yes, if not checked**: fake L1 data gives fake deposits |
| `l2PreRoot` | starting L2 output root | last stored output (`:395`) | the root registered for the prior state (§3.3) | **yes, if not anchored** (finding 1) |
| `claimRoot` (`l2PostRoot`) | claimed L2 output root | proposer input (`:396`) | the candidate state's root | chosen by the prover; correctness comes from the proof |
| `claimBlockNum` (`l2BlockNumber`) | L2 height of `claimRoot` | proposer input (`:397`) | the candidate's L2 height, and `candidate.turnNum` `[INFERENCE]` | chosen by the prover; GHSA-5jh4 showed a mis-binding (§6.3) |
| `rollupConfigHash` | SHA-256 of the kona `RollupConfig` JSON | contract config (`:398`); `utils/client/src/boot.rs:17-29` | immutable constant | **yes, if not pinned**: another chain or deposit contract |
| `rangeVkeyCommitment` (`multiBlockVKey`) | identifies the range program that the aggregation verified | contract config (`:399`); `main.rs:47,86` | immutable constant | **yes, if not pinned**: the attacker's own range program |
| `proverAddress` | anti-front-running tag | caller input (`:400`) | ignore, or bind to `msg.sender` in the non-view entry point | yes, harmless `[INFERENCE]` |
| `exitCode` | must be 0 | proof bytes | checked by the verifier (`SP1VerifierPlonk.sol`, same as Groth16 `:74-76`) | no |
| `vkRoot` | identifies the recursion circuits | proof bytes | checked against `VK_ROOT()` by the verifier | no |
| `nonce` | proof nonce | proof bytes (`SP1VerifierPlonk.sol:67,74`) | not checked | yes; Succinct calls `proof_nonce` usage experimental (`web` security-model page); harmless `[INFERENCE]` |

The aggregation program checks these links (`ops` `programs/aggregation/src/main.rs`):

- consecutive ranges chain by root, and use one `rollupConfigHash` (`:27-38`);
- each range proof verifies under `multi_block_vkey` (`:41-48`);
- each range's L1 head is an ancestor of the committed `l1Head`, through a header chain (`:50-71`).

### 3.2 RISC Zero path: Kailua FPVM journal

Kailua packs the journal into 220 bytes (`kailua` `crates/kona/src/journal.rs:23-45,91-106`). The contract rebuilds it
and calls `RISC_ZERO_VERIFIER.verify(seal, FPVM_IMAGE_ID, sha256(journal))` (`kailua`
`crates/contracts/foundry/src/KailuaVerifier.sol:337-367`).

| Input | Binding in `L2BridgeApp` | Free for an attacker? |
|---|---|---|
| `imageId` = `fpvm_image_id` (also in the journal) | immutable constant | no, if pinned |
| `l1_head` | stored L1 checkpoint (§4) | **yes, if not checked** |
| `agreed_l2_output_root` | registered root of the prior state | **yes, if not anchored** |
| `claimed_l2_output_root` | candidate root | prover's choice; proof gives correctness |
| `claimed_l2_block_number` (`u64`) | candidate L2 height | prover's choice |
| `config_hash` = hash of rollup config and L1 config (`journal.rs:63-66`) | immutable constant | **yes, if not pinned** |
| `precondition_hash` | constant `0`. Zero means "fault proving", a plain transition (`journal.rs:26-32`) | **yes, if not fixed**: it changes the statement |
| `payout_recipient` | ignore or bind to `msg.sender` | yes, harmless `[INFERENCE]` |
| `CONTROL_ROOT`, `BN254_CONTROL_ID` | fixed in the pinned base verifier (`RiscZeroGroth16Verifier.sol:70-72`) | no |
| exit code, input, assumptions | fixed by `ReceiptClaimLib.ok` (`IRiscZeroVerifier.sol:74-82`) | no |

OP Succinct and Kailua compute the config hash in different ways (`ops` `utils/client/src/boot.rs:17-29`; `kailua`
`crates/kona/src/journal.rs:63-66`). So `L2BridgeApp` stores two constants.

### 3.3 Binding to the prior state and to the app configuration

The binding below is the design that this document proposes. It is `[INFERENCE]` except where a citation is given.

1. **A proven-root registry.** `L2BridgeApp` stores `provenRoot[l2Height]`. The constructor sets one genesis root. A
   permissionless, non-`view` function `proveRoot(…)` does these steps:
   - It requires that `l2PreRoot` is registered at a lower height.
   - It requires that `l1Head` is a stored L1 checkpoint (§4).
   - It builds the public values from storage and constants, as OP's `ZKDisputeGame.prove` does (`specs`
     `specs/fault-proof/stage-one/zk/zk-interface.md:58-87`: "no caller-supplied data beyond `_proofBytes` affects what
     the verifier checks").
   - It calls the pinned verifier.
   - It stores the root. If a different root is already stored at that height, it sets `frozen = true` (§7).
   - It rejects a height whose L2 timestamp is in the future. OP Succinct has the same check (`ops`
     `OPSuccinctL2OutputOracle.sol:364-367`).
2. **`stateIsSupported` reads storage only.** It returns true when all of these hold:
   - `frozen` is false;
   - the candidate's root is registered at the candidate's L2 height;
   - the candidate's L2 height is greater than the prior state's L2 height;
   - `candidate.turnNum` equals the candidate's L2 height;
   - one participant signed the candidate (design).
3. **Why the registry is needed.**
   - Finding 1: the prior state alone is not evidence, because the proxy slot can sign it. The registry anchors every
     pre-root to genesis through proven links.
   - Finding 3: the L1-head check happens once, in `proveRoot`. Then support no longer depends on the age of `l1Head`.
   - Finding 4: the freeze needs a storage write.
   - Turn number equal to height: otherwise an attacker can sign an old registered root with a large `turnNum`.
     ForceMove requires a higher turn number to replace a state (`nitro` `src/nitro/ForceMove.sol:48-55,100`). So such a
     state blocks newer roots.
4. **This design keeps CON decision 18.** The proofs give correctness. The ForceMove window still decides which proven
   state is newest (`laconic` `labbook/2026-10-18-contract-owned-notes-and-zk-channel-apps.md:207-209`).
5. **The verifier calls stay `view`.** They now run inside `proveRoot`, not inside `stateIsSupported`. The constraint of
   the design still holds. Every pinned verifier call is `view` (§1.1).

---

## 4. L1 head check

### 4.1 Mechanisms

| Mechanism | Window | Gas class | Fact |
|---|---|---|---|
| `blockhash(n)` | last 256 blocks; returns 0 outside | opcode | EIP-2935 keeps `BLOCKHASH_SERVE_WINDOW = 256` and does not change `BLOCKHASH` (`web` https://eips.ethereum.org/EIPS/eip-2935) |
| EIP-2935 history contract `0x0000F90827F1C53a10cb7A02335B175320002935` | last 8191 blocks; `get` reverts outside | STATICCALL with 32-byte calldata | `HISTORY_SERVE_WINDOW = 8191`; `get` reverts outside `[block.number-8191, block.number-1]` (`web` EIP-2935). `chain`: the code is present; `get(N-8000)` returned the same hash as `cast block N-8000`; `get(N-9000)` reverted |
| Stored checkpoint | permanent | one SSTORE when written | OP Succinct `checkpointBlockHash` stores `blockhash(n)` in `historicBlockHashes`, and `proposeL2Output` reads it (`ops` `OPSuccinctL2OutputOracle.sol:388-391,506-512`) |
| L1 head fixed at creation | n/a | n/a | OP `DisputeGameFactory.create()` stores `blockhash(block.number - 1)` (`specs` `zk-interface.md:85-87`). Kailua takes `l1Head` from a known proposal contract (`kailua` `KailuaTournament.sol:509-545,731-743`) |

With 12-second slots, 256 blocks is about 51 minutes, and 8191 blocks is about 27.3 hours `[INFERENCE]`. EIP-2935 also
gives "about a day of coverage" (`web` EIP-2935, rationale).

### 4.2 Recommended check

1. A permissionless function `checkpointL1(n)` reads `blockhash(n)`. If that returns 0, it reads the EIP-2935 contract.
   It stores the hash. It reverts if both sources fail `[INFERENCE]`.
2. `proveRoot` requires `l1Head` to be a stored checkpoint (§3.3).
3. A prover first stores a recent checkpoint. Then it proves against that hash. The proof time is then not limited by
   the window. OP Succinct uses the same flow (`ops` `OPSuccinctL2OutputOracle.sol:388-391,506-512`).

### 4.3 Failure modes

1. **Proof older than the window.** If `stateIsSupported` checks the L1 head directly, a state becomes unsupported
   after 256 or 8191 blocks. An honest party may need that state later, for example to answer a challenge. ForceMove
   calls the app at that later time (`nitro` `src/nitro/ForceMove.sol:59,103`). So the honest party can lose the
   challenge `[INFERENCE]`. The stored checkpoint removes this mode.
2. **Late checkpoint.** If nobody stores a checkpoint for block `n` within 8191 blocks, no proof against `n` can ever be
   used. The prover must make a new proof against a newer checkpoint. This costs time. It does not cost safety
   `[INFERENCE]`.
3. **Proof against an L1 fork.** The bridge runs on L1. `blockhash` and EIP-2935 return hashes of ancestors of the
   executing block only (`web` EIP-2935). So a hash from an orphaned block fails the check. If L1 reorgs after a
   checkpoint, the checkpoint transaction leaves the chain too. A reorg deeper than finality is outside this model
   `[INFERENCE]`.
4. **Fake L1 head.** Without the check, an attacker supplies a hash of a made-up L1 header chain. The aggregation program
   only checks that the headers hash to `l1Head` (`ops` `programs/aggregation/src/main.rs:56-66`). So made-up deposits
   become L2 funds, and withdrawals then pay against them `[INFERENCE]`.
5. **Future protocol change.** A later fork can change EIP-2935 (address, window, or semantics). An immutable contract
   cannot follow. `BLOCKHASH` has kept its 256-block semantics since launch. So `blockhash` is the primary source, and
   EIP-2935 is a second source `[INFERENCE]`.
6. **Deposits after `l1Head`.** A proof against an older `l1Head` simply consumes fewer deposits. OP derivation adds a
   user deposit in the first L2 block of the epoch whose L1 origin holds the deposit (`specs`
   `specs/protocol/derivation.md:1014-1021`). So a later proof includes it `[INFERENCE]`.

---

## 5. Program identity and reproducibility

L2BEAT's 2026 Stage 1 rule 3 asks for instructions to regenerate the on-chain verifiers and their keys. Rule 4 asks for
instructions to regenerate the program commitments from source (`web`
https://forum.l2beat.com/t/new-stage-1-requirements-for-l2-proving-systems/413).

### 5.1 SP1 PLONK verifier v6.1.0 (verifier key)

Source: `web` https://l2beat.com/zk-catalog/sp1hypercube.md, "SP1 Hypercube Plonk v6.1.0", verification steps.

```
git clone https://github.com/succinctlabs/sp1.git
cd sp1/crates/prover
git checkout v6.1.0   # commit d454975ac7c1126097e36eceda9bce2cb9899da4
export SP1_ALLOW_DEPRECATED_HOOKS=true
make build-circuits
```

The output in `build/plonk` holds the verifier contract, the verifier key, and the verifier hash. L2BEAT reports about
100 GiB peak memory. Compare the result with `0xc3c6…AF2A` and with `VERIFIER_HASH()`.

### 5.2 SP1 Groth16 verifier v6.1.0 (for completeness; excluded by §1.2)

Same source. After `make build-circuits`, regenerate the keys from the ceremony transcript:

```
git clone https://github.com/succinctlabs/semaphore-gnark-11.git   # tested at 6d6ebc3608e609ec879e9ba99abee6b6b97d937d
curl "https://sp1-circuits.s3-us-east-2.amazonaws.com/v6.1.0-trusted-setup.tar.gz" -o trusted-setup.tar.gz
tar -xzf trusted-setup.tar.gz && go build
./semaphore-gnark-11 key trusted-setup/phase1 trusted-setup/phase2-17 trusted-setup/evals /path/to/sp1/crates/prover/build/groth16/groth16_circuit.bin
shasum -a 256 vk
```

### 5.3 OP Succinct program keys

Two keys matter: `aggregationVkey` and `rangeVkeyCommitment`.

- The repo document is `ops` `book/advanced/verify-binaries.md:16-57`:
  `sp1up -v 6.10.0`, `just build-range-elfs`, `just build-agg-elf`, `cargo run --bin config --release`. The last command
  prints the rollup config hash, the aggregation key, and the range key commitment.
- L2BEAT verified OP Succinct `v3.14.1` (commit `35d98244c9b4ccc2b93f016be249788c5d6d5372`) with these commands (`web`
  https://l2beat.com/zk-catalog/sp1hypercube.md, "Aggregation program of OP Succinct"):
  ```
  sp1up -v v6.8.1
  cd programs/range/ethereum && cargo prove build --elf-name range-elf-embedded --docker --tag v6.8.1 --output-directory ../../../elf
  cd programs/aggregation && cargo prove build --elf-name aggregation-elf --docker --tag v6.8.1 --output-directory ../../elf
  cargo run --release --bin config
  ```
  L2BEAT's results: aggregation `0x00070992a03760456403e800b0bd1a0e907a85a5c2cffca292b5994b56b5363b`, range
  `0x5ae13fed447a94f57d6ee8fc44cce8075ef994485121dcd27e75577d264f75b1`.
- The rollup config hash is SHA-256 of `serde_json::to_string_pretty(config)` (`ops` `utils/client/src/boot.rs:17-29`).
  A change in kona's JSON field order or field set changes the hash. So the hash must be regenerated with the pinned
  release, not with a later one `[INFERENCE]`.

### 5.4 RISC Zero Groth16 verifier 3.0.0

Source: `web` https://l2beat.com/zk-catalog/risc0.md, "RiscZero v3.0.0", verification steps (36 GiB of memory, about
14 GiB of setup files).

```
# circom v2.1.0 from source, snarkjs from npm
git clone https://github.com/risc0/risc0.git && cd risc0
git checkout v2.0.0   # 3f26f9d4c2fb8a7e5eb830ae2433c8eae67f5a38
git lfs install && git lfs pull --include=groth16_proof/groth16/stark_verify.circom
cd groth16_proof/groth16 && circom stark_verify.circom --r1cs
shasum -a 256 stark_verify.r1cs   # 84d3c34b7c0eb55ad1b16b24f75e0b9de307f7b74089ea4a20a998390ee24178
wget https://storage.googleapis.com/zkevm/ptau/powersOfTau28_hez_final_23.ptau
wget https://risc0-artifacts.s3.us-west-2.amazonaws.com/tsc/2024-04-04/stark_verify_final.zkey
snarkjs zkey verify stark_verify.r1cs powersOfTau28_hez_final_23.ptau stark_verify_final.zkey
snarkjs zkey export solidityverifier stark_verify_final.zkey verifier.sol
git checkout v3.0.0   # 96eef1c3c9c6dde65dcc2fc86791a4358fcc8e4d
RUST_LOG=info cargo xtask bootstrap
```

`bootstrap` prints `allowed_control_root` `0xa54dc85ac99f851c92d7c96d7318af41dbe7c0194edfcc37eb4d422a998c1f56` and
`bn254_identity_control_id` `0x04446e66…7ac0` (byte order reversed). `r0e` also names `cargo xtask bootstrap` as the
source of these values (`contracts/src/groth16/RiscZeroGroth16Verifier.sol:66-69`). The Kailua book prints the same
`CONTROL_ROOT` for RISC Zero 3.0.6 (`kailua` `book/src/setup.md:42-51`).

### 5.5 Kailua FPVM image ID

- The build target is `just build-fpvm`, which runs `RISC0_USE_DOCKER=1 cargo build … -F rebuild-fpvm …` (`kailua`
  `justfile:42-43`; `build/risczero/build.rs:34-35`).
- `kailua-cli config --op-node-url … --op-geth-url … --eth-rpc-url …` prints `KAILUA_FPVM_KONA_ID`, `CONTROL_ROOT` and
  `ROLLUP_CONFIG_HASH` (`kailua` `book/src/setup.md:33-58`).
- L2BEAT has **not** verified an upstream Kailua image ID. Its only Kailua entry (MegaETH) is "unsuccessful", because
  the sources are not published (`web` https://l2beat.com/zk-catalog/risc0.md, "Program Hashes"). So the Kailua path
  needs our own reproduction before deployment.

### 5.6 What `L2BridgeApp` must publish

For each pinned value, the deployment must publish the source commit, the toolchain tag, and the command that
regenerates it. The values are: the verifier addresses, `VERIFIER_HASH`/`VK_ROOT`/`SELECTOR`, the program keys, and the
config hashes. This follows rules 3 and 4 above `[INFERENCE]`.

---

## 6. Soundness history

All rows come from `ghsa` unless another source is named. "On-chain" means that the advisory, or the vendor's freeze,
shows an effect on the on-chain verifier path.

### 6.1 SP1 and its wrappers

| Published | ID | Severity | Summary | Affected | Fixed | On-chain |
|---|---|---|---|---|---|---|
| 2023-10-16 | gnark GHSA-7p92-x423-vwj6 | critical | gnark PLONK verifier contract: KZG multi-point randomness lets a third party derive a new valid proof for the same inputs | gnark ≤ 0.9.0 | 0.9.1 | yes, PLONK verifier contract (malleability, not a false statement) |
| 2024-09-06 | gnark GHSA-q3hw-3gm4-w5cr | medium | Groth16 commitment extension unsound with more than one commitment | gnark ≤ 0.10.0 | 0.11.0 | the v6.1.0 SP1 Groth16 verifier has only `PUB_0..PUB_4` and no commitment key (`sp1c` `Groth16Verifier.sol:80-89`) `[INFERENCE]`: not exposed |
| 2024-10-28 | SP1 GHSA-f77q-r5qm-w4m8 | high | BabyBear range checks missing in the gnark recursion circuit | < 1.2.0 | 1.2.0 | yes; the v1.1.0 verifier was frozen on 2024-09-04 (advisory text) |
| 2024-11-08 | SP1 GHSA-8m24-3cfx-9fjw | medium | cumulative sum not observed in Fiat-Shamir | < 3.0.0 | 3.0.0 | versions deprecated |
| 2025-01-15 | SP1 GHSA-c873-wfhp-wx5m | high | `chip_ordering` unchecked; `is_complete` underconstrained; Plonky3 evaluation claims not observed | < 4.0.0 | 4.0.0 | yes for deferred proofs; older verifiers frozen (advisory text) |
| 2025-06-03 | SP1 GHSA-6248-228x-mmvh | high | Plonky3 vulnerability; Rust verifier `vk_root`; embedded allocator | < 5.0.0 | 5.0.0 | yes: "affected both the Rust verifier and the on-chain verifiers" (advisory text) |
| 2026-04-11 | SP1 GHSA-63x8-x938-vx33 | high | V6 recursion circuit: row-count binding gap in the jagged PCS verifier | 6.0.0-6.0.2 | 6.1.0 | yes `[INFERENCE]`: recursion feeds the wrap; `chain`: the v6.0.0 gateway routes are frozen |
| 2026-08-24 | gnark GHSA-3mvx-pp85-pm65 | critical | unconstrained hint outputs in `gnark/std` gadgets (emulated arithmetic, recursion, hashes) | gnark < 0.16.2 | 0.16.2 | unknown: this document did not confirm whether the SP1 wrap circuit uses the affected gadgets |
| 2026-10-11 | SP1 GHSA-p7v2-jjv3-8p8f | high | V6 native Rust verifier: tensor length not checked | 6.0.0-6.8.0 | 6.8.1 | no: "Groth16 and Plonk verification are unaffected" (advisory text) |

Also relevant: Succinct states that SP1 Turbo relies on proximity-gap conjectures, and SP1 Hypercube does not (`web`
https://docs.succinct.xyz/docs/sp1/security/security-model). The sp1-contracts repository has no advisories (`ghsa`).

### 6.2 RISC Zero and its verifiers

| Published | ID | Severity | Summary | Affected | Fixed | On-chain |
|---|---|---|---|---|---|---|
| 2024-07-15 | GHSA-5xgj-pmjj-gw49 | low | notes on zero knowledge | all | n/a | no |
| 2024-09-25 | GHSA-5c79-r6x7-3jx9 | high | insufficient validation of multi-step instruction modes | < 1.1.1 | 1.1.1 | `[INFERENCE]` yes; RISC Zero later removed the 1.1.0-rc.3 and 1.2.0 routes "on Nov 10, 2025" (`r0e` `deployment.toml:27-40`) |
| 2024-11-13 | GHSA-349p-x622-29x6 | low | syscall abuse by malicious proving hosts | < 1.1.3 | 1.1.3 | no data |
| 2025-06-18 | GHSA-g3qg-6746-3mg9 | critical | rv32im circuit: missing constraint on 3-register instructions | zkvm 2.0.0 to < 2.1.0 | 2.1.0 | yes; 2.0.0-rc.3 proxy stopped (`r0e` `deployment.toml:45-49`; `chain`) |
| 2025-06-24 | risc0-ethereum GHSA-gjv3-89hh-9xq2 | low | Steel accepts a zero commitment digest | Steel < 2.1.1 | 2.1.1 | Steel only; not used here |
| 2025-08-04 | GHSA-f6rc-24x4-ppxp | medium | signed division underconstrained | zkvm 2.0 to < 2.2 | 2.2.0 (circuit 3.0.0) | yes; 2.1.0 proxy stopped (`r0e` `deployment.toml:53-57`; `chain`) |
| 2025-10-01 | GHSA-jqq4-c7wq-36h7 | critical | `sys_read` memory safety: a malicious host writes anywhere in guest memory | zkvm < 2.3.2, < 3.0.3 | 2.3.2, 3.0.3 | guest library bug: the fix changes the image ID, not the verifier `[INFERENCE]` |

L2BEAT lists the same three 2025 incidents (`web` https://l2beat.com/zk-catalog/risc0.md, "Milestones & Incidents").
Kailua has no published advisories (`ghsa`, empty list). Kailua has five Veridise audits, dated 2025-02-18 to 2026-02-12
(`kailua` `audits/`).

### 6.3 STF program layer (OP Succinct and kona)

Question 6 asks for zkVM and verifier bugs. This layer is added because it holds most of the recent soundness bugs, and
both candidate programs share it.

| Published | ID | Severity | Summary | Fixed |
|---|---|---|---|---|
| 2025-07-12 | GHSA-8pwr-p276-45mf | low | DoS in range programs (agreed root equal to claimed root shortcut) | 2.3.0 |
| 2025-08-20 | GHSA-4cqq-rfmm-f597 | high | KZG blob proofs not checked in derivation: invalid state transition, "putting user funds at risk" | 3.1.0, 2.3.1 |
| 2026-01-21 | GHSA-x8rg-xfvg-9fq3 | high | precompile lookup in `DELEGATECALL` context: blocks unprovable | none listed |
| 2026-04-24 | GHSA-5jmw-h5pf-383c | high | upstream kona: derivation divergence; pipeline cursor started at `B256::ZERO` instead of the agreed root, "breaking soundness" | 3.8.0, 4.3.0 |
| 2026-05-06 | GHSA-5jh4-3p33-85xc | high | range proof can commit `(root(N), T)` with `N < T` when derivation halts early | 4.3.1, 3.8.1 |
| 2026-05-14 | GHSA-wq56-mx95-8mw9 | high | upstream kona-executor: wrong state trie for `DestroyedChanged` accounts | 4.4.0 |
| 2026-06-28 | GHSA-pq4w-5vv8-gxhr | high | EIP-4844 point-evaluation precompile accepts invalid KZG openings in the zkVM | 3.10.0, 4.5.0 |

GHSA-5jmw and GHSA-wq56 originate in kona. Kailua also uses kona (`kailua` `Cargo.toml:104-107`). Whether each kona bug
also affected Kailua is not confirmed here `[INFERENCE]`.

---

## 7. The freeze rule

### 7.1 Statement

The rule: if two valid proofs give different roots for one L2 height, the bridge stops paying against proven roots
(design). L2BEAT names "two valid contradicting zk proofs" as an on-chain bug detection system for Stage 2 (`web`
https://l2beat.com/stages, "Do users have at least 30 days to exit").

### 7.2 Soundness of the trigger (no false freeze)

The trigger is sound when all of these hold `[INFERENCE]`:

1. Both proofs start from registered roots (§3.3). Without anchoring, an attacker proves from a made-up pre-root. Two
   valid proofs from different made-up pre-roots can then differ honestly. This is a false freeze, so a griefing attack.
2. Both `l1Head` values are checkpoints on the current L1 chain (§4).
3. Both use the same rollup config hash and the same program keys.
4. OP derivation is a function of the canonical L1 chain, and it extends only as L1 extends. A missing batch makes the
   pipeline wait for a full sequencing window. It then fills the gap with empty batches (`specs`
   `specs/protocol/derivation.md:615-617`). So two canonical L1 prefixes give the same safe block at a height that both
   reach.

Under these conditions, a conflict proves that at least one accepted statement is false. Note one consequence: any
statement-binding bug becomes a freeze lever. GHSA-5jh4 (§6.3) let anyone pair a canonical root with a wrong height.
Succinct rated its impact as a withdrawal stall only. Under the freeze rule, the same bug lets anyone freeze the whole
bridge `[INFERENCE]`.

### 7.3 Completeness of the trigger (every exploit causes a freeze)

The trigger is **not complete**.

- An attacker with a soundness bug proves one false root, at a height `h`. No conflict exists until an honest prover
  proves the canonical root at the same height `h`. The trigger compares equal heights only `[INFERENCE]`.
- If withdrawals pay against a root in the same block in which the root is registered, the funds leave before any
  honest proof can arrive. The freeze then comes too late `[INFERENCE]`.
- So the trigger protects funds only under three conditions `[INFERENCE]`:
  1. Payouts against a root wait for a delay `D` after registration. In this design, the ForceMove challenge duration
     gives that delay. `challenge` sets `finalizesAt = block.timestamp + challengeDuration` (`nitro`
     `src/nitro/ForceMove.sol:66-80`).
  2. A live honest prover watches every registered root, and proves the same height within `D`.
  3. Honest proving at height `h` is possible within `D`. That needs L1 data for `h` and a working zkVM.
- An attacker who uses the bug only once still causes a conflict, if condition 2 holds. A single use does not avoid
  detection. It only bets on condition 2 failing. With no honest watcher, nothing detects the false root.
- A second trigger helps for proof-system bugs. RISC Zero's circuit breaker freezes on a valid proof of the all-zero
  claim digest (`r0e` `RiscZeroVerifierEmergencyStop.sol:45-59`). `L2BridgeApp` can accept the same evidence against
  its pinned verifier and set `frozen`. This needs a public exploit. It does not catch a bug in rv32im semantics or in
  the STF program, because such a bug does not forge an arbitrary claim `[INFERENCE]`. SP1 has no equivalent contract
  (`sp1c` `contracts/src/`).

### 7.4 What a multi-proof design changes

- **2-of-2.** One zkVM bug no longer produces a registered false root. The attacker needs two independent zkVM bugs at
  the same time. The freeze rule then matters only for a program-layer bug, or for a bug in both systems. The shared
  kona base is the remaining correlated risk (§1.4). Liveness falls: if one system cannot prove a block, the bridge
  stops registering roots. GHSA-x8rg made some blocks unprovable in OP Succinct (§6.3). Users then exit through Nitro
  (design) `[INFERENCE]`.
- **1-of-2 plus freeze.** Detection improves, because the honest counter-proof can come from either system. The
  completeness gap in §7.3 stays. Safety is the safety of the weaker system plus a live watcher `[INFERENCE]`.
- Neither rule removes the program-layer risk. Only a second, independent STF implementation removes it. OP Succinct
  and Kailua do not provide one (finding 2) `[INFERENCE]`.

---

## 8. The trusted primitive in Lean

The 2026-10-28 ruling keeps the ZK verifier trusted. Step 4 does not refine the verifier to `VerifyAccepts` (`laconic`
`labbook/2026-10-28-verification-chain-rulings.md:7-8`; `engineering/build-plan.md:136-138`). The pool model already
shows the pattern:

- `VerifyAccepts` is an abstract `axiom … → Prop` (`lean` `ArmadaLean/PoolSafety.lean:578-581`).
- Soundness is the hypothesis `Constrained w pub` that each theorem takes directly (`lean`
  `ArmadaLean/PoolOwned.lean:93-99`).
- Completeness is a named structure that only "can happen" theorems take (`lean` `ArmadaLean/PoolSafety.lean:583-589`;
  `ArmadaLean/PoolOwned.lean:108-119`).

The bridge model should name these hypotheses. They are written in plain words. Names in code font are proposals.

1. **`L2VerifyAccepts pub` (abstract).** Some proof over the public values `pub` makes the pinned verifier return
   without a revert, under the pinned program key. Left abstract, as `VerifyAccepts` is.
2. **zkVM soundness (`ZkvmSound`).** If `L2VerifyAccepts pub` holds, then the program named by the pinned key ran to a
   successful exit on some private input. That run committed exactly `pub`. The rejected 2-of-2 variant (§9.2) would
   take `ZkvmSound_A ∨ ZkvmSound_B`, and acceptance would mean that both verifiers accept.
3. **Program correctness (`ProgramImplementsDerivation`).** If the pinned program commits `pub` on any private input,
   then OP Stack derivation, started at `pub.preRoot`, reads the L1 chain that ends at `pub.l1Head` under the rollup
   config named by `pub.configHash`. It reaches height `pub.l2Height`, and the output root there is `pub.postRoot`.
   This hypothesis is separate from item 2, because the kona bugs in §6.3 break item 3 and not item 2. It is shared by
   both systems in a multi-proof design.
4. **Key uniqueness (`ProgramKeyUnique`).** The program key identifies one program. Two different programs do not share a
   key. OP states the same assumption as aZKG-002 (`specs` `zk-dispute-game.md:369-372`). The range-key commitment
   inside the aggregation statement also needs this property.
5. **Digest binding.** The public-values digest does not collide on the values that a run uses. SP1 uses SHA-256 with 3
   bits cleared, so 253 bits (`sp1c` `SP1VerifierGroth16.sol:43-45`). RISC Zero uses SHA-256 of the journal (`kailua`
   `KailuaVerifier.sol:367`). The model can treat these as symbolic hash terms, as it treats Keccak and Poseidon
   (`laconic` `engineering/build-plan.md:123`).
6. **Config binding.** The config hash identifies one rollup config. That config names the deposit contract that holds
   the L1 funds and the `SystemConfig` address. OP derivation reads deposits from the deposit contract (`specs`
   `specs/protocol/deposits.md:340-344,360-365`). It reads the batcher from `SystemConfig` updates
   (`specs/protocol/system-config.md:176,348-352`).
7. **Derivation determinism (a lemma about the derivation model, not a crypto assumption).** For one L1 chain, derivation
   gives at most one output root at each height. A longer L1 prefix does not change a root that a shorter prefix
   already derived. The freeze soundness in §7.2 needs this property.
8. **L1 oracle.** A stored checkpoint holds only the hash of an ancestor of the block in which it was written. This
   models `BLOCKHASH` and EIP-2935.
9. **Completeness (`L2ProofCompleteness`), for "can happen" and liveness theorems only.** For every correct transition
   from a registered root, an honest prover can build an accepted proof. The proof fits the 2^24 gas cap (`laconic`
   `labbook/2026-10-28-verification-chain-rulings.md:46-49`). The prover delivers it within the delay `D` of §7.3. The
   rejected 2-of-2 variant would need this hypothesis for **both** systems.
10. **Freeze evidence.** If the freeze flag is set, then two accepted statements exist with equal start conditions and
    different roots at one height. Or a valid circuit-breaker proof exists. This is a theorem about the contract, not an
    assumption.

Two model changes follow from §3 and §4 `[INFERENCE]`:

- The Lean rule `Rules.supports` is a computable `Bool` function. The model states that `Rules` "has no access to chain
  state" (`lean-app` `ArmadaLean/Nitro/Channel.lean:130-136`). With the registry and the freeze flag, support depends on
  contract storage. That is a change to the model, not only to the app. Registration only adds roots, so support only
  grows, until a freeze. A freeze removes all support. The model needs a storage parameter for `supports`. The liveness
  theorems must take "not frozen" as a hypothesis, or must state the Nitro exit as the outcome after a freeze.
- The same model comment says that go-nitro declares `stateIsSupported` `pure` (`lean-app` `Channel.lean:132-133`).
  armada-nitro already declares it `view` (`nitro` `src/apps/ZkForceMoveApp.sol:58`). A `view` app that reads storage
  breaks the "no chain state" premise. Even a direct `blockhash` check breaks it.
- `Rules.NoUnilateralSupport` requires that every participant signed some state in the support proof (`lean-app`
  `ArmadaLean/Nitro/Channel.lean:204-206`). The bridge supports a state that one participant signed (design). So
  `finalizable_or_two_enabled` and `single_enabled_is_finalizable` probably do not apply to the bridge channel
  (`laconic` `labbook/2026-10-28-verification-chain-rulings.md:121-124`).

---

## 9. Recommendation

### 9.1 Choice (operator decision, 2026-11-01)

Use **RISC Zero 3.0.0 Groth16 alone**, with the Kailua FPVM program. Pin the base verifier `RiscZeroGroth16Verifier`
`0x2a098988600d87650Fb061FfAff08B97149Fa84D` directly. Do not call `RiscZeroVerifierRouter`. Bind all inputs through
the proven-root registry of §3.3, with stored L1 checkpoints (§4.2). Freeze on a conflict at one L2 height, and on the
RISC Zero circuit-breaker evidence (§7.3).

Reasons:

1. **The circuit-free exit limits the damage of a zkVM bug.** Users exit value held in Nitro channels by ForceMove with
   their latest co-signed state (exit path 3 of the design). That path uses no zkVM and no SNARK for plain tokens. A
   shielded payout uses one Circom Groth16 proof (`TransactOwned`), which a phone can build from public data
   (`lean` `ArmadaLean/PoolSafety.lean:1060`, `owned_spend_witness_exists`). So a zkVM bug can reach only the value
   that depends on proven roots: canonical withdrawals and L2 contract state. 2-of-2 protects only that value
   `[INFERENCE]`.
2. **The trusted setup passes Stage 1.** L2BEAT rates the RISC Zero Groth16 setup yellow (§1.2). The 2026 Stage 1 rule
   forbids only red setups.
3. **The pinned verifier has no admin.** The base verifier has no owner, no router and no pause (§1.1, `chain`).
4. **A circuit-breaker trigger exists.** `RiscZeroVerifierEmergencyStop` accepts a valid proof of the all-zero claim
   digest as evidence of a proof-system bug (`r0e` `RiscZeroVerifierEmergencyStop.sol:45-59`). `L2BridgeApp` can accept
   the same evidence against its pinned verifier. SP1 has no equivalent contract (§7.3).
5. **Proof requests can go on chain.** The Boundless market takes a request either in a transaction to the market
   contract or through an off-chain order stream. The docs name the on-chain path for "censorship-resistance
   requirements" (`web` https://docs.boundless.network/developers/proof-lifecycle). Locking provers post collateral
   that is slashed if they miss the deadline, and the slashed collateral pays another prover to fulfil the request
   (same source). So a user or a watcher can buy a proof of a missing root without the cooperation of any operator.
   This keeps exit path 1 alive when no honest prover runs. Succinct also runs a prover market, but requests go
   through an off-chain auctioneer service that the docs compare to an L2 sequencer (`web`
   https://docs.succinct.xyz/docs/protocol/spn/architecture).
6. **Cost.** 254,073 gas for each registered root (§1.1, `gas`), paid once for each L2 output, not for each user. One
   system also halves proving cost and removes the second liveness dependency of 2-of-2 (§7.4).

Residual risks that this choice keeps:

- **One zkVM bug can register a false root.** RISC Zero had three on-chain-relevant soundness advisories in 12 months
  (§6). The defence is the payout delay `D`, a live honest watcher, and the freeze (§7.3). The freeze trigger is not
  complete.
- **A kona bug is not detected** (§1.4, §6.3). 2-of-2 would not detect it either.
- **The setup is yellow, not green.** It passes Stage 1. It is weaker than SP1 PLONK on Aztec Ignition (green).
- **The Boundless market is young and has its own dependencies** `[INFERENCE]`:
  - The market contract runs on Base mainnet (chain ID 8453), not on Ethereum L1 (`web`
    https://docs.boundless.network/provers/quick-start). An on-chain request therefore depends on Base liveness.
  - Lock collateral is in ZKC (`web` https://docs.boundless.network/developers/proof-lifecycle). Market liquidity and
    prover participation are unmeasured here.
  - The default delivered seal is a Merkle inclusion proof into an aggregated Groth16 proof. The docs recommend
    `RiscZeroVerifierRouter` to verify it (same source). A pinned base Groth16 verifier cannot check that seal. A
    request for `L2BridgeApp` must set `Requirements.selector` to the plain Groth16 type (same source, `Requirements`).
- **The Kailua image ID still needs our own reproduction** (§5).

### 9.2 Rejected alternatives

- **2-of-2 multi-proof (SP1 PLONK and RISC Zero Groth16).** It stops one zkVM bug from registering a false root by
  itself (§7.4). The operator rejected it, for three reasons: exit path 3 already bounds the damage of a zkVM bug
  (§9.1, reason 1); it gives no protection against a kona bug; and it doubles proving cost and liveness dependencies
  (§1.1, §7.4).
- **SP1 PLONK alone.** This is the strongest single-system alternative. Its setup is green (Aztec Ignition, §1.2). It is
  OP's own ZK backend (`specs` `specs/fault-proof/stage-one/zk/zk-dispute-game.md:105`). L2BEAT verified the most
  OP Succinct artifacts (§5). It loses on three points: no circuit-breaker contract (§7.3), a request path that goes
  through an off-chain auctioneer (§9.1, reason 5), and higher gas (291,639, §1.1). Its advisory count is four in 18
  months, against three in 12 months for RISC Zero (§6). The two rates are similar `[INFERENCE]`.
- **SP1 Groth16.** Excluded, because its setup is red (§1.2).

### 9.3 Open questions for the operator

1. **Anchoring.** Does the operator accept the proven-root registry (§3.3)? Without it, the starting root is free
   (finding 1). The other way is a recursive proof from genesis. That needs a custom program.
2. **Payout delay.** What is the delay `D` between registering a root and paying against it (§7.3)? It must exceed the
   time to buy an honest proof on Boundless. This document did not find measured proving times for Kailua.
3. **Vendor stop signal.** May `L2BridgeApp` freeze when RISC Zero's emergency stop stops the pinned verifier's proxy
   (§2.2, option 3)? This trusts the vendor in the stop direction only. The RISC Zero guardian is an EOA (`chain`).
4. **Program diversity.** Is the kona risk acceptable? Kailua uses kona for the state transition (finding 2).
5. **Deposit contract and `SystemConfig`.** Which L1 contract emits `TransactionDeposited` for this L2? Who owns its
   `SystemConfig`? The owner can change the batcher (`specs` `specs/protocol/system-config.md:348-352`). An owner is an
   admin over derivation. That conflicts with the Stage 2 target `[INFERENCE]`.
6. **`conclude` bypasses the app.** `ForceMove._conclude` finalizes a unanimously signed final state and does not call
   `stateIsSupported` (`nitro` `src/nitro/ForceMove.sol:140-170`). If anyone can sign for the proxy slot, the
   custodian alone can conclude any outcome. The proxy slot must refuse every `isFinal` state, or apply the registry
   checks to it `[INFERENCE]`.
7. **Freeze as a lever.** Any statement-binding bug becomes a bridge-wide freeze (§7.2). Is a freeze, followed by the
   Nitro exit, an acceptable outcome for a liveness-class bug?
8. **Kailua for a non-Kailua contract.** Can Kailua produce a single transition proof with `precondition_hash = 0` for
   any range that `L2BridgeApp` names? This document did not confirm it.
9. **Boundless on Base.** Is a dependency on Base liveness acceptable for buying a missing proof? The other way is to
   run an honest prover directly, which needs proving hardware or a contract with a prover.

### 9.4 Questions that no primary source answered

- Whether each kona advisory also affected Kailua (§6.3).
- Measured proving times of Kailua for a given L2 range, on Boundless and on dedicated hardware.
- Measured prover participation and fill rates on the Boundless market.
- The source repository of the program behind OP `ZKDisputeGame`'s `absolutePrestate`. The spec names SP1 only.
- For the SP1 alternatives only: whether the SP1 v6.1.0 wrap circuit uses the `gnark/std` gadgets of
  GHSA-3mvx-pp85-pm65 (§6.1), the gas of the v6.1.0 SP1 contracts (§1.1), and which phase-1 ceremony SP1 v6.1.0 used
  (§1.2).
