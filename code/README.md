# Vendored code (shallow snapshots)

These are **shallow, history-free snapshots** of the **Gitea-only** (`git.vdb.to`) repositories
cited in [`../armada_watcher_party_architecture.html`](../armada_watcher_party_architecture.html),
vendored so the code stays reviewable without Gitea access (e.g. if `git.vdb.to` is down).

GitHub-hosted repos referenced in the document (`go-nitro`, `watcher-ts`, `mobymask`, `ts-nitro`,
`ipld-eth-server`, `ipld-eth-state-snapshot`, `plugeth-statediff`) are **linked, not vendored**.
`plugeth` (a full geth fork, ~200 MB) is **linked only**.

| Path | Upstream | Pinned commit |
|---|---|---|
| `chain-signatures/` | `git.vdb.to/cerc-io/chain-signatures` | `9016a7c` |
| `laconicd/` | `git.vdb.to/cerc-io/laconicd` (branch `roysc/nitro-integration`) | `d130608` |
| `laconic-wallet/` | `git.vdb.to/cerc-io/laconic-wallet` | `bb5223a` |
| `laconic-wallet-web/` | `git.vdb.to/cerc-io/laconic-wallet-web` | `2a4a478` |

Snapshots only — **no git history**. For full history or current state, use the upstream Gitea repos.
