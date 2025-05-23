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
	Route.NewSub(server.NewRoute("/edit", func(w http.ResponseWriter, r *http.Request) {
		type Page struct {
			Name, Title, Vers            string
			ServersCount, Ongoing, Slots int
			Users, Online                int
			Attack                       int
			Remotes                      map[string]*servers.Server
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
			Title:        "Admin Management",
			Vers:         models.Config.Vers,
			ServersCount: len(servers.Servers) + len(apis.Apis),
			Ongoing:      database.Container.GlobalRunning(),
			Slots:        servers.Slots()[0],
			Attack:       database.Container.GlobalRunning() + models.Config.Fake.Attacks,
			Users:        database.Container.Users(),
			Remotes:      servers.Servers,
			Online:       sessions.Count() + sess.Count(),
			Session:      user,
		}, w, "admin", "editUser.html")
	}))
}
