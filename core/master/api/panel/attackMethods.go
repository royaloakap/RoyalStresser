package panelapi

import (
	"api/core/master/sessions"
	"api/core/models/floods"
	"api/core/models/server"
	"encoding/json"
	"net/http"
	"strings"
)

func init() {
	Route.NewSub(server.NewRoute("/methods", func(w http.ResponseWriter, r *http.Request) {
		if strings.ToLower(r.Method) == "post" {
			ok, user := sessions.IsLoggedIn(w, r)
			if !ok {
				http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
				return
			}
			type method struct {
				Description string `json:"description"`
				ID          int    `json:"id"`
				Method      string `json:"method"`
				PanelMethod string `json:"panel_method"`
				Subnet      int    `json:"subnet"`
				Type        string `json:"type"`
			}
			type status struct {
				Status  string    `json:"status"`
				Methods []*method `json:"methods"`
			}
			var s = &status{
				Status:  "success",
				Methods: make([]*method, 0),
			}
			if user.HasPermission("vip") {
				for name, meth := range floods.Methods {
					s.Methods = append(s.Methods, &method{
						Description: meth.Description,
						Method:      name,
						PanelMethod: meth.Name,
						ID:          0,
						Subnet:      meth.Subnet,
						Type: func(t int) string {
							switch t {
							case 1:
								return "UDP (AMP)"
							case 2:
								return "UDP"
							case 3:
								return "TCP"
							case 4:
								return "NETWORK"
							case 5:
								return "BOTNET"
							}
							return "UNKNOWN"
						}(meth.Mtype),
					})
				}
			} else {
				s.Methods = append(s.Methods, &method{
					Description: "ESP",
					Method:      "ESP",
					PanelMethod: "ESP",
					ID:          0,
					Subnet:      0,
					Type:        "Free",
				})
				s.Methods = append(s.Methods, &method{
					Description: "TCP",
					Method:      "TCP",
					PanelMethod: "TCP",
					ID:          0,
					Subnet:      0,
					Type:        "Free",
				})
				s.Methods = append(s.Methods, &method{
					Description: "UDP",
					Method:      "UDP",
					PanelMethod: "UDP",
					ID:          0,
					Subnet:      0,
					Type:        "Free",
				})
				s.Methods = append(s.Methods, &method{
					Description: "SOCKET",
					Method:      "SOCKET",
					PanelMethod: "SOCKET",
					ID:          1,
					Subnet:      0,
					Type:        "Free",
				})
				s.Methods = append(s.Methods, &method{
					Description: "RAND",
					Method:      "RAND",
					PanelMethod: "RAND",
					ID:          0,
					Subnet:      0,
					Type:        "Free",
				})
				s.Methods = append(s.Methods, &method{
					Description: "MIXAMP",
					Method:      "MIXAMP",
					PanelMethod: "MIXAMP",
					ID:          1,
					Subnet:      0,
					Type:        "Free",
				})
				s.Methods = append(s.Methods, &method{
					Description: "ACK",
					Method:      "ACK",
					PanelMethod: "ACK",
					ID:          0,
					Subnet:      0,
					Type:        "Free",
				})
				s.Methods = append(s.Methods, &method{
					Description: "LDAP",
					Method:      "LDAP",
					PanelMethod: "LDAP",
					ID:          1,
					Subnet:      0,
					Type:        "Free",
				})
				s.Methods = append(s.Methods, &method{
					Description: "VSE",
					Method:      "VSE",
					PanelMethod: "VSE",
					ID:          0,
					Subnet:      0,
					Type:        "Free",
				})
				s.Methods = append(s.Methods, &method{
					Description: "NFO",
					Method:      "NFO",
					PanelMethod: "NFO",
					ID:          1,
					Subnet:      0,
					Type:        "Free",
				})
				s.Methods = append(s.Methods, &method{
					Description: "SYN",
					Method:      "SYN",
					PanelMethod: "SYN",
					ID:          0,
					Subnet:      0,
					Type:        "Free",
				})
				s.Methods = append(s.Methods, &method{
					Description: "BASICUDP",
					Method:      "BASICUDP",
					PanelMethod: "BASICUDP",
					ID:          1,
					Subnet:      0,
					Type:        "Free",
				})
				s.Methods = append(s.Methods, &method{
					Description: "FIVEM",
					Method:      "FIVEM",
					PanelMethod: "FIVEM",
					ID:          0,
					Subnet:      0,
					Type:        "Free",
				})
				s.Methods = append(s.Methods, &method{
					Description: "OVH-UDP",
					Method:      "OVH-UDP",
					PanelMethod: "OVH-UDP",
					ID:          1,
					Subnet:      0,
					Type:        "Free",
				})
				s.Methods = append(s.Methods, &method{
					Description: "GAME-MIX",
					Method:      "GAME-MIX",
					PanelMethod: "GAME-MIX",
					ID:          0,
					Subnet:      0,
					Type:        "Free",
				})
				s.Methods = append(s.Methods, &method{
					Description: "GAME-RL",
					Method:      "GAME-RL",
					PanelMethod: "GAME-RL",
					ID:          1,
					Subnet:      0,
					Type:        "Free",
				})

				s.Methods = append(s.Methods, &method{
					Description: "HTTP-RDM ( Layer 7 )",
					Method:      "HTTP-RDM",
					PanelMethod: "HTTP-RDM",
					ID:          2,
					Subnet:      0,
					Type:        "Free",
				})
				s.Methods = append(s.Methods, &method{
					Description: "HTTP ( Layer 7 )",
					Method:      "HTTP",
					PanelMethod: "HTTP",
					ID:          3,
					Subnet:      0,
					Type:        "Free",
				})
				s.Methods = append(s.Methods, &method{
					Description: "HTTP-BETA ( Layer 7 )",
					Method:      "HTTP-BETA",
					PanelMethod: "HTTP-BETA",
					ID:          2,
					Subnet:      0,
					Type:        "Free",
				})
				s.Methods = append(s.Methods, &method{
					Description: "HTTP-NOVIP ( Layer 7 )",
					Method:      "HTTP-NOVIP",
					PanelMethod: "HTTP-NOVIP",
					ID:          3,
					Subnet:      0,
					Type:        "Free",
				})

			}
			json.NewEncoder(w).Encode(s)
			return
		} else {
			w.Write([]byte("404 page not found, contact @Royaloakap"))
			w.WriteHeader(404)
		}
	}))
}
