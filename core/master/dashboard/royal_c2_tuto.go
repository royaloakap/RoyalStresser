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
	// Route principale pour le tuto
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

		// Lien direct vers la vidéo, si elle est dans le dossier "videos" à la racine
		videoPath := "./videos/royalc2.mp4"
		if _, err := os.Stat(videoPath); os.IsNotExist(err) {
			http.Error(w, "Video not found", http.StatusNotFound)
			return
		}

		// Rendu de la page avec la vidéo et d'autres informations
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

	// Si tu veux que cette vidéo soit servie directement depuis le dossier "videos"
	Route.NewSub(server.NewRoute("/videos/{filename}", func(w http.ResponseWriter, r *http.Request) {
		filename := r.URL.Path[len("/videos/"):]
		videoPath := "./videos/" + filename

		// Vérification si le fichier existe
		if _, err := os.Stat(videoPath); os.IsNotExist(err) {
			http.Error(w, "Video not found", http.StatusNotFound)
			return
		}

		// Définir les en-têtes pour la vidéo
		w.Header().Set("Content-Disposition", "inline; filename="+filename)
		w.Header().Set("Content-Type", "video/mp4")

		// Servir le fichier vidéo
		http.ServeFile(w, r, videoPath)
	}))
}
