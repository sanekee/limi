package admin

import "net/http"

type Admin struct{}

func (a Admin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Admin portal")) // nolint:errcheck
}
