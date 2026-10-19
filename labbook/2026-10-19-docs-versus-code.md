# The docs versus the ship-armada code

Lab book entry. Session 2026-10-19 19:43 UTC to 23:36 UTC.
Session id `01a14aa3-1d15-74fd-9f8d-df18ee172022`.
Repos affected: `armada-laconic` (docs). Repos read: `armada-poc`,
`armada-circuits`, `armada-deployments`, `armada-nitro`, `armada-sdk`,
`armada-client`, `armada-relayer`, `armada-interface`.

A sibling entry, `2026-10-19-unified-stack-and-verified-note-sync.md`, records
the client and indexing decisions of the same session.

## 1. The question

The session started with two names in §8.2, "fork-lite" and "fork-full". It then
asked whether Armada offers a bridge from Railgun. The answer showed that the
docs carry designs that no one intends to build. The session then asked how much
of the docs still describes the code, now that the team has the ship-armada
repos.

## 2. Evidence

### What the session read or measured

- Session `01a150cb` (2026-10-18, 23:03 to 23:26 UTC) holds the argument against
  a dispute game inside a proof. ADR-0005 kept only "research-grade".
- ADR-0006 D6.1 described a Railgun "onboarding bridge". D6.2 described
  cross-pool membership proofs that "inherit Railgun's crowd". The only source
  found on this machine is session `01a06ee5`, 2026-09-05 00:20 to 00:22 UTC.
  The original design session `01a058db` is not on this machine and not in the
  ingest-sessions index (0 records).
- `armada-poc` `contracts/railgun/README.md` says the Railgun contracts "are
  **not deployed** in the POC". No contract or script imports, migrates or
  bridges from Railgun (search of `contracts/` on `design/rendered-baseline`
  `a73fd11b`).
- `armada-nitro/src/HubCustodian.sol:4` imports `FROST` from
  `frost-secp256k1-evm`. Lines 98 and 156 verify with it. The docs name
  `chain-signatures` threshold Schnorr with kyber DKG (ADR-0016 D16.2,
  `shielded-nitro-bridge-design.md`, `nitro-bridge-audit.md`, T2.4, T3.2).
- A two-pass audit compared every doc claim with the code. Pass 1 built
  inventories with printed citations: `local://inv-InvPool.md`,
  `inv-InvCircuitsDeploy.md`, `inv-InvNitroSdk.md`, `inv-InvClientOps.md`.
  Pass 2 compared the docs: `local://cmp-CmpOnchain.md`, `cmp-CmpArc42.md`,
  `cmp-CmpNitroHub.md`, `cmp-CmpClientOps.md`. The refs read were
  `armada-poc@102f17b8`, `armada-circuits@ef2682f`,
  `armada-deployments@bac3639`, `armada-sdk@db5757f`,
  `armada-client@78a8fdc`, `armada-relayer@bbf5e62`,
  `armada-interface@6968cd7` and `armada-nitro@55939ef`.
- An earlier scout audit (four agents on `google/gemini-3.8-flash`) was
  discarded. Its tables cited doc lines that do not hold the quoted claims.
- The lead checked these audit facts by hand:
  - The Phase-2 ceremony is not finished. Tag `v1.0.0-ceremony` (`c663a1f`) is
    not on `main`. Its `ceremony/manifest.json` lists the shapes with `"status":
    "open"` and no beacon. `main` has only `scripts/setup-dev.sh`.
  - Mainnet instance `armada-1` (2026-10-08) deployed the ARM crowdfund,
    governance and CCTP. Its manifest says "shielded pool ships in a later
    launch".
  - `PrivacyPool.sol:39` sets `MAX_SHIELD_FEE_BPS = 1000`. The pool routes calls
    to four modules by `delegatecall` (lines 20 to 25). It has no proxy.
  - The selective-disclosure circuit exists only on unmerged branches:
    `armada-circuits` `disclosure-circuit` (`7175b46`,
    `src/operations/NullifierDisclosure.circom`) and `armada-sdk`
    `selective-disclosure` (`0aa2ba1`).
  - `armada-nitro` has a watchtower in its e2e gate harness
    (`env/gates/watchtower/watch.ts`). No client watchtower exists.

### Operator rulings, verbatim

- 2026-10-19T19:43:53Z — fork-lite and fork-full are confusing names and isn't
  the argument against fork-full better documented than that? rather it was
  discussed in depth in a session and that rational should survive as a
  footnote.
- 2026-10-19T19:59:02Z — The bridge is intended more as a UX thing than some
  sort of protcol level functionality.
- 2026-10-19T20:03:56Z — the idea is that we'd somehow get added security from
  the existing railgun pool. But this is impossible for the reasons you explain
  in design 1, right?
- 2026-10-19T20:05:47Z — this needs to be cleaned up so that there are no
  references to the bad design in these docs and instead we explain that we just
  need to bootstrap the new pool with new users and that we will make it easy to
  move over in our wallet, with all the ceveats around how that impacts security
- 2026-10-19T20:08:29Z — also we are not using PoI we are using selective
  disclosure.
