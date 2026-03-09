# DCIP L1 — 토크노믹스 통일 수정

저장소: https://github.com/sinmb79/Decentralized-Collective-Intelligence-Protocol

아래 3개 파일을 수정하고 commit & push 하라.

---

## 수정 1: token/acl/acl.go

```go
// 변경 전
TotalSupply     uint64  = 2_100_000_000_000_000
InitialReward   uint64  = 5_000_000_000
HalvingInterval uint64  = 210_000

// 변경 후
TotalSupply     uint64  = 2_100_000_000 * 100_000_000  // 2.1B ACL × 10^8
InitialReward   uint64  = 84 * 100_000_000              // 84 ACL × 10^8
HalvingInterval uint64  = 12_614_400                    // ~4년 (10초 블록타임 기준)
```

그리고 보상 나머지 처리를 명시적으로 수정하라:

```go
// DistributeReward 함수에서 변경 전
participantPool := reward * 40 / 100
perParticipant := participantPool / uint64(len(participants))
proposerShare := reward - (perParticipant * uint64(len(participants)))

// 변경 후 (60/40 명시 + 나머지는 burn)
proposerShare := reward * 60 / 100
participantPool := reward * 40 / 100
perParticipant := participantPool / uint64(len(participants))
remainder := reward - proposerShare - (perParticipant * uint64(len(participants)))
// remainder는 소각 (burned에 추가)
l.Balances[proposer] += proposerShare
for _, participant := range participants {
    l.Balances[participant] += perParticipant
}
l.Burned += remainder
l.rewardPool -= reward
```

---

## 수정 2: README.md

토크노믹스 테이블을 아래로 교체하라:

```markdown
| Property | Value |
|----------|-------|
| Total Supply | 2,100,000,000 ACL (fixed) |
| Decimals | 8 (smallest unit: 1 satoshi = 10⁻⁸ ACL) |
| Consensus | Proof of Inference (PoI) |
| Block Reward | 84 ACL (halves every 12,614,400 blocks ≈ 4 years) |
| Block Time | ~10 seconds |
| Node Selection | ECVRF (verifiable random, manipulation-proof) |
| Fee Burn | 50% of transaction fees burned |
| Pre-mine | 0 — Fair Launch |
```

---

## 수정 3: consensus/vrf/vrf.go

Prove 함수 위에 주석 추가:

```go
// Prove returns a deterministic output and proof for a given input.
// NOTE: Phase 1 approximation using Ed25519 signatures as a VRF substitute.
// Replace with a RFC 9381-compliant ECVRF implementation in Phase 3.
func (v *VRF) Prove(input []byte) ([]byte, []byte, error) {
```

---

## 수정 4: token/reward/reward_test.go 업데이트

상수 변경에 맞게 테스트의 기댓값도 수정하라.
InitialReward = 84 * 10^8 기준으로 테스트 통과 확인.

---

## 실행 순서

```bash
# 파일 수정 후
go build ./...
go test ./...   # 전체 테스트 통과 확인

git add token/acl/acl.go token/reward/reward.go README.md consensus/vrf/vrf.go
git commit -m "fix: unify tokenomics with Worldland spec (2.1B ACL / 84 ACL / 12,614,400 halving)"
git push origin main
```

---

## 완료 조건

- [ ] `go test ./...` 전체 통과
- [ ] acl.go 상수 3개 수정 완료
- [ ] 분배 나머지 → burn 처리 명시
- [ ] README 토크노믹스 테이블 업데이트
- [ ] VRF Phase 1 주석 추가
- [ ] GitHub push 완료
