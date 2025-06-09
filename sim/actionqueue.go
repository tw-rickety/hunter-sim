package sim

import (
	"fmt"
	"math"
	"math/rand"
)

type ActionQueue struct {
	lastShot    string
	queuedSpell string
}

func (a *ActionQueue) Process(c *Clock, h *Hunter, r *SimResult) {
	h.PopRapidFireIfReady(c)

	// // Initial multishot
	// if c.IsFresh() {
	// 	a.StartMultishot(c, h)
	// 	return
	// }

	// If we are casting multi or steady, don't queue any abilities and check if we are clipping autoshot
	if c.IsCasting() {
		if c.Timers.ReloadingTime <= 0 {
			r.TotalClippingTime += c.TickSize
			c.Timers.NextShotClippingTime += c.TickSize
		}
		return
	}

	// If we are already aiming an autoshot, don't queue any abilities
	if c.Timers.AimingTime > 0 {
		return
	}

	// Start aiming an autoshot if we're done reloading
	if c.Timers.ReloadingTime == 0 {
		a.StartAiming(c, h)
		return
	}

	// a.QueueRotationASMA(c, h)
	a.QueueRotationAMAS(c, h)
}

// auto -> multi -> auto -> steady rotation
func (a *ActionQueue) QueueRotationAMAS(c *Clock, h *Hunter) {
	if c.Timers.Gcd == 0 && c.Timers.MultishotCooldown == 0 && a.lastShot == "autoshot" {
		a.StartMultishot(c, h)
		return
	}

	if c.Timers.Gcd == 0 && a.lastShot == "autoshot" {
		a.StartSteadyshot(c, h)
		return
	}
}

// auto -> steady -> multi -> auto rotation
func (a *ActionQueue) QueueRotationASMA(c *Clock, h *Hunter) {
	if c.Timers.Gcd == 0 && a.lastShot == "autoshot" {
		a.StartSteadyshot(c, h)
		return
	}

	if c.Timers.Gcd == 0 && c.Timers.MultishotCooldown == 0 && a.lastShot == "steadyshot" {
		a.StartMultishot(c, h)
		return
	}
}

func (a *ActionQueue) QueueAbilities(c *Clock, h *Hunter) {
	if c.Timers.Gcd == 0 && a.lastShot == "autoshot" {
		a.StartSteadyshot(c, h)
		return
	}

}

func (a *ActionQueue) StartAiming(c *Clock, h *Hunter) {
	c.Timers.AimingTime = AIMING_TIME
}

func (a *ActionQueue) StartMultishot(c *Clock, h *Hunter) {
	c.Timers.Gcd = GCD
	c.Timers.MultishotCooldown = h.GetMultishotCooldown()
	// We add ping to the cast time because that's functionality how it works - i.e for 150ms ping multishot (500ms cast time):
	// Client starts casting at 0ms, Server registers cast at 150ms, Server finishes cast at 650ms.
	// This means we can simulate by assuming that multishot took 650ms to cast instead of 500ms.
	c.Timers.MultishotCastTime = h.GetMultishotCastTime() + (h.Ping / 1000)
}

func (a *ActionQueue) StartSteadyshot(c *Clock, h *Hunter) {
	c.Timers.Gcd = GCD
	c.Timers.SteadyshotCastTime = h.GetSteadyshotCastTime() + (h.Ping / 1000)
}

func (a *ActionQueue) FireAutoshot(c *Clock, h *Hunter, r *SimResult) {
	a.lastShot = "autoshot"
	c.Timers.ReloadingTime = h.GetReloadingTime()
	a.AutoshotDamageDealt(c, h, r, false)

	// check if quickshots or endless quiver procced
	h.RollQuickshotsProc(c, r)
	extraShot := h.CheckEndlessQuiver()
	if extraShot {
		a.AutoshotDamageDealt(c, h, r, true)
	}
}

