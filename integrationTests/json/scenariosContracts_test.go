package vmjsonintegrationtest

import (
	"testing"
)

func TestRustAdder(t *testing.T) {
	runAllTestsInFolder(t, "adder/scenarios")
}

func TestRustErc20(t *testing.T) {
	runAllTestsInFolder(t, "erc20-rust/scenarios")
}

func TestCErc20(t *testing.T) {
	runAllTestsInFolder(t, "erc20-c")
}

func TestMultisig(t *testing.T) {
	runAllTestsInFolder(t, "multisig/scenarios")
}

// func TestDnsContract(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("not a short test")
// 	}

// 	runAllTestsInFolder(t, "dns")
// }

// func TestCrowdfundingDcdt(t *testing.T) {
// 	runAllTestsInFolder(t, "crowdfunding-dcdt")
// }

func TestRewaDcdtSwap(t *testing.T) {
	runAllTestsInFolder(t, "rewa-dcdt-swap")
}

func TestPingPongRewa(t *testing.T) {
	runAllTestsInFolder(t, "ping-pong-rewa")
}

func TestRustAttestation(t *testing.T) {
	if testing.Short() {
		t.Skip("not a short test")
	}

	runAllTestsInFolder(t, "attestation-rust")
}

func TestCAttestation(t *testing.T) {
	if testing.Short() {
		t.Skip("not a short test")
	}
	runAllTestsInFolder(t, "attestation-c")
}
