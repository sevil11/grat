package api

import (
	"encoding/json"
	"net/http"
	"sync"
)

var (
	settings = map[string]interface{}{
		"server_name":     "G-RAT C2",
		"beacon_interval": 60,
		"max_agents":      100,
		"log_to_file":     false,
	}
	settingsMutex sync.RWMutex
)

// HandleSettings handles getting and updating server settings
func HandleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Return current settings
		settingsMutex.RLock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(settings)
		settingsMutex.RUnlock()
		
	case http.MethodPost:
		// Update settings
		var newSettings map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&newSettings); err != nil {
			http.Error(w, "Invalid settings data", http.StatusBadRequest)
			return
		}
		
		// Update only provided settings
		settingsMutex.Lock()
		for key, value := range newSettings {
			settings[key] = value
		}
		settingsMutex.Unlock()
		
		// Return updated settings
		settingsMutex.RLock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(settings)
		settingsMutex.RUnlock()
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
