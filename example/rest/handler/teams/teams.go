package teams

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/sanekee/limi"
	"github.com/sanekee/limi/example/rest/db"
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
	w.Write(body)
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
	w.Write(mBody)

}

type Team struct {
	limi     struct{} `path:"/teams/{id:[0-9]+}"` //lint:ignore U1000 field parsed by limi
	DBClient DBClient
}

func (t Team) Get(w http.ResponseWriter, req *http.Request) {
	idStr := limi.GetURLParam(req.Context(), "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	merchant, err := t.DBClient.GetTeamByID(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(merchant)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (t Team) Put(w http.ResponseWriter, req *http.Request) {
	idStr := limi.GetURLParam(req.Context(), "id")
	id, err := strconv.Atoi(idStr)
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

	merchant, err := t.DBClient.UpdateTeamByID(id, updateParams)
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
	w.Write(mBody)

}

func (t Team) Delete(w http.ResponseWriter, req *http.Request) {
	idStr := limi.GetURLParam(req.Context(), "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := t.DBClient.DeleteTeamByID(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type TeamMerchants struct {
	limi     struct{} `path:"/teams/{id:[0-9]+}/merchants"` //lint:ignore U1000 field parsed by limi
	DBClient DBClient
}

func (t TeamMerchants) Get(w http.ResponseWriter, req *http.Request) {
	idStr := limi.GetURLParam(req.Context(), "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	merchants, err := t.DBClient.GetMerchantsByTeamID(id)
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
	w.Write(body)
}
