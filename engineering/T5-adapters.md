# T5 · Adapters

**Status:** draft · tier doc · 2026-09-05
**Parent:** [`05-building-block-view.md`](./05-building-block-view.md) · **Tier:** T5
**Release:** v1 (T5.0, T5.1)
**Owner:** work package **C · yield & exchange** ([`00-work-packages.md`](./00-work-packages.md))
**Depends on / Blocks:** consumes T4.5, T4.3, T4.2, T4.0; settles over T0.3; sits ON T0.0 · blocks nothing downstream

T5 is the **adapter layer of Armada**, and per [ADR-0013](./09-architecture-decisions.md#adr-0013) it is **in scope** — one engineering team builds the whole Armada product, so the old "Armada-built vs Laconic-built" ownership line no longer applies here. This tier is **owned by work package C** (yield & exchange). The adapters are **applications** that ride the T4 venue and settle over the T0.3 rail; they are not part of the shielded pool and take no custody. The two sections below name the pieces: T5.0 is the set of adapters (what is reused vs net-new), and T5.1 is the interface note describing how the T4 venue routes to and from them. Neither introduces a new settlement contract — settlement is T0.3, work package A.

## T5.0 Adapters

**What it is.** The governance-added adapters that ride Armada's shielded pool: **CCTP** (cross-chain USDC in), **Aave-v4 yield**, and **Swaps**. Each is an **application** on the execution/settlement layer, not part of the pool and not a new settlement contract. Per [ADR-0013](./09-architecture-decisions.md#adr-0013) all three are in scope for work package C; the boundary here is **reuse vs net-new**, not org ownership.

**Boundary — what this tier is and is not.**
- **Applications on the layer.** As the index diagram states, the venue is a private execution/settlement layer **beneath** the adapters, and Swaps and yield are applications on it. `ex_net`/the matcher is *not* a sibling adapter; that is T4.6, which sits under this tier.
- **Pool stays immutable.** Every adapter operates at the adapter/execution tier, so nothing here alters Armada's core Railgun contracts (T0.0) or takes custody. The shielded pool remains immutable — a single anonymity set, untouched.
- **Non-custodial.** Value only ever moves under the user's own key. The value-moving mechanism an adapter uses to touch an external protocol is **Railgun's own RelayAdapt / Cookbook recipe** — an atomic `unshield → external call → reshield` in one transaction (see `yield-clearing.html` §3) — never a held vault. Per [ADR-0010](./09-architecture-decisions.md#adr-0010), that recipe is **Railgun's own**, distinct from our T0.3 deposit/payout contract; the T0.3 rail is not an "adapter."

**The three adapters — reuse vs net-new.**
- **CCTP.** Cross-chain USDC (Circle's CCTP), **already built** — this is **reuse**. A CCTP mint is wrapped into the shield atomically, so cross-chain USDC arrives as a shielded note.
- **Aave-v4 yield.** This adapter turns a shielded USDC note into a shielded aUSDC ("aave") note. It is the **mechanism for USDC yield**, which — since Armada is privacy-for-USDC — is the **priority capability** (work package C open decision; see [ADR-0013](./09-architecture-decisions.md#adr-0013) and the C open-decisions list in [`00-work-packages.md`](./00-work-packages.md)). It is **the team's to build**, or to **reuse Railgun's Cookbook / RelayAdapt Aave recipe** if the **license permits** — an **OPEN gate**, C-scope.
- **Swaps.** A token-for-token exchange exposed as an application on the venue. Like the yield recipe, it is **build-or-reuse** against Railgun's Cookbook (same license gate); the swap itself clears over Nitro like any other T4 trade. Routing is T5.1.

**USDC-yield mechanism — open C-scope decision.** The Aave-v4 adapter delivers USDC yield, but *how* is not yet fixed: a **v1** direct-adapter recipe with a **public Design-A boundary** (amounts exposed only at the transparent Aave supply) versus the **LP-buffered private rail** (T4.5). That v1/v2 split is an **open work-package-C decision** and **may amend [ADR-0011](./09-architecture-decisions.md#adr-0011)**, which currently cuts LP-buffered USDC yield (T4.5) and the venue solver (T4.2) to v2.

**Interface exposed / consumed.** T5.0 exposes each adapter as a named application endpoint that the T4 venue routes into (T5.1). It consumes only the settlement/routing contract that T4 (work package A/C) already provides over Nitro. Amount privacy for shielded-to-shielded activity is solved inside the pool; amounts are exposed only at the transparent boundary (CCTP mint, Aave supply) per Design A (`yield-clearing.html` §4).

## T5.1 Routing interface

**What it is.** The **interface note** describing how the T4 venue routes to and from the T5.0 adapters. It is *interface*, v1 — **not a new settlement contract**. The interface already exists in T4's contracts and clearing; this section only names how those pieces reach the adapters and settle over Nitro.

**How the venue routes.**
- **USDC yield → Aave-v4 adapter.** The venue's USDC-yield rail, **T4.5** (LP-buffered, batched Aave, v2), is the routing point into the Aave-v4 yield adapter. When the venue rebalances buffered USDC into Aave, it does so through the adapter's Railgun recipe (`unshield USDC → Aave supply → reshield aUSDC`), atomically and in the taker's custody. Saturation is handled inside T4.5 (dribble / become-LP), so a rebalance never forces a public per-user Aave deposit (`yield-clearing.html` §5). (Which mechanism ships in v1 is the open C-scope decision noted in T5.0.)
- **Swaps → applications on T4.** A swap routes as an application on the venue: priced by the posted-price contract (**T4.0**) and, in v1, filled from the venue's static inventory; automated hedging/internalization by the venue solver (**T4.2**) is v2. Swaps clear and settle over Nitro like any other T4 trade, and fee-split allocation (**T4.3**) applies unchanged.
- **CCTP → shield boundary.** A CCTP mint is wrapped into the shield by the adapter, so from the venue's side it appears as a fresh shielded USDC note entering a channel — no special venue routing beyond the normal deposit/settlement path (**T0.3**).

**Interface consumed / exposed.** T5.1 **consumes** T4.5 (Aave rail), T4.2 (solver), and T4.0 (posted price), and **exposes** those as the routing targets for T5.0's Swaps and Aave-v4 adapters. Integrator revenue-share is delivered as a clearing allocation on channel state (T4.3), not as a new mechanism here.

**Shared dependency — its own audit boundary.** Putting one execution layer *under* several adapters makes it a **shared dependency**: a larger surface than any single adapter, and one that needs **its own audit**. This is a live engineering point, not an org boundary — it changes nothing about custody or the pool (the layer operates at the adapter/execution tier and never alters the core contracts, T0.0), but the shared-surface audit is real, and it belongs to whoever owns the shared T4 layer (work package C). The mitigation is to keep the venue's contract footprint small (posted-price + deposit/payout) so the shared surface stays auditable.

**Testing / risks (inline).** The interface carries no new contract, so its correctness is exercised by T4's own tests (posted-price, solver, USDC rail) plus the adapter-side tests. The standing risk is the shared-dependency audit surface above.

## Sources

- `index.html` — Adapters tier (Swaps · Aave-v4 yield · CCTP; governance-added), the "How it plugs into Armada" diagram (adapters as applications *on* the execution/settlement layer; pool immutable; non-custodial), the tier map (T5; CCTP built), and the shared-substrate ⇒ own-audit note.
- `yield-clearing.html` — §3 the value-moving recipe is Railgun's RelayAdapt / Cookbook (atomic `unshield → external call → reshield`, turning shielded USDC into shielded aUSDC, wrapping a CCTP mint into the shield); §4 amount privacy at the transparent boundary (Design A); §5 USDC yield via LP-buffered batched Aave and saturation handling.
- Railgun RelayAdapt / Cookbook: https://github.com/Railgun-Community/cookbook
- [ADR-0013](./09-architecture-decisions.md#adr-0013) — single team; T5 in scope, owned by work package C. [ADR-0011](./09-architecture-decisions.md#adr-0011) — market-making (T4.2) and LP-buffered USDC yield (T4.5) cut to v2, may be amended by the C-scope USDC-yield decision. [ADR-0010](./09-architecture-decisions.md#adr-0010) — RelayAdapt/Cookbook is Railgun's recipe, distinct from the T0.3 deposit/payout contract.
