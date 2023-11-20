package teams

import (
	"encoding/json"
	"io"
	"net/http"

	"rest/db"

	"github.com/sanekee/limi"
)

type DBClient interface {
	ListTeams() (db.Teams, error)
	GetTeamByID(int) (db.Team, error)
	CreateTeam(db.CreateTeamParams) (db.Team, error)
	UpdateTeamByID(int, db.UpdateTeamParams) (db.Team, error)
	DeleteTeamByID(int) error
	GetMerchantsByTeamID(int) (db.Merchants, error)
}

type Teams struct {
	DBClient DBClient
}

func (t Teams) Get(w http.ResponseWriter, req *http.Request) {
	l, err := t.DBClient.ListTeams()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(l)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body) // nolint:errcheck
}

func (t Teams) Post(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var createParams db.CreateTeamParams
	if err := json.Unmarshal(body, &createParams); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	merchant, err := t.DBClient.CreateTeam(createParams)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	mBody, err := json.Marshal(merchant)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(mBody) // nolint:errcheck

}

type Team struct {
	_        teamParams `limi:"path=/teams/{id:[0-9]+}"`
	DBClient DBClient
}

type teamParams struct {
	id int `limi:"param"`
}

func (t Team) Get(w http.ResponseWriter, req *http.Request) {
	params, err := limi.GetParams[teamParams](req.Context())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	team, err := t.DBClient.GetTeamByID(params.id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(team)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body) // nolint:errcheck
}

func (t Team) Put(w http.ResponseWriter, req *http.Request) {
	params, err := limi.GetParams[teamParams](req.Context())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var updateParams db.UpdateTeamParams
	if err := json.Unmarshal(body, &updateParams); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	merchant, err := t.DBClient.UpdateTeamByID(params.id, updateParams)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	mBody, err := json.Marshal(merchant)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(mBody) // nolint:errcheck

}

func (t Team) Delete(w http.ResponseWriter, req *http.Request) {
	params, err := limi.GetParams[teamParams](req.Context())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := t.DBClient.DeleteTeamByID(params.id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type TeamMerchants struct {
	_        teamMerchantsParams `limi:"path=/teams/{id:[0-9]+}/merchants"`
	DBClient DBClient
}

type teamMerchantsParams struct {
	id int `limi:"param"`
}

func (t TeamMerchants) Get(w http.ResponseWriter, req *http.Request) {
	params, err := limi.GetParams[teamMerchantsParams](req.Context())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	merchants, err := t.DBClient.GetMerchantsByTeamID(params.id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(merchants)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body) // nolint:errcheck
}
