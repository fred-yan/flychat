package model

import "flychat/platform"

func InstallDB() {
	db := platform.DB
	if err := db.AutoMigrate(
		&User{},
		&Message{},
		&Story{}); err != nil {
		panic(err)
	}
}
