package scbe

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"bytes"

	"github.com/IBM/ubiquity/utils"
)

type ScbeRestClient interface {
	Login() (string, error)
}

type scbeRestClient struct {
	logger         *log.Logger
	baseURL        string
	authURL        string
	referrer       string
	connectionInfo ConnectionInfo
	httpClient     *http.Client
}

func NewScbeRestClient(logger *log.Logger, conInfo ConnectionInfo, baseURL string, authURL string, referrer string) (ScbeRestClient, error) {

	return &scbeRestClient{logger: logger, connectionInfo: conInfo, baseURL: baseURL, authURL: authURL, referrer: referrer, httpClient: &http.Client{}}, nil
}

func (s *scbeRestClient) Login() (string, error) {
	activateURL := utils.FormatURL(s.baseURL, s.authURL)

	payload, err := json.MarshalIndent(s.connectionInfo.credentialInfo, "", " ")
	if err != nil {

		s.logger.Printf(fmt.Sprintf(ErrorMarshallingCredentialInfo+"%s", err))
		return "", fmt.Errorf(ErrorMarshallingCredentialInfo)
	}

	request, err := http.NewRequest("POST", activateURL, bytes.NewBuffer(payload))
	if err != nil {
		s.logger.Printf("Error in creating request %#v", err)
		return "", fmt.Errorf("Error in creating request")
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("referer", s.referrer)
	request.Header.Del("Authorization")

	response, err := s.httpClient.Do(request)
	fmt.Printf("response %#v", response)
	if err != nil {
		s.logger.Printf("Error in executing remote call %#v", err)
		return "", fmt.Errorf("Error in executing remote call")
	}
	var loginResponse = LoginResponse{}
	err = utils.UnmarshalResponse(response, &loginResponse)
	if err != nil {
		s.logger.Printf("Error in unmarshalling response %#v", err)
		return "", fmt.Errorf("Error in unmarshalling response")
	}
	return loginResponse.Token, nil
}
