# Prior formal work on UTXO channels, for the Armada adjudicator specification

Date: 2026-10-26. Scope: read only. This note answers three questions. Which formal work on UTXO payment channels and UTXO ledgers exists? Which work covers the four adjudicator issues? What can the Lean specification reuse?

The four issues, with the labels that this note uses:

- **I1 spend authority.** Who may spend or merge the outputs that back a channel.
- **I2 phase rules.** Deposits and merges occur only before finalization. Transfers occur only after finalization.
- **I3 non-interference.** No concurrent call can make a pending honest settlement stale. After finalization, only transfers that the outcome permits can spend the channel's notes.
- **I4 attribution.** The value of a channel equals the value of its own outputs. Each output belongs to at most one channel.

Local copies of every downloaded source are in `.scratch/utxo-lit/` (gitignored). The Hydra and LeanDY clones are there too.

Pinned versions:

- BOLTs: `lightning/bolts` master at `191462825ae341f0a27098765c075f01ca9bc6c2` (2026-10-19). Line numbers below refer to this commit.
- Hydra: `cardano-scaling/hydra` master at `e016d0a20d59ba1e7d34b7074ac0eeeed65201d1` (2026-10-22), directory `spec/`.
- LeanDY: `SecPriv/leandy` at `e0705986c5dea547848bf80c10b5c4f88ae6868c` (2026-07-07).
- Zcash protocol specification: version v2026.7.0-344-g2bdf39 [NU6.2], dated 2026-10-21.

Page numbers: "p." is the printed page number. For arXiv preprints and the EUTXO preprints, the PDF page and the printed page are the same. For the two theses, I give the printed page; the PDF page is larger by 18 (Aumayr) and by 8 (Melkonian).

## (a) Sources

Column "Formal": F = formal definitions and theorems; S = normative specification (MUST/MUST NOT rules); I = informal prose. Column "Tool": the proof tool, or "none". A cell in I1–I4 is "yes" when the source states a rule or theorem about the issue, "part" when it covers part of the issue, and blank when it does not cover it.

