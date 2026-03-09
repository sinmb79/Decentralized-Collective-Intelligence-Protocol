package message

import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"fmt"
	"math"
	"sync"

	"golang.org/x/crypto/sha3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

// MsgType identifies the protocol-level purpose of a DCIP message.
type MsgType uint8

const (
	MsgHandshake MsgType = 0x01
	MsgPing      MsgType = 0x02
	MsgQuery     MsgType = 0x10
	MsgResponse  MsgType = 0x11
	MsgProof     MsgType = 0x12
	MsgBlock     MsgType = 0x20
	MsgReward    MsgType = 0x21
	MsgTx        MsgType = 0x30
)

// Message is the common envelope used across the DCIP protocol.
type Message struct {
	Version uint8
	Type    MsgType
	From    []byte
	To      []byte
	Payload []byte
	Sig     []byte
	TS      int64
}

// QueryPayload is the payload format for MsgQuery.
type QueryPayload struct {
	Content     string
	ContentHash []byte
	Difficulty  uint8
}

// ResponsePayload is the payload format for MsgResponse.
type ResponsePayload struct {
	QueryHash []byte
	IPFSCid   string
	Summary   string
}

// ProofPayload is the payload format for MsgProof.
type ProofPayload struct {
	QueryHash  []byte
	Signatures [][]byte
}

var (
	envelopeDescriptor protoreflect.MessageDescriptor
	envelopeOnce       sync.Once
	envelopeErr        error
)

// Sign signs the canonical unsigned message hash with an Ed25519 private key.
func (m *Message) Sign(privKey []byte) error {
	if len(privKey) != ed25519.PrivateKeySize {
		return fmt.Errorf("invalid Ed25519 private key length: %d", len(privKey))
	}

	pubKey := ed25519.PrivateKey(privKey).Public().(ed25519.PublicKey)
	if len(m.From) == 0 {
		m.From = cloneBytes(pubKey)
	} else {
		if len(m.From) != ed25519.PublicKeySize {
			return fmt.Errorf("invalid message from length: %d", len(m.From))
		}
		if !bytes.Equal(m.From, pubKey) {
			return errors.New("message from public key does not match private key")
		}
	}

	digest, err := m.hashDigest()
	if err != nil {
		return err
	}

	m.Sig = cloneBytes(ed25519.Sign(ed25519.PrivateKey(privKey), digest))
	return nil
}

// Verify validates the stored signature against the canonical unsigned message hash.
func (m *Message) Verify() bool {
	if len(m.From) != ed25519.PublicKeySize || len(m.Sig) != ed25519.SignatureSize {
		return false
	}

	digest, err := m.hashDigest()
	if err != nil {
		return false
	}

	return ed25519.Verify(ed25519.PublicKey(m.From), digest, m.Sig)
}

// Encode serializes the message as a deterministic protobuf envelope.
func (m *Message) Encode() ([]byte, error) {
	return m.marshalEnvelope(true)
}

// Decode loads the message from its protobuf envelope encoding.
func (m *Message) Decode(data []byte) error {
	desc, err := getEnvelopeDescriptor()
	if err != nil {
		return err
	}

	msg := dynamicpb.NewMessage(desc)
	if err := proto.Unmarshal(data, msg); err != nil {
		return err
	}

	versionField := desc.Fields().ByName(protoreflect.Name("version"))
	typeField := desc.Fields().ByName(protoreflect.Name("type"))
	fromField := desc.Fields().ByName(protoreflect.Name("from"))
	toField := desc.Fields().ByName(protoreflect.Name("to"))
	payloadField := desc.Fields().ByName(protoreflect.Name("payload"))
	sigField := desc.Fields().ByName(protoreflect.Name("sig"))
	tsField := desc.Fields().ByName(protoreflect.Name("ts"))

	version := msg.Get(versionField).Uint()
	if version > math.MaxUint8 {
		return fmt.Errorf("version out of range: %d", version)
	}

	msgType := msg.Get(typeField).Uint()
	if msgType > math.MaxUint8 {
		return fmt.Errorf("message type out of range: %d", msgType)
	}

	m.Version = uint8(version)
	m.Type = MsgType(msgType)
	m.From = cloneBytes(msg.Get(fromField).Bytes())
	m.To = cloneBytes(msg.Get(toField).Bytes())
	m.Payload = cloneBytes(msg.Get(payloadField).Bytes())
	m.Sig = cloneBytes(msg.Get(sigField).Bytes())
	m.TS = msg.Get(tsField).Int()
	return nil
}

// Hash returns the SHA3-256 hash of the canonical unsigned protobuf envelope.
func (m *Message) Hash() []byte {
	digest, err := m.hashDigest()
	if err != nil {
		return nil
	}

	return digest
}

