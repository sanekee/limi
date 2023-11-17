package merchants

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"rest/db"

	"github.com/sanekee/limi"
)

type DBClient interface {
	ListMerchants() (db.Merchants, error)
	GetMerchantByID(int) (db.Merchant, error)
	CreateMerchant(db.CreateMerchantParams) (db.Merchant, error)
	UpdateMerchantByID(int, db.UpdateMerchantParams) (db.Merchant, error)
	DeleteMerchantByID(int) error
}

type Merchants struct {
	DBClient DBClient
}

func (m Merchants) Get(w http.ResponseWriter, req *http.Request) {
	l, err := m.DBClient.ListMerchants()
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

func (m Merchants) Post(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var createParams db.CreateMerchantParams
	if err := json.Unmarshal(body, &createParams); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	merchant, err := m.DBClient.CreateMerchant(createParams)
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

type Merchant struct {
	_        struct{} `limi:"path={id}"`
	DBClient DBClient
}

func (m Merchant) Get(w http.ResponseWriter, req *http.Request) {
	idStr := limi.GetURLParam(req.Context(), "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	merchant, err := m.DBClient.GetMerchantByID(id)
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
	w.Write(body) // nolint:errcheck
}

func (m Merchant) Put(w http.ResponseWriter, req *http.Request) {
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

	var updateParams db.UpdateMerchantParams
	if err := json.Unmarshal(body, &updateParams); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	merchant, err := m.DBClient.UpdateMerchantByID(id, updateParams)
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

func (m Merchant) Delete(w http.ResponseWriter, req *http.Request) {
	idStr := limi.GetURLParam(req.Context(), "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := m.DBClient.DeleteMerchantByID(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
