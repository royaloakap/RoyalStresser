package dashboard

import (
	"api/core/database"
	"api/core/master/sessions"
	"api/core/models"
	"api/core/models/apis"
	"api/core/models/functions"
	"api/core/models/server"
	"api/core/models/servers"
	"net/http"
	"os"
)

func init() {
	Route.NewSub(server.NewRoute("/tuto/stresser", func(w http.ResponseWriter, r *http.Request) {
		type Page struct {
			Name, Title, Vers            string
			ServersCount, Ongoing, Slots int
			Users                        int
			Remotes                      map[string]*servers.Server
			*sessions.Session
		}
		ok, user := sessions.IsLoggedIn(w, r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		videoPath := "./videos/stresser.mp4"
		if _, err := os.Stat(videoPath); os.IsNotExist(err) {
			http.Error(w, "Video not found", http.StatusNotFound)
			return
		}
		functions.Render(Page{
			Name:         models.Config.Name,
			Title:        "Manager",
			Vers:         models.Config.Vers,
			ServersCount: len(servers.Servers) + len(apis.Apis),
			Ongoing:      database.Container.GlobalRunning(),
			Slots:        servers.Slots()[0],
			Users:        database.Container.Users() + models.Config.Fake.Users,
			Remotes:      servers.Servers,
			Session:      user,
		}, w, "tuto", "stresser.html")
	}))
	Route.NewSub(server.NewRoute("/videos/{filename}", func(w http.ResponseWriter, r *http.Request) {
		filename := r.URL.Path[len("/videos/"):]
		videoPath := "./videos/" + filename

		if _, err := os.Stat(videoPath); os.IsNotExist(err) {
			http.Error(w, "Video not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Disposition", "inline; filename="+filename)
		w.Header().Set("Content-Type", "video/mp4")
		http.ServeFile(w, r, videoPath)
	}))
}
