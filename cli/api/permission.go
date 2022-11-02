package api

import (
	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/filecoin-project/venus-auth/core"
)

func PermissionedAPI(api LocalAPI) LocalAPI {
	var out LocalAPIStruct
	auth.PermissionedProxy(core.PermArr, []auth.Permission{}, api, &out.Internal)
	return &out
}