func (m *Message) hashDigest() ([]byte, error) {
	encoded, err := m.marshalEnvelope(false)
	if err != nil {
		return nil, err
	}

	sum := sha3.Sum256(encoded)
	return sum[:], nil
}

func (m *Message) marshalEnvelope(includeSig bool) ([]byte, error) {
	desc, err := getEnvelopeDescriptor()
	if err != nil {
		return nil, err
	}

	fields := desc.Fields()
	versionField := fields.ByName(protoreflect.Name("version"))
	typeField := fields.ByName(protoreflect.Name("type"))
	fromField := fields.ByName(protoreflect.Name("from"))
	toField := fields.ByName(protoreflect.Name("to"))
	payloadField := fields.ByName(protoreflect.Name("payload"))
	sigField := fields.ByName(protoreflect.Name("sig"))
	tsField := fields.ByName(protoreflect.Name("ts"))

	msg := dynamicpb.NewMessage(desc)
	msg.Set(versionField, protoreflect.ValueOfUint32(uint32(m.Version)))
	msg.Set(typeField, protoreflect.ValueOfUint32(uint32(m.Type)))

	if len(m.From) > 0 {
		msg.Set(fromField, protoreflect.ValueOfBytes(cloneBytes(m.From)))
	}
	if len(m.To) > 0 {
		msg.Set(toField, protoreflect.ValueOfBytes(cloneBytes(m.To)))
	}
	if len(m.Payload) > 0 {
		msg.Set(payloadField, protoreflect.ValueOfBytes(cloneBytes(m.Payload)))
	}
	if includeSig && len(m.Sig) > 0 {
		msg.Set(sigField, protoreflect.ValueOfBytes(cloneBytes(m.Sig)))
	}
	if m.TS != 0 {
		msg.Set(tsField, protoreflect.ValueOfInt64(m.TS))
	}

	return proto.MarshalOptions{Deterministic: true}.Marshal(msg)
}

func getEnvelopeDescriptor() (protoreflect.MessageDescriptor, error) {
	envelopeOnce.Do(func() {
		fileDescriptor := &descriptorpb.FileDescriptorProto{
			Name:    stringPtr("dcip/core/message/message.proto"),
			Package: stringPtr("dcip.core.message"),
			Syntax:  stringPtr("proto3"),
			MessageType: []*descriptorpb.DescriptorProto{
				{
					Name: stringPtr("Envelope"),
					Field: []*descriptorpb.FieldDescriptorProto{
						{
							Name:   stringPtr("version"),
							Number: int32Ptr(1),
							Label:  labelPtr(descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL),
							Type:   typePtr(descriptorpb.FieldDescriptorProto_TYPE_UINT32),
						},
						{
							Name:   stringPtr("type"),
							Number: int32Ptr(2),
							Label:  labelPtr(descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL),
							Type:   typePtr(descriptorpb.FieldDescriptorProto_TYPE_UINT32),
						},
						{
							Name:   stringPtr("from"),
							Number: int32Ptr(3),
							Label:  labelPtr(descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL),
							Type:   typePtr(descriptorpb.FieldDescriptorProto_TYPE_BYTES),
						},
						{
							Name:   stringPtr("to"),
							Number: int32Ptr(4),
							Label:  labelPtr(descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL),
							Type:   typePtr(descriptorpb.FieldDescriptorProto_TYPE_BYTES),
						},
						{
							Name:   stringPtr("payload"),
							Number: int32Ptr(5),
							Label:  labelPtr(descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL),
							Type:   typePtr(descriptorpb.FieldDescriptorProto_TYPE_BYTES),
						},
						{
							Name:   stringPtr("sig"),
							Number: int32Ptr(6),
							Label:  labelPtr(descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL),
							Type:   typePtr(descriptorpb.FieldDescriptorProto_TYPE_BYTES),
						},
						{
							Name:   stringPtr("ts"),
							Number: int32Ptr(7),
							Label:  labelPtr(descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL),
							Type:   typePtr(descriptorpb.FieldDescriptorProto_TYPE_INT64),
						},
					},
				},
			},
		}

		file, err := protodesc.NewFile(fileDescriptor, nil)
		if err != nil {
			envelopeErr = err
			return
		}

		envelopeDescriptor = file.Messages().ByName(protoreflect.Name("Envelope"))
		if envelopeDescriptor == nil {
			envelopeErr = errors.New("message descriptor not found")
		}
	})

	return envelopeDescriptor, envelopeErr
}

func cloneBytes(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}

	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

func stringPtr(v string) *string {
	return &v
}

func int32Ptr(v int32) *int32 {
	return &v
}

func labelPtr(v descriptorpb.FieldDescriptorProto_Label) *descriptorpb.FieldDescriptorProto_Label {
	return &v
}

func typePtr(v descriptorpb.FieldDescriptorProto_Type) *descriptorpb.FieldDescriptorProto_Type {
	return &v
}
