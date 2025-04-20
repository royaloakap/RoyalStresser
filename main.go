package main

import (
	"api/core"
	"api/core/database"
	"api/core/master"
	"api/core/models"
	"api/core/models/ranks"
	"api/core/models/servers"
	"api/core/net"
	"api/core/net/commands"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/janeczku/go-spinner"
)

func main() {
	s := spinner.StartNew("Initializing")
	log.Println("Royal Projets BEST LEAKS IN COM (:")
	core.Initialize()

	// Initialize the database
	if err := database.New(); err != nil {
		log.Println("failed to initialize database", err)
		return
	}

	// Create a new user in the database
	database.Container.NewUser(&database.User{
		ID:         0,
		Username:   "royal",
		Key:        []byte("floconparadise11"),
		Membership: "admin",
		Ranks: []*ranks.Rank{
			ranks.GetRole("admin", true),
			ranks.GetRole("vip", true),
			ranks.GetRole("api", true),
			ranks.GetRole("cnc", true),
		},
		Concurrents: 10,
		Duration:    500,
		Servers:     10,
		Balance:     10000,
		Expiry:      -1,
	})
	s.Stop()
	// If server configurations are enabled
	if models.Config.Server.Enabled {
		// Start necessary routines
		go net.Listener()   // Start net listener
		go servers.Listen() // Start server listener
		clearScreen()
		time.Sleep(5 * time.Millisecond)
		go commands.Init() // Initialize commands
		time.Sleep(5 * time.Millisecond)
		s := spinner.StartNew("running")
		master.NewV2() // Initialize webhandler
		s.Stop()
	} else {
		// Print message indicating CnC turned off
		fmt.Printf("[main] %s main.go CnC Turned Off!\n", time.Now().Format("15:04:05"))

		// Start server listener
		go servers.Listen()
		clearScreen()
		s := spinner.StartNew("running")
		master.NewV2() // Initialize webhandler
		s.Stop()
	}
}

// clear screen
func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
