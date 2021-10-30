package main

import log "github.com/sirupsen/logrus"

type UserContext struct {
	current *User
}

func NewUserContext() *UserContext {
	return &UserContext{}
}

func (um *UserContext) Reset() {
	um.current = nil
}

func (um *UserContext) Set(user *User) {
	um.current = user
	log.WithField("uuid", user.UUID).Info("set current user")
}

func (um *UserContext) Get() *User {
	return um.current
}

func (um *UserContext) GetUUID() string {
	if um.current == nil {
		log.Warn("no current user")
		return PM5_USER_UUID
	}

	return um.current.UUID
}
