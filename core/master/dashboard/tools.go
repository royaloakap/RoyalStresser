package dashboard

import (
	"api/core/master/sessions"
	"api/core/models"
	"api/core/models/functions"
	"api/core/models/server"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func lookupCFXCode(cfxCode string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://cfx-resolver.site/fivem?cfx=%s", cfxCode)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Erreur lors de la requête CFX: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Erreur de lecture de la réponse: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("Erreur lors du décodage JSON: %v", err)
	}

	if serverIP, ok := result["server_ip"].(string); ok {
		serverPort := result["server_port"].(string)
		serverName := result["server_name"].(string)
		maxClients := result["max_clients"].(float64)
		currentClients := result["current_clients"].(float64)

		resultData := map[string]interface{}{
			"ServerIP":        serverIP,
			"Port":            serverPort,
			"ServerName":      serverName,
			"Players":         fmt.Sprintf("%v / %v", currentClients, maxClients),
			"MaxClients":      maxClients,
			"Gametype":        result["gametype"].(string),
			"ServerVersion":   result["server_version"].(string),
			"ResourcesCount":  result["resources_count"].(float64),
			"OwnerName":       result["owner_name"].(string),
			"Credits":         result["credits"].(string),
			"ProjectDesc":     result["project_desc"].(string),
			"DiscordLink":     result["discord_link"].(string),
			"Map":             result["map"].(string),
			"EnhancedHosting": result["enhanced_hosting"].(string),
			"ISP":             "N/A",
		}

		if isp, err := lookupIPInfo(serverIP); err == nil {
			resultData["ISP"] = isp
		}

		return resultData, nil
	}

	return nil, fmt.Errorf("Code CFX NOT FOUND.")
}

func lookupIPInfo(ipAddress string) (map[string]interface{}, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s", ipAddress)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Erreur lors de la requête IP: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Erreur de lecture de la réponse: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("Erreur lors du décodage JSON: %v", err)
	}

	return result, nil
}

func init() {
	Route.NewSub(server.NewRoute("/tools/lookup-cfx", func(w http.ResponseWriter, r *http.Request) {
		type Page struct {
			Name, Title string
			Result      map[string]interface{}
			Error       string
			*sessions.Session
		}

		ok, user := sessions.IsLoggedIn(w, r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		switch strings.ToLower(r.Method) {
		case "get":
			functions.Render(Page{
				Name:    models.Config.Name,
				Title:   "CFX Lookup",
				Session: user,
			}, w, "tools", "lookupcfx.html")
		case "post":
			cfxCode := r.FormValue("cfx_code")
			var result map[string]interface{}
			var errMsg string
			if cfxCode == "" {
				errMsg = "Code CFX invalide ou vide."
			} else {
				cfxResult, err := lookupCFXCode(cfxCode)
				if err != nil {
					errMsg = err.Error()
				} else {
					result = cfxResult
				}
			}

			functions.Render(Page{
				Name:    models.Config.Name,
				Title:   "CFX Lookup",
				Result:  result,
				Error:   errMsg,
				Session: user,
			}, w, "tools", "lookupcfx.html")
		}
	}))
}

func isValidIP(ip string) bool {
	re := regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	return re.MatchString(ip)
}
func lookupIPDetails(ip string) (map[string]interface{}, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Erreur lors de la requête IP: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Erreur de lecture de la réponse: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("Erreur lors du décodage JSON: %v", err)
	}

	return result, nil
}

