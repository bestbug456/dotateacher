package main

import "testing"

func TestBitmaskCreation(t *testing.T) {
	heros := []int{89, 106, 101, 12, 21}
	ris := createBitmasksForTeam(heros)
	for i := 0; i < len(heros); i++ {
		if ris[heros[i]] != 1 {
			t.Fatalf("Error: exspected 1 for the position %d but having: %f", heros[i], ris[heros[i]])
		}
	}
}

func TestOrderPickAndCreateBitmask(t *testing.T) {
	pick := []int{89, 51, 52, 106, 64, 101, 38, 12, 21, 61}
	expected := []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	finalBitmask := orderPickByTeamAndCreateBitmask(pick)
	if len(finalBitmask) != len(expected) {
		t.Fatalf("Error: len mismatch expected %d but having %d", len(expected), len(finalBitmask))
	}
	for i := 0; i < len(finalBitmask); i++ {
		if finalBitmask[i] != expected[i] {
			t.Fatalf("Error: exspected %f for the position %d but having: %f", expected[i], i, finalBitmask[i])
		}
	}
}
