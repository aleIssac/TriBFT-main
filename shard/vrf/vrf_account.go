package vrf

import (
	"crypto/ecdsa"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
)

var hasherPool = sync.Pool{
	New: func() interface{} { return sha3.NewLegacyKeccak256() },
}

type VrfAccount struct {
	privateKey  *ecdsa.PrivateKey
	pubKey      *ecdsa.PublicKey
	accountAddr common.Address
	keyDir      string
}

func NewVrfAccount(nodeDatadir string) *VrfAccount {
	_privateKey := newPrivateKey()

	vrfAccount := &VrfAccount{
		privateKey: _privateKey,
		keyDir:     filepath.Join(nodeDatadir, "KeyStoreDir"),
	}
	vrfAccount.pubKey = &vrfAccount.privateKey.PublicKey
	vrfAccount.accountAddr = crypto.PubkeyToAddress(*vrfAccount.pubKey)
	// fmt.Printf("create addr: %v\n", vrfAccount.accountAddr)
	return vrfAccount
}

func newPrivateKey() *ecdsa.PrivateKey {
	// Select elliptic curve, using secp256k1 curve
	s, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		log.Error("generate private key fail", "err", err)
	}
	privateKey := s.ToECDSA()
	return privateKey
}

func (vrfAccount *VrfAccount) SignHash(hash []byte) []byte {
	sig, err := crypto.Sign(hash, vrfAccount.privateKey)
	if err != nil {
		log.Error("signHashFail", "err", err)
		return []byte{}
	}
	return sig
}

// VerifySignature verifies signature correctness
// This method is called by accounts other than the signer to verify signature
// Cannot directly get public key and address, need to recover from signature
func VerifySignature(msgHash []byte, sig []byte, expected_addr common.Address) bool {
	// Recover public key
	pubKeyBytes, err := crypto.Ecrecover(msgHash, sig)
	if err != nil {
		log.Error("ecrecover fail", "err", err)
		// fmt.Printf("ecrecover err: %v\n", err)
	}

	pubkey, err := crypto.UnmarshalPubkey(pubKeyBytes)
	if err != nil {
		log.Error("UnmarshalPubkey fail", "err", err)
		// fmt.Printf("UnmarshalPubkey err: %v\n", err)
	}

	recovered_addr := crypto.PubkeyToAddress(*pubkey)
	return recovered_addr == expected_addr
}

// GenerateVRFOutput receives a random seed, uses private key to generate random output and corresponding proof
func (vrfAccount *VrfAccount) GenerateVRFOutput(randSeed []byte) *VRFResult {
	// vrfResult := utils.GenerateVRF(vrfAccount.privateKey, randSeed)
	sig := vrfAccount.SignHash(randSeed)
	vrfResult := &VRFResult{
		RandomValue: sig,
		Proof:       vrfAccount.accountAddr[:],
	}
	return vrfResult
}

// VerifyVRFOutput receives random output and proof, uses public key to verify if the random output is valid
func (vrfAccount *VrfAccount) VerifyVRFOutput(vrfResult *VRFResult, randSeed []byte) bool {
	// return utils.VerifyVRF(vrfAccount.pubKey, randSeed, vrfResult)
	return VerifySignature(randSeed, vrfResult.RandomValue, common.BytesToAddress(vrfResult.Proof))
}

func printAccounts(vrfAccount *VrfAccount) {
	fmt.Println("node account", "keyDir", vrfAccount.keyDir, "address", vrfAccount.accountAddr)
	fmt.Printf("privateKey: %x", crypto.FromECDSA(vrfAccount.privateKey))
}

func (vrfAccount *VrfAccount) GetAccountAddress() *common.Address {
	return &vrfAccount.accountAddr
}

// rlpHash first RLP encodes the struct, then hashes it
// Note: struct must NOT contain int type variables!
func rlpHash(x interface{}) (h common.Hash, err error) {
	sha := hasherPool.Get().(crypto.KeccakState)
	defer hasherPool.Put(sha)
	sha.Reset()
	err = rlp.Encode(sha, x)
	if err != nil {
		return common.Hash{}, err
	}
	_, err = sha.Read(h[:])
	return h, err
}

func RlpHash(x interface{}) (h common.Hash, err error) {
	return rlpHash(x)
}