| # | Title | Authors | Year | Venue / URL that I opened | Formal | Tool | I1 | I2 | I3 | I4 |
|---|---|---|---|---|---|---|---|---|---|---|
| 1 | The Bitcoin Lightning Network: Scalable Off-Chain Instant Payments (draft 0.5.9.2) | Poon, Dryja | 2016 | https://lightning.network/lightning-network-paper.pdf | I | none | yes | | part | part |
| 2 | BOLT #2 Peer Protocol (channel establishment v1/v2, quiescence, splicing, close) | Lightning spec authors | 2026 (commit above) | https://raw.githubusercontent.com/lightning/bolts/master/02-peer-protocol.md | S | none | yes | yes | part | part |
| 3 | BOLT #3 Bitcoin Transaction and Script Formats (anchor outputs) | Lightning spec authors | 2026 | https://raw.githubusercontent.com/lightning/bolts/master/03-transactions.md | S | none | part | | yes | |
| 4 | BOLT #5 On-chain Transaction Handling | Lightning spec authors | 2026 | https://raw.githubusercontent.com/lightning/bolts/master/05-onchain.md | S | none | | | yes | part |
| 5 | Full Disclosure: CVE-2023-40231..40234 (replacement cycling), with the reply by Osuntokun (Lightning Labs) | Riard; Osuntokun | 2023 | https://www.mail-archive.com/lightning-dev@lists.linuxfoundation.org/msg03369.html | I | none | | | yes | |
| 6 | Announcing lnd 0.12-beta (anchor channels) | Lightning Labs | 2021 | https://lightning.engineering/posts/2021-01-28-lnd-v0.12 | I | none | | | part | |
| 7 | eltoo: A Simple Layer2 Protocol for Bitcoin | Decker, Russell, Osuntokun | 2018 | https://blockstream.com/eltoo.pdf | I | none | | | part | |
| 8 | Foundations of Bitcoin-Compatible Scalability Protocols (PhD thesis). Ch. 2 = Sleepy Channels; Ch. 3 = Generalized Channels; Ch. 6 = Bitcoin-Compatible Virtual Channels; Ch. 7 = Domino attack and Donner | Aumayr | 2024 | TU Wien reposiTUm: https://repositum.tuwien.at/bitstream/20.500.12708/197528/1/Aumayr%20Lukas%20-%202024%20-%20Foundations%20of%20bitcoin-compatible%20scalability%20protocols.pdf | F (UC) | none | part | part | part | yes |
| 9 | Sleepy Channels: Bi-directional Payment Channels without Watchtowers (CCS 2022) | Aumayr, Thyagarajan, Malavolta, Moreno-Sanchez, Maffei | 2022 | https://www.pure.ed.ac.uk/ws/portalfiles/portal/660834034/AumayrEtalCCS2022SleepyChannels.pdf | F (UC) | none | | | | |
| 10 | Daric: A Storage Efficient Payment Channel With Penalization Mechanism (abstract only) | Mirzaei, Sakzad, Yu, Steinfeld | 2022 | https://eprint.iacr.org/2022/1295 | F (UC) | none | | | part | |
| 11 | A Formally Verified Lightning Network | Fabiański, Stefański, Litos | 2025 | https://arxiv.org/abs/2503.07200 (v1 PDF) | F | Why3 | yes | | part | yes |
| 12 | Model Checking the Security of the Lightning Network | Grundmann, Hartenstein | 2025 | https://arxiv.org/pdf/2505.15568 | F | TLA+ / TLC | part | | part | |
| 13 | Payout Races and Congested Channels: A Formal Analysis of Security in the Lightning Network (CCS 2024) | Weintraub, Kumble, Nita-Rotaru, Roos | 2024 | https://arxiv.org/pdf/2405.02147v1 | F | Spin | | | part | |
| 14 | Lightning Network – formal verification of a payment protocol (talk slides) | Louistisserand, Jeanteur, Maffei | 2024 | https://gtmfs2024.sciencesconf.org/data/pages/Lightning_Network.pdf | F (slides only) | ProVerif | | | | |
| 15 | LeanDY, case study `Protocols/PaymentChannel` | SecPriv (TU Wien) | 2026 | https://github.com/SecPriv/leandy | F | Lean 4 | yes | | part | yes |
| 16 | Verified compilation from BitML to Bitcoin: an Agda odyssey (PhD thesis; mechanises BitML of Bartoletti and Zunino 2018) | Melkonian | 2024 | http://omelkonian.github.io/data/phd-thesis.pdf | F | Agda | yes | part | | part |
| 17 | The Extended UTXO Model (FC 2020 WTSC, preprint) | Chakravarty, Chapman, MacKenzie, Melkonian, Peyton Jones, Wadler | 2020 | https://omelkonian.github.io/data/publications/eutxo.pdf | F | Agda | part | | | part |
| 18 | Native Custom Tokens in the Extended UTXO Model (preprint) | Chakravarty, Chapman, MacKenzie, Melkonian, Müller, Peyton Jones, Vinogradova, Wadler | 2020 | https://omelkonian.github.io/data/publications/eutxoma.pdf | F | Agda | | | | yes |
| 19 | Hydra Head protocol specification (literate Agda) and its documentation page | Hydra team (IOG / cardano-scaling) | 2026 (commit above) | https://github.com/cardano-scaling/hydra/tree/master/spec ; https://hydra.family/head-protocol/docs/dev/agda | F | Agda | yes | yes | part | yes |
| 20 | Zcash Protocol Specification | Hopwood, Bowe, Hornby, Wilcox | 2026 | https://zips.z.cash/protocol/protocol.pdf | S | none | part | | | |
| 21 | RAILGUN developer guide: State Structure | RAILGUN contributors | n.d. | https://docs.railgun.org/developer-guide/engine-1/state-structure | I | none | part | | | |
| 22 | Aztec.nr API: `check_nullifier_exists` | Aztec | n.d. | https://docs.aztec.network/aztec-nr-api/mainnet/noir_aztec/oracle/nullifiers/fn.check_nullifier_exists | I | none | | | part | |
| 23 | Fuel Specifications: transaction input format | Fuel Labs | n.d. | https://docs.fuel.network/docs/specs/tx-format/input/ | S | none | | | yes | part |
| 24 | Concurrency and all that: Cardano smart contracts and the EUTXO model | Hryniuk (IOHK) | 2021 | https://www.essentialcardano.io/article/concurrency-and-all-that-cardano-smart-contracts-and-the-eutxo-model | I | none | | | part | |

Sources that I could not access in full text are listed at the end. I do not cite their content.

## (b) Definitions and theorems, per source

### 1. Poon and Dryja, Lightning paper (2016)

- I1. §3.1.1, p.7–8: "The output for this Funding Transaction is a single 2-of-2 multisignature script with both participants in this channel." Also p.8: "both signatures are needed to spend from the Funding Transaction, in other words, both Alice and Bob need to agree to spend from the Funding Transaction."
- I3, I4. §3.1.4, p.12: "These two Commitment Transactions spend from the same output (Funding Transaction), and have different contents; only one can be broadcast on the blockchain, as both pairs of Commitment Transactions spend from the same Funding Transaction."
- Revocation. §3.3.4, p.23: "This invalidation occurs by having both parties sign a Breach Remedy Transaction (BR1), which supersedes the Revocable Delivery Transaction (RD1)."
- The paper is prose. It has no definitions or theorems.

### 2. BOLT #2 (peer protocol)

