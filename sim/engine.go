package sim

import (
	"fmt"
	"math"
	"sync"
)

const (
	DEBUG = false
)

type InputParameters struct {
	AP                     int
	Crit                   float64
	Hit                    int
	ItemHaste              float64
	QuiverHaste            float64
	ArrowDPS               float64
	Bow                    *Bow
	Talents                *Talents
	Race                   *Race
	Ping                   float64
	NumberOfSims           int
	FightDurationInSeconds float64

	// TODO - move to "special?"
	MultishotCooldown float64
}

type SimResult struct {
	DPS                 float64
	DPSRangeMin         float64
	DPSRangeMax         float64
	TotalDamage         float64
	AutoDamage          float64
	MultishotDamage     float64
	SteadyshotDamage    float64
	EndlessQuiverDamage float64
	PiercingShotsDamage float64
	TotalClippingTime   float64
	RapidFireUptime     float64
	QuickshotsUptime    float64

	Report []string
}

type DebugResult struct {
	MinDamageAutoshot   float64
	MaxDamageAutoshot   float64
	AutoshotSpeed       float64
	MinDamageMultishot  float64
	MaxDamageMultishot  float64
	MinDamageSteadyshot float64
	MaxDamageSteadyshot float64
}

func RunBasicSim(params *InputParameters) (*SimResult, error) {
	if params.NumberOfSims > 100000 {
		return nil, fmt.Errorf("invalid number of simulations: %d (maximum allowed: 100000)", params.NumberOfSims)
	}
	result := runParallelSims(params)
	return &result, nil
}

func runSingleSim(params *InputParameters) SimResult {
	hunter, clock, actionQueue := SetupSim(params)
	simResult := &SimResult{}

	for !clock.IsDone() {
		clock.Tick(actionQueue, hunter, simResult)
	}

	simResult.DPS = simResult.TotalDamage / clock.EndTime

	return *simResult
}

func SetupSim(params *InputParameters) (*Hunter, *Clock, *ActionQueue) {
	hunter := &Hunter{
		AP:                params.AP,
		Crit:              params.Crit,
		Hit:               params.Hit,
		ItemHaste:         params.ItemHaste,
		QuiverHaste:       params.QuiverHaste,
		ArrowDPS:          params.ArrowDPS,
		Bow:               params.Bow,
		Talents:           params.Talents,
		Race:              params.Race,
		Ping:              params.Ping,
		MultishotCooldown: params.MultishotCooldown,
		BonusStats: &HunterBonusStats{
			TrinketAP:       0,
			BonusCrit:       0,
			RapidFireHaste:  HasteBuff{RemainingTime: 0, Haste: 1.4},
			QuickshotsHaste: HasteBuff{RemainingTime: 0, Haste: 1.15},
		},
	}
	clock := &Clock{
		Time:     0,
		EndTime:  params.FightDurationInSeconds,
		TickSize: 0.01,
		Timers:   &Timers{},
	}
	actionQueue := &ActionQueue{}
	return hunter, clock, actionQueue
}

func runParallelSims(params *InputParameters) SimResult {
	var wg sync.WaitGroup
	results := make(chan SimResult, params.NumberOfSims)

	// Launch goroutines for each simulation
	for i := 0; i < params.NumberOfSims; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- runSingleSim(params)
		}()
	}

	// Close the channel when all goroutines are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and average results
	var totalDamage, autoDamage, multishotDamage, steadyshotDamage, endlessQuiverDamage, piercingShotsDamage, dps, clippingTime, rapidFireUptime, quickshotsUptime float64
	var dpsValues []float64 // Add slice to store individual DPS values
	var minDPS, maxDPS float64
	count := 0

	for result := range results {
		totalDamage += result.TotalDamage
		autoDamage += result.AutoDamage
		multishotDamage += result.MultishotDamage
		steadyshotDamage += result.SteadyshotDamage
		endlessQuiverDamage += result.EndlessQuiverDamage
		piercingShotsDamage += result.PiercingShotsDamage
		dps += result.DPS
		dpsValues = append(dpsValues, result.DPS) // Store each DPS value

		// Track min and max DPS
		if count == 0 {
			minDPS = result.DPS
			maxDPS = result.DPS
		} else {
			if result.DPS < minDPS {
				minDPS = result.DPS
			}
			if result.DPS > maxDPS {
				maxDPS = result.DPS
			}
		}

		clippingTime += result.TotalClippingTime
		rapidFireUptime += result.RapidFireUptime
		quickshotsUptime += result.QuickshotsUptime
		count++
	}

	// Calculate averages
	avgDPS := dps / float64(count)

	// Calculate standard deviation
	var sumSquaredDiff float64
	for _, value := range dpsValues {
		diff := value - avgDPS
		sumSquaredDiff += diff * diff
	}
	stdDev := math.Sqrt(sumSquaredDiff / float64(count))

	avgResult := SimResult{
		DPS:                 avgDPS,
		DPSRangeMin:         minDPS,
		DPSRangeMax:         maxDPS,
		TotalDamage:         totalDamage / float64(count),
		AutoDamage:          autoDamage / float64(count),
		MultishotDamage:     multishotDamage / float64(count),
		SteadyshotDamage:    steadyshotDamage / float64(count),
		EndlessQuiverDamage: endlessQuiverDamage / float64(count),
		PiercingShotsDamage: piercingShotsDamage / float64(count),
		TotalClippingTime:   clippingTime / float64(count),
		RapidFireUptime:     rapidFireUptime / float64(count),
		QuickshotsUptime:    quickshotsUptime / float64(count),
	}

	// Generate report
	avgResult.Report = []string{
		fmt.Sprintf("Average Total Damage: %f", avgResult.TotalDamage),
		fmt.Sprintf("Average Auto Damage: %f (%.2f%%)", avgResult.AutoDamage, avgResult.AutoDamage/avgResult.TotalDamage*100),
		fmt.Sprintf("Average Multishot Damage: %f (%.2f%%)", avgResult.MultishotDamage, avgResult.MultishotDamage/avgResult.TotalDamage*100),
		fmt.Sprintf("Average Steadyshot Damage: %f (%.2f%%)", avgResult.SteadyshotDamage, avgResult.SteadyshotDamage/avgResult.TotalDamage*100),
		fmt.Sprintf("Average Endless Quiver Damage: %f (%.2f%%)", avgResult.EndlessQuiverDamage, avgResult.EndlessQuiverDamage/avgResult.TotalDamage*100),
		fmt.Sprintf("Average Piercing Shots Damage: %f (%.2f%%)", avgResult.PiercingShotsDamage, avgResult.PiercingShotsDamage/avgResult.TotalDamage*100),
		fmt.Sprintf("Average DPS: %f (±%.2f) [%.2f - %.2f]", avgResult.DPS, stdDev, minDPS, maxDPS),
		fmt.Sprintf("Average Total Clipping Time: %fs (%.2f%%)", avgResult.TotalClippingTime, avgResult.TotalClippingTime/params.FightDurationInSeconds*100),
		fmt.Sprintf("Average Rapid Fire Uptime: %fs (%.2f%%)", avgResult.RapidFireUptime, avgResult.RapidFireUptime/params.FightDurationInSeconds*100),
		fmt.Sprintf("Average Quickshots Uptime: %fs (%.2f%%)", avgResult.QuickshotsUptime, avgResult.QuickshotsUptime/params.FightDurationInSeconds*100),
		fmt.Sprintf("Number of Simulations: %d", count),
	}

	return avgResult
}