- Ask answer, 2026-10-19 — ADR-0006: "Edit ADR-0006 in place".
- 2026-10-19T20:17:25Z — we will be using the ship-armada repos including
  armada-poc
- 2026-10-19T20:19:04Z — these docs were written before we had access to
  armada-poc so that stuff actually already happened.
- 2026-10-19T20:20:30Z — the licensing is neither open nor a risk and not
  something we need to focus on or discuss in these docs.
- 2026-10-19T20:22:02Z — the dss/frost thing has already been resolved in the
  code.

## 3. Findings and decisions

1. §8.2 describes the two channel-privacy designs by what they do. A footnote
   keeps the names fork-lite and fork-full and the reasons from session
   `01a150cb`. Committed as `2dc133f` (ruling 19:43:53Z).
2. Cross-pool membership proofs cannot work. A spend hides only among the notes
   it could consume. A Railgun note that our pool accepts lets its owner spend
   the same value in both pools, and our pool pays tokens it never received. A
   circuit that refuses Railgun notes gives no extra hiding. No design that uses
   Railgun's tree adds privacy or security (ruling 20:03:56Z).
3. The move from Railgun is a wallet feature in T6.7: a Railgun unshield, then a
   shield into our pool. No contract of ours reads or touches Railgun's pool. The
   crowd grows only from deposits into our own pool (rulings 19:59:02Z and
   20:05:47Z).
4. The move has these security costs, which the wallet shows before the user
   confirms: the token, the amount and the time are public on both pools; the
   unshield recipient is public; the value leaves Railgun's larger crowd for our
   smaller one; a later exit can link back; Railgun charges its unshield fee.
5. Selective disclosure (ADR-0020, accepted) replaces Proofs of Innocence. No
   gate blocks a deposit, a transfer, a withdrawal or channel funding (ruling
   20:08:29Z).
6. Items 2 to 5 are committed as `c67c3af`. ADR-0002, ADR-0006, ADR-0013 and
   ADR-0014 were edited in place, against D18.1, on the operator's choice.
7. The ship-armada code already built what the docs plan as T0.0 to T0.3:
   - T0.0: `armada-poc` `PrivacyPool` with four modules, no proxy, a token
     blocklist, a shield fee (50 bps in the Sepolia deploy script) and
     `ArmadaFeeModule`.
   - T0.1: `armada-circuits`, 19 `Transact` shapes. The ceremony is open, not
     finished.
   - T0.2 and T0.3: `ArmadaNitroAdjudicator`. It funds from a note with
     `depositFromNote` and pays out with allocation type `SHIELD`. Six commits
     change the vendored go-nitro. No separate `fundChannel` contract exists.
   The docs must describe this code, not plan it (ruling 20:19:04Z).
8. Licensing is out of the docs (ruling 20:20:30Z).
9. The hub custodian signs with FROST over secp256k1, not with `chain-signatures`
   (ruling 20:22:02Z).
10. Of the 30 items in §5, the audit found 4 built, 8 partly built, 17 not built
    and 1 not checkable. 11 of the not-built items assume the Laconic substrate:
    T1.0, T1.1, T2.0 to T2.4, T3.0 to T3.2, T6.0, T6.3 and T6.5.
11. The docs omit the governance contracts in `armada-poc`: `ArmadaGovernor`,
    `ArmadaTreasuryGov`, `TreasurySteward`, `ArmadaToken`, `ArmadaCrowdfund`,
    `ArmadaWindDown`, `ArmadaRedemption` and the revenue contracts.

## 4. The check for each decision

| Decision | Check | Result |
|---|---|---|
| 1 | Pages build of `2dc133f`; the footnote text on the live page | built; 1 match |
| 2 to 6 | Pages build of `c67c3af`; every published page searched for "cross-pool", "Proof of Innocence", "POI", "onboarding bridge", "import bridge" | built; 0 matches |
| 7 | Inventories pass 1, comparison pass 2, lead spot checks listed in §2 | done |
| 9 | `HubCustodian.sol:4,98,156` | confirmed |
| 10, 11 | `local://cmp-CmpArc42.md` T-item table | recorded; not yet in the docs |

## 5. Rejected options

- **A Railgun onboarding bridge as a protocol component.** Rejected. The move is
  a wallet feature (ruling 19:59:02Z).
- **Cross-pool membership proofs.** Rejected. They either break solvency or add
  no hiding (decision 2).
- **Proofs of Innocence as an entry gate.** Rejected (ruling 20:08:29Z).
- **A new ADR that supersedes D6.1 and D6.2, under D18.1.** Rejected on the
  operator's choice. ADR-0006 was edited in place.

## 6. Open items

- The rewrite scope is not decided: a targeted fix of T0.0 to T0.3, A.2 to A.5,
  ADR-0014 and ADR-0016, or a full re-anchor of §5 and the tier docs on the code.
- `mobile-privacy.md` and `architecture.md` still call the pool Railgun's and
  "consumed, not built".
- The `armada-nitro` README pins `lib/armada-poc` at `532bc064`. The submodule is
  at `2db1d338`.