- I4. Lines 61–62 (§"Definition of `channel_id`"): the `channel_id` is "derived from the funding transaction by combining the `funding_txid` and the `funding_output_index`". So the identity of a v1 channel is its single funding output.
- Dual funding. Lines 1137–1142: "It changes the previous protocol to allow the `accept_channel2` peer (the *accepter*/*non-initiator*) to contribute inputs to the funding transaction, via the interactive transaction construction protocol."
- I2 (splicing is a deposit or withdrawal in a live channel). Lines 1559–1563: "Splicing is the term given for replacing the funding transaction with a new one. For simplicity, splicing takes place once a channel is quiescent."
- I2. Lines 1668–1674, requirements on `splice_init`:
  "MUST NOT send `splice_init` if the channel is not quiescent."
  "MUST NOT send `splice_init` before sending and receiving `channel_ready`."
  "MUST NOT send `splice_init` while another splice is being negotiated."
  "MUST NOT send `splice_init` if it has previously sent `shutdown`."
- I2. Line 2156, requirement on `shutdown`: "MUST NOT send `shutdown` if there is a splice transaction that isn't locked yet."
- I1. Lines 1781–1782: "The splice transaction must spend the current channel funding output." Line 1895: "Spending the channel funding output requires a signature from both peers."
- I3 (contention between the channel's own transactions). Lines 1957–1960: "Since splice transactions always spend the current channel funding output, the RBF attempts automatically double-spend each other. We thus disallow RBF when `option_zeroconf` has been negotiated, because that creates a risk of losing funds."
- I3 (griefing during funding). Lines 267–279, "Liquidity griefing": "Malicious remote nodes could delay messages or stop responding, which can result in a partially created transaction that cannot be broadcast by the honest node." The recommended reply uses double-spends: "keep UTXOs unlocked and actively reuse them in concurrent sessions, which guarantees that transactions created with honest nodes double-spend pending transactions with malicious nodes".

### 3. BOLT #3 (anchor outputs)

- I3. Lines 160–162: "Each party has its own anchor output that locks to their funding key. This is to prevent a malicious peer from attaching child transactions with a low fee density to an anchor and thereby blocking the victim from getting the commit tx confirmed in time. [...] This is also the reason that every non-anchor output on the commit tx is CSV locked."
- I1 (a third party may spend a channel-adjacent output). Lines 164–165 and 174: "any anchor that remains unspent can be spent by anyone after the commitment tx confirms." "After 16 blocks, anyone can sweep the anchor".
- I3. Line 181: the `shared_anchor` output "is only added when the commitment transaction is using version 3 (`zero_fee_commitments`) and prevents pinning attacks on the commitment transaction."

### 4. BOLT #5 (on-chain handling)

- I4-adjacent vocabulary. Lines 53–60: "Any unspent output is considered to be *unresolved* and can be *resolved* [...]". "Outputs that are *resolved* are considered *irrevocably resolved* once the remote's *resolving* transaction is included in a block at least 100 deep".
- I3. Lines 552–556: honest node "SHOULD *resolve* revoked outputs in their own, separate penalty transactions. A previous penalty transaction claiming multiple revoked outputs at once may be blocked from confirming because of a transaction pinning attack."
- I3. Lines 567–574 (rationale): "the cheating node can pin spends of its HTLC-timeout/HTLC-success outputs. Using a single penalty transaction for all revoked outputs is thus unsafe".

### 5. Riard, replacement cycling disclosure (2023)

- I3. Section "Problem" of the disclosure: "A malicious channel counterparty can broadcast its HTLC-preimage transaction with a higher absolute fee and higher feerate than the honest HTLC-timeout of the victim lightning node and triggers a replacement." Then: "This replacement cycling tricks can be repeated for each rebroadcast attempt of the HTLC-timeout by the honest lightning node until expiration of the inbound HTLC timelock I."
- This is the Lightning analogue of the Armada denial of service. A competing spend of a shared output removes the honest pending spend, again and again, until a deadline passes. The attacker in Lightning is the counterparty. In Armada the attacker is any caller who can spend or merge the adjudicator's notes.
- Osuntokun's reply in the same thread states that "all of lnd's relevant mitigations were in place by lnd v0.16.1-beta". The mitigations are operational (mempool watching). They are not formal.

### 6. Lightning Labs, lnd 0.12 post (2021)

- I3, part. "Anchor output-based channels [...] allow a node to dynamically increase the fee of a pending commitment transaction using Child Pays for Parent (CPFP)." This is the engineering origin of the BOLT #3 anchor rules. It has no formal content.

### 7. eltoo (Decker, Russell, Osuntokun 2018)

- I3, part. §4, p.9: "This is achieved by allowing update transactions to bind to any of the previous update transactions." §4.1, p.9: "We call transactions that can be attached to any output of one of its predecessors a floating transaction."
- §4.1.1, p.10: "SIGHASH_NOINPUT instructs the signature creation and the signature verification code to blank the previous output field of the input that is being signed." So a signed eltoo update does not go stale when a different predecessor output is on chain. The paper has no theorems.

### 8. Aumayr thesis (2024): Generalized Channels (Ch. 3) and Virtual Channels (Ch. 6)

- I1, I2. §3.4, p.55, the properties of the ideal functionality F_L:
  "Consensus on creation: A generalized channel γ is successfully created only if all parties in γ.users agree with the creation."
  "Consensus on update: A generalized channel γ is successfully updated only if both parties in γ.users agree with the update."
- I3, part. §3.4, p.55: "Instant finality with punish: An honest party P ∈ γ.users has the guarantee that either the current state of the channel can be enforced on the ledger, or P can enforce a state where she gets all γ.cash coins."
- I4. §3.4, p.56 (Create): F_L "expects a channel funding transaction to appear on the ledger L within ∆ rounds. Such a transaction must spend both funding sources [...] and contain one output of the value γ.cash."
- I3, the assumption that Armada breaks. §3.4, p.56 (Close): "a transaction, spending the channel funding transaction and whose output corresponds to the latest channel state γ.st, should appear on L within ∆ rounds." The functionality assumes bounded delivery. It does not model a third party who can spend the funding output first.
- I3. §3.4, p.57 (Punish): "The functionality checks if a funding transaction of some channel was spent. If yes, then it expects one of the following to happen: (i) a punish transaction appears on L within ∆ rounds, assigning γ.cash coins to the honest party".
- Ch. 6, §6.3, p.141: "(V1) Balance security: If γ is a virtual channel and γ.Ingrid is honest, she never loses coins, even if γ.Alice and γ.Bob collude."
- The model is UC, on paper. The proofs are in Appendix B.8 and E. No proof assistant is used.

### 9. Sleepy Channels (CCS 2022)

- p.3: the protocol uses "a novel collateral technique" that "Enables the punishment of a misbehaving channel user within a predetermined time". It does not address I1–I4. I list it because the brief names it.

### 10. Daric (2022, abstract only)

- I3, part. Abstract: "eltoo fails to achieve bounded closure, enabling a dishonest channel party to significantly delay the channel closure process." The abstract states: "We also formalise Daric and prove its security in the Universal Composability model." I did not read the body.

### 11. Fabiański, Stefański, Litos, "A Formally Verified Lightning Network" (Why3, 2025)

- Mechanised: 6060 lines of program code and 7420 lines of verification code (§1, p.1). The main result is lemma `honestPartyWins` in `honestPartyVsAdversary.mlw` (§2, p.10).
- I3, the delivery assumption. §1.7, p.5: the adversary's "control over the network is encumbered with the responsibility to deliver transactions to the ledger within a specific time bound."
- Definition 1 (Funds Security), §2, p.9–10: "The winning state for an adversary is one that satisfies the following: 1. closeOrderTime ≠ ⊥, 2. time ≥ closeOrderTime + channelClosingTime, 3. The honest party has less coins than expectedAmount on-chain".
- Adversary move 2, §2, p.8: "SendMsgToGamma: The adversary generates an arbitrary message which is immediately processed by Γ. This models the ability of a *rushing* adversary to add transactions to the blockchain at any time, 'skiping the queue'." So the model has front-running. The model still proves funds security, because only the two parties can spend the 2-of-2 output.
- I1. §3, p.11: "The 2-out-of-2 multisig can be spent in exactly one way, namely by a commitment transaction that has been signed by both counterparties." p.12: "Since our security experiment assumes that at least one of the two channel counterparties is honest, only the aforementioned way to spend the 2-out-of-2 multisig is possible in our protocol."
- I4. §4.4, p.14: "we consider as owned those coins that a party can reliably and unilaterally move to its public key account". "Let extractable value be the value of such a game when both parties play optimally. [...] We then say that the funds owned by the party are equal to the extractable value." The proof replaces this with an Evaluator, `partyExpectationsFull` in `gamma.mlw`, "which provides a pessimistic under-approximation of the funds owned by the party" (p.15).
- Proof structure. §4.5, p.15: a relation `goodTransition` in `partyInterface.mlw`. Step 1 proves that any update that satisfies `goodTransition` preserves evaluator solvency (`totalEvaluatorMinusTotalBalanceMonotone`). Step 2 proves that the honest party's transitions satisfy `goodTransition`.

### 12. Grundmann and Hartenstein, TLA+ (2025)

- Definition 1 (informal security), p.6: "An honest user will finally get paid out at least the user's correct balance."
- I1, I3, part. p.6: the adversary publishes transactions "by finding all outputs that are spendable for the adversary and publishing a new transaction that spends the output to an output that is owned by the adversary". The model omits arbitrary messages: "We do not model that the adversary sends messages with arbitrary content".
- The check is model checking with TLC, over models with up to four hops and two concurrent payments (abstract, p.1). Two refinement steps are proved (Theorem 9, p.30 and the time-skip proof).

### 13. Weintraub et al., Spin model (CCS 2024)

- I3, part. Property 4, p.8: "HTLC Congestion. A channel cannot have max_accepted_htlcs open. [...] □(l + r < max_accepted_htlcs)". The paper finds that the model violates two of its five properties (p.8). One violation reproduces the congestion attack. The other is the new "Payout Race" (abstract, p.1).

### 14. Louistisserand, Jeanteur, Maffei, ProVerif slides (2024)

- Slide 14: "Transaction is represented by a quadruplet tx = (pk1, pk2, b, h)". The slides contain titles only for the properties. I could not extract definitions. They give no evidence on I1–I4.

### 15. LeanDY payment-channel case study (Lean 4, Dolev–Yao)

This is the nearest source in tool and method: Lean 4, symbolic cryptography, a UTXO (SegWit-style) ledger.

- I4. `Protocols/PaymentChannel/Scripts.lean` lines 473–475: `Blockchain.Pointer.points_to_common_fund` holds when `b.get_output ptr = P2WSH.mkOutput (b.get_output ptr).amount (Script.CommonFundS {pka:=pks.a, pkb:=pks.b})`. A channel's backing is one pointer to one output.
- I4, conservation. `Protocols/PaymentChannel/Security.lean` lines 214–217: invariant `PcStateInvariant₁` has the field `(tr.get_blockchain.get_output cf_ptr h).amount = s.balance[self] + s.balance[self.other]`.
- I1. `Protocols/PaymentChannel/Proofs/Soundness.lean` lines 422–437, lemma `spent`: "Any valid transaction spending a CommonFund output contains an authorized signature by the selected channel participant on the stripped transaction." The conclusion is `sign_tx_pred_base tr (...) (nonce (Nonce.Sk ps[self])) ∨ TraceEvent.Corrupt ps[self] ∈ tr`.
- I1, I3 (safety half). Same file, lines 670–684, lemma `spent_by_base_transaction`: if `common_fund.spent b`, then there is a transaction `tx ∈ b` with `tx.spends common_fund`, and its stripped form equals `(mk_base_transaction ps.vk self common_fund balance hrk).strip b 0 ...`; or else `TraceEvent.Corrupt ps.a ∈ tr ∧ TraceEvent.Corrupt ps.b ∈ tr`. In words: only a protocol-shaped commitment transaction can spend the channel's output, unless both parties are corrupt.
- The case study has no deposit after opening and no splice. It does not state I2. It does not state the liveness half of I3.

### 16. Melkonian thesis: BitML in Agda (2024)

- I1 (merge authority). §4.7.5, p.70, rules [DEP-AuthJoin] and [DEP-Join]: the source configuration is `⟨ A has v ⟩at x ∣ ⟨ A has v′ ⟩at y ∣ A auth[ x ↔ y ▷⟨ A , v + v′ ⟩ ] ∣ Γ`. The prose: "For the authorisation rule [DEP-AuthJoin], we requite the source configuration to contain two distinct deposits owned by A and augment the configuration with the authorisation of A to join them in the target. For the execution rule [DEP-Join], we require the source to contain the two deposits and the authorisation, all of which are replaced by a single merged deposit in the target." So only the owner can merge two deposits. [DEP-Divide] is the inverse (p.70).
- I2, I4. p.73, [C-Init]: "Once all persistent deposits required by the contract's preconditions have been authorised for spending, [C-Init] finishes the job by actually consuming them as authorised by [C-AuthInit], and replacing the contract advertisement ` ad with an active contract ⟨ C , v ⟩at x." The funds of an active contract are one term at one name x.
- I2 contrast. p.75, [C-PutRev]: a `put` command absorbs volatile deposits into a running contract. BitML therefore allows funding during execution. The contract code decides when.
- Status of the main theorem. §7.4, p.205: the thesis states computational soundness over "(partial) mechanisations of parsing [...] and unparsing". The Agda text on p.205 contains open holes (`{!!}`). So the compiler-correctness theorem is not fully closed.

### 17. The Extended UTXO Model (2020)

- I1. §2, p.3: an output is (ν, value, δ). A spending input supplies a redeemer ρ, and "we check that ν(value, δ, ρ, tx) = true". The validator ν decides who may spend.
- I4, the open problem. §2, p.3: a state machine needs "(b) [...] to be able to enforce that the same contract code is used along the entire sequence of transactions". §4.1, p.11: "There is no notion of initial state, since we would not be able to enforce it on the blockchain level." The paper proves a weak bisimulation between Constraint Emitting Machines and ledger transitions in Agda (§1, p.2; footnote 9 names `formal-utxo/Bisimulation.agda` at commit a1574e6).

### 18. Native Custom Tokens in the EUTXO Model (2020)

- I4, the problem. §3.1, p.6: "Multiple instances of a single state machine (using the same validator code) generate state machine outputs that look alike. Hence, for a given output locked by the state machine's validator, we cannot tell whether it belongs to one or another run of that state machine."
- I4, the fix. §3.1, p.6: "a unique non-fungible token, called a state thread token, which is minted in an initial CEM state. From the initial state on, it uniquely identifies the transaction sequence implementing the progression of one specific instance of the token's state machine."
- Proposition 2 (Global Preservation), p.10: for a valid ledger l, Σ_{t∈l} t.forge = Σ_{o∈unspentOutputs(l)} o.value.
- Proposition 4 (Initiality), p.12: "Given a valid ledger l and an unspent output carrying the thread token, we can always trace it back to a single origin, which forges the token and abides by the forging policy".
- Proposition 5 (Extraction), p.13: "Given a valid ledger l and a singular provenance of the (non-fungible) thread token, we can extract a rooted state machine trace." Corollary 1, p.13: "Any predicate that always holds for a CEM trace also holds for the one extracted from the on-chain trace."
- The paper states that these results are mechanised in Agda (§4, p.9).

### 19. Hydra Head specification (literate Agda, 2026)

This is the most relevant mechanised source. A Hydra head is a channel whose backing is UTxO. The head opens empty and grows by deposits while open.

- Trust status. Documentation page, "The trust story": the security results and the on-chain theorems "(non-stuckness, value conservation, bounded contest window) are closed Agda proofs over the abstract model". The model "rests on 65" postulates, which `spec/check-trust-ledger.sh` enumerates.
- I4. `spec/src/Hydra/Protocol/Overview.lagda.typ` lines 40–41: "A _state thread token (ST)_ minted in _init_ marks the head output and ensures contract continuity". `OnChain.lagda.typ` lines 546–547: "State token is sent to the head validator" and "All participation tokens are sent to the head output alongside the state token".
- I4, deposit attribution. `OnChain.lagda.typ` lines 309–313: "Hashing C alone would not identify a deposit: a deposit datum is unauthenticated data anyone can copy into a look-alike deposit holding less value, which hashes identically and would accept the same signature." The fix is `depositCommitsHashOf ctx ref = hash (depositDatumCommitsHash ctx ref ‖ OutputRef.txId ref)`. `ReferenceBridge.agda` lines 321–327 add the own-head check `claimedByOwnHead`: the deposit datum's head id equals the spent head's id.
- I4, conservation theorem. `Solvency.lagda.typ` line 47 states the invariant: η = accUTxO(U) and valHead = r₀ ⊕ Σ_{o∈U} val(o). Lines 395–397 state the theorem:
  `solvency : ∀ {r₀ d U w} → SolventReach r₀ d U w → (OC.ηOf d ≡ OC.accUTxO U) × (w ≡ r₀ +ᵛ sumValue U)`.
  Lines 608–613 state the payoff `fanout-covered`: a solvent head with a valid fanout distributes exactly U (`OC.distributedOuts ctx m ≡ᵉ U`).
- I2. `Solvency.lagda.typ` lines 243–246: the increment step `s-inc` takes `OC.IncrementValid ... (OC.Open ...) (OC.Open ...)`. A deposit enters the head only by a transition from Open to Open. `SecurityProofs.lagda.typ` lines 1021–1023: "a deposit can be absorbed into the head neither without every party's signature nor after it has become recoverable." `Overview.lagda.typ` lines 100–104: "A deadline is associated with the deposit [...]. Should a deposit have not been picked up in time, a recover transaction allows anyone to unlock the original UTxO."
- I1, I3, part. `OnChain.lagda.typ` lines 247–251: recover "requires exactly one" deposit input, because otherwise "a second deposit whose C hashes alike would be satisfied by the same output set, leaving its value free for the transaction author to direct anywhere." This rule stops double satisfaction. It does not address contention on the head output.
- Counter-model pattern. `SolvencyCounterModel.agda` lines 12–17 and 64–80: a machine-checked miniature of the deposit-aliasing attack. `attack : AcceptsOld fake 0 1` shows that the old digest accepts a look-alike deposit. `insolvent : 1 ≡ 0 + credited → ⊥` shows the result. `fixed : digestNew fake ≡ digestNew real → ⊥` shows that the new digest rejects it.
- The specification does not state a theorem about concurrent transactions that race for the head output. [INFERENCE] Each transition spends the single head output, so the ledger serialises all head transitions.

### 20. Zcash Protocol Specification

- I1, part. §1.2, p.9: "To each note there is cryptographically associated a note commitment. Once the transaction creating a note has been mined, the note is associated with a fixed note position in a tree of note commitments, and with a nullifier unique to that note. Computing the nullifier requires the associated private spending key (or the nullifier deriving key for Sapling-pool or Orchard-pool notes)." The specification has no contract-owned notes and no channels.

### 21. RAILGUN developer guide

- I1, part. "The RAILGUN contract verifies a a set of transaction circuits (JoinSplit with varying input and output UTXO counts) which are the only circuits allowed to write state changes." External contracts can use "the RAILGUN JoinSplit circuit with the adapt fields to add additional verification to state transitions." The page is informal.

### 22. Aztec.nr documentation

- I3, part. `check_nullifier_exists`: "a `false` value should not be relied upon since other transactions may emit this nullifier before the current transaction is included in a block." This is an informal statement of the race between a pending proof and a concurrent spend of the same note.

### 23. Fuel transaction format

- I3. `InputContract`: "when signing a transaction, `txID`, `outputIndex`, `balanceRoot`, `stateRoot`, and `txPointer` are set to zero." A signed call to a contract therefore does not name the contract's current UTXO. A concurrent call that updates the contract does not make the signed call stale.
- I4, part. `InputContract`: "Transaction is invalid if: there is not exactly one output of type `OutputType.Contract` with `inputIndex` equal to this input's index". Each contract has one live UTXO.

### 24. IOHK, "Concurrency and all that" (2021)

- I3, part. "Given that users can face contention issues trying to access the same UTXO, it is important to use many different UTXOs." "DApps should split up their on-chain state across many UTXOs." The essay names order books and batchers. It has no formal content.

## (c) Reuse

### Can reuse directly

1. **Non-interference statement shape (I1, I3 safety half), from LeanDY.** `spent_by_base_transaction` (Soundness.lean 670–684) has the shape that I3 needs: "if a channel output is spent, the spend is a protocol-shaped transaction, or the parties are corrupt". The Armada form: if a note that the channel owns is nullified after finalization, then the nullifying call is an outcome-permitted transfer. LeanDY is Lean 4 with a Dolev–Yao trace, the same setting as the Armada model. The proof method also transfers: an inductive invariant over the trace, plus a signing predicate (`sign_tx_pred_base`) that ties each signature to a trace event. [INFERENCE] The Armada analogue of the signing predicate is a predicate on the adjudicator's calls into the pool.
2. **Attribution and conservation invariant (I4), from Hydra.** `SolventReach` and `solvency` (Solvency.lagda.typ 237–255, 395–397) carry the off-chain set U as ghost state and prove valHead = r₀ ⊕ Σ U by induction over transitions. The Armada form: for each channel c, value(c) = Σ value(notes(c)), carried as ghost state over adjudicator calls. The step relation lists each transition. The documentation states that a CI script fails if a transition is neither consumed by a step nor excluded with a reason. That gate is reusable as a discipline.
3. **Counter-model of the current design, from Hydra.** `SolvencyCounterModel.agda` is a template. The Armada form: a small Lean module with `attack` (a third-party merge or spend makes the honest payout proof stale), `stuck` (the honest payout cannot complete), and `fixed` (under the phase-typed channel, the attack is not well-typed). This records the denial-of-service finding as a checked artefact.
4. **Phase typing (I2), from Hydra and BOLT #2.** Hydra types an increment as Open → Open (Solvency.lagda.typ 243–246). BOLT #2 forbids `splice_init` after `shutdown` and forbids `shutdown` while a splice is unlocked (lines 1674, 2156). Both match the proposed rule "deposits and merges only before finalization". The Lean form: index the channel type by phase, and give `deposit` and `merge` the type `Channel .funding → ...` or `Channel .open → ...`, never `Channel .finalized`.
5. **Identity binding for deposits (I4), from Hydra.** A deposit must bind its own identity, not only its content (OnChain.lagda.typ 309–313). The Armada form: a deposit into channel c must bind c's identifier and the deposit's commitment. A look-alike note with copied fields must not count toward c.
6. **Thread-token provenance (I4), from EUTXOma.** Propositions 4 and 5 (p.12–13) give a statement shape: every live output that carries channel c's token traces back to c's creation. Corollary 1 gives the method to move a state-machine invariant onto ledger traces.
7. **Ownership as extractable value (I4), from the Why3 Lightning proof.** §4.4 (p.14) defines owned funds as what a party can unilaterally extract against an obstructor. The Evaluator (p.15) is a decidable under-approximation. The split into `goodTransition` plus an implementation proof (§4.5) suits a Lean interface between the adjudicator model and the channel logic.
8. **Property vocabulary, from Generalized Channels.** "Consensus on creation", "Consensus on update" and "Instant finality with punish" (p.55) are clear names for Armada statements. The F_L Create rule "contain one output of the value γ.cash" (p.56) is the single-output case of I4.
9. **Owner-only merge (I1), from BitML.** [DEP-Join] (p.70) requires the owner's authorisation for a merge. The Armada rule: only the adjudicator, acting for channel c, may merge c's notes.

### Cannot reuse directly, and why

1. **Bounded-delivery ledgers.** Generalized Channels (p.56–57) and the Why3 proof (§1.7, p.5) assume that an honest transaction reaches the ledger within ∆. The Armada denial of service is a violation of that assumption: a concurrent call stales the proof, so the payout never executes. A model with this assumption cannot express the attack. The Armada model must make delivery conditional on the absence of a conflicting nullifier.
2. **Script and mempool mechanisms.** CSV locks, anchors, CPFP and RBF rules (BOLT #3 lines 160–181; BOLT #5 lines 552–574; Riard) defend against relay-level pinning. The Armada problem is at contract level: a conflicting call that the pool accepts. [INFERENCE] The mempool defences do not transfer, but the attack pattern does.
3. **Floating spends.** eltoo (§4.1.1, p.10) and Fuel (`InputContract`) avoid staleness because the signed message does not name the spent output. A pool spend proof must name the nullifiers of the notes it spends. [INFERENCE] So the Armada pool cannot use this mechanism without a change to the pool's circuits. The design lesson does transfer: one live output per channel (Fuel's "exactly one `OutputContract`") serialises all spends through the channel's own state machine.
4. **Native tokens.** EUTXOma and Hydra put the thread token in the ledger's multi-asset value. The Armada pool has no native token layer. [INFERENCE] The channel identity must live in a note field or in the adjudicator's own storage, with a separate proof that the binding is unforgeable.
5. **UC proofs on paper.** Generalized Channels, Virtual Channels, Sleepy Channels and Daric give UC proofs that are not mechanised. Their functionalities can guide statement names. Their proofs cannot be imported into Lean.
6. **Model checking.** The TLA+ and Spin results are bounded state-space searches. They give counterexamples, not general theorems.
7. **BitML soundness.** The compiler-correctness theorem has open holes (p.205). Only the semantics rules are fully checked.

## (d) Gaps: issues that no source covers

1. **I3, liveness half: freedom from griefing of a pending settlement by a third party.** No source states or proves that a pending honest settlement cannot be made stale by a concurrent call from an arbitrary caller. The formal sources either assume bounded delivery (Generalized Channels p.56; Why3 §1.7) or restrict spends of the channel output to the two parties (Why3 §3 p.11–12; LeanDY `spent_by_base_transaction`). The relay-level attacks (pinning, replacement cycling) are informal. Search terms used: "payment channel" with "griefing", "pinning", "replacement cycling", "non-interference", "front-running"; "formal verification Lightning Tamarin OR ProVerif OR TLA+ OR Isabelle OR Coq"; "EUTXO concurrency contention".
2. **A channel backed by a set of discrete outputs.** Every channel source backs a channel with exactly one output (Lightning funding output; Generalized Channels "one output of the value γ.cash"; LeanDY `points_to_common_fund`; Hydra head output with ST). No source formalises a channel whose value is a set of separately spendable outputs, or the merge of those outputs. Search terms used: "splicing formal", "dual-funded channel formal", "multiple funding outputs channel".
3. **Shielded pools with contract-owned notes.** No source found that formalises a note pool (commitments and nullifiers) in a proof assistant, or a pool in which a contract owns notes. The Zcash specification (§1.2) and the Railgun and Aztec documentation are prose. Search terms used: "shielded pool formal verification Lean 4 OR Coq OR Isabelle nullifier"; "Tornado Cash OR Zerocash formal verification symbolic model ProVerif Tamarin nullifier".
4. **Staleness of zero-knowledge proofs that name nullifiers.** Only the Aztec documentation mentions the race, informally. No formal source covers it. Search terms used: as in gap 3, plus "nullifier race pending transaction".
5. **I2 as a stated theorem.** Hydra enforces the phase rule through the types of its transitions and through the deposit deadline. BOLT #2 states it as MUST NOT rules. No source states the phase rule as a standalone theorem of the form "no deposit or merge after finalization". [INFERENCE] The Armada specification needs to state it explicitly.

## Sources not accessed in full

These sources sit behind a Cloudflare challenge (eprint.iacr.org PDFs, research.ed.ac.uk files) or a paywall. The `read` tool returned only the title and abstract. I do not cite their content.

- Kiayias, Litos, "A Composable Security Treatment of the Lightning Network", CSF 2020, eprint 2019/778, DOI 10.1109/CSF49147.2020.00031. Open-access copy at research.ed.ac.uk returned HTTP 403.
- Chakravarty et al., "Hydra: Fast Isomorphic State Channels", FC 2021, eprint 2020/299. The current Agda specification (source 19) replaces it for this note.
- Bartoletti, Zunino, "BitML: A Calculus for Bitcoin Smart Contracts", CCS 2018, eprint 2018/122. The Melkonian thesis (source 16) mechanises its semantics and cites its sections.
- Aumayr et al., "Generalized Channels from Limited Blockchain Scripts and Adaptor Signatures", ASIACRYPT 2021, eprint 2020/476; and "Bitcoin-Compatible Virtual Channels", IEEE S&P 2021, eprint 2020/554. I read both as chapters 3 and 6 of the Aumayr thesis (source 8).
- Mirzaei et al., Daric, eprint 2022/1295: abstract only (source 10).
- Aumayr et al., "Breaking and Fixing Virtual Channels: Domino Attack and Donner", eprint 2021/855. It is chapter 7 of the Aumayr thesis. I did not read that chapter.
- DyLean and Veil: the brief lists them as already found. I did not review them here.
