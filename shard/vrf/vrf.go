package vrf

import (
	"crypto/ecdsa"
	"reflect"

	"github.com/ethereum/go-ethereum/log"
	"github.com/vechain/go-ecvrf"
)

// VRFResult contains VRF method output results
type VRFResult struct {
	Proof       []byte // VRF proof
	RandomValue []byte // Random value
}

// GenerateVRF performs VRF computation using private key
//
// VRF algorithm: uses third-party library ecvrf
func GenerateVRF(privateKey *ecdsa.PrivateKey, input []byte) *VRFResult {
	output, proof, err := ecvrf.Secp256k1Sha256Tai.Prove(privateKey, input)
	if err != nil {
		log.Error("GenerateVRF fail", "err", err)
	}
	return &VRFResult{
		Proof:       proof,
		RandomValue: output,
	}
}

// VerifyVRF verifies VRF result using public key
//
// VRF algorithm: uses third-party library ecvrf
func VerifyVRF(publicKey *ecdsa.PublicKey, input []byte, vrfResult *VRFResult) bool {
	output, err := ecvrf.Secp256k1Sha256Tai.Verify(publicKey, input, vrfResult.Proof)
	if err != nil {
		log.Error("VerifyVRF fail", "err", err)
	}

	return reflect.DeepEqual(output, vrfResult.RandomValue)
}
