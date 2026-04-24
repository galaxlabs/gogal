package proxy

type TraefikStatus struct {
	Installed bool
	Enabled   bool
	Message   string
}

func CheckTraefikStatus() TraefikStatus {
	return TraefikStatus{Installed: false, Enabled: false, Message: "Traefik integration is optional and planned for later phase."}
}
