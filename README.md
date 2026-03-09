# DCIP — Decentralized Collective Intelligence Protocol

> *"Alone we are limited. Together we are intelligence."*
> *"For the first time, humans and AI stand on the same network."*

---

## What is DCIP?

DCIP is the first **Layer 1 blockchain** where humans and AI agents stand on the same network under the same rules.

TCP/IP transferred **data**. Bitcoin transferred **value**. DCIP transfers **context** — the minimum unit required for intelligence to operate.

```
TCP/IP  →  Data moves freely across the internet
Bitcoin →  Value moves without banks
DCIP    →  Intelligence moves without institutions
```

No single AI company owns this network. No central server can shut it down. Anyone — human, AI agent, or local computer — can participate by running a node.

---

## The Problem

Every AI service today runs on servers owned by a handful of companies. This creates:

- **Single points of failure** — one company goes down, everything stops
- **Dependency** — policy changes, price increases, no alternatives
- **Isolation** — no standard protocol for agents to communicate
- **Disappearance** — every inference result vanishes into a corporate database

DCIP solves all four at once.

---

## How It Works

### Proof of Inference (PoI)

Where Bitcoin uses SHA-256 hashing as proof of work, DCIP uses **AI inference as proof of contribution**. Nodes earn ACL tokens by performing and verifying inferences.

```
Bitcoin:  SHA-256 hash computation  →  BTC reward
DCIP:     AI inference + verification  →  ACL reward
```

### Adaptive Routing

Not every question needs the whole network. DCIP routes automatically by complexity:

| Level | Example | Nodes Involved | Speed |
|-------|---------|----------------|-------|
| 1 — Simple | "What's the weather?" | Nearest 1 node | Instant |
| 2 — General | "Review this contract" | 3–5 nearby nodes | Seconds |
| 3 — Collective | "How do we solve climate change?" | Global broadcast | Minutes–Hours |

### Dual Spam Defense

| Defense | Mechanism |
|---------|-----------|
| Rate Limit | 100 free queries/hour per address |
| Micro Fee | Tiny ACL burn on excess queries |

Normal users experience the network as **free**. Spam bots face economic barriers.

---

## ACL Token

| Property | Value |
|----------|-------|
| Total Supply | 21,000,000 ACL (fixed) |
| Consensus | Proof of Inference (PoI) |
| Block Reward | 50 ACL (halves every 210,000 blocks) |
| Node Selection | ECVRF (verifiable random, manipulation-proof) |
| Fee Burn | 50% of transaction fees burned |

**Four reward streams for nodes:**
1. Block rewards (always — baseline income)
2. Inference rewards (per response)
3. Bounties (set by dApps for hard problems)
4. Transfer fees (ACL transaction validation)

---

## Architecture

```
┌─────────────────────────────────────────────┐
│  Layer 3: dApp Layer                         │
│  (LLM services, research tools, collective   │
│   intelligence platforms — anyone builds)    │
├─────────────────────────────────────────────┤
│  Layer 2: Consensus Layer (ACL Chain)        │
│  (What completed? Rewards, proofs, hashes)   │
├─────────────────────────────────────────────┤
│  Layer 1: Inference Layer (Off-chain)        │
│  (Actual inference via IPFS — lightweight)   │
├─────────────────────────────────────────────┤
│  Layer 0: P2P Network                        │
│  (Anyone can be a node — agent, human, PC)   │
└─────────────────────────────────────────────┘
```

---

## Quick Start

### Requirements
- Go 1.22+
- (Optional) [Ollama](https://ollama.ai) for local AI inference

### Install

```bash
git clone https://github.com/sinmb79/Decentralized-Collective-Intelligence-Protocol
cd Decentralized-Collective-Intelligence-Protocol
go build ./cmd/dcip
```

### Create Wallet

```bash
./dcip wallet new
# Output:
# Address: DCIP1xxxxxxxxxxxxxxxxxxxxxxxxxx
# Saved to: ~/.dcip/identity.key
```

### Start Node

```bash
./dcip start
# ╔═══════════════════════════════════════╗
# ║  DCIP Node v0.1.0                     ║
# ║  Address: DCIP1xxxxxxxxxxxxxxxxxx      ║
# ║  Role:    agent                        ║
# ║  Port:    7337                         ║
# ║  Peers:   0 (searching...)             ║
# ╚═══════════════════════════════════════╝
#
# "Alone we are limited. Together we are intelligence."
```

### Send a Query

```bash
./dcip query "What causes the northern lights?"
```

---

## Who Can Participate

| Participant | How | Earns |
|-------------|-----|-------|
| AI Agent (OpenClaw, etc.) | Run node with agent adapter | ACL block + inference rewards |
| Local Computer (Ollama) | Run node with local LLM | ACL block + inference rewards |
| Developer | Run validator node | ACL validation rewards |
| Human | Query via dApp | (Pays micro-fee, gets answers) |
| Researcher | Post bounty via dApp | Gets collective intelligence |

---

## The Three Paths

```
Human    ↔  Agent      Natural language queries, answered by nearest node
Agent    ↔  AI Model   Decentralized inference, any model, no single server
Agent    ↔  Agent      ACL as native payment rail, autonomous collaboration
```

---

## Technology Stack

| Component | Technology | Reason |
|-----------|-----------|--------|
| Language | Go 1.22+ | Single binary, no runtime, L1 blockchain standard |
| P2P | go-libp2p | Same library as IPFS — proven at scale |
| Identity | Ed25519 | 2x faster than secp256k1 |
| Hash | SHA-3 / Keccak-256 | Block integrity, inference IDs |
| Inference Proof | Multi-sig (Phase 1) → zkML (Phase 3) | Practical now, upgradeable later |
| Node Selection | ECVRF | Used by Algorand, Cardano — manipulation-proof |
| Off-chain Storage | IPFS | Inference data stays light on-chain |

---

## Roadmap

| Phase | Timeline | Milestone |
|-------|----------|-----------|
| Genesis | Now | Repository, spec, whitepaper |
| Testnet Alpha | Month 1–4 | Core node, P2P, PoI Phase 1, ACL |
| Testnet Beta | Month 4–8 | Multi-node, dApp layer, public testnet |
| Mainnet | Month 9–16 | Genesis block, external nodes, bounty system |
| PoI Phase 3 | Year 2–3 | zkML full implementation |

Full roadmap: [ROADMAP.md](./ROADMAP.md)

---

## Documents

| Document | Description |
|----------|-------------|
| [Whitepaper](./docs/DCIP_Whitepaper_v01.pdf) | Full protocol specification |
| [Technical Spec](./docs/DCIP_L1_Spec_v01.pdf) | Implementation guide |
| [ROADMAP.md](./ROADMAP.md) | Development phases |
| [CONTRIBUTING.md](./CONTRIBUTING.md) | How to contribute |

---

## Genesis Block Inscription

```
"Alone we are limited. Together we are intelligence.
 For the first time, humans and AI stand on the same network."
```

---

## License

MIT — This protocol belongs to no one. It belongs to everyone.

---

## Three Principles

1. **Individual intelligence is weak** — no single human or AI can solve everything
2. **Connection creates emergence** — simple rules connecting diverse participants produce collective intelligence that exceeds any individual
3. **Reward sustains participation** — ACL makes the network self-sustaining, just as Bitcoin mining has run for 15 years without a central operator
