package sim

import (
	"fmt"
	"math"
	"math/rand"
)

const (
	SECONDS_PER_MINUTE = 60

	NORMALIZED_SPEED           = 2.8
	CAST_TIME_OFFSET           = 0.5
	STEADY_SHOT_BASE_CAST_TIME = 1
	AIMING_TIME                = 0.65
	GCD                        = 1.5
	CRIT_MULTIPLIER            = 2.3
	QUICKSHOTS_DURATION        = 12

	PIERCING_SHOTS_DURATION    = 8
	PIERCING_SHOTS_MULTIPLIER  = 0.3
	PIERCING_SHOTS_TICKS_EVERY = 2

	// TODO - adjust these based on config
	RAPID_FIRE_DURATION = 19
	RAPID_FIRE_COOLDOWN = 5 * SECONDS_PER_MINUTE
)

type HunterBonusStats struct {
	TrinketAP       int
	BonusCrit       float64
	RapidFireHaste  HasteBuff
	QuickshotsHaste HasteBuff
}

type HasteBuff struct {
	RemainingTime float64
	Haste         float64
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

type DoT struct {
	DamagePerTick float64
	DurationLeft  float64
	TicksEvery    float64
}

// Haste is multiplicative - with the exception of multiple items with +haste%. For example, if you have 5% haste total
// from items, 15% from quiver, and 1% from NE racial, the total haste is 1.05 * 1.15 * 1.01 = 1.21825.
// for that reason, we have to store different haste values separately.
type Hunter struct {
	AP                int
	Crit              float64
	Hit               int
	ItemHaste         float64
	QuiverHaste       float64
	ArrowDPS          float64
	Bow               *Bow
	Talents           *Talents
	Race              *Race
	BonusStats        *HunterBonusStats
	Ping              float64
	Rotation          string
	MultishotCooldown float64

	StandardizedDamage bool

	PiercingShotsDoT DoT

	// todo - cooldowns array or struct

	// todo - kings buffs, consumes, etc? or in different struct?
}

func (h *Hunter) GetTotalHaste() float64 {
	totalHaste := h.ItemHaste * h.QuiverHaste * h.Race.Haste * h.Talents.SwiftReflexesHaste

	if h.BonusStats.RapidFireHaste.RemainingTime > 0 {
		totalHaste *= h.BonusStats.RapidFireHaste.Haste
	}

	if h.BonusStats.QuickshotsHaste.RemainingTime > 0 {
		totalHaste *= h.BonusStats.QuickshotsHaste.Haste
	}

	return totalHaste
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

	var randomizedDamage float64
	if h.StandardizedDamage {
		randomizedDamage = (maxDamage + minDamage) / 2
	} else {
		randomizedDamage = minDamage + rand.Float64()*(maxDamage-minDamage)
	}

	if didCrit {
		randomizedDamage *= CRIT_MULTIPLIER
	}

	return int(math.Round(randomizedDamage)), int(math.Floor(minDamage)), int(math.Ceil(maxDamage))
}

// GetMultishotDamage returns the randomized, minimum, and maximum, and randomized damage for a multishot.
func (h *Hunter) GetMultishotDamage(didCrit bool) (int, int, int) {
	damageBonusArrows := float64(h.Bow.Speed * h.ArrowDPS)
	damageBonusScope := float64(h.Bow.ScopeDamage)
	damageBonusAP := float64(h.GetTotalAP()) / 14.0 * NORMALIZED_SPEED
	damageBonusSkill := float64(172.5)

	// TODO - move to config/hunter
	barrageMultiplier := 1.15

	totalDamageBonus := damageBonusArrows + damageBonusScope + damageBonusAP + damageBonusSkill

	// note - according to cMangos, RWS does not apply to multishot
	minDamage := (float64(h.Bow.MinDamage) + totalDamageBonus) * barrageMultiplier
	maxDamage := (float64(h.Bow.MaxDamage) + totalDamageBonus) * barrageMultiplier

	var randomizedDamage float64
	if h.StandardizedDamage {
		randomizedDamage = (maxDamage + minDamage) / 2
	} else {
		randomizedDamage = minDamage + rand.Float64()*(maxDamage-minDamage)
	}

	if didCrit {
		randomizedDamage *= CRIT_MULTIPLIER
	}

	return int(math.Round(randomizedDamage)), int(math.Floor(minDamage)), int(math.Ceil(maxDamage))
}

// GetSteadyshotDamage returns the randomized, minimum, and maximum, and randomized damage for a steadyshot.
func (h *Hunter) GetSteadyshotDamage(didCrit bool) (int, int, int) {
	damageBonusArrows := float64(h.Bow.Speed * h.ArrowDPS)
	damageBonusScope := float64(h.Bow.ScopeDamage)
	damageBonusAP := float64(h.GetTotalAP()) / 14.0 * NORMALIZED_SPEED
	damageBonusSkill := float64(50)

	// TODO - move to config/hunter
	improvedSteadyshotMultiplier := 1.15

	totalDamageBonus := damageBonusArrows + damageBonusScope + damageBonusAP + damageBonusSkill

	// note - according to checking many target dummy logs, RWS does seem apply to steadyshot on turtleWoW.
	minDamage := (float64(h.Bow.MinDamage) + totalDamageBonus) * improvedSteadyshotMultiplier * h.Talents.RangedWeaponSpec
	maxDamage := (float64(h.Bow.MaxDamage) + totalDamageBonus) * improvedSteadyshotMultiplier * h.Talents.RangedWeaponSpec

	var randomizedDamage float64
	if h.StandardizedDamage {
		randomizedDamage = (maxDamage + minDamage) / 2
	} else {
		randomizedDamage = minDamage + rand.Float64()*(maxDamage-minDamage)
	}

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
	return h.MultishotCooldown
}

func (h *Hunter) GetReloadingTime() float64 {
	return (h.Bow.Speed / h.GetTotalHaste()) - AIMING_TIME
}

func (h *Hunter) RollQuickshotsProc(c *Clock, r *SimResult) {
	if rand.Float64() < 0.10 {
		h.BonusStats.QuickshotsHaste.RemainingTime = QUICKSHOTS_DURATION
		if DEBUG {
			fmt.Printf("%f - Quickshots procced!\n", c.Time)
		}
	}
}

func (h *Hunter) PopRapidFireIfReady(c *Clock) {
	if c.Timers.RapidFireCooldown <= 0 {
		h.BonusStats.RapidFireHaste.RemainingTime = RAPID_FIRE_DURATION
		c.Timers.RapidFireCooldown = RAPID_FIRE_COOLDOWN
		if DEBUG {
			fmt.Printf("%f - Rapid fire popped!\n", c.Time)
		}
	}
}

func (h *Hunter) CheckEndlessQuiver() bool {
	// 6% chance to fire another shot
	return rand.Float64() < 0.06
}
