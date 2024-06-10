package main

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func openDB(p string) gorm.Dialector {
	return sqlite.Open(p)
}
