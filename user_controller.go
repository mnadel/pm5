package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type UserController struct {
	db     *Database
	config *Configuration
}

func NewUserController(config *Configuration, db *Database) *UserController {
	return &UserController{
		db:     db,
		config: config,
	}
}

func (uc *UserController) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	user := CurrentUser.Get()
	if user == nil {
		http.Error(w, "no current user", http.StatusForbidden)
		return
	}

	fmt.Fprintf(w, "%s\n", user.UUID)
}

func (uc *UserController) HandleSetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	CurrentUser.Reset()

	vars := mux.Vars(r)
	uuid := vars["uuid"]

	user, err := uc.db.GetUser(uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, fmt.Sprintf("not found: %s", uuid), http.StatusForbidden)
		return
	}

	CurrentUser.Set(user)

	fmt.Fprint(w, "👍\n")
}
