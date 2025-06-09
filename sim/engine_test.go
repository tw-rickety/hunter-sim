package sim

import (
	"testing"
)

// Test for calculating autoshot min/max damage.
// Test case is for my hunter's gear with no consumes.
// Expected values are taken from the Character pane in-game.
func TestGetAutoshotDamage(t *testing.T) {
	talents := &Talents{
		RangedWeaponSpec:   1.05,
		SwiftReflexesHaste: 1.02,
	}
	race := &Race{
		Haste: 1.01,
	}
	bonusStats := &HunterBonusStats{
		TrinketAP:       0,
		BonusCrit:       0,
		RapidFireHaste:  HasteBuff{RemainingTime: 0, Haste: 1.4},
		QuickshotsHaste: HasteBuff{RemainingTime: 0, Haste: 1.15},
	}
	bow := &Bow{
		MinDamage:   144,
		MaxDamage:   255,
		ScopeDamage: 7,
		Speed:       3.1,
	}
	hunter := &Hunter{
		AP:          1751,
		Crit:        31.48,
		Hit:         7,
		ItemHaste:   1.0,
		QuiverHaste: 1.15,
		ArrowDPS:    20,
		Bow:         bow,
		Talents:     talents,
		Race:        race,
		BonusStats:  bonusStats,
		Ping:        150,
	}

	_, minDamage, maxDamage := hunter.GetAutoshotDamage(false)

	expectedMinDamage := 630
	expectedMaxDamage := 748

	if minDamage != expectedMinDamage {
		t.Errorf("Expected min damage to be %v, got %v", expectedMinDamage, minDamage)
	}

	if maxDamage != expectedMaxDamage {
		t.Errorf("Expected max damage to be %v, got %v", expectedMaxDamage, maxDamage)
	}
}
