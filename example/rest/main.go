package main

import (
	"fmt"
	"net/http"

	"github.com/sanekee/limi"
	"github.com/sanekee/limi/example/rest/db"
	"github.com/sanekee/limi/example/rest/handler/merchants"
	"github.com/sanekee/limi/example/rest/handler/teams"
)

func main() {
	r, err := limi.NewRouter("/", limi.WithProfiler())
	if err != nil {
		panic(err)
	}

	dbClient := db.NewMemClient()
	if err := r.AddHandlers([]limi.Handler{
		merchants.Merchants{DBClient: dbClient},
		merchants.Merchant{DBClient: dbClient},
		teams.Teams{DBClient: dbClient},
		teams.Team{DBClient: dbClient},
		teams.TeamMerchants{DBClient: dbClient},
	}); err != nil {
		panic(err)
	}

	fmt.Println("rest example listening on :3333")
	http.ListenAndServe(":3333", r)
}
