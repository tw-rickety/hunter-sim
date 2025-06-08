package api

import (
	"encoding/json"
	"net/http"

	"github.com/tw-rickety/hunter-sim/sim"
)

func StartServer() {
	http.HandleFunc("/simulate", func(w http.ResponseWriter, r *http.Request) {
		result := sim.RunBasicSim()
		json.NewEncoder(w).Encode(result)
	})
	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		result := sim.DebugValues()
		json.NewEncoder(w).Encode(result)
	})

	http.ListenAndServe(":8080", nil)
}
