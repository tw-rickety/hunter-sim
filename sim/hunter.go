package sim

import (
	"math"
	"math/rand"
)

const (
	NORMALIZED_SPEED           = 2.8
	CAST_TIME_OFFSET           = 0.5
	STEADY_SHOT_BASE_CAST_TIME = 1
	MULTISHOT_COOLDOWN         = 9
	AIMING_TIME                = 0.65
	GCD                        = 1.5
	CRIT_MULTIPLIER            = 2.3
)

type HunterBonusStats struct {
	TrinketAP       int
	BonusCrit       float64
	RapidFireHaste  float64
	QuickshotsHaste float64
}

type Bow struct {
	MinDamage   int
	MaxDamage   int
	ScopeDamage int
	Speed       float64
}

type Talents struct {
	// multiplier for RWS - 1.00 - 1.05
	RangedWeaponSpec   float64
	SwiftReflexesHaste float64
}

type Race struct {
	Haste float64
}

// Haste is multiplicative - with the exception of multiple items with +haste%. For example, if you have 5% haste total
// from items, 15% from quiver, and 1% from NE racial, the total haste is 1.05 * 1.15 * 1.01 = 1.21825.
// for that reason, we have to store different haste values separately.
type Hunter struct {
	AP          int
	Crit        float64
	Hit         int
	ItemHaste   float64
	QuiverHaste float64
	ArrowDPS    float64
	Bow         *Bow
	Talents     *Talents
	Race        *Race
	BonusStats  *HunterBonusStats
	Ping        float64
	// todo - kings buffs, consumes, etc? or in different struct?
}

func (h *Hunter) GetTotalHaste() float64 {
	return h.ItemHaste * h.QuiverHaste * h.Race.Haste * h.Talents.SwiftReflexesHaste *
		h.BonusStats.RapidFireHaste * h.BonusStats.QuickshotsHaste
}

func (h *Hunter) GetTotalAP() int {
	return h.AP + h.BonusStats.TrinketAP
}

func (h *Hunter) GetTotalCrit() float64 {
	return h.Crit + h.BonusStats.BonusCrit
}

// This has been confirmed to produce correct results by querying the lua API
// GetAutoshotDamage returns the randomized, minimum, and maximum, and randomized damage for an autoshot.
func (h *Hunter) GetAutoshotDamage(didCrit bool) (int, int, int) {
	damageBonusArrows := float64(h.Bow.Speed * h.ArrowDPS)
	damageBonusScope := float64(h.Bow.ScopeDamage)
	damageBonusAP := float64(h.GetTotalAP()) / 14.0 * h.Bow.Speed

	totalDamageBonus := damageBonusArrows + damageBonusScope + damageBonusAP

	minDamage := (float64(h.Bow.MinDamage) + totalDamageBonus) * h.Talents.RangedWeaponSpec
	maxDamage := (float64(h.Bow.MaxDamage) + totalDamageBonus) * h.Talents.RangedWeaponSpec

	randomizedDamage := minDamage + rand.Float64()*(maxDamage-minDamage)

	if didCrit {
		randomizedDamage *= CRIT_MULTIPLIER
	}

	return int(math.Round(randomizedDamage)), int(math.Floor(minDamage)), int(math.Ceil(maxDamage))
}

// TODO - verify formulas here
// GetMultishotDamage returns the randomized, minimum, and maximum, and randomized damage for a multishot.
func (h *Hunter) GetMultishotDamage(didCrit bool) (int, int, int) {
	damageBonusArrows := float64(h.Bow.Speed * h.ArrowDPS)
	damageBonusScope := float64(h.Bow.ScopeDamage)
	damageBonusAP := float64(h.GetTotalAP()) / 14.0 * NORMALIZED_SPEED
	damageBonusSkill := float64(172.5)

	barrageMultiplier := 1.15

	totalDamageBonus := damageBonusArrows + damageBonusScope + damageBonusAP + damageBonusSkill

	// note - according to cMangos, RWS does not apply to multishot
	minDamage := (float64(h.Bow.MinDamage) + totalDamageBonus) * barrageMultiplier
	maxDamage := (float64(h.Bow.MaxDamage) + totalDamageBonus) * barrageMultiplier

	randomizedDamage := minDamage + rand.Float64()*(maxDamage-minDamage)

	if didCrit {
		randomizedDamage *= CRIT_MULTIPLIER
	}

	return int(math.Round(randomizedDamage)), int(math.Floor(minDamage)), int(math.Ceil(maxDamage))
}

// TODO - verify formulas here
// GetSteadyshotDamage returns the randomized, minimum, and maximum, and randomized damage for a steadyshot.
func (h *Hunter) GetSteadyshotDamage(didCrit bool) (int, int, int) {
	damageBonusArrows := float64(h.Bow.Speed * h.ArrowDPS)
	damageBonusScope := float64(h.Bow.ScopeDamage)
	damageBonusAP := float64(h.GetTotalAP()) / 14.0 * NORMALIZED_SPEED
	damageBonusSkill := float64(50)

	improvedSteadyshotMultiplier := 1.15

	totalDamageBonus := damageBonusArrows + damageBonusScope + damageBonusAP + damageBonusSkill

	// note - according to target dummy logs, RWS does apply to steadyshot on turtleWoW.
	minDamage := (float64(h.Bow.MinDamage) + totalDamageBonus) * improvedSteadyshotMultiplier * h.Talents.RangedWeaponSpec
	maxDamage := (float64(h.Bow.MaxDamage) + totalDamageBonus) * improvedSteadyshotMultiplier * h.Talents.RangedWeaponSpec

	randomizedDamage := minDamage + rand.Float64()*(maxDamage-minDamage)

	if didCrit {
		randomizedDamage *= CRIT_MULTIPLIER
	}

	return int(math.Round(randomizedDamage)), int(math.Floor(minDamage)), int(math.Ceil(maxDamage))
}

func (h *Hunter) GetSteadyshotCastTime() float64 {
	steadyshotCastTime := (STEADY_SHOT_BASE_CAST_TIME / h.GetTotalHaste()) + CAST_TIME_OFFSET
	return steadyshotCastTime
}

func (h *Hunter) GetMultishotCastTime() float64 {
	return CAST_TIME_OFFSET
}

func (h *Hunter) GetMultishotCooldown() float64 {
	return MULTISHOT_COOLDOWN
}

func (h *Hunter) GetReloadingTime() float64 {
	return (h.Bow.Speed / h.GetTotalHaste()) - AIMING_TIME
}
