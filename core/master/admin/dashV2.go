package admin

import (
	"api/core/database"
	"api/core/master/sessions"
	"api/core/models"
	"api/core/models/apis"
	"api/core/models/functions"
	"api/core/models/server"
	"api/core/models/servers"
	sess "api/core/net/sessions"
	"net/http"
)

func init() {
	Route.NewSub(server.NewRoute("/dash", func(w http.ResponseWriter, r *http.Request) {
		type Page struct {
			Name, Title, Vers            string
			ServersCount, Ongoing, Slots int
			Users, Remotes               int
			*sessions.Session
		}
		ok, user := sessions.IsLoggedIn(w, r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		if !user.HasPermission("admin") {
			http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
			return
		}
		functions.Render(Page{
			Name:         models.Config.Name,
			Title:        "Admin Test",
			Vers:         models.Config.Vers,
			ServersCount: len(servers.Servers) + len(apis.Apis),
			Ongoing:      database.Container.GlobalRunning(),
			Slots:        servers.Slots()[0],
			Users:        database.Container.Users(),
			Remotes:      sessions.Count() + sess.Count(),
			Session:      user,
		}, w, "admin", "dashV2.html")
	}))
}