type StatEquivalenceResult struct {
	CritApEquivalence    string
	AgilityApEquivalence string
}

func EstimateStatEquivalence(params *InputParameters) (*StatEquivalenceResult, error) {
	if params.NumberOfSims > 100000 {
		return nil, fmt.Errorf("invalid number of simulations: %d (maximum allowed: 100000)", params.NumberOfSims)
	}
	// we check +10% crit instead of +1% crit, it was yielding much more accurate results with less simulations
	paramsPlusOneCrit := *params
	paramsPlusOneCrit.Crit = params.Crit + 1.0
	critPlusOneResults := runParallelSims(&paramsPlusOneCrit)

	critPlusOneDPS := critPlusOneResults.DPS

	apEquivalenceMin := 10.0
	apEquivalenceMax := 60.0
	cursor := (apEquivalenceMax + apEquivalenceMin) / 2.0
	var diff float64

	// run a max of 20 fake "binary searches", in case something goes wrong to avoid an infinite loop
	for i := 0; i < 20; i++ {
		paramsPlusAp := *params
		paramsPlusAp.AP = params.AP + int(cursor)
		apPlusResults := runParallelSims(&paramsPlusAp)
		apPlusDPS := apPlusResults.DPS
		diff = math.Abs(apPlusDPS - critPlusOneDPS)

		fmt.Printf("Iteration: %d, Cursor: %f, diff: %f\n", i+1, cursor, diff)
		fmt.Printf("bonus: %f, apPlusDPS: %f, critPlusOneDPS: %f\n", cursor, apPlusDPS, critPlusOneDPS)
		if diff < 0.1 {
			break
		}
		if apPlusDPS > critPlusOneDPS {
			apEquivalenceMax = cursor
		} else {
			apEquivalenceMin = cursor
		}
		cursor = (apEquivalenceMax + apEquivalenceMin) / 2.0
	}

	apPerAgility := 2.2
	critPerAgility := 1.0 / 53.0
	agilityApEquivalence := apPerAgility + (critPerAgility * cursor)

	result := &StatEquivalenceResult{
		CritApEquivalence:    fmt.Sprintf("at this gear level, +1%% crit is approximately equal to +%f AP", cursor),
		AgilityApEquivalence: fmt.Sprintf("at this gear level, +1 agility is approximately equal to +%f AP (with Blessing of Kings)", agilityApEquivalence),
	}
	return result, nil
}

func DebugValues(params *InputParameters) DebugResult {
	hunter, _, _ := SetupSim(params)
	_, minDamageAuto, maxDamageAuto := hunter.GetAutoshotDamage(false)
	_, minDamageMultishot, maxDamageMultishot := hunter.GetMultishotDamage(false)
	_, minDamageSteadyshot, maxDamageSteadyshot := hunter.GetSteadyshotDamage(false)
	return DebugResult{
		MinDamageAutoshot:   float64(minDamageAuto),
		MaxDamageAutoshot:   float64(maxDamageAuto),
		AutoshotSpeed:       hunter.GetReloadingTime() + AIMING_TIME,
		MinDamageMultishot:  float64(minDamageMultishot),
		MaxDamageMultishot:  float64(maxDamageMultishot),
		MinDamageSteadyshot: float64(minDamageSteadyshot),
		MaxDamageSteadyshot: float64(maxDamageSteadyshot),
	}
}
