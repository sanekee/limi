package db

import (
	"fmt"
	"sync"
)

// simple in memory db with teams and merchants (1 to many) objects store
type memClient struct {
	merchants map[int]Merchant
	teams     map[int]Team

	rwLock     sync.RWMutex
	merchantID int
	teamID     int
}

func NewMemClient() *memClient {
	return &memClient{
		merchants: make(map[int]Merchant),
		teams:     make(map[int]Team),
	}
}

func (m *memClient) ListMerchants() (Merchants, error) {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()

	var merchants []Merchant
	for _, mc := range m.merchants {
		merchants = append(merchants, mc)
	}
	return Merchants{Merchants: merchants}, nil
}

func (m *memClient) GetMerchantByID(id int) (Merchant, error) {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()

	if mc, ok := m.merchants[id]; ok {
		return mc, nil
	}

	return Merchant{}, fmt.Errorf("merchant %d not found", id)
}

func (m *memClient) CreateMerchant(params CreateMerchantParams) (Merchant, error) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	_, ok := m.teams[params.TeamID]
	if !ok {
		return Merchant{}, fmt.Errorf("team %d not found", params.TeamID)
	}
	m.merchantID++

	merchant := Merchant{
		ID:     m.merchantID,
		Name:   params.Name,
		TeamID: params.TeamID,
	}
	m.merchants[m.merchantID] = merchant

	return merchant, nil
}

func (m *memClient) UpdateMerchantByID(id int, params UpdateMerchantParams) (Merchant, error) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	if _, ok := m.merchants[id]; !ok {
		return Merchant{}, fmt.Errorf("merchant %d not found", id)
	}

	merchant := Merchant{
		ID:     id,
		Name:   params.Name,
		TeamID: params.TeamID,
	}
	m.merchants[id] = merchant

	return merchant, nil
}

func (m *memClient) DeleteMerchantByID(id int) error {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	if _, ok := m.merchants[id]; !ok {
		return fmt.Errorf("merchant %d not found", id)
	}

	delete(m.merchants, id)
	return nil
}

func (m *memClient) ListTeams() (Teams, error) {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()

	var teams []Team
	for id, team := range m.teams {
		merchants, err := m.GetMerchantsByTeamID(id)
		if err != nil {
			return Teams{}, fmt.Errorf("error loading merchants for team %d %w", id, err)
		}

		team.Merchants = merchants.Merchants
		teams = append(teams, team)
	}

	return Teams{Teams: teams}, nil
}

func (m *memClient) GetTeamByID(id int) (Team, error) {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()

	team, ok := m.teams[id]
	if !ok {
		return Team{}, fmt.Errorf("team %d not found", id)
	}

	merchants, err := m.GetMerchantsByTeamID(id)
	if err != nil {
		return Team{}, fmt.Errorf("error loading merchants for team %d %w", id, err)
	}

	team.Merchants = merchants.Merchants
	return team, nil

}

func (m *memClient) CreateTeam(params CreateTeamParams) (Team, error) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	m.teamID++

	team := Team{
		ID:   m.teamID,
		Name: params.Name,
	}
	m.teams[m.teamID] = team

	return team, nil
}

func (m *memClient) UpdateTeamByID(id int, params UpdateTeamParams) (Team, error) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	if _, ok := m.teams[id]; !ok {
		return Team{}, fmt.Errorf("team %d not found", id)
	}

	team := Team{
		ID:   id,
		Name: params.Name,
	}
	m.teams[id] = team

	return team, nil
}

func (m *memClient) DeleteTeamByID(id int) error {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	if _, ok := m.teams[id]; !ok {
		return fmt.Errorf("team %d not found", id)
	}

	for mID, merc := range m.merchants {
		if merc.TeamID == id {
			delete(m.merchants, mID)
		}
	}
	delete(m.teams, id)
	return nil
}

func (m *memClient) GetMerchantsByTeamID(teamID int) (Merchants, error) {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()

	var merchants []Merchant
	for _, mc := range m.merchants {
		if mc.TeamID == teamID {
			merchants = append(merchants, mc)
		}
	}
	return Merchants{Merchants: merchants}, nil
}
