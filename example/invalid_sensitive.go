package example

import (
	"log"
)

type invalidSensitiveVoid struct {
	password string
}

func invalidSensitive() {
	void := invalidSensitiveVoid{
		password: "password",
	}
	log.Printf("info - %v", void.password)
}