func (a *ActionQueue) AutoshotDamageDealt(c *Clock, h *Hunter, r *SimResult, extraShot bool) {
	didCrit := (100 * rand.Float64()) < h.GetTotalCrit()
	damage, _, _ := h.GetAutoshotDamage(didCrit)
	r.TotalDamage += float64(damage)
	if extraShot {
		r.EndlessQuiverDamage += float64(damage)
	} else {
		r.AutoDamage += float64(damage)
	}

	if DEBUG {
		if extraShot {
			if didCrit {
				fmt.Printf("%f - endless quiver CRIT for %d damage.\n", c.Time, damage)
			} else {
				fmt.Printf("%f - endless quiver hit for %d damage.\n", c.Time, damage)
			}
			return
		}

		if didCrit {
			fmt.Printf("%f - autoshot CRIT for %d damage.", c.Time, damage)
		} else {
			fmt.Printf("%f - autoshot hit for %d damage.", c.Time, damage)
		}

		if c.Timers.NextShotClippingTime > 0 {
			fmt.Printf("Shot clipped by: %fms\n", c.Timers.NextShotClippingTime*1000)
		} else {
			fmt.Printf("\n")
		}
	}

	c.Timers.NextShotClippingTime = 0
}

func (a *ActionQueue) FireMultishot(c *Clock, h *Hunter, r *SimResult) {
	a.lastShot = "multishot"
	a.MultishotDamageDealt(c, h, r)

	// check if endless quiver procced
	extraShot := h.CheckEndlessQuiver()
	if extraShot {
		a.AutoshotDamageDealt(c, h, r, true)
	}
}

func (a *ActionQueue) MultishotDamageDealt(c *Clock, h *Hunter, r *SimResult) {
	didCrit := (100 * rand.Float64()) < h.GetTotalCrit()
	damage, _, _ := h.GetMultishotDamage(didCrit)
	r.TotalDamage += float64(damage)
	r.MultishotDamage += float64(damage)

	// piercing shots procced, overwrite previous piercing shots DoT
	if didCrit {
		h.PiercingShotsDoT.DurationLeft = PIERCING_SHOTS_DURATION
		h.PiercingShotsDoT.TicksEvery = PIERCING_SHOTS_TICKS_EVERY
		h.PiercingShotsDoT.DamagePerTick = math.Round(float64(damage) * PIERCING_SHOTS_MULTIPLIER * PIERCING_SHOTS_TICKS_EVERY / PIERCING_SHOTS_DURATION)
	}

	if DEBUG {
		if didCrit {
			fmt.Printf("%f - multishot CRIT for %d damage\n", c.Time, damage)
		} else {
			fmt.Printf("%f - multishot hit for %d damage\n", c.Time, damage)
		}
	}
}

func (a *ActionQueue) FireSteadyshot(c *Clock, h *Hunter, r *SimResult) {
	a.lastShot = "steadyshot"
	a.SteadyshotDamageDealt(c, h, r)

	// check if endless quiver procced
	extraShot := h.CheckEndlessQuiver()
	if extraShot {
		a.AutoshotDamageDealt(c, h, r, true)
	}
}

func (a *ActionQueue) SteadyshotDamageDealt(c *Clock, h *Hunter, r *SimResult) {
	a.lastShot = "steadyshot"
	didCrit := (100 * rand.Float64()) < h.GetTotalCrit()
	damage, _, _ := h.GetSteadyshotDamage(didCrit)
	r.TotalDamage += float64(damage)
	r.SteadyshotDamage += float64(damage)

	// piercing shots procced, overwrite previous piercing shots DoT
	if didCrit {
		h.PiercingShotsDoT.DurationLeft = PIERCING_SHOTS_DURATION
		h.PiercingShotsDoT.DamagePerTick = float64(damage) * PIERCING_SHOTS_MULTIPLIER * PIERCING_SHOTS_TICKS_EVERY / PIERCING_SHOTS_DURATION
		h.PiercingShotsDoT.TicksEvery = PIERCING_SHOTS_TICKS_EVERY
	}

	if DEBUG {
		if didCrit {
			fmt.Printf("%f - steadyshot CRIT for %d damage\n", c.Time, damage)
		} else {
			fmt.Printf("%f - steadyshot hit for %d damage\n", c.Time, damage)
		}
	}
}
