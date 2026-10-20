# Governance audits compared: threat model and coverage

<p class="lede">Two audits examined the Armada governance contracts: the Armada external audit and the Laconic audit. This page compares what each audit assumed, what each audit examined, and how each audit proved its results. It does not describe individual open findings.</p>

<p>
  <span class="tag">governance</span>
  <span class="tag">coverage comparison</span>
  <span class="tag">no finding details</span>
</p>

<div class="note">
<strong>Scope of this page.</strong> This page compares coverage and method only. It does not give attack steps, function names, line references, or timelines for findings that are open. The governance code and its specification changed after both audits, so each result needs a new check against the current code.
</div>

## 1. The two audits {#audits}

| | Armada external audit | Laconic audit |
|---|---|---|
| Contracts | Crowdfund, token, governor, treasury, steward, wind-down, redemption, revenue counter, revenue lock, adapter registry, pause controller | Governor, timelock, token delegation, treasury, revenue counter, wind-down, pause controller, monitoring specification |
| Question | Does the code do what the specification says, and is the code safe? | Can a voting bloc take the treasury, and what can honest holders do to stop it? |
| Threat model | A bloc that can pass governance proposals is a trusted party | A bloc that can pass governance proposals is the attacker |
| Proof | Executable tests for each finding. The auditors also checked the fixes. Formal verification of selected invariants. | Code reading with cited lines. The findings are candidates. No fork tests yet. |
| Severity scale | Impact × likelihood matrix | Severity and difficulty scales, applied to each candidate |

## 2. The main difference: who is the attacker {#threat-model}

The Armada external audit lists governance majority power as a centralization risk. In that model, the timelock delays, the Security Council veto, and the delay on outflow increases make sure that no single role can take value alone.

The Laconic audit starts from the opposite assumption. It asks what happens when a bloc controls enough votes to pass one proposal of the highest tier. It then asks whether each check in the list above still holds against that bloc.

The two conclusions do not contradict each other. They answer different questions. The external audit result is valid only while the governance majority stays honest. The Laconic audit examines the case where the majority does not stay honest.

## 3. Coverage by area {#coverage}

| Area | Armada external audit | Laconic audit |
|---|---|---|
| Capture of governance by a voting bloc | Recorded as a trust assumption | Primary subject |
| What the execution layer can reach | Not examined as a separate area | Examined |
| Security Council veto and ratification | Correctness of the proposal state machine | Whether the veto holds against a bloc with quorum |
| Vote delegation | Agreement with the specification | Misuse of the delegation authority |
| Proposal classification | Bypasses of the classification rules, consistency of the registries | Changes that governance can make to its own rules |
| Wind-down | Correctness of the trigger, risk that governance delays it | Availability of the holder exit during a capture |
| Ownership of governance-owned contracts | Initialization checks, storage layout | Access control after deployment |
| Monitoring | Event and function names in the specification that do not match the code | Absence of governance alerts |
| Completeness of the specification | Item-by-item comparison of the specification and the code | Missing list of the powers that governance holds |
| Steward spending channel | Examined | Examined as part of the capture model |
| Revenue accounting | Examined | Out of scope |
| Redemption arithmetic | Examined | Out of scope |
| Crowdfund | Examined | Out of scope |

## 4. Where the audits agree {#agreement}

Both audits identified the proposal classification rules and the wind-down parameters as areas of concern. The external audit recorded them as centralization risks. The Laconic audit recorded them as findings. An executable test from the external audit shows that each of these code paths exists.

## 5. Where the audits differ in method {#method}

- **Proof.** The external audit gives a test for each finding. The Laconic audit gives cited code lines and lists the fork test that each candidate needs.
- **Fix status.** The external audit records the response of the Armada team to each finding and checks the fixes. The Laconic audit has no fix cycle yet.
- **Use of the specification.** The external audit measures the code against the specification. The Laconic audit also asks whether the specification states the requirements that the code must meet.

## 6. What follows from the comparison {#next}

1. The external audit conclusion depends on an honest governance majority. A reader must not read it as proof that the treasury is safe against a captured majority.
2. The Laconic findings are candidates. Each one needs a fork test before it is reported as confirmed.
3. The governance code and its specification changed after both audits. The changes include the quorum floor, the quiet period, the veto ratification period, the Security Council pause, and the outflow floors. Both sets of results need a new check against the current code.
4. The two audits together cover more than either audit alone. The external audit covers revenue, redemption, and the crowdfund. The Laconic audit covers the governance capture case.
