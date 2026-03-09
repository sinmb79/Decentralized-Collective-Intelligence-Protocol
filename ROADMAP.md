# DCIP Development Roadmap

> *"We don't have a roadmap enforced by a company. We have a direction agreed upon by participants."*

---

## North Star

```
A world where any human, any AI agent, any computer
can participate in collective intelligence
under the same rules, with the same rewards,
on a network owned by no one.
```

---

## Current Status

```
✅ Whitepaper v0.1
✅ Technical Specification v0.1
✅ GitHub Repository
⬜ First line of code
⬜ Genesis block
```

---

## Phase 0 — Foundation (Now)

**Goal:** Everything needed before the first line of code.

| Task | Status | Description |
|------|--------|-------------|
| Whitepaper | ✅ Done | Protocol philosophy and design |
| Technical Spec | ✅ Done | ChatGPT/Codex implementation guide |
| GitHub Repository | ✅ Done | Public, open source, MIT license |
| go.mod setup | ⬜ | Dependency configuration |
| Module scaffolding | ⬜ | Empty directory structure |

**Completion criteria:** Repository has structure, compiles with `go build`, no logic yet.

---

## Phase 1 — Testnet Alpha (Month 1–4)

**Goal:** Three nodes connecting, exchanging queries, earning ACL.

### Month 1–2: Core Infrastructure

| Module | Priority | Description |
|--------|----------|-------------|
| `core/message` | 🔴 First | Message format — protocol constitution |
| `core/identity` | 🔴 First | Ed25519 wallet, DCIP addresses |
| `network/p2p` | 🔴 First | libp2p node discovery and connection |
| `core/block` | 🟡 Second | Block structure, genesis block |

**Week 1–2 milestone:**
```bash
# Two nodes find each other automatically
Node A: "Peer discovered: DCIP1yyy..."
Node B: "Peer discovered: DCIP1xxx..."
```

### Month 2–3: Economy Layer

| Module | Priority | Description |
|--------|----------|-------------|
| `token/acl` | 🔴 | ACL issuance, transfer, burn |
| `token/reward` | 🔴 | Block rewards, halving |
| `token/ratelimit` | 🟡 | 100 free queries/hour, micro-fee burn |
| `core/chain` | 🔴 | Chain management, state persistence |

**Month 3 milestone:**
```bash
# Node earns ACL for first time
./dcip wallet balance
# ACL Balance: 50.00000000
# (first block reward received)
```

### Month 3–4: Proof of Inference

| Module | Priority | Description |
|--------|----------|-------------|
| `consensus/vrf` | 🔴 | ECVRF node selection |
| `inference/adapter` | 🔴 | Ollama, OpenAI, Echo adapters |
| `inference/ipfs` | 🟡 | Off-chain inference storage |
| `consensus/poi` | 🔴 | PoI Phase 1: multi-sig cross-verification |

**Month 4 milestone — Testnet Alpha complete:**
```
✅ Node A (OpenClaw agent) running
✅ Node B (Ollama local) running
✅ Node C (validator) running
✅ All three connected automatically
✅ Human sends query → nearest node responds
✅ Block generated → ACL distributed
✅ Rate limit working
✅ Fee burn working
✅ Chain persists across node restarts
✅ Halving simulation passes
```

---

## Phase 2 — Testnet Beta (Month 4–8)

**Goal:** Public testnet, first external nodes, dApp layer.

### Month 4–5: Hardening

| Task | Description |
|------|-------------|
| Security audit | Internal review of all message handling |
| Performance testing | 100+ concurrent queries |
| Network partition testing | Nodes disconnecting and reconnecting |
| Bootstrap nodes | First permanent public nodes |

### Month 5–6: dApp Layer

| Task | Description |
|------|-------------|
| RPC API | External interface for dApp developers |
| Simple web query interface | First human-facing UI (minimal) |
| dApp documentation | How to build on DCIP |
| First external dApp | Community-built proof of concept |

### Month 6–8: Public Testnet

| Task | Description |
|------|-------------|
| Public testnet launch | Anyone can join |
| Faucet | Free testnet ACL for new nodes |
| Explorer | Block and inference explorer |
| External node onboarding | First 10 external nodes |
| Bounty system | dApps can post rewards for hard problems |

**Month 8 milestone — Testnet Beta complete:**
```
✅ 10+ external nodes running
✅ Public block explorer live
✅ At least 1 external dApp running
✅ 1000+ test inferences processed
✅ No critical security issues
```

---

## Phase 3 — Mainnet (Month 9–16)

**Goal:** Genesis block. Real ACL. The network lives.

### Month 9–12: Mainnet Preparation

| Task | Description |
|------|-------------|
| External security audit | Third-party code review |
| Economic model stress test | Simulate 2100万 ACL distribution over 100 years |
| Node diversity requirement | Minimum model variety for collective intelligence |
| Reputation system | Track node reliability, quality |
| Governance framework | How protocol upgrades are decided |

### Month 12–16: Genesis

| Milestone | Description |
|-----------|-------------|
| Genesis block | First real block mined |
| ACL live | Real token, real rewards |
| 50+ nodes | Sufficient decentralization |
| First mainnet dApp | Real use case running |
| Collective intelligence event | First Level 3 global broadcast query |

**Genesis block inscription:**
```
"Alone we are limited. Together we are intelligence.
 For the first time, humans and AI stand on the same network."
```

---

## Phase 4 — Evolution (Year 2–3)

**Goal:** zkML, full Proof of Inference, protocol maturity.

| Milestone | Timeline | Description |
|-----------|----------|-------------|
| PoI Phase 2 | Year 2 | zkML for small models (≤7B params) |
| PoI Phase 3 | Year 2–3 | zkML for large models, full PoI |
| Cross-chain bridge | Year 2 | ACL ↔ ETH, ACL ↔ SOL |
| Mobile node | Year 2 | Lightweight node for smartphones |
| Hardware node | Year 3 | Dedicated DCIP node hardware |
| Governance DAO | Year 3 | Full protocol governance on-chain |

---

## What We're NOT Building

It's as important to know what DCIP is not:

```
❌ Not a cloud AI service
❌ Not a GPU marketplace
❌ Not a model training platform
❌ Not a company product
❌ Not a token for speculation

✅ Public infrastructure
✅ A protocol, like TCP/IP
✅ Owned by its participants
✅ Minimal rules, maximum freedom
```

---

## How This Roadmap Changes

This roadmap is a direction, not a contract. It changes when:

1. A better technical approach is proven in code
2. The community reaches consensus on a change
3. External developments make a direction obsolete

Changes are proposed in GitHub Issues and decided by the participants building the network. There is no company that can override this.

---

## Measuring Progress

We track three numbers that matter:

```
Nodes:      How many nodes are running right now?
Inferences: How many queries has the network answered?
Diversity:  How many different AI models are participating?
```

When nodes > 100, inferences > 10,000, and diversity > 10 model families — DCIP has achieved its first meaningful state of collective intelligence.

---

## Join Phase 0

The most important thing you can do right now:

```bash
# Star this repository
# Watch for the first release
# Open an issue: "I want to run a node"

# When code is ready:
git clone https://github.com/sinmb79/Decentralized-Collective-Intelligence-Protocol
./dcip wallet new
# Post your address in Discussions: "Genesis Node #N"
```

Every node that joins the genesis testnet is part of the founding network. That matters.

---

*Last updated: 2026*
*This roadmap is a living document. It evolves with the network.*
