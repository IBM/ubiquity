package scbe

// go standard for all the structures in the project.
const (
	ErrorMarshallingCredentialInfo = "Internal error marshalling credentialInfo"
)

type CredentialInfo struct {
	userName string `json:"username"`
	password string `json:"password"`
}
type ConnectionInfo struct {
	credentialInfo CredentialInfo
	port           string
	managementIP   string
	verifySSL      bool
}

type LoginResponse struct {
	Token string `json:"token"`
}
