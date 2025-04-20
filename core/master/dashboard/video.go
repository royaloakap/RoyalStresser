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
	// Route pour afficher le tutoriel Royal C2
	Route.NewSub(server.NewRoute("/tuto/royal_c2", func(w http.ResponseWriter, r *http.Request) {
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
		}, w, "tuto", "royal_c2.html")
	}))

	// Route pour afficher la vidéo tutoriel depuis le dossier "videos"
	Route.NewSub(server.NewRoute("/tuto/royal_c2/video/{filename}", func(w http.ResponseWriter, r *http.Request) {
		// Récupérer le nom du fichier vidéo depuis l'URL
		filename := r.URL.Path[len("/tuto/royal_c2/video/"):]
		videoPath := "./videos/" + filename

		// Vérifier si le fichier existe
		if _, err := os.Stat(videoPath); os.IsNotExist(err) {
			http.Error(w, "Video not found", http.StatusNotFound)
			return
		}

		// Définir les en-têtes pour la réponse vidéo
		w.Header().Set("Content-Disposition", "inline; filename="+filename)
		w.Header().Set("Content-Type", "video/mp4")

		// Servir le fichier vidéo
		http.ServeFile(w, r, videoPath)
	}))
}
