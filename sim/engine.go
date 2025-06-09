package sim

import (
	"fmt"
	"sync"
)

type ReportType []string

type SimResult struct {
	TotalDamage         float64
	AutoDamage          float64
	MultishotDamage     float64
	SteadyshotDamage    float64
	EndlessQuiverDamage float64
	PiercingShotsDamage float64
	DPS                 float64
	TotalClippingTime   float64
	RapidFireUptime     float64
	QuickshotsUptime    float64

	Report []string
}

type DebugResult struct {
	MinDamageAutoshot   float64
	MaxDamageAutoshot   float64
	MinDamageMultishot  float64
	MaxDamageMultishot  float64
	MinDamageSteadyshot float64
	MaxDamageSteadyshot float64
}

const (
	DEBUG = false
)

func RunBasicSim() SimResult {
	// return runParallelSims(1)
	return runParallelSims(100000)
}

func runSingleSim() SimResult {
	hunter, clock, actionQueue := SetupSim()
	simResult := &SimResult{}

	for !clock.IsDone() {
		clock.Tick(actionQueue, hunter, simResult)
	}

	simResult.DPS = simResult.TotalDamage / clock.EndTime

	return *simResult
}

func runParallelSims(numSims int) SimResult {
	var wg sync.WaitGroup
	results := make(chan SimResult, numSims)

	// Launch goroutines for each simulation
	for i := 0; i < numSims; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- runSingleSim()
		}()
	}

	// Close the channel when all goroutines are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and average results
	var totalDamage, autoDamage, multishotDamage, steadyshotDamage, endlessQuiverDamage, piercingShotsDamage, dps, clippingTime, rapidFireUptime, quickshotsUptime float64
	count := 0

	for result := range results {
		totalDamage += result.TotalDamage
		autoDamage += result.AutoDamage
		multishotDamage += result.MultishotDamage
		steadyshotDamage += result.SteadyshotDamage
		endlessQuiverDamage += result.EndlessQuiverDamage
		piercingShotsDamage += result.PiercingShotsDamage
		dps += result.DPS
		clippingTime += result.TotalClippingTime
		rapidFireUptime += result.RapidFireUptime
		quickshotsUptime += result.QuickshotsUptime
		count++
	}

	// Calculate averages
	avgResult := SimResult{
		TotalDamage:         totalDamage / float64(count),
		AutoDamage:          autoDamage / float64(count),
		MultishotDamage:     multishotDamage / float64(count),
		SteadyshotDamage:    steadyshotDamage / float64(count),
		EndlessQuiverDamage: endlessQuiverDamage / float64(count),
		PiercingShotsDamage: piercingShotsDamage / float64(count),
		DPS:                 dps / float64(count),
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
		fmt.Sprintf("Average DPS: %f", avgResult.DPS),
		fmt.Sprintf("Average Total Clipping Time: %fs", avgResult.TotalClippingTime),
		// TODO - stop hardcoding 60s
		fmt.Sprintf("Average Total Clipping Percentage: (%.2f%%)", avgResult.TotalClippingTime/60*100),
		fmt.Sprintf("Average Rapid Fire Uptime: %fs (%.2f%%)", avgResult.RapidFireUptime, avgResult.RapidFireUptime/60*100),
		fmt.Sprintf("Average Quickshots Uptime: %fs (%.2f%%)", avgResult.QuickshotsUptime, avgResult.QuickshotsUptime/60*100),
		fmt.Sprintf("Number of Simulations: %d", count),
	}

	return avgResult
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
	clock := &Clock{
		Time:     0,
		EndTime:  60,
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
