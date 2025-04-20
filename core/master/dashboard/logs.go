package dashboard

import (
	"api/core/database"
	"api/core/master/sessions"
	"api/core/models"
	"api/core/models/floods"
	"api/core/models/functions"
	"api/core/models/server"
	"net/http"
	"time"
)

func init() {
	Route.NewSub(server.NewRoute("/logs", func(w http.ResponseWriter, r *http.Request) {
		type Page struct {
			Attacks                      []*floods.Attack
			Error                        string
			IsAdmin                      bool
			Username                     string
			SearchResult                 []*floods.Attack
			Name, Title, Vers            string
			ServersCount, Ongoing, Slots int
			Users                        int
			User                         *sessions.Session
			*sessions.Session
		}
		ok, user := sessions.IsLoggedIn(w, r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		isAdmin := user.HasPermission("admin")
		if !user.HasPermission("admin") {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		var attacks []*floods.Attack
		var errMsg string
		if isAdmin {
			attacks, errMsg = getAllAttacks()
		} else {
			attacks, errMsg = getUserAttacks(user.Username)
		}
		searchQuery := r.FormValue("search_target")
		if searchQuery != "" {
			searchResult, err := database.Container.SearchAttacks(searchQuery)
			if err != nil {
				errMsg = err.Error()
			}
			attacks = searchResult
		}
		functions.Render(Page{
			Name:         models.Config.Name,
			Title:        "Logs",
			Attacks:      attacks,
			Error:        errMsg,
			Session:      user,
			IsAdmin:      isAdmin,
			SearchResult: attacks,
			Username:     user.Username, // Passage de Username ici
			User:         user,
		}, w, "user", "logs.html")
	}))
}

func getAllAttacks() ([]*floods.Attack, string) {
	attacks, err := database.Container.GetAllAttacks()
	if err != nil {
		return nil, err.Error()
	}

	for _, attack := range attacks {
		user, err := database.Container.GetUserByID(attack.Parent)
		if err == nil {
			attack.Username = user.Username
		}
		attack.FormattedCreated = time.Unix(attack.Created, 0).Format("2006-01-02 15:04:05")
	}

	return attacks, ""
}

func getUserAttacks(username string) ([]*floods.Attack, string) {
	attacks, err := database.Container.GetUserAttacks(username)
	if err != nil {
		return nil, err.Error()
	}

	for _, attack := range attacks {
		user, err := database.Container.GetUserByID(attack.Parent)
		if err == nil {
			attack.Username = user.Username
		}
		attack.FormattedCreated = time.Unix(attack.Created, 0).Format("2006-01-02 15:04:05")
	}

	return attacks, ""
}
