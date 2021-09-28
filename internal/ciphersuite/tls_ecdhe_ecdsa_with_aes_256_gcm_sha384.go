package ciphersuite

import (
	"crypto/sha512"
	"fmt"
	"hash"
	"sync/atomic"

	"github.com/pion/dtls/v2/pkg/crypto/ciphersuite"
	"github.com/pion/dtls/v2/pkg/crypto/clientcertificate"
	"github.com/pion/dtls/v2/pkg/crypto/prf"
	"github.com/pion/dtls/v2/pkg/protocol/recordlayer"
)

// TLSEcdheEcdsaWithAes256GcmSha384  represents a TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 CipherSuite
type TLSEcdheEcdsaWithAes256GcmSha384 struct {
	gcm atomic.Value // *cryptoGCM
}

// CertificateType returns what type of certficate this CipherSuite exchanges
func (c *TLSEcdheEcdsaWithAes256GcmSha384) CertificateType() clientcertificate.Type {
	return clientcertificate.ECDSASign
}

// ID returns the ID of the CipherSuite
func (c *TLSEcdheEcdsaWithAes256GcmSha384) ID() ID {
	return TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
}

func (c *TLSEcdheEcdsaWithAes256GcmSha384) String() string {
	return "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
}

// HashFunc returns the hashing func for this CipherSuite
func (c *TLSEcdheEcdsaWithAes256GcmSha384) HashFunc() func() hash.Hash {
	return sha512.New384
}

// AuthenticationType controls what authentication method is using during the handshake
func (c *TLSEcdheEcdsaWithAes256GcmSha384) AuthenticationType() AuthenticationType {
	return AuthenticationTypeCertificate
}

// IsInitialized returns if the CipherSuite has keying material and can
// encrypt/decrypt packets
func (c *TLSEcdheEcdsaWithAes256GcmSha384) IsInitialized() bool {
	return c.gcm.Load() != nil
}

// Init initializes the internal Cipher with keying material
func (c *TLSEcdheEcdsaWithAes256GcmSha384) Init(masterSecret, clientRandom, serverRandom []byte, isClient bool) error {
	const (
		prfMacLen = 0
		prfKeyLen = 32
		prfIvLen  = 4
	)

	keys, err := prf.GenerateEncryptionKeys(masterSecret, clientRandom, serverRandom, prfMacLen, prfKeyLen, prfIvLen, c.HashFunc())
	if err != nil {
		return err
	}

	var gcm *ciphersuite.GCM
	if isClient {
		gcm, err = ciphersuite.NewGCM(keys.ClientWriteKey, keys.ClientWriteIV, keys.ServerWriteKey, keys.ServerWriteIV)
	} else {
		gcm, err = ciphersuite.NewGCM(keys.ServerWriteKey, keys.ServerWriteIV, keys.ClientWriteKey, keys.ClientWriteIV)
	}
	c.gcm.Store(gcm)

	return err
}

// Encrypt encrypts a single TLS RecordLayer
func (c *TLSEcdheEcdsaWithAes256GcmSha384) Encrypt(pkt *recordlayer.RecordLayer, raw []byte) ([]byte, error) {
	gcm := c.gcm.Load()
	if gcm == nil {
		return nil, fmt.Errorf("%w, unable to encrypt", errCipherSuiteNotInit)
	}

	return gcm.(*ciphersuite.GCM).Encrypt(pkt, raw)
}

// Decrypt decrypts a single TLS RecordLayer
func (c *TLSEcdheEcdsaWithAes256GcmSha384) Decrypt(raw []byte) ([]byte, error) {
	gcm := c.gcm.Load()
	if gcm == nil {
		return nil, fmt.Errorf("%w, unable to decrypt", errCipherSuiteNotInit)
	}

	return gcm.(*ciphersuite.GCM).Decrypt(raw)
}
