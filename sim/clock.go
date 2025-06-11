package sim

import (
	"fmt"
	"math"
)

type Clock struct {
	Time     float64
	EndTime  float64
	TickSize float64
	Timers   *Timers
}

type Timers struct {
	Gcd                float64
	ReloadingTime      float64
	AimingTime         float64
	MultishotCastTime  float64
	SteadyshotCastTime float64
	MultishotCooldown  float64
	RapidFireCooldown  float64

	// tracks how much the next autoshot was clipped by (not overall total clipping time)
	NextShotClippingTime float64
}

func (c *Clock) Tick(a *ActionQueue, h *Hunter, r *SimResult) {
	c.Time += c.TickSize
	c.decrementTimers(a, h, r)
	c.decrementBuffs(h, r)
	c.decrementDoTs(h, r)
	a.Process(c, h, r)
}

func (c *Clock) decrementTimers(a *ActionQueue, h *Hunter, r *SimResult) {
	wasMultishotCastActive := c.Timers.MultishotCastTime > 0
	wasSteadyshotCastActive := c.Timers.SteadyshotCastTime > 0
	wasAimingActive := c.Timers.AimingTime > 0

	// always tick cooldowns and cast times
	c.Timers.Gcd = math.Max(c.Timers.Gcd-c.TickSize, 0)
	c.Timers.MultishotCooldown = math.Max(c.Timers.MultishotCooldown-c.TickSize, 0)
	c.Timers.MultishotCastTime = math.Max(c.Timers.MultishotCastTime-c.TickSize, 0)
	c.Timers.SteadyshotCastTime = math.Max(c.Timers.SteadyshotCastTime-c.TickSize, 0)
	c.Timers.AimingTime = math.Max(c.Timers.AimingTime-c.TickSize, 0)
	c.Timers.ReloadingTime = math.Max(c.Timers.ReloadingTime-c.TickSize, 0)
	c.Timers.RapidFireCooldown = math.Max(c.Timers.RapidFireCooldown-c.TickSize, 0)

	if wasMultishotCastActive && c.Timers.MultishotCastTime <= 0 {
		a.FireMultishot(c, h, r)
	}
	if wasSteadyshotCastActive && c.Timers.SteadyshotCastTime <= 0 {
		a.FireSteadyshot(c, h, r)
	}
	if wasAimingActive && c.Timers.AimingTime <= 0 {
		a.FireAutoshot(c, h, r)
	}
}

func (c *Clock) decrementBuffs(h *Hunter, r *SimResult) {
	wasQuickshotsHasteActive := h.BonusStats.QuickshotsHaste.RemainingTime > 0
	wasRapidFireHasteActive := h.BonusStats.RapidFireHaste.RemainingTime > 0

	if h.BonusStats.QuickshotsHaste.RemainingTime > 0 {
		r.QuickshotsUptime += c.TickSize
	}
	if h.BonusStats.RapidFireHaste.RemainingTime > 0 {
		r.RapidFireUptime += c.TickSize
	}

	h.BonusStats.QuickshotsHaste.RemainingTime = math.Max(h.BonusStats.QuickshotsHaste.RemainingTime-c.TickSize, 0)
	h.BonusStats.RapidFireHaste.RemainingTime = math.Max(h.BonusStats.RapidFireHaste.RemainingTime-c.TickSize, 0)

	if h.DebugCombatLog {
		if wasQuickshotsHasteActive && h.BonusStats.QuickshotsHaste.RemainingTime <= 0 {
			h.BonusStats.QuickshotsHaste.RemainingTime = 0
			r.Report = append(r.Report, fmt.Sprintf("%.2fs: Quickshots expired", c.Time))
		}
		if wasRapidFireHasteActive && h.BonusStats.RapidFireHaste.RemainingTime <= 0 {
			h.BonusStats.RapidFireHaste.RemainingTime = 0
			r.Report = append(r.Report, fmt.Sprintf("%.2fs: Rapid fire expired", c.Time))
		}
	}
}

func (c *Clock) decrementDoTs(h *Hunter, r *SimResult) {
	if h.PiercingShotsDoT.DurationLeft > 0 && h.PiercingShotsDoT.DurationLeft < PIERCING_SHOTS_DURATION {
		if isDivisible(h.PiercingShotsDoT.DurationLeft, h.PiercingShotsDoT.TicksEvery) {
			r.PiercingShotsDamage += h.PiercingShotsDoT.DamagePerTick
			if h.DebugCombatLog {
				r.Report = append(r.Report, fmt.Sprintf("%.2fs: Piercing shots DoT tick for %0.f damage", c.Time, h.PiercingShotsDoT.DamagePerTick))
			}
		}
	}
	h.PiercingShotsDoT.DurationLeft = math.Max(h.PiercingShotsDoT.DurationLeft-c.TickSize, 0)
}

func (c *Clock) IsDone() bool {
	return c.Time >= c.EndTime
}

func (c *Clock) IsFresh() bool {
	return c.Timers.Gcd == 0 &&
		c.Timers.MultishotCooldown == 0 &&
		c.Timers.MultishotCastTime == 0 &&
		c.Timers.SteadyshotCastTime == 0 &&
		c.Timers.ReloadingTime == 0 &&
		c.Timers.AimingTime == 0
}

func (c *Clock) IsCasting() bool {
	return c.Timers.MultishotCastTime > 0 ||
		c.Timers.SteadyshotCastTime > 0
}

func isDivisible(a float64, b float64) bool {
	// check if its whole number first
	if a-float64(int64(a)) > 0.000001 {
		return false
	}
	return int64(a)%int64(b) == 0
}
