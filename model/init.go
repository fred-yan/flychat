package model

import "flychat/platform"

func InstallDB() {
	db := platform.DB
	if err := db.AutoMigrate(
		&User{},
		&Message{}); err != nil {
		panic(err)
	}
}
