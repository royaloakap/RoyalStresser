package dashboard

import (
	"api/core/models/server"
	"log"
	"net/http"
)

func init() {
	// Route pour vérifier les requêtes suspectes
	Route.NewSub(server.NewRoute("/check-devtools", func(w http.ResponseWriter, r *http.Request) {
		// Vérifier des entêtes spécifiques qui pourraient indiquer l'utilisation de l'inspecteur
		userAgent := r.Header.Get("User-Agent")
		if userAgent == "" {
			http.Error(w, "No user-agent provided", http.StatusBadRequest)
			return
		}

		// Par exemple, si on détecte un comportement suspect via un user-agent ou un autre header
		if containsSuspiciousHeader(userAgent) {
			// Effectuer une redirection
			http.Redirect(w, r, "/redirect-url", http.StatusFound)
			log.Println("Suspected devtools activity detected, redirecting.")
			return
		}
		w.Write([]byte("Request processed normally"))
	}))
}

// Fonction pour vérifier des headers suspects ou des comportements
func containsSuspiciousHeader(userAgent string) bool {
	// Ajouter des conditions sur l'agent utilisateur ou des headers
	if userAgent == "DevTools" {
		return true
	}
	return false
}
