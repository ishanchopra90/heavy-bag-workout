package models

import (
	"strings"
	"testing"
)

func TestMoveMapping_GetMappingDescriptionWithStance_Orthodox(t *testing.T) {
	mm := NewMoveMapping()
	desc := mm.GetMappingDescriptionWithStance(Orthodox)

	if !strings.Contains(desc, "orthodox") {
		t.Errorf("expected description to contain 'orthodox', got: %s", desc)
	}

	// Check for stance-specific punch names for orthodox
	if !strings.Contains(desc, "left hook") {
		t.Errorf("expected description to contain 'left hook' for orthodox, got: %s", desc)
	}
	if !strings.Contains(desc, "right hook") {
		t.Errorf("expected description to contain 'right hook' for orthodox, got: %s", desc)
	}
	if !strings.Contains(desc, "left uppercut") {
		t.Errorf("expected description to contain 'left uppercut' for orthodox, got: %s", desc)
	}
	if !strings.Contains(desc, "right uppercut") {
		t.Errorf("expected description to contain 'right uppercut' for orthodox, got: %s", desc)
	}

	// Check for technical names
	if !strings.Contains(desc, "Lead Hook") {
		t.Errorf("expected description to contain technical name 'Lead Hook', got: %s", desc)
	}
	if !strings.Contains(desc, "Rear Hook") {
		t.Errorf("expected description to contain technical name 'Rear Hook', got: %s", desc)
	}
}

func TestMoveMapping_GetMappingDescriptionWithStance_Southpaw(t *testing.T) {
	mm := NewMoveMapping()
	desc := mm.GetMappingDescriptionWithStance(Southpaw)

	if !strings.Contains(desc, "southpaw") {
		t.Errorf("expected description to contain 'southpaw', got: %s", desc)
	}

	// Check for stance-specific punch names for southpaw
	// For southpaw, Lead Hook becomes right hook, Rear Hook becomes left hook
	if !strings.Contains(desc, "right hook") {
		t.Errorf("expected description to contain 'right hook' for southpaw (lead hook), got: %s", desc)
	}
	if !strings.Contains(desc, "left hook") {
		t.Errorf("expected description to contain 'left hook' for southpaw (rear hook), got: %s", desc)
	}
	if !strings.Contains(desc, "right uppercut") {
		t.Errorf("expected description to contain 'right uppercut' for southpaw (lead uppercut), got: %s", desc)
	}
	if !strings.Contains(desc, "left uppercut") {
		t.Errorf("expected description to contain 'left uppercut' for southpaw (rear uppercut), got: %s", desc)
	}
}

func TestMoveMapping_GetMappingDescriptionWithStance_PunchMappings(t *testing.T) {
	mm := NewMoveMapping()
	desc := mm.GetMappingDescriptionWithStance(Orthodox)

	// Verify all punches are included
	punches := AllPunches()
	for _, punch := range punches {
		punchName := punch.NameForStance(Orthodox)
		if !strings.Contains(desc, punchName) {
			t.Errorf("expected description to contain punch name '%s', got: %s", punchName, desc)
		}
	}
}

func TestMoveMapping_GetMappingDescriptionWithStance_DefensiveMoves(t *testing.T) {
	mm := NewMoveMapping()
	desc := mm.GetMappingDescriptionWithStance(Orthodox)

	// Verify all defensive moves are included
	defensiveMoves := AllDefensiveMoves()
	for _, move := range defensiveMoves {
		moveName := move.String()
		if !strings.Contains(desc, moveName) {
			t.Errorf("expected description to contain defensive move '%s', got: %s", moveName, desc)
		}
	}
}

func TestMoveMapping_GetMappingDescriptionWithStance_ContainsNumbers(t *testing.T) {
	mm := NewMoveMapping()
	desc := mm.GetMappingDescriptionWithStance(Orthodox)

	// Verify punch numbers (1-6) are mentioned
	for i := 1; i <= 6; i++ {
		// Check if number appears in the description (as part of mapping)
		if !strings.Contains(desc, "1") || !strings.Contains(desc, "6") {
			t.Errorf("expected description to contain punch numbers 1-6")
			break
		}
	}

	// Verify defensive move numbers (7-12) are mentioned
	if !strings.Contains(desc, "7") || !strings.Contains(desc, "12") {
		t.Errorf("expected description to contain defensive move numbers 7-12")
	}
}
