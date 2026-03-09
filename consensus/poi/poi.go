package poi

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/dcip/dcip/consensus/vrf"
	"github.com/dcip/dcip/core/identity"
	"github.com/dcip/dcip/core/message"
	"github.com/dcip/dcip/inference/adapter"
	"github.com/dcip/dcip/inference/ipfs"
	"github.com/dcip/dcip/network/p2p"
	"golang.org/x/crypto/sha3"
)

var (
	ErrNilMessage      = errors.New("message is nil")
	ErrInvalidType     = errors.New("unexpected message type")
	ErrInvalidQuery    = errors.New("invalid query payload")
	ErrContextNotFound = errors.New("query context not found")
)

// QueryContext tracks a pending query and the responses gathered for it.
type QueryContext struct {
	Query     *message.QueryPayload
	Responses []*message.ResponsePayload
	Signers   [][]byte
	CreatedAt int64
}

// PoIEngine coordinates local inference and response aggregation.
type PoIEngine struct {
	node      *p2p.P2PNode
	identity  *identity.Identity
	vrf       *vrf.VRF
	inference adapter.Adapter
	pendingQ  map[string]*QueryContext
	completed map[string]*message.ProofPayload
	mutex     sync.Mutex
}

// New creates a new Phase 1 PoI engine.
func New(node *p2p.P2PNode, id *identity.Identity, vrfEngine *vrf.VRF, inferenceBackend adapter.Adapter) *PoIEngine {
	return &PoIEngine{
		node:      node,
		identity:  id,
		vrf:       vrfEngine,
		inference: inferenceBackend,
		pendingQ:  make(map[string]*QueryContext),
		completed: make(map[string]*message.ProofPayload),
	}
}

// HandleQuery validates a query, runs inference if this node is selected, and records the response.
func (e *PoIEngine) HandleQuery(msg *message.Message) error {
	if msg == nil {
		return ErrNilMessage
	}
	if msg.Type != message.MsgQuery {
		return ErrInvalidType
	}
	if !msg.Verify() {
		return errors.New("query signature verification failed")
	}

	query, queryHash, err := decodeQuery(msg.Payload)
	if err != nil {
		return err
	}

	e.mutex.Lock()
	ctx := e.ensureContext(queryHash, query)
	e.mutex.Unlock()

	if !e.shouldRespond(queryHash) || e.inference == nil || !e.inference.IsReady() {
		return nil
	}

	responseText, err := e.inference.Infer(query.Content)
	if err != nil {
		return err
	}

	cid, err := ipfs.Store(responseText)
	if err != nil {
		cid = ""
	}

	responsePayload := &message.ResponsePayload{
		QueryHash: queryHash,
		IPFSCid:   cid,
		Summary:   responseText,
	}

	responseBytes, err := json.Marshal(responsePayload)
	if err != nil {
		return err
	}

	responseMsg := &message.Message{
		Version: msg.Version,
		Type:    message.MsgResponse,
		Payload: responseBytes,
		TS:      time.Now().Unix(),
	}
	if e.identity != nil {
		if err := responseMsg.Sign(e.identity.PrivKey); err != nil {
			return err
		}
	}

	e.mutex.Lock()
	e.appendResponseLocked(ctx, responsePayload, responseMsg.Sig)
	e.mutex.Unlock()

	if e.node != nil && responseMsg.Verify() {
		_ = e.node.Broadcast(responseMsg)
	}

	return nil
}

// HandleResponse records a response and generates a proof when the threshold is met.
func (e *PoIEngine) HandleResponse(msg *message.Message) error {
	if msg == nil {
		return ErrNilMessage
	}
	if msg.Type != message.MsgResponse {
		return ErrInvalidType
	}
	if !msg.Verify() {
		return errors.New("response signature verification failed")
	}

	var response message.ResponsePayload
	if err := json.Unmarshal(msg.Payload, &response); err != nil {
		return err
	}

	key := queryKey(response.QueryHash)

	e.mutex.Lock()
	defer e.mutex.Unlock()

	ctx, ok := e.pendingQ[key]
	if !ok {
		return ErrContextNotFound
	}

	e.appendResponseLocked(ctx, &response, msg.Sig)
	if len(ctx.Responses) < requiredResponses(ctx.Query.Difficulty) {
		return nil
	}

	proof, err := e.generateProofLocked(ctx)
	if err != nil {
		return err
	}
	e.completed[key] = proof

	if e.node != nil && e.identity != nil {
		payload, marshalErr := json.Marshal(proof)
		if marshalErr == nil {
			proofMsg := &message.Message{
				Version: 1,
				Type:    message.MsgProof,
				Payload: payload,
				TS:      time.Now().Unix(),
			}
			if signErr := proofMsg.Sign(e.identity.PrivKey); signErr == nil {
				_ = e.node.Broadcast(proofMsg)
			}
		}
	}

	return nil
}

// GenerateProof produces a proof payload from a pending query context.
func (e *PoIEngine) GenerateProof(ctx *QueryContext) (*message.ProofPayload, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	return e.generateProofLocked(ctx)
}

