package p2p

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/dcip/dcip/core/identity"
	"github.com/dcip/dcip/core/message"
	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	DefaultPort = 7337
	ProtocolID  = protocol.ID("/dcip/1.0.0")
)

// BootstrapPeers holds optional static peers configured by the operator.
var BootstrapPeers = []string{}

// P2PNode wraps a libp2p host and DCIP message transport.
type P2PNode struct {
	Host      host.Host
	DHT       *dht.IpfsDHT
	Identity  *identity.Identity
	Port      int
	Peers     map[string]peer.ID
	OnMessage func(msg *message.Message)

	mutex  sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	mdns   mdns.Service
}

// NewP2PNode creates a libp2p host for the given identity and port.
func NewP2PNode(id *identity.Identity, port int) (*P2PNode, error) {
	if id == nil {
		return nil, errors.New("identity is nil")
	}
	if len(id.PrivKey) == 0 {
		return nil, errors.New("identity private key is empty")
	}
	if port <= 0 {
		port = DefaultPort
	}

	privKey, err := crypto.UnmarshalEd25519PrivateKey(id.PrivKey)
	if err != nil {
		return nil, err
	}

	host, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)),
	)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	node := &P2PNode{
		Host:     host,
		Identity: id,
		Port:     port,
		Peers:    make(map[string]peer.ID),
		ctx:      ctx,
		cancel:   cancel,
	}
	node.Host.SetStreamHandler(ProtocolID, node.handleStream)
	return node, nil
}

// Start enables discovery and connects to configured bootstrap peers.
func (n *P2PNode) Start() error {
	var errs []error
	discoveryReady := false

	if n.DHT == nil {
		dhtInstance, err := dht.New(n.ctx, n.Host)
		if err != nil {
			errs = append(errs, fmt.Errorf("dht init: %w", err))
		} else {
			n.DHT = dhtInstance
			if err := n.DHT.Bootstrap(n.ctx); err != nil {
				errs = append(errs, fmt.Errorf("dht bootstrap: %w", err))
			} else {
				discoveryReady = true
			}
		}
	}

	if n.mdns == nil {
		service := mdns.NewMdnsService(n.Host, string(ProtocolID), n)
		if err := service.Start(); err != nil {
			errs = append(errs, fmt.Errorf("mdns start: %w", err))
		} else {
			n.mdns = service
			discoveryReady = true
		}
	}

	for _, addr := range BootstrapPeers {
		if err := n.Connect(addr); err != nil {
			errs = append(errs, fmt.Errorf("bootstrap connect %s: %w", addr, err))
		}
	}

	if discoveryReady {
		return nil
	}
	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

// Close releases network resources.
func (n *P2PNode) Close() error {
	if n == nil {
		return nil
	}

	n.cancel()

	var errs []error
	if n.mdns != nil {
		if err := n.mdns.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if n.DHT != nil {
		if err := n.DHT.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if n.Host != nil {
		if err := n.Host.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// HandlePeerFound satisfies the mDNS notifee interface.
func (n *P2PNode) HandlePeerFound(info peer.AddrInfo) {
	if info.ID == n.Host.ID() {
		return
	}

	n.Host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	if err := n.Host.Connect(n.ctx, info); err == nil {
		n.rememberPeer(info.ID)
	}
}

// Connect dials a specific peer multiaddress.
func (n *P2PNode) Connect(addr string) error {
	peerAddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return err
	}

	info, err := peer.AddrInfoFromP2pAddr(peerAddr)
	if err != nil {
		return err
	}

	n.Host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	if err := n.Host.Connect(n.ctx, *info); err != nil {
		return err
	}

	n.rememberPeer(info.ID)
	return nil
}

// Broadcast sends a message to all connected peers.
func (n *P2PNode) Broadcast(msg *message.Message) error {
	peers := n.peerIDs()

	var firstErr error
	for _, peerID := range peers {
		if err := n.Send(peerID, msg); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// Send sends a message to a specific peer.
func (n *P2PNode) Send(peerID peer.ID, msg *message.Message) error {
	if msg == nil {
		return errors.New("message is nil")
	}

	data, err := msg.Encode()
	if err != nil {
		return err
	}

	stream, err := n.Host.NewStream(n.ctx, peerID, ProtocolID)
	if err != nil {
		return err
	}
	defer stream.Close()

	if err := binary.Write(stream, binary.BigEndian, uint32(len(data))); err != nil {
		return err
	}
	if _, err := stream.Write(data); err != nil {
		return err
	}

	n.rememberPeer(peerID)
	return nil
}

// PeerCount returns the number of known peers.
func (n *P2PNode) PeerCount() int {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return len(n.Peers)
}

// PeerList returns known peer IDs as strings.
func (n *P2PNode) PeerList() []string {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	peers := make([]string, 0, len(n.Peers))
	for key := range n.Peers {
		peers = append(peers, key)
	}

	return peers
}

// ListenAddresses returns the host listen addresses with embedded peer ID.
func (n *P2PNode) ListenAddresses() []string {
	info := peer.AddrInfo{
		ID:    n.Host.ID(),
		Addrs: n.Host.Addrs(),
	}

	addrs, err := peer.AddrInfoToP2pAddrs(&info)
	if err != nil {
		return nil
	}

	values := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		values = append(values, addr.String())
	}

	return values
}

func (n *P2PNode) handleStream(stream network.Stream) {
	defer stream.Close()
	n.rememberPeer(stream.Conn().RemotePeer())

	var size uint32
	if err := binary.Read(stream, binary.BigEndian, &size); err != nil {
		return
	}
	if size == 0 {
		return
	}

	data := make([]byte, size)
	if _, err := io.ReadFull(stream, data); err != nil {
		return
	}

	var msg message.Message
	if err := msg.Decode(data); err != nil {
		return
	}
	if !msg.Verify() {
		return
	}

	if n.OnMessage != nil {
		n.OnMessage(&msg)
	}
}

func (n *P2PNode) rememberPeer(peerID peer.ID) {
	if peerID == "" || peerID == n.Host.ID() {
		return
	}

	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.Peers[peerID.String()] = peerID
}

func (n *P2PNode) peerIDs() []peer.ID {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	peers := make([]peer.ID, 0, len(n.Peers))
	for _, peerID := range n.Peers {
		peers = append(peers, peerID)
	}

	return peers
}
