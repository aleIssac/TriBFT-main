package vrf

import (
	"fmt"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/ethereum/go-ethereum/log"
)

func TestVRF(t *testing.T) {
	// Select elliptic curve, using secp256k1 curve
	s, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		log.Error("generate private key fail", "err", err)
	}
	privateKey := s.ToECDSA()

	for i := 0; i < 3; i++ {
		// Construct input data
		inputData := []byte(fmt.Sprintf("This is some input data %d", i))

		// Perform VRF computation
		vrfResult := GenerateVRF(privateKey, inputData)

		// Output VRF result
		fmt.Printf("VRF Proof: %x\n", vrfResult.Proof)
		fmt.Printf("Random Value: %x\n", vrfResult.RandomValue)

		// Verify VRF result
		isValid := VerifyVRF(&privateKey.PublicKey, inputData, vrfResult)
		fmt.Println("VRF Verification:", isValid)
	}
}