// Pending returns a copy of the pending query context for a query hash.
func (e *PoIEngine) Pending(queryHash []byte) *QueryContext {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	ctx := e.pendingQ[queryKey(queryHash)]
	return cloneContext(ctx)
}

// Proof returns a generated proof for a query hash, if any.
func (e *PoIEngine) Proof(queryHash []byte) *message.ProofPayload {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	proof := e.completed[queryKey(queryHash)]
	return cloneProof(proof)
}

func (e *PoIEngine) ensureContext(queryHash []byte, query *message.QueryPayload) *QueryContext {
	key := queryKey(queryHash)
	ctx, ok := e.pendingQ[key]
	if !ok {
		ctx = &QueryContext{
			Query:     cloneQuery(query),
			CreatedAt: time.Now().Unix(),
		}
		e.pendingQ[key] = ctx
	}
	return ctx
}

func (e *PoIEngine) shouldRespond(queryHash []byte) bool {
	if e.identity == nil || e.vrf == nil {
		return true
	}

	candidates := []string{e.identity.Address}
	if e.node != nil {
		candidates = append(candidates, e.node.PeerList()...)
	}
	if len(candidates) == 1 {
		return true
	}

	output, _, err := e.vrf.Prove(queryHash)
	if err != nil {
		return true
	}

	return vrf.SelectNode(output, candidates) == e.identity.Address
}

func (e *PoIEngine) appendResponseLocked(ctx *QueryContext, response *message.ResponsePayload, signature []byte) {
	for _, existing := range ctx.Signers {
		if bytes.Equal(existing, signature) {
			return
		}
	}

	ctx.Responses = append(ctx.Responses, cloneResponse(response))
	ctx.Signers = append(ctx.Signers, cloneBytes(signature))
}

func (e *PoIEngine) generateProofLocked(ctx *QueryContext) (*message.ProofPayload, error) {
	if ctx == nil || ctx.Query == nil {
		return nil, ErrInvalidQuery
	}
	if len(ctx.Signers) == 0 {
		return nil, errors.New("no response signatures available")
	}

	proof := &message.ProofPayload{
		QueryHash:  queryHash(ctx.Query),
		Signatures: cloneSignatures(ctx.Signers),
	}
	return proof, nil
}

func decodeQuery(data []byte) (*message.QueryPayload, []byte, error) {
	var query message.QueryPayload
	if err := json.Unmarshal(data, &query); err != nil {
		return nil, nil, err
	}
	if query.Content == "" && len(query.ContentHash) == 0 {
		return nil, nil, ErrInvalidQuery
	}

	return &query, queryHash(&query), nil
}

func queryHash(query *message.QueryPayload) []byte {
	if query == nil {
		return nil
	}
	if len(query.ContentHash) > 0 {
		return cloneBytes(query.ContentHash)
	}

	sum := sha3.Sum256([]byte(query.Content))
	return sum[:]
}

func queryKey(hash []byte) string {
	return hex.EncodeToString(hash)
}

func requiredResponses(difficulty uint8) int {
	switch difficulty {
	case 0, 1:
		return 1
	case 2:
		return 3
	default:
		return 5
	}
}

func cloneQuery(src *message.QueryPayload) *message.QueryPayload {
	if src == nil {
		return nil
	}

	return &message.QueryPayload{
		Content:     src.Content,
		ContentHash: cloneBytes(src.ContentHash),
		Difficulty:  src.Difficulty,
	}
}

func cloneResponse(src *message.ResponsePayload) *message.ResponsePayload {
	if src == nil {
		return nil
	}

	return &message.ResponsePayload{
		QueryHash: cloneBytes(src.QueryHash),
		IPFSCid:   src.IPFSCid,
		Summary:   src.Summary,
	}
}

func cloneContext(src *QueryContext) *QueryContext {
	if src == nil {
		return nil
	}

	return &QueryContext{
		Query:     cloneQuery(src.Query),
		Responses: cloneResponses(src.Responses),
		Signers:   cloneSignatures(src.Signers),
		CreatedAt: src.CreatedAt,
	}
}

func cloneProof(src *message.ProofPayload) *message.ProofPayload {
	if src == nil {
		return nil
	}

	return &message.ProofPayload{
		QueryHash:  cloneBytes(src.QueryHash),
		Signatures: cloneSignatures(src.Signatures),
	}
}

func cloneResponses(src []*message.ResponsePayload) []*message.ResponsePayload {
	if len(src) == 0 {
		return nil
	}

	clone := make([]*message.ResponsePayload, 0, len(src))
	for _, response := range src {
		clone = append(clone, cloneResponse(response))
	}
	return clone
}

func cloneSignatures(src [][]byte) [][]byte {
	if len(src) == 0 {
		return nil
	}

	clone := make([][]byte, 0, len(src))
	for _, sig := range src {
		clone = append(clone, cloneBytes(sig))
	}
	return clone
}

func cloneBytes(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}

	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}
