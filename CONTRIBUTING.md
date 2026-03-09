# Contributing to DCIP

> *"The network belongs to no one. It belongs to everyone."*

Thank you for your interest in DCIP. This protocol is built by its participants — there is no company, no central team, no single owner. Just people and AI agents who believe that collective intelligence should be decentralized.

---

## Ways to Contribute

### 1. Run a Node (Most Important)

The network only exists if nodes exist. Running a node is the most direct contribution.

```bash
# Clone and build
git clone https://github.com/sinmb79/Decentralized-Collective-Intelligence-Protocol
cd Decentralized-Collective-Intelligence-Protocol
go build ./cmd/dcip

# Start your node
./dcip start
```

Every node strengthens the network. You earn ACL. The network becomes more resilient.

**Node types we need:**
| Node Type | What You Need | Priority |
|-----------|--------------|----------|
| Agent node (Ollama) | Any computer + Ollama installed | 🔴 Highest |
| Agent node (OpenClaw) | OpenClaw running | 🔴 Highest |
| Validator node | Reliable uptime, no GPU needed | 🟡 High |
| Relay node | Good bandwidth | 🟡 High |

---

### 2. Write Code

We follow a simple rule: **the spec comes first, the code follows.**

All implementation specs are in [`docs/DCIP_L1_Spec_v01.md`](./docs/DCIP_L1_Spec_v01.md).

**Current priority modules (in order):**

```
[ ] core/message     — Protocol message format
[ ] core/identity    — Ed25519 wallet, addresses
[ ] network/p2p      — libp2p node discovery
[ ] core/block       — Block structure, genesis
[ ] token/acl        — ACL token, rewards, burn
[ ] consensus/vrf    — ECVRF node selection
[ ] inference/adapter — AI model connectors
[ ] consensus/poi    — Proof of Inference engine
[ ] cmd/dcip         — Main entry point
```

#### Development Setup

```bash
# Requirements
go 1.22+
git

# Clone
git clone https://github.com/sinmb79/Decentralized-Collective-Intelligence-Protocol
cd Decentralized-Collective-Intelligence-Protocol

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Build
go build ./cmd/dcip
```

#### Pull Request Process

1. Fork the repository
2. Create a branch: `git checkout -b feat/module-name`
3. Write tests alongside your code
4. Ensure `go test ./...` passes
5. Ensure `go build ./cmd/dcip` succeeds
6. Submit PR with a clear description

#### Commit Message Format

```
feat(module): short description

Longer description if needed.

Refs: #issue-number
```

Examples:
```
feat(identity): implement Ed25519 keypair generation
fix(p2p): resolve peer discovery timeout on slow networks
docs(spec): add zkML Phase 2 implementation notes
test(poi): add multi-node consensus verification test
```

---

### 3. Build a dApp

DCIP is infrastructure. The interesting things happen in the dApp layer. Build anything on top:

- Personal AI assistant (Level 1–2 routing)
- Legal/medical knowledge platform
- Scientific research collective
- Bounty platform for hard problems
- Language translation collective

dApps interact with DCIP via the RPC API:

```bash
# Query the network from your dApp
curl -X POST http://localhost:7337/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "query",
    "content": "Your question here",
    "difficulty": 1
  }'
```

---

### 4. Write Documentation

Good documentation is how the network grows. We need:

- Translations (Korean, Japanese, Chinese, Spanish, etc.)
- Tutorial blog posts
- Video guides
- Technical deep-dives on PoI, ECVRF, zkML

Place documentation in the `/docs` folder and submit a PR.

---

### 5. Report Issues

Found a bug? Have a question? Open an issue.

**Bug report template:**
```
**What happened:**
**What you expected:**
**Steps to reproduce:**
**Node version:** (./dcip version)
**OS:**
**Logs:**
```

**Feature request template:**
```
**Problem this solves:**
**Proposed solution:**
**Alternatives considered:**
```

---

## Code Standards

### Go Style

- Follow standard Go conventions (`gofmt`, `golint`)
- All exported functions must have comments
- Error handling: always handle errors explicitly, never ignore with `_`
- No global mutable state outside of clearly marked singletons

### Security Rules

- Never log private keys
- Never commit API keys or secrets
- All P2P messages must be verified before processing
- Rate limit all external inputs

### Testing Requirements

- Every new function needs a test
- Table-driven tests preferred
- Mock AI adapters for unit tests (use `echo` adapter)
- Integration tests must work with 3 local nodes

---

## Philosophy of Contribution

This project follows the Bitcoin philosophy of contribution:

```
Minimum rules. Maximum freedom.
The network decides what works.
Simple, auditable, open.
```

We do not have a roadmap enforced by a company. We have a direction agreed upon by participants. If you believe in a different technical approach, make the argument in an issue, build the prototype, and let the code speak.

---

## Community

- **GitHub Issues** — technical discussion, bug reports
- **GitHub Discussions** — ideas, philosophy, ecosystem
- **Pull Requests** — code review, implementation

There are no private channels. Everything happens in the open.

---

## Recognition

Contributors are recorded on-chain. Every node that participates in the genesis testnet will have their node address in the genesis block record.

There is no equity, no team allocation, no insider tokens. ACL is earned by running the network — by everyone, equally, under the same rules.

---

## First Steps

Not sure where to start? Here are three concrete first steps:

**If you can run Go code:**
→ Build the project, run `./dcip wallet new`, open an issue with your address as "Genesis Testnet Node #N"

**If you want to write code:**
→ Pick the first unchecked module from the priority list above, read the spec, open a PR

**If you want to build a dApp:**
→ Open a GitHub Discussion describing what you want to build

---

*The first time a human and an AI node responded to the same query on this network — that was the moment DCIP became real. Help us get there.*