func init() {
	Route.NewSub(server.NewRoute("/tools/lookup-ip", func(w http.ResponseWriter, r *http.Request) {
		type Page struct {
			Name, Title string
			Result      map[string]interface{}
			Error       string
			*sessions.Session
		}

		ok, user := sessions.IsLoggedIn(w, r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		switch strings.ToLower(r.Method) {
		case "get":
			functions.Render(Page{
				Name:    models.Config.Name,
				Title:   "IP Lookup",
				Session: user,
			}, w, "tools", "lookupip.html")
		case "post":
			ip := r.FormValue("ip_address")
			var result map[string]interface{}
			var errMsg string

			if ip == "" || !isValidIP(ip) {
				errMsg = "Adresse IP invalide ou vide."
			} else {
				ipDetails, err := lookupIPDetails(ip)
				if err != nil {
					errMsg = err.Error()
				} else if ipDetails["status"] == "fail" {
					errMsg = "IP non trouvée ou invalide."
				} else {
					result = map[string]interface{}{
						"IP":           ipDetails["query"],
						"Country":      ipDetails["country"],
						"Region":       ipDetails["regionName"],
						"City":         ipDetails["city"],
						"ISP":          ipDetails["isp"],
						"Organization": ipDetails["org"],
						"Timezone":     ipDetails["timezone"],
					}
				}
			}

			functions.Render(Page{
				Name:    models.Config.Name,
				Title:   "IP Lookup",
				Result:  result,
				Error:   errMsg,
				Session: user,
			}, w, "tools", "lookupip.html")
		}
	}))
}

func cleanURL(url string) string {
	return strings.TrimPrefix(strings.TrimPrefix(url, "http://"), "https://")
}

func isValidIPOrDomain(input string) bool {
	if strings.Contains(input, ":") {
		input = strings.Split(input, ":")[0]
	}
	_, err := http.Get(fmt.Sprintf("https://api.mcsrvstat.us/3/%s", input))
	return err == nil
}

func lookupMinecraftServer(serverIP string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.mcsrvstat.us/3/%s", serverIP)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Error during Minecraft request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading the response: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("Error decoding JSON: %v", err)
	}
	if result["online"] != nil && result["online"].(bool) {
		return map[string]interface{}{
			"IP":       result["ip"],
			"Port":     result["port"],
			"Hostname": result["hostname"],
			"Online":   result["online"],
			"Players":  result["players"],
			"Version":  result["version"],
			"Motd":     result["motd"],
			"Status":   result["status"],
			"ISP":      "N/A",
		}, nil
	} else {
		return map[string]interface{}{
			"IP":       result["ip"],
			"Port":     result["port"],
			"Hostname": result["hostname"],
			"Online":   result["online"],
			"Players":  "N/A",
			"Version":  "N/A",
			"Motd":     "N/A",
			"Status":   "Offline",
			"ISP":      "N/A",
		}, nil
	}

	return nil, fmt.Errorf("Minecraft server not found or offline.")
}

func init() {
	Route.NewSub(server.NewRoute("/tools/lookup-mc", func(w http.ResponseWriter, r *http.Request) {
		type Page struct {
			Name, Title string
			Result      map[string]interface{}
			Error       string
			*sessions.Session
		}

		ok, user := sessions.IsLoggedIn(w, r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		switch strings.ToLower(r.Method) {
		case "get":
			functions.Render(Page{
				Name:    models.Config.Name,
				Title:   "Minecraft Lookup",
				Session: user,
			}, w, "tools", "lookupmc.html")
		case "post":
			serverIP := cleanURL(r.FormValue("server_ip"))
			var result map[string]interface{}
			var errMsg string
			if serverIP == "" || !isValidIPOrDomain(serverIP) {
				errMsg = "IP ou domaine invalide."
			} else {
				mcResult, err := lookupMinecraftServer(serverIP)
				if err != nil {
					errMsg = err.Error()
				} else {
					result = mcResult

					if mcResult["IP"] != nil {
						if ipInfo, err := lookupIPInfo(mcResult["IP"].(string)); err == nil {
							result["ISP"] = ipInfo["isp"]
							result["Country"] = ipInfo["country"]
							result["Region"] = ipInfo["regionName"]
							result["City"] = ipInfo["city"]
							result["Organization"] = ipInfo["org"]
						} else {
							result["ISP"] = "ISP N/A"
						}
					}
				}
			}

			functions.Render(Page{
				Name:    models.Config.Name,
				Title:   "Minecraft Lookup",
				Result:  result,
				Error:   errMsg,
				Session: user,
			}, w, "tools", "lookupmc.html")
		}
	}))
}

func lookupMinecraftIPInfo(ipAddress string) (map[string]interface{}, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s", ipAddress)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Error during IP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading the IP response: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("Error decoding IP response: %v", err)
	}

	if isp, ok := result["isp"].(string); ok {
		return map[string]interface{}{
			"ISP":          isp,
			"Country":      result["country"],
			"Region":       result["regionName"],
			"City":         result["city"],
			"Timezone":     result["timezone"],
			"Organization": result["org"],
		}, nil
	}

	return nil, fmt.Errorf("ISP information not found.")
}
func lookupPing(ip string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("https://cr2off.site/ping?ip=%s", ip)

	// Requête HTTP GET
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Erreur lors de la requête Ping : %v", err)
	}
	defer resp.Body.Close()

	// Lecture de la réponse
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Erreur de lecture de la réponse : %v", err)
	}

	// Décodage JSON
	var jsonResponse struct {
		Results []map[string]interface{} `json:"results"`
	}
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, fmt.Errorf("Erreur lors du décodage JSON : %v", err)
	}

	// Filtrer les résultats pour ne pas afficher les IPs de proxy
	var filteredResults []map[string]interface{}
	for _, result := range jsonResponse.Results {
		if proxy, ok := result["proxy"].(string); ok && proxy != "" {
			// Masquer les proxies ou filtrer ici
			result["proxy"] = "N/A" // ou omettre ce champ
		}

		// Ajouter un champ 'time' formaté si 'response_time' existe
		if responseTime, ok := result["response_time"].(string); ok {
			result["time"] = responseTime
		} else {
			result["time"] = "N/A" // Si pas de réponse, afficher "N/A"
		}

		filteredResults = append(filteredResults, result)
	}

	return filteredResults, nil
}

