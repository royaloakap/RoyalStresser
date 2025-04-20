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
	Route.NewSub(server.NewRoute("/products/Royal_Ascii", func(w http.ResponseWriter, r *http.Request) {
		type Page struct {
			Name, Title, Vers            string
			ServersCount, Ongoing, Slots int
			Users                        int
			Remotes                      map[string]*servers.Server
			*sessions.Session
		}

		// Vérification si l'utilisateur est connecté
		ok, user := sessions.IsLoggedIn(w, r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		// Rendu de la page
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
		}, w, "", "Royal_Ascii.zip")
	}))

	// Route pour télécharger le fichier ZIP
	Route.NewSub(server.NewRoute("/products/Royal_Ascii.zip", func(w http.ResponseWriter, r *http.Request) {
		// Définir le chemin vers le fichier ZIP
		zipPath := "./Royal_Ascii.zip" // Assure-toi que le fichier est à la racine du projet

		// Vérifier si le fichier existe
		if _, err := os.Stat(zipPath); os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Définir les en-têtes pour le téléchargement du fichier
		w.Header().Set("Content-Disposition", "attachment; filename=Royal_Ascii.zip")
		w.Header().Set("Content-Type", "application/zip")

		// Servir le fichier
		http.ServeFile(w, r, zipPath)
	}))
}
