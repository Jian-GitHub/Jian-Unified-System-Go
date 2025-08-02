package types

type WebAuthnConf struct {
	RPID          string   `json:",optional"`
	RPDisplayName string   `json:",optional"`
	RPOrigins     []string `json:",optional"`
}
