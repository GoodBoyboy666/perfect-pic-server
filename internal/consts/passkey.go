package consts

import "time"

type PasskeySessionType string

const (
	PasskeySessionRegistration PasskeySessionType = "registration"
	PasskeySessionLogin        PasskeySessionType = "login"
	PasskeySessionTTL                             = 5 * time.Minute
	MaxUserPasskeyCount                           = 10
	PasskeyNameMaxRunes                           = 64
)
