# Trusted primitives and blocked witnesses: the options, and a survey of concrete instances

Date: 2026-11-14. Session `01a195d3`.
Related: review item 3 on ship-armada/armada-lean#1 (coax1d, 2026-11-09); `labbook/2026-11-11-library-selection.md` (decision D3: Poseidon is a trusted primitive).

## The problem
armada-lean declares three primitives as `axiom`s: `EdDSAVerify` (`Transact.lean`), `VerifyAccepts` (`PoolSafety.lean`) and `VerifyOwnedAccepts` (`PoolOwned.lean`). Each has no definition. A witness for a theorem whose premise mentions one of them must show that the primitive holds of a concrete input. No such input can be built, so `scripts/check.sh` reports the witness as `WITNESS_BLOCKED`.

At `integrated` `f753341b` there are 14 blocked witnesses, in three groups:
- **Refinement (3):** `Pool.apply_forward_refinement`, `Pool.apply_backward_refinement`, `Pool.apply_backward_refinement_run`. Their premises (`verify.Realizes`, `TransactProofSoundness`, `OwnedProofSoundness`) are about the verifier.
- **Can-happen (5):** `transact_can_happen`, `transact_spend_can_happen`, `unshield_can_happen`, `owned_spend_can_happen`, `owned_transact_can_happen`. They need a verifier that accepts a real proof.
- **Arithmetic (6):** `transact_value_conserved`, `transact_inputs_distinct_nullifiers`, `transact_public_outputs_determined`, `transact_not_blocked`, `transact_spends_only_owned`, `transact_unshield_exact`. They take `Constrained` (`Transact.lean:142`), which bundles the field `.sig : EdDSAVerify` with the arithmetic constraints. The proof of `transact_value_conserved` does not use `.sig` (armada-lean `aln-c67`). The other five are not yet checked.

The reviewer proposed to make the primitives parameters instead of axioms.

## The options

1. **A symbolic model of the primitives.** Define each primitive in the free algebra, the same treatment `Term` gives Poseidon. A signature is the term `sign(sk, m)`; `EdDSAVerify (pk sk) m s` holds exactly when `s = sign(sk, m)`. A proof term carries its witness; `VerifyAccepts` holds exactly when that witness satisfies the circuit. Unforgeability and knowledge soundness hold by construction. The symbolic model is then the assumption, and the docs state it, as for the hashes (review item 5).
2. **Parameters with laws.** A structure bundles each primitive with its laws. Completeness: an honest signature or an honest proof verifies. This excludes a verifier that accepts nothing. Soundness: acceptance implies a satisfying witness or a signature by the key holder. This excludes a verifier that accepts everything. Each theorem quantifies over every scheme that satisfies the laws it uses. The symbolic model of option 1 is one instance, and each witness uses it. A bare parameter, without the laws, is the reviewer's proposal. It is not sufficient: a witness could choose a verifier that accepts nothing, and the can-happen theorems would then say nothing.
3. **Concrete instances.** Real EdDSA on BabyJubJub and a real Groth16 verifier on BN254, in Lean. Each witness evaluates a real verification.

A fourth form, laws added as further `axiom`s, is excluded by armada-lean's own rule (`docs/spec-module-design.md`: no witness uses an axiom that asserts a property).

## Decision
Option 2, with option 1 as its instance, is the order of work. It closes review item 3 for all 14 blocked theorems. A theorem that does not need the signature law does not take it, so the arithmetic-family question becomes part of the same restatement. It is a statement change (process.md rule 8.3). The operator's approval of the statement texts is still required.

Option 3 is out of scope now (operator, 2026-11-14). It does not compete with option 2: a concrete scheme is a second instance of the same law structure. It would prove the completeness laws for the real schemes. Soundness stays an assumption in every option: unforgeability and knowledge soundness are computational properties, and Lean cannot prove them for the real schemes.

## Survey: concrete instances in Lean (2026-11-14)

| Piece | Exists? | Source |
|---|---|---|
| BN254 scalar field, with a machine-checked Pratt primality certificate | Yes | [CompPoly](https://github.com/Verified-zkEVM/CompPoly), `CompPoly/Fields/BN254.lean` |
| circomlib Poseidon circuit, sound and complete against its specification, arity 1 | Yes | [Clean](https://github.com/Verified-zkEVM/clean), `Clean/Circomlib/Poseidon.lean`; [zkSecurity, 2026-05-04](https://blog.zksecurity.xyz/posts/poseidon-clean) |
| Computable elliptic-curve arithmetic | Partly | [daira/CompElliptic](https://github.com/daira/CompElliptic): short Weierstrass over Pallas and Vesta, with proven fast projective arithmetic and multi-scalar multiplication. Pairings are a stated aim, not done. Its `Jubjub` field is Zcash's Jubjub, not BabyJubJub. |
| Elliptic-curve group law, for proofs (not computable) | Yes | Mathlib, `WeierstrassCurve.Affine.Point` |
| SNARK framework | Yes, no Groth16 | [ArkLib](https://github.com/Verified-zkEVM/ArkLib): interactive oracle reductions, Sumcheck with soundness and completeness. Its status page (last reviewed 2026-11-04) says no protocol under `ArkLib/ProofSystem/` instantiates its knowledge-soundness results yet. |
| BabyJubJub (twisted Edwards) or EdDSA-Poseidon | Not found | |
| BN254 pairing or a Groth16 verifier | Not found | |

## Cost of option 3
- **Arithmetic in Lean is not the obstacle.** `Nat` is arbitrary precision. The compiled code and the kernel both use GMP for `Nat` arithmetic on literals. `BitVec` and `bv_decide` cover bit-level facts.
- **The size of the code:** EdDSA verification is a few twisted-Edwards scalar multiplications and one Poseidon hash. A Groth16 verifier on BN254 needs the Fp2/Fp6/Fp12 tower, a Miller loop and a final exponentiation. Neither exists in Lean (see the survey).
- **The cost of checked evaluation:** a witness must run the verification. In the kernel (`decide`, `rfl`), every definition is unfolded. A pairing is thousands of Fp12 multiplications. Not measured. CompElliptic's trust rules allow `native_decide` only where the kernel cannot feasibly run a check, and they name the order of a fixed root of unity as one such fact. `native_decide` adds the axiom `Lean.ofReduceBool`, which armada-lean's axiom allowlist rejects.
- **Where it would go:** after the pool and Nitro emit Solidity, or upstream, in CompPoly or CompElliptic, as a BabyJubJub curve form and a BN254 pairing.
