package internal

import (
	"api/core/database"
	"api/core/master/sessions"
	"api/core/models"
	"api/core/models/functions"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Fonction d'envoi de journaux Telegram
func sendTelegramLog(message, username, ip, timestamp string, success bool) {
	token := "6924113960:AAHxSCJJ0zbDHh8zvbd9iZaYg6e85GHqAy0"
	chatID := "-4718955831"

	status := "Success"
	if !success {
		status = "Failure"
	}

	telegramMessage := fmt.Sprintf(
		"*Royal Stresser Login Log*\n"+
			"Status: %s\n"+
			"Username: %s\n"+
			"IP: %s\n"+
			"Timestamp: %s\n"+
			"Details: %s",
		status, username, ip, timestamp, message,
	)

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := map[string]string{
		"chat_id":    chatID,
		"text":       telegramMessage,
		"parse_mode": "Markdown",
	}

	jsonPayload, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Printf("Failed to send Telegram log: %v\n", err)
	} else {
		resp.Body.Close()
	}
}

// Fonction de connexion
func Login(w http.ResponseWriter, r *http.Request) {
	type Page struct {
		Name   string
		Title  string
		Script template.HTML
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	ip := KeyByRealIP(r) // Récupération de l'IP réelle

	username := ""
	if len(r.Form["login-username"]) > 0 {
		username = r.Form["login-username"][0]
	}

	// Vérifiez que "login-username" existe et n'est pas vide
	if username == "" {
		sendTelegramLog("Missing username in login form.", username, ip, time.Now().Format(time.RFC3339), false)
		functions.Render(Page{
			Name:  models.Config.Name,
			Title: "Login",
			Script: template.HTML(functions.Toast(functions.Toastr{
				Icon:  "error",
				Title: "Error!",
				Text:  "Username is required.",
			})),
		}, w, "login", "login.html")
		return
	}

	// Vérifiez que "login-password" existe et n'est pas vide
	if len(r.Form["login-password"]) == 0 || len(r.Form["login-password"][0]) == 0 {
		sendTelegramLog("Missing password in login form.", username, ip, time.Now().Format(time.RFC3339), false)
		functions.Render(Page{
			Name:  models.Config.Name,
			Title: "Login",
			Script: template.HTML(functions.Toast(functions.Toastr{
				Icon:  "error",
				Title: "Error!",
				Text:  "Password is required.",
			})),
		}, w, "login", "login.html")
		return
	}

	// Récupérez l'utilisateur par le nom d'utilisateur
	user, err := database.Container.GetUser(username)
	if err != nil || user == nil {
		sendTelegramLog("Invalid username or user not found.", username, ip, time.Now().Format(time.RFC3339), false)
		functions.Render(Page{
			Name:  models.Config.Name,
			Title: "Login",
			Script: template.HTML(functions.Toast(functions.Toastr{
				Icon:  "error",
				Title: "Error!",
				Text:  "Invalid credentials.",
			})),
		}, w, "login", "login.html")
		return
	}

	// Vérifiez que le mot de passe est correct
	if !user.IsKey([]byte(r.Form["login-password"][0])) {
		sendTelegramLog("Invalid password.", username, ip, time.Now().Format(time.RFC3339), false)
		functions.Render(Page{
			Name:  models.Config.Name,
			Title: "Login",
			Script: template.HTML(functions.Toast(functions.Toastr{
				Icon:  "error",
				Title: "Error!",
				Text:  "Invalid credentials.",
			})),
		}, w, "login", "login.html")
		return
	}

	// Créez un token de session
	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(30 * time.Minute)
	if _, remember := r.Form["remember-me"]; remember {
		expiresAt = time.Now().Add(48 * time.Hour)
	}

	// Stockez la session
	sessions.Sessions[sessionToken] = sessions.Session{
		User:   user,
		Expiry: expiresAt,
	}

	// Définissez le cookie de session
	http.SetCookie(w, &http.Cookie{
		Name:    "session-token",
		Value:   sessionToken,
		Expires: expiresAt,
	})

	// Envoyez un log Telegram pour la connexion réussie
	sendTelegramLog("Login successful.", username, ip, time.Now().Format(time.RFC3339), true)

	// Redirigez vers le tableau de bord
	http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
}
