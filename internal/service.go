package internal

import "github.com/sirupsen/logrus"

type Database interface {
}

type App struct {
	log *logrus.Logger
	db  Database
}

func NewApp(log *logrus.Logger, db Database) *App {
	return &App{
		log: log,
		db:  db,
	}
}
