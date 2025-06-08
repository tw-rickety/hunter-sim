package sim

import (
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
	MultishotCooldown  float64
	MultishotCastTime  float64
	SteadyshotCastTime float64
	ReloadingTime      float64
	AimingTime         float64
}

func (c *Clock) Tick(a *ActionQueue, h *Hunter, r *SimResult) {
	c.Time += c.TickSize
	c.decrementTimers(a, h, r)
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