func init() {
	// Route pour Ping Lookup
	Route.NewSub(server.NewRoute("/tools/ping", func(w http.ResponseWriter, r *http.Request) {
		type Page struct {
			Name, Title string
			Result      []map[string]interface{}
			Error       string
			*sessions.Session
		}

		// Vérification de la session utilisateur
		ok, user := sessions.IsLoggedIn(w, r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		switch strings.ToLower(r.Method) {
		case "get":
			// Chargement de la page
			functions.Render(Page{
				Name:    models.Config.Name,
				Title:   "Ping Lookup",
				Session: user,
			}, w, "tools", "ping.html")

		case "post":
			// Gestion de la requête POST
			ip := r.FormValue("ip")
			var results []map[string]interface{}
			var errMsg string

			// Validation de l'entrée
			if ip == "" {
				errMsg = "Adresse IP invalide ou vide."
			} else {
				// Exécution du Ping
				pingResults, err := lookupPing(ip)
				if err != nil {
					errMsg = err.Error()
				} else if len(pingResults) == 0 {
					errMsg = "Impossible de récupérer les données pour cette IP."
				} else {
					results = pingResults
				}
			}

			// Rendu de la page avec les résultats ou une erreur
			functions.Render(Page{
				Name:    models.Config.Name,
				Title:   "Ping Lookup",
				Result:  results,
				Error:   errMsg,
				Session: user,
			}, w, "tools", "ping.html")
		}
	}))
}

func lookupPaping(ip string, port int) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("https://cr2off.site/paping?ip=%s&port=%d", ip, port)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Erreur lors de la requête Paping: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Erreur de lecture de la réponse: %v", err)
	}

	var jsonResponse struct {
		Results []map[string]interface{} `json:"results"`
	}

	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, fmt.Errorf("Erreur lors du décodage JSON: %v", err)
	}

	return jsonResponse.Results, nil
}

func init() {
	Route.NewSub(server.NewRoute("/tools/paping", func(w http.ResponseWriter, r *http.Request) {
		type Page struct {
			Name, Title string
			Result      []map[string]interface{}
			Error       string
			*sessions.Session
		}

		ok, user := sessions.IsLoggedIn(w, r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		switch strings.ToLower(r.Method) {
		case "get":
			functions.Render(Page{
				Name:    models.Config.Name,
				Title:   "Paping Lookup",
				Session: user,
			}, w, "tools", "paping.html")

		case "post":
			ip := r.FormValue("ip")
			port, _ := strconv.Atoi(r.FormValue("port"))

			var results []map[string]interface{}
			var errMsg string

			if ip == "" || port < 1 || port > 65535 {
				errMsg = "IP ou port invalide."
			} else {
				papingResults, err := lookupPaping(ip, port)
				if err != nil {
					errMsg = err.Error()
				} else {
					results = papingResults
				}
			}

			functions.Render(Page{
				Name:    models.Config.Name,
				Title:   "Paping Lookup",
				Result:  results,
				Error:   errMsg,
				Session: user,
			}, w, "tools", "paping.html")
		}
	}))
}

func init() {
	Route.NewSub(server.NewRoute("/tools/send-time", func(w http.ResponseWriter, r *http.Request) {
		type Page struct {
			Name, Title string
			Session     *sessions.Session
		}

		ok, user := sessions.IsLoggedIn(w, r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		switch strings.ToLower(r.Method) {
		case "get":
			functions.Render(Page{
				Name:    models.Config.Name,
				Title:   "Send Time",
				Session: user,
			}, w, "tools", "sendtime.html")
		case "post":
			r.ParseForm()
			target := r.PostFormValue("target")
			timeToSendStr := r.PostFormValue("time")

			// Validation des champs
			if target == "" || timeToSendStr == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Invalid target or time"))
				return
			}

			// Conversion de `timeToSendStr` (string) en `timeToSend` (int)
			timeToSend, err := strconv.Atoi(timeToSendStr)
			if err != nil || timeToSend <= 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Invalid time format, must be a positive integer"))
				return
			}

			// Lancement de la tâche en arrière-plan
			go func(target string, timeToSend int) {
				sendCount := 0
				for sendCount < timeToSend {
					sendCount++
					time.Sleep(2 * time.Second)
				}
			}(target, timeToSend)

			// Réponse de confirmation
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Send Time initiated successfully"))
		}
	}))
}
