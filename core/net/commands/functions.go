package commands

import (
	"api/core/models"
	"api/core/net/sessions"
	"fmt"
	"time"
)

func admin(session *sessions.Session, args []string) {
	fmt.Fprintf(session.Conn, "\033c")
	fmt.Fprintf(session.Conn, "+Welcome to %s, %s!\r\n", models.Config.Name, session.User.Username)
	fmt.Fprintf(session.Conn, "+--------------------------------------------+\r\n")
	for _, v := range Commands {
		fmt.Fprintf(session.Conn, "%-20s | %s\r\n", v.Name, v.Description)
	}
}
func credits(session *sessions.Session, args []string) {
	fmt.Fprintf(session.Conn, "\n\r")
}
func logout(session *sessions.Session, args []string) {
	fmt.Fprintf(session.Conn, "Have a good day!\n\r")
	time.Sleep(5 * time.Second)
	session.Conn.Close()
}
