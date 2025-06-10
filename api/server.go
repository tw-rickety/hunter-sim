package api

import (
	"encoding/json"
	"net/http"

	"github.com/tw-rickety/hunter-sim/sim"
)

func StartServer() {
	http.HandleFunc("/simulate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var params sim.InputParameters
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		result, err := sim.RunBasicSim(&params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(result)
	})
	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var params sim.InputParameters
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		result := sim.DebugValues(&params)
		json.NewEncoder(w).Encode(result)
	})

	http.ListenAndServe(":8080", nil)
}
