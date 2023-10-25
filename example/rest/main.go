package main

import (
	"fmt"
	"net/http"

	"rest/db"
	"rest/handler/merchant"
	"rest/handler/team"

	"github.com/sanekee/limi"
)

func main() {
	r, err := limi.NewRouter("/", limi.WithProfiler())
	if err != nil {
		panic(err)
	}

	dbClient := db.NewMemClient()
	if err := r.AddHandlers([]limi.Handler{
		merchant.Merchants{DBClient: dbClient},
		merchant.Merchant{DBClient: dbClient},
		team.Teams{DBClient: dbClient},
		team.Team{DBClient: dbClient},
		team.TeamMerchants{DBClient: dbClient},
	}); err != nil {
		panic(err)
	}

	fmt.Println("rest example listening on :3333")
	http.ListenAndServe(":3333", r)
}
