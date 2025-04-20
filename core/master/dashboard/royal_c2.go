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
	"strings"
)

func init() {
	Route.NewSub(server.NewRoute("/products/royal_c2", func(w http.ResponseWriter, r *http.Request) {
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
		filename := strings.TrimPrefix(r.URL.Path, "royalsrc")
		imagePath := "./videos/" + filename + ".png"
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			imagePath = "" // Si l'image n'existe pas, on laisse vide
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
		}, w, "products", "royal_c2.html")
	}))
	Route.NewSub(server.NewRoute("/videos/{filename}", func(w http.ResponseWriter, r *http.Request) {
		filename := strings.TrimPrefix(r.URL.Path, "/videos/")

		imagePath := "./videos/" + filename

		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Disposition", "inline; filename="+filename)
		w.Header().Set("Content-Type", "image/png")

		http.ServeFile(w, r, imagePath)
	}))
}
