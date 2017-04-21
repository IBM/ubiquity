package scbe

// go standard for all the structures in the project.

type CredentialInfo struct {
	UserName string `json:"username"`
	Password string `json:"password"`
	Group    string `json:"group"`
}

type ConnectionInfo struct {
	CredentialInfo CredentialInfo
	Port           string
	ManagementIP   string
	VerifySSL      bool
}

type LoginResponse struct {
	Token string `json:"token"`
}
