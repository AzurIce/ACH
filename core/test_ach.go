package core

import (
	"ach/utils"
)

// TestRouter ...
func (ach *ACHCore) TestRouter() {
	ach.router.Run(":8888")
}

func (ach *ACHCore) TestRun() {
	// utils.TestLogin()
	// utils.TestGetToken()
	// utils.TestGetXboxLiveToken()
	// utils.TestGetXSTSToken()
	// utils.TestGetMCToken()
	utils.TestGetPlayerInfo()
}