package model

import "flychat/platform"

func InstallDB() {
	db := platform.DB
	if err := db.AutoMigrate(&User{}); err != nil {
		panic(err)
	}
}
