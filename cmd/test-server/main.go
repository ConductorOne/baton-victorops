package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

const (
	testAPIId  = "test-api-id"
	testAPIKey = "test-api-key"
	listenAddr = ":8765"
)

type Server struct {
	state *State
}

func main() {
	state := NewState()
	server := &Server{state: state}
	mux := http.NewServeMux()

	// Doc URL: https://portal.victorops.com/public/api-docs.html#!/Users/get_api_public_v1_user
	mux.HandleFunc("GET /api-public/v1/user", server.requireAuth(server.handleListUsers))

	// Doc URL: https://portal.victorops.com/public/api-docs.html#!/Users/get_api_public_v1_user_user
	mux.HandleFunc("GET /api-public/v1/user/{username}", server.requireAuth(server.handleGetUser))

	// Doc URL: https://portal.victorops.com/public/api-docs.html#!/Teams/get_api_public_v1_team
	mux.HandleFunc("GET /api-public/v1/team", server.requireAuth(server.handleListTeams))

	// Doc URL: https://portal.victorops.com/public/api-docs.html#!/Teams/get_api_public_v1_team_team_members
	mux.HandleFunc("GET /api-public/v1/team/{slug}/members", server.requireAuth(server.handleListTeamMembers))

	// Doc URL: https://portal.victorops.com/public/api-docs.html#!/Teams/post_api_public_v1_team_team_members
	mux.HandleFunc("POST /api-public/v1/team/{slug}/members", server.requireAuth(server.handleAddTeamMember))

	// Doc URL: https://portal.victorops.com/public/api-docs.html#!/Teams/delete_api_public_v1_team_team_members_user
	mux.HandleFunc("DELETE /api-public/v1/team/{slug}/members/{username}", server.requireAuth(server.handleRemoveTeamMember))

	// Doc URL: https://portal.victorops.com/public/api-docs.html#!/Teams/get_api_public_v1_team_team_admins
	mux.HandleFunc("GET /api-public/v1/team/{slug}/admins", server.requireAuth(server.handleListTeamAdmins))

	// Doc URL: https://portal.victorops.com/public/api-docs.html#!/On45call/get_api_public_v1_oncall_current
	mux.HandleFunc("GET /api-public/v1/oncall/current", server.requireAuth(server.handleOnCallCurrent))

	log.Printf("baton-victorops test server listening on %s", listenAddr)
	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
