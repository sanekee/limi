package blog

import (
	"net/http"

	"github.com/sanekee/limi"
)

type Blog struct{}

// Get handles HTTP GET request
func (s Blog) Get(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("A cool story.")) // nolint:errcheck
}

// Post handles HTTP POST request
func (s Blog) Post(w http.ResponseWriter, req *http.Request) {
	// create story
	w.WriteHeader(http.StatusOK)
}

type Author struct {
	_ authorParams `limi:"path={storyId:[0-9]+}/author"` // custom relative path
}

type authorParams struct {
	storyID int `limi:"param=storyId"`
}

// Get handles HTTP GET request
func (s Author) Get(w http.ResponseWriter, req *http.Request) {
	params, err := limi.GetParams[authorParams](req.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// retrieve author and response
	author := getAuthor(params.storyID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("The auther is " + author)) // nolint:errcheck
}

func getAuthor(int) string {
	return "me"
}

type Copyright struct {
	_ struct{} `limi:"path=/copyright"` // custom absolute path
}

// Get handles HTTP GET request
func (s Copyright) Get(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Copyright ©️ 2023")) // nolint:errcheck
}
