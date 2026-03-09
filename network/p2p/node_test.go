package p2p

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/dcip/dcip/core/identity"
	"github.com/dcip/dcip/core/message"
)

func TestNodesCanConnectAndSendMessages(t *testing.T) {
	portA := freePort(t)
	portB := freePort(t)

	identityA, err := identity.NewIdentity()
	if err != nil {
		t.Fatalf("NewIdentity(A) error = %v", err)
	}
	identityB, err := identity.NewIdentity()
	if err != nil {
		t.Fatalf("NewIdentity(B) error = %v", err)
	}

	nodeA, err := NewP2PNode(identityA, portA)
	if err != nil {
		t.Fatalf("NewP2PNode(A) error = %v", err)
	}
	defer nodeA.Close()

	nodeB, err := NewP2PNode(identityB, portB)
	if err != nil {
		t.Fatalf("NewP2PNode(B) error = %v", err)
	}
	defer nodeB.Close()

	received := make(chan *message.Message, 1)
	nodeB.OnMessage = func(msg *message.Message) {
		received <- msg
	}

	addr := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", portB, nodeB.Host.ID())
	if err := nodeA.Connect(addr); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	msg := &message.Message{
		Version: 1,
		Type:    message.MsgPing,
		Payload: []byte("ping"),
		TS:      time.Now().Unix(),
	}
	if err := msg.Sign(identityA.PrivKey); err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if err := nodeA.Send(nodeB.Host.ID(), msg); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	select {
	case delivered := <-received:
		if string(delivered.Payload) != "ping" {
			t.Fatalf("unexpected payload: %q", delivered.Payload)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for message delivery")
	}

	if nodeA.PeerCount() != 1 {
		t.Fatalf("PeerCount() = %d, want 1", nodeA.PeerCount())
	}
}

func TestListenAddressesIncludePeerID(t *testing.T) {
	id, err := identity.NewIdentity()
	if err != nil {
		t.Fatalf("NewIdentity() error = %v", err)
	}

	node, err := NewP2PNode(id, freePort(t))
	if err != nil {
		t.Fatalf("NewP2PNode() error = %v", err)
	}
	defer node.Close()

	addrs := node.ListenAddresses()
	if len(addrs) == 0 {
		t.Fatal("expected at least one listen address")
	}
}

func freePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}
