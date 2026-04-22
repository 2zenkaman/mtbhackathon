package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

type ActionRequest struct {
	Type string `json:"type"`
}

type CheckinResponse struct {
	State       PlayerState `json:"state"`
	DidCheckin  bool        `json:"didCheckin"`
	RewardMoney int64       `json:"rewardMoney"`
	Message     string      `json:"message"`
}

type PlayerState struct {
	Money       int64 `json:"money"`
	VaultStored int64 `json:"vaultStored"`
	VaultCap    int64 `json:"vaultCap"`

	Cashiers int64 `json:"cashiers"`
	ATMs     int64 `json:"atms"`

	Queue int64 `json:"queue"`

	Streak         int64  `json:"streak"`
	LastCheckinDay string `json:"lastCheckinDay"` // YYYY-MM-DD (UTC)

	LastSeenUnix int64 `json:"lastSeenUnix"`
}

type GameBalance struct {
	ClientArrivePerSec float64
	CashierServePerSec float64

	IncomePerCustomer int64
	ATMIncomePerSec   float64

	CostCashier int64
	CostATM     int64
	CostVault   int64

	VaultUpgradeAmount int64
}

func defaultBalance() GameBalance {
	return GameBalance{
		ClientArrivePerSec: 0.18,
		CashierServePerSec: 0.10,

		IncomePerCustomer: 12,
		ATMIncomePerSec:   0.35,

		CostCashier: 150,
		CostATM:     250,
		CostVault:   200,

		VaultUpgradeAmount: 250,
	}
}

func main() {
	bal := defaultBalance()
	store := NewMemoryStore(bal)

	mux := http.NewServeMux()

	// API
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /api/state", func(w http.ResponseWriter, r *http.Request) {
		state := store.Get(r)
		writeJSON(w, http.StatusOK, state)
	})

	mux.HandleFunc("POST /api/action", func(w http.ResponseWriter, r *http.Request) {
		var req ActionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		state, err := store.ApplyAction(r, req.Type)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, state)
	})

	mux.HandleFunc("POST /api/checkin", func(w http.ResponseWriter, r *http.Request) {
		state, did, reward, msg := store.Checkin(r)
		writeJSON(w, http.StatusOK, CheckinResponse{
			State:       state,
			DidCheckin:  did,
			RewardMoney: reward,
			Message:     msg,
		})
	})

	srv := &http.Server{
		Addr:              ":8088",
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("game-demo API server listening on %s", srv.Addr)
	log.Printf("health: GET http://localhost%s/healthz", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For demo simplicity. In prod: strict allowlist + proper auth tokens.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,X-Player-Id")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
