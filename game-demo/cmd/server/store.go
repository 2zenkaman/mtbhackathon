package main

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

type MemoryStore struct {
	mu sync.Mutex

	bal GameBalance

	players map[string]PlayerState
}

func NewMemoryStore(bal GameBalance) *MemoryStore {
	return &MemoryStore{
		bal:     bal,
		players: make(map[string]PlayerState),
	}
}

func (s *MemoryStore) playerID(r *http.Request) string {
	if v := r.Header.Get("X-Player-Id"); v != "" {
		return v
	}
	if v := r.URL.Query().Get("playerId"); v != "" {
		return v
	}
	return "demo"
}

func (s *MemoryStore) Get(r *http.Request) PlayerState {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.playerID(r)
	st := s.getOrInitLocked(id)
	st = s.applyOfflineProgressLocked(st, time.Now().UTC())
	s.players[id] = st
	return st
}

func (s *MemoryStore) ApplyAction(r *http.Request, actionType string) (PlayerState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.playerID(r)
	now := time.Now().UTC()
	st := s.getOrInitLocked(id)
	st = s.applyOfflineProgressLocked(st, now)

	switch actionType {
	case "buy_cashier":
		if st.Money < s.bal.CostCashier {
			return st, Err("not enough money")
		}
		st.Money -= s.bal.CostCashier
		st.Cashiers++
	case "buy_atm":
		if st.Money < s.bal.CostATM {
			return st, Err("not enough money")
		}
		st.Money -= s.bal.CostATM
		st.ATMs++
	case "upgrade_vault":
		if st.Money < s.bal.CostVault {
			return st, Err("not enough money")
		}
		st.Money -= s.bal.CostVault
		st.VaultCap += s.bal.VaultUpgradeAmount
	case "collect_vault":
		st.Money += st.VaultStored
		st.VaultStored = 0
	case "serve_manual":
		// micro-action: player "helps" serve a client instantly
		if st.Queue <= 0 {
			break
		}
		st.Queue--
		st = s.addEarningsLocked(st, s.bal.IncomePerCustomer)
	default:
		return st, Err("unknown action type: " + actionType)
	}

	st.LastSeenUnix = now.Unix()
	s.players[id] = st
	return st, nil
}

func (s *MemoryStore) Checkin(r *http.Request) (PlayerState, bool, int64, string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.playerID(r)
	now := time.Now().UTC()
	st := s.getOrInitLocked(id)
	st = s.applyOfflineProgressLocked(st, now)

	today := now.Format("2006-01-02")
	if st.LastCheckinDay == today {
		st.LastSeenUnix = now.Unix()
		s.players[id] = st
		return st, false, 0, "already checked in today"
	}

	// If missed at least 1 full day, streak resets to 1.
	if st.LastCheckinDay != "" {
		if last, err := time.ParseInLocation("2006-01-02", st.LastCheckinDay, time.UTC); err == nil {
			days := int(now.Sub(last).Hours() / 24)
			if days >= 2 {
				st.Streak = 0
			}
		}
	}
	st.Streak++
	st.LastCheckinDay = today

	reward := int64(80 + 15*st.Streak)
	st.Money += reward

	st.LastSeenUnix = now.Unix()
	s.players[id] = st
	return st, true, reward, "check-in reward granted"
}

func (s *MemoryStore) getOrInitLocked(id string) PlayerState {
	if st, ok := s.players[id]; ok {
		return st
	}
	now := time.Now().UTC()
	st := PlayerState{
		Money:        250,
		VaultStored:  0,
		VaultCap:     400,
		Cashiers:     1,
		ATMs:         0,
		Queue:        4,
		Streak:       0,
		LastSeenUnix: now.Unix(),
	}
	s.players[id] = st
	return st
}

func (s *MemoryStore) applyOfflineProgressLocked(st PlayerState, now time.Time) PlayerState {
	last := time.Unix(st.LastSeenUnix, 0).UTC()
	if now.Before(last) {
		st.LastSeenUnix = now.Unix()
		return st
	}

	elapsed := now.Sub(last).Seconds()
	if elapsed <= 0.25 {
		return st
	}

	// Arrivals & service are simulated in aggregate, capped to avoid insane jumps in demo.
	if elapsed > 60*60*6 {
		elapsed = 60 * 60 * 6
	}

	arrivals := int64(s.bal.ClientArrivePerSec * elapsed)
	st.Queue += arrivals

	serveRate := s.bal.CashierServePerSec * float64(maxInt64(st.Cashiers, 1))
	served := int64(serveRate * elapsed)
	if served > st.Queue {
		served = st.Queue
	}
	st.Queue -= served
	if served > 0 {
		st = s.addEarningsLocked(st, served*s.bal.IncomePerCustomer)
	}

	atmEarn := int64(s.bal.ATMIncomePerSec * float64(st.ATMs) * elapsed)
	if atmEarn > 0 {
		st = s.addEarningsLocked(st, atmEarn)
	}

	st.LastSeenUnix = now.Unix()
	return st
}

func (s *MemoryStore) addEarningsLocked(st PlayerState, amount int64) PlayerState {
	if amount <= 0 {
		return st
	}
	free := st.VaultCap - st.VaultStored
	if free <= 0 {
		return st
	}
	if amount > free {
		amount = free
	}
	st.VaultStored += amount
	return st
}

type Err string

func (e Err) Error() string { return string(e) }

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// Optional helper: parse int query param for demo tweaks.
func qInt64(r *http.Request, key string) (int64, bool) {
	v := r.URL.Query().Get(key)
	if v == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}
