package sim

import (
	"fmt"
	"math/rand"
)

type ActionQueue struct {
	lastShot    string
	queuedSpell string
}

func (a *ActionQueue) Process(c *Clock, h *Hunter, r *SimResult) {
	// Queueing
	if c.IsFresh() {
		a.StartMultishot(c, h)
		return
	}

	if c.IsCasting() && c.Timers.ReloadingTime <= 0 {
		// TODO - Clipped for x seconds
		// fmt.Println("CLIPPING")
	}

	if c.IsCasting() || c.Timers.AimingTime > 0 {
		return
	}

	if c.Timers.ReloadingTime == 0 {
		a.StartAiming(c, h)
		return
	}

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
	c.Timers.MultishotCooldown = MULTISHOT_COOLDOWN
	// There is ping to start the cast, and then ping for client to see cast as finished, so we multiply by 2
	c.Timers.MultishotCastTime = h.GetMultishotCastTime() + (h.Ping / 1000 * 2)
	c.Timers.Gcd = GCD
}

func (a *ActionQueue) StartSteadyshot(c *Clock, h *Hunter) {
	// There is ping to start the cast, and then ping for client to see cast as finished, so we multiply by 2
	c.Timers.SteadyshotCastTime = h.GetSteadyshotCastTime() + (h.Ping / 1000 * 2)
	c.Timers.Gcd = GCD
}

func (a *ActionQueue) FireAutoshot(c *Clock, h *Hunter, r *SimResult) {
	a.lastShot = "autoshot"
	didCrit := (100 * rand.Float64()) < h.GetTotalCrit()
	damage, _, _ := h.GetAutoshotDamage(didCrit)
	r.TotalDamage += float64(damage)
	r.AutoDamage += float64(damage)

	c.Timers.ReloadingTime = h.GetReloadingTime()

	if didCrit {
		fmt.Printf("%f - autoshot CRIT for %d damage\n", c.Time, damage)
	} else {
		fmt.Printf("%f - autoshot hit for %d damage\n", c.Time, damage)
	}
}

func (a *ActionQueue) FireMultishot(c *Clock, h *Hunter, r *SimResult) {
	a.lastShot = "multishot"
	didCrit := (100 * rand.Float64()) < h.GetTotalCrit()
	damage, _, _ := h.GetMultishotDamage(didCrit)
	r.TotalDamage += float64(damage)
	r.MultishotDamage += float64(damage)

	c.Timers.MultishotCooldown = h.GetMultishotCooldown()

	if didCrit {
		fmt.Printf("%f - multishot CRIT for %d damage\n", c.Time, damage)
	} else {
		fmt.Printf("%f - multishot hit for %d damage\n", c.Time, damage)
	}
}

func (a *ActionQueue) FireSteadyshot(c *Clock, h *Hunter, r *SimResult) {
	a.lastShot = "steadyshot"
	didCrit := (100 * rand.Float64()) < h.GetTotalCrit()
	damage, _, _ := h.GetSteadyshotDamage(didCrit)
	r.TotalDamage += float64(damage)
	r.SteadyshotDamage += float64(damage)

	if didCrit {
		fmt.Printf("%f - steadyshot CRIT for %d damage\n", c.Time, damage)
	} else {
		fmt.Printf("%f - steadyshot hit for %d damage\n", c.Time, damage)
	}
}
