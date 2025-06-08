package sim

type SimResult struct {
	TotalDamage      float64
	AutoDamage       float64
	MultishotDamage  float64
	SteadyshotDamage float64
	DPS              float64
}

type DebugResult struct {
	MinDamageAutoshot   float64
	MaxDamageAutoshot   float64
	MinDamageMultishot  float64
	MaxDamageMultishot  float64
	MinDamageSteadyshot float64
	MaxDamageSteadyshot float64
}

func RunBasicSim() SimResult {
	hunter, clock, actionQueue := SetupSim()
	simResult := &SimResult{}

	for !clock.IsDone() {
		clock.Tick(actionQueue, hunter, simResult)
	}

	simResult.DPS = simResult.TotalDamage / clock.EndTime

	return *simResult
}

func SetupSim() (*Hunter, *Clock, *ActionQueue) {
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
		RapidFireHaste:  1.0,
		QuickshotsHaste: 1.0,
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
	clock := &Clock{
		Time:     0,
		EndTime:  6000,
		TickSize: 0.01,
		Timers:   &Timers{},
	}
	actionQueue := &ActionQueue{}
	return hunter, clock, actionQueue
}

func DebugValues() DebugResult {
	hunter, clock, actionQueue := SetupSim()
	simResult := &SimResult{}
	clock.Tick(actionQueue, hunter, simResult)
	_, minDamageAuto, maxDamageAuto := hunter.GetAutoshotDamage(false)
	_, minDamageMultishot, maxDamageMultishot := hunter.GetMultishotDamage(false)
	_, minDamageSteadyshot, maxDamageSteadyshot := hunter.GetSteadyshotDamage(false)
	return DebugResult{
		MinDamageAutoshot:   float64(minDamageAuto),
		MaxDamageAutoshot:   float64(maxDamageAuto),
		MinDamageMultishot:  float64(minDamageMultishot),
		MaxDamageMultishot:  float64(maxDamageMultishot),
		MinDamageSteadyshot: float64(minDamageSteadyshot),
		MaxDamageSteadyshot: float64(maxDamageSteadyshot),
	}
}
