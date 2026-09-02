# Mobile ZK proving — research note

Supporting research for [`mobile-privacy.html`](./mobile-privacy.html) crux #2 ("Mobile Railgun prover"). Compiled 2026-09-02 from two source-verified research passes plus the mopro benchmark page. Distinguishes **fact** from **[inference]**. This is a research record, not a site page.

## Question

Railgun proves shielded transactions with Circom + **Groth16 / BN254**. Client-side proving on a phone is the dominant mobile difficulty. Are there constructions or prover runtimes that make this feasible on mobile while staying **Ethereum-verifiable**, and who is working on it?

## Verdict

**Mobile client-side proving is a mostly-solved *engineering* problem for small-to-medium circuits (shipped at scale by World App), but Railgun-scale circuits sit at the memory frontier.** The pragmatic win is a better **prover runtime** (keep Railgun's Groth16, run it natively on-device), not a different pool — every Ethereum-L1 privacy pool still puts a per-user Circom/Groth16 proof on the device (Privacy Pools uses the *same* stack as Railgun, so it is no lighter). The only constructions that truly remove the heavy per-user proof do so by offloading to an MPC/relayer cluster (Renegade) or by moving to their own L2 (Aztec).

## mopro performance (source: zkmopro.org/docs/performance, v0.3)

mopro = the Ethereum Foundation / PSE mobile-prover toolkit: native Rust bindings (Swift / Kotlin / React Native / Flutter) wrapping **Circom, Halo2, Noir**; bundles **rapidsnark** for native Groth16. Ethereum-compat: Circom+Groth16 has a standard Solidity `Groth16Verifier`, so proofs verify on L1 directly.

### Circom + rapidsnark, iPhone 16 Pro (2024) — best-adapter witness-gen + proof-gen

| Circuit | Witness gen (best) | Proof gen (rapidsnark) | ≈ end-to-end | vs browser snarkjs |
|---|---|---|---|---|
| Semaphore-32 | ~15 ms (circom-witnesscalc) | 143 ms | **~0.16 s** | ~6.9× |
| SHA256 | 41 ms (witnesscalc) | 187 ms | **~0.23 s** | ~8.2× |
| Keccak256 | 75 ms (circom-witnesscalc) | 630 ms | **~0.7 s** | ~8.2× |
| RSA | 153 ms (witnesscalc) | 749 ms | **~0.9 s** | ~8.8× |
| Anon Aadhaar (RSA-heavy) | 285 ms (witnesscalc) | 3,132 ms | **~3.4 s** | — |

Android (Samsung S23 Ultra, 2023) is the same order; proof-gen is ~13–15× faster than browser snarkjs (e.g. Anon Aadhaar 3.4 s via rapidsnark vs **51.5 s** snarkjs). The "up to ~20×" headline is driven by native rapidsnark replacing browser snarkjs.

### The memory wall (the Railgun-relevant risk)

- **Halo2** Keccak256: ~11 s on iPhone 15 Pro. **Halo2 RSA crashes** on both iPhone 15 Pro and Pixel 6 Pro — the circuit needs ~5 GB while phones cap app memory at ~3 GB.
- **Noir / UltraHonk** (bb, MacBook Air M3 / Pixel 6): Keccak 349 ms iOS, RSA 312 ms iOS, zkEmail 1.3 s iOS, Anon Aadhaar 2.2 s iOS, Semaphore 828 ms iOS — competitive; this is Aztec's client-side path.

**Implication for Railgun:** none of these is a Railgun *transaction* circuit (large Merkle note-scan + shielded transfer), which is bigger than these micro-benchmarks. The proving *engine* is fast enough; the binding constraint is **memory / circuit size**. A Railgun-scale circuit likely lands multi-second and memory-pressured — the same wall that crashes Halo2-RSA. mopro's docs explicitly advise benchmarking the *actual* circuit. [inference, high-confidence]

## Two levers

### Lever A — keep Railgun's construction, change the prover runtime (recommended)

| Option | What it improves | Ethereum-verifiable | Maturity | Who |
|---|---|---|---|---|
| **mopro** | native Circom/Halo2/Noir prover for iOS/Android/RN; ~8–20× vs snarkjs | yes (standard Groth16 verifier) | active lib | EF / PSE, 0xPARC |
| **rapidsnark / ProveKit** | native ARM Groth16; ProveKit adds WHIR→gnark **Groth16 recursive wrapper** for on-chain; targets large-circuit memory | yes | prod (World App) / ProveKit pre-audit | World Foundation |
| **ICICLE + Metal / IMP1** | mobile-GPU MSM/NTT/Sumcheck on Apple Silicon (v3.6, Mar 2025) | yes (accelerates under Groth16) | lib (commercial license for prod) | Ingonyama |
| **Collaborative / MPC proving (coSNARK)** | offloads proving off-device without revealing the witness; output is standard Groth16/PLONK | yes | Renegade mainnet (L2); TACEO testnet | Renegade, TACEO |
| **Folding (Nova/HyperNova, Sonobe)** | incremental proving fits Merkle note-scan; compresses to Groth16 EVM verifier (~600–750k gas) | yes | **experimental, unaudited** | PSE + 0xPARC |

Real deployment proof: **World App** generates Semaphore/Groth16 proofs on-device for millions via native rapidsnark. Coordinating effort: **PSE "Client-Side Proving"** with a public mobile benchmark suite (ethproofs.org/csp-benchmarks).

### Lever B — alternative Ethereum-compatible pools (mostly not lighter, or not L1)

| Project | Construction | Client proving vs Railgun | Ethereum | Status |
|---|---|---|---|---|
| **Privacy Pools** (0xbow) | Circom + Groth16/BN254 | **same class — not lighter** | L1 mainnet | live (Mar 2025) |
| **Renegade** | MPC + collaborative SNARK (relayer proves) | **no heavy per-user proof** | Arbitrum L2 | live (Sep 2024); dark-pool DEX, SDK-only, no mobile app |
| **Aztec** (Noir/UltraHonk, PXE) | client-side proofs, own rollup | client-side, mobile-optimized | L2 rollup → L1 | testnet; V5 proving vuln (Aug 2026) → V6 |
| **Nocturne** | account-based ZK + stealth | — | L1 | **shut down (Jun 2024)** |
| **Tornado Nova** | Circom+Groth16 | same class | L1 (immutable) | no active team; sanctions lifted Mar 2025 |
| **Panther** | zAsset UTXO SNARK | same class | Polygon (not L1) | mainnet on Polygon |
| **Firn** | account-based ZK, Sigma/Bulletproof | lighter *sync* (sub-KB), not lighter proving | L1 | browser wallet, small team |
| **Nightfall_4** (EY) | zk-rollup | rollup, enterprise | L1 via rollup | permissioned |

## Recommendation for our design

Keep Railgun and its L1 Groth16 verifier. Adopt **mopro + rapidsnark** for on-device proving, **ICICLE-Metal** to accelerate MSM/NTT, and a **coSNARK server-assist** fallback for weak/low-memory devices. Longer term, **ProveKit's WHIR→Groth16 wrapper** and **Sonobe folding** target the large/incremental-circuit memory wall but are pre-audit. Do not switch pools for proving reasons — the win is the runtime, not the construction. Must benchmark the *actual* Railgun circuit on target devices before committing.

## Caveats

- Railgun's specific bottleneck (large note-scan + Merkle in Groth16) is at the upper edge of on-device provers today; folding or server-assist is the direct mitigation but least mature / adds a non-collusion assumption.
- Sonobe folding and much WHIR/Spartan mobile-prover work is experimental/unaudited (2026).
- Renegade has the right offload model but is a dark-pool DEX on L2, not a general pool; Aztec is L1-settled as a rollup, not per-tx L1 verification.
- ICICLE GPU backends default to an R&D license; production needs a commercial license.

## Citations

- mopro benchmarks — https://zkmopro.org/docs/performance ; repo https://github.com/zkmopro/mopro
- ProveKit — https://github.com/worldfnd/ProveKit (WHIR + gnark Groth16 recursive verifier)
- rapidsnark — https://github.com/iden3/rapidsnark ; witnesscalc — https://github.com/0xPolygonID/witnesscalc
- ICICLE Metal — https://github.com/ingonyama-zk/icicle (v3.6, Mar 2025)
- coSNARKs — https://docs.taceo.io ; Renegade — https://docs.renegade.fi/concepts/mpc-zkp
- Folding — https://sonobe.pse.dev (Decider → Groth16 EVM verifier)
- PSE client-side proving — https://pse.dev/projects/client-side-proving ; benchmarks https://ethproofs.org
- Aztec — Noir + Barretenberg UltraHonk; PXE client-side proving
- Privacy Pools — https://docs.privacypools.com/layers/zk
- Full agent transcripts: `history://MobileProvingTech`, `history://EthPrivacyPoolAlternatives`
