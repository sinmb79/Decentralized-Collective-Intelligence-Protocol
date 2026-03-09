package poi

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/dcip/dcip/consensus/vrf"
	"github.com/dcip/dcip/core/identity"
	"github.com/dcip/dcip/core/message"
	"github.com/dcip/dcip/inference/adapter"
)

func TestHandleQueryCreatesLocalResponse(t *testing.T) {
	requester, err := identity.NewIdentity()
	if err != nil {
		t.Fatalf("NewIdentity(requester) error = %v", err)
	}
	responder, err := identity.NewIdentity()
	if err != nil {
		t.Fatalf("NewIdentity(responder) error = %v", err)
	}

	vrfEngine, err := vrf.FromIdentity(responder)
	if err != nil {
		t.Fatalf("FromIdentity() error = %v", err)
	}

	engine := New(nil, responder, vrfEngine, adapter.NewEchoAdapter())

	queryPayload := message.QueryPayload{
		Content:    "hello collective intelligence",
		Difficulty: 1,
	}
	data, _ := json.Marshal(queryPayload)
	msg := &message.Message{
		Version: 1,
		Type:    message.MsgQuery,
		Payload: data,
		TS:      time.Now().Unix(),
	}
	if err := msg.Sign(requester.PrivKey); err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if err := engine.HandleQuery(msg); err != nil {
		t.Fatalf("HandleQuery() error = %v", err)
	}

	queryHash := queryPayload.ContentHash
	if len(queryHash) == 0 {
		queryHash = hashContent(queryPayload.Content)
	}

	ctx := engine.Pending(queryHash)
	if ctx == nil {
		t.Fatal("expected pending query context")
	}
	if len(ctx.Responses) != 1 {
		t.Fatalf("len(ctx.Responses) = %d, want 1", len(ctx.Responses))
	}
	if ctx.Responses[0].Summary != queryPayload.Content {
		t.Fatalf("unexpected response summary: %q", ctx.Responses[0].Summary)
	}
}

func TestHandleResponseGeneratesProofAtThreshold(t *testing.T) {
	requester, err := identity.NewIdentity()
	if err != nil {
		t.Fatalf("NewIdentity(requester) error = %v", err)
	}
	local, err := identity.NewIdentity()
	if err != nil {
		t.Fatalf("NewIdentity(local) error = %v", err)
	}

	vrfEngine, err := vrf.FromIdentity(local)
	if err != nil {
		t.Fatalf("FromIdentity() error = %v", err)
	}

	engine := New(nil, local, vrfEngine, adapter.NewEchoAdapter())
	queryPayload := message.QueryPayload{
		Content:    "assemble three answers",
		Difficulty: 2,
	}
	queryBytes, _ := json.Marshal(queryPayload)
	queryMsg := &message.Message{
		Version: 1,
		Type:    message.MsgQuery,
		Payload: queryBytes,
		TS:      time.Now().Unix(),
	}
	if err := queryMsg.Sign(requester.PrivKey); err != nil {
		t.Fatalf("Sign(query) error = %v", err)
	}

	if err := engine.HandleQuery(queryMsg); err != nil {
		t.Fatalf("HandleQuery() error = %v", err)
	}

	queryHash := hashContent(queryPayload.Content)
	for _, summary := range []string{"second", "third"} {
		peerIdentity, err := identity.NewIdentity()
		if err != nil {
			t.Fatalf("NewIdentity(response) error = %v", err)
		}

		responsePayload := message.ResponsePayload{
			QueryHash: queryHash,
			Summary:   summary,
		}
		responseBytes, _ := json.Marshal(responsePayload)
		responseMsg := &message.Message{
			Version: 1,
			Type:    message.MsgResponse,
			Payload: responseBytes,
			TS:      time.Now().Unix(),
		}
		if err := responseMsg.Sign(peerIdentity.PrivKey); err != nil {
			t.Fatalf("Sign(response) error = %v", err)
		}

		if err := engine.HandleResponse(responseMsg); err != nil {
			t.Fatalf("HandleResponse() error = %v", err)
		}
	}

	proof := engine.Proof(queryHash)
	if proof == nil {
		t.Fatal("expected generated proof")
	}
	if len(proof.Signatures) != 3 {
		t.Fatalf("len(proof.Signatures) = %d, want 3", len(proof.Signatures))
	}
}

func hashContent(content string) []byte {
	payload := message.QueryPayload{Content: content}
	return queryHash(&payload)
}
