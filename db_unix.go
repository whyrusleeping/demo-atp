package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openDB(p string) gorm.Dialector {
	return sqlite.Open(p)
}
