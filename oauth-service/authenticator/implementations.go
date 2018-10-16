package authenticator

import "github.com/darren-west/app/oauth-service/auth"

// Map is used to store Authenticator implementation. Use domain of oauth server to avoid clashes.
var Map = map[string]auth.Authenticator{}
