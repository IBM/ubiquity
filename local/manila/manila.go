package manila

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.ibm.com/almaden-containers/ubiquity/resources"

	"github.com/openstack/golang-client/openstack"
)

type ShareIds map[string]string

type ManilaBackend struct {
	logger        *log.Logger
	nfsClientCIDR string
	configPath    string
	osConfig      OpenstackConfig
	osAuth        authContainer
	shareIds      ShareIds
}

type OpenstackConfig struct {
	AuthUrl   string
	ManilaUrl string // TODO: Get Manila endpoint URL automatically from Keystone service
	ProjectId string
	UserId    string
	Password  string
}

func NewManilaBackend(logger *log.Logger, osConfig OpenstackConfig, configPath, nfsClientCIDR string) *ManilaBackend {
	osAuth := getKeystoneV3Auth(osConfig)
	manilaBackend := ManilaBackend{logger: logger, osConfig: osConfig, osAuth: osAuth, configPath: configPath, nfsClientCIDR: nfsClientCIDR}
	manilaBackend.retrieveShareIds()
	return &manilaBackend
}

func (m *ManilaBackend) GetServices() []resources.Service {
	plan1 := resources.ServicePlan{
		Name:        "bronze",
		Id:          "manila-bronze-plan-1",
		Description: "Default Plan for Manila/NFS Storage Service providing 1GiB of shared storage",
		Metadata:    nil,
		Free:        true,
	}

	plan2 := resources.ServicePlan{
		Name:        "silver",
		Id:          "manila-silver-plan-10",
		Description: "Default Plan for Manila/NFS Storage Service providing 10GiB of shared storage",
		Metadata:    nil,
		Free:        true,
	}

	plan3 := resources.ServicePlan{
		Name:        "gold",
		Id:          "manila-gold-plan-100",
		Description: "Default Plan for Manila/NFS Storage Service providing 100GiB of shared storage",
		Metadata:    nil,
		Free:        true,
	}

	service := resources.Service{
		Name:            "manila-nfs",
		Id:              "manila-nfs-guid",
		Description:     "Provides the Manila/NFS volume service",
		Bindable:        true,
		PlanUpdateable:  false,
		Tags:            []string{"manila", "nfs"},
		Requires:        []string{"volume_mount"},
		Metadata:        nil,
		Plans:           []resources.ServicePlan{plan1, plan2, plan3},
		DashboardClient: nil,
	}

	return []resources.Service{service}
}

func (m *ManilaBackend) CreateVolume(serviceInstance resources.ServiceInstance, name string, opts map[string]interface{}) error {
	m.logger.Println("ManilaBackend: CreateVolume start")
	defer m.logger.Println("ManilaBackend: CreateVolume end")

	if _, ok := m.shareIds[name]; ok == true {
		msg := fmt.Sprintf("ManilaBackend: CreateVolume: Volume %s already exists", name)
		m.logger.Println(msg)
		return fmt.Errorf(msg)
	}

	// FIXME: clean up usage of planId for getting plan information
	planIdSplit := strings.Split(serviceInstance.PlanId, "-")
	planSize := planIdSplit[len(planIdSplit)-1]
	size, _ := strconv.Atoi(planSize)

	manilaShare, err := m.manilaCreateShare(name, size)
	if err != nil {
		return err
	}
	if err := m.manilaAllowAccess(manilaShare.Id); err != nil {
		return err
	}
	fmt.Sprintf("ManilaBackend: CreateVolume: Created volume %s. Client CIDR: %s", name, m.nfsClientCIDR)
	m.shareIds[name] = manilaShare.Id
	return m.persistShareIds()
}

func (m *ManilaBackend) RemoveVolume(serviceInstance resources.ServiceInstance, name string) error {
	m.logger.Println("ManilaBackend: RemoveVolume start")
	defer m.logger.Println("ManilaBackend: RemoveVolume end")

	shareId, err := m.getShareId(name)
	if err != nil {
		return err
	}

	if err := m.manilaDeleteShare(*shareId); err != nil {
		return err
	}
	delete(m.shareIds, name)
	return m.persistShareIds()
}

func (m *ManilaBackend) ListVolumes(serviceInstance resources.ServiceInstance) ([]resources.VolumeMetadata, error) {
	m.logger.Println("ManilaBackend: ListVolumes start")
	defer m.logger.Println("ManilaBackend: ListVolumes end")

	manilaShares, err := m.manilaGetShares()
	if err != nil {
		return []resources.VolumeMetadata{}, err
	}

	volumeMetaData := make([]resources.VolumeMetadata, len(manilaShares))
	for i, manilaShare := range manilaShares {
		volumeMetaData[i] = resources.VolumeMetadata{
			Name:       manilaShare.Name,
			Mountpoint: "N/A",
		}
	}
	return volumeMetaData, nil
}

func (m *ManilaBackend) GetVolume(serviceInstance resources.ServiceInstance, name string) (volumeMetadata *resources.VolumeMetadata, clientDriverName *string, config *map[string]interface{}, err error) {
	m.logger.Println("ManilaBackend: GetVolume start")
	defer m.logger.Println("ManilaBackend: GetVolume end")

	clientDriver := "nfs-plugin"

	shareId, err := m.getShareId(name)
	if err != nil {
		return nil, nil, nil, err
	}

	exportLocations, err := m.manilaGetShareExportLocations(*shareId)
	if err != nil {
		return nil, nil, nil, err
	}

	volumeMetadata = &resources.VolumeMetadata{
		Name:       name,
		Mountpoint: "N/A",
	}

	configMap := make(map[string]interface{})
	nfsShare := exportLocations[0].Path
	configMap["nfs_share"] = nfsShare
	m.logger.Printf("Adding nfs_share %s to bind config for share %s (%s)\n", nfsShare, name, shareId)
	return volumeMetadata, &clientDriver, &configMap, nil
}

func (m *ManilaBackend) retrieveShareIds() error {
	m.logger.Println("ManilaBackend: retrievePersistedShareIds start")
	defer m.logger.Println("ManilaBackend: retrievePersistedShareIds end")
	shareIdsFile, err := os.Open(path.Join(m.configPath, ".ibm-storage-broker-manila-shares.json"))
	if err != nil {
		m.logger.Println(err.Error())
		if os.IsNotExist(err) == true {
			m.logger.Println("file does not exist")
			m.shareIds = ShareIds{}
			err = m.persistShareIds()
			if err != nil {
				return fmt.Errorf("error initializing config file (%s)", err.Error())
			}
			return nil
		} else {
			return fmt.Errorf("error opening config file (%s)", err.Error())
		}
	}
	jsonParser := json.NewDecoder(shareIdsFile)
	var shareIds ShareIds
	if err = jsonParser.Decode(&shareIds); err != nil {
		return fmt.Errorf("error parsing config file (%s)", err.Error())
	}
	m.shareIds = shareIds
	return nil
}

func (m *ManilaBackend) persistShareIds() error {
	m.logger.Println("ManilaBackend: persistShareIds start")
	defer m.logger.Println("ManilaBackend: persistShareIds end")
	data, err := json.Marshal(&m.shareIds)
	if err != nil {
		return fmt.Errorf("Error marshalling mapping config to file: %s", err.Error())
	}
	err = ioutil.WriteFile(path.Join(m.configPath, ".ibm-storage-broker-manila-shares.json"), data, 0644)
	if err != nil {
		return fmt.Errorf("Error writing json spec: %s", err.Error())
	}
	return nil
}

func (m *ManilaBackend) getShareId(shareName string) (*string, error) {
	shareId, ok := m.shareIds[shareName]
	if ok == false {
		errMsg := fmt.Sprintf("Could not retrieve share id for share %s", shareName)
		fmt.Printf("Error: %s\n", errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	return &shareId, nil
}

type authContainer struct {
	Auth authRequestContainer `json:"auth"`
}

type authRequestContainer struct {
	Identity identityContainer `json:"identity"`
	Scope    scopeContainer    `json:"scope,omitempty"`
}

type identityContainer struct {
	Methods  []string      `json:"methods"`
	Password passwordCreds `json:"password"`
}

type passwordCreds struct {
	User userCreds `json:"user"`
}

type userCreds struct {
	Id       string `json:"id"`
	Password string `json:"password"`
}

type scopeContainer struct {
	Project userProject `json:"project,omitempty"`
}

type userProject struct {
	Id string `json:"id,omitempty"`
}

type authRef struct {
	Token     string
	Endpoint  string
	ProjectId string
}

func (ar authRef) GetToken() string {
	return ar.Token
}
func (ar authRef) GetExpiration() time.Time {
	// TODO: Get real expiration time of token and re-auth if necessary
	return time.Date(2099, 1, 1, 1, 1, 1, 1, time.UTC)
}
func (ar authRef) GetEndpoint(string, string) (string, error) {
	return ar.Endpoint, nil
}
func (ar authRef) GetProject() string {
	return ar.ProjectId
}

type manilaShare struct {
	Status string `json:"status,omitempty"`
	Links  []struct {
		Href string `json:"href,omitempty"`
		Rel  string `json:"rel,omitempty"`
	} `json:"links,omitempty"`
	ExportLocation     string      `json:"export_location,omitempty"`
	Name               string      `json:"name,omitempty"`
	ShareType          interface{} `json:"share_type"`
	AvailabilityZone   interface{} `json:"availability_zone"`
	ConsistencyGroupId interface{} `json:"consistency_group_id"`
	CreatedAt          string      `json:"created_at,omitempty"`
	Description        interface{} `json:"description"`
	ExportLocations    []string    `json:"export_locations,omitempty"`
	ShareProto         string      `json:"share_proto"`
	ShareNetworkID     interface{} `json:"share_network_id"`
	Metadata           interface{} `json:"metadata"`
	ShareServerID      string      `json:"share_server_id,omitempty"`
	Host               string      `json:"host,omitempty"`
	VolumeType         string      `json:"volume_type,omitempty"`
	SnapshotID         interface{} `json:"snapshot_id"`
	IsPublic           bool        `json:"is_public"`
	ProjectID          string      `json:"project_id,omitempty"`
	Id                 string      `json:"id,omitempty"`
	Size               int         `json:"size,omitempty"`
}

type manilaCreateShareRequest struct {
	Share manilaShare `json:"share"`
}

type manilaCreateShareResponse struct {
	Share manilaShare `json:"share"`
}

type manilaShowShareResponse struct {
	Share manilaShare `json:"share"`
}

type manilaGetSharesResponse struct {
	Shares []manilaShare `json:"shares"`
}

type manilaAccess struct {
	ShareID     string `json:"share_id,omitempty"`
	AccessType  string `json:"access_type"`
	AccessTo    string `json:"access_to"`
	AccessLevel string `json:"access_level"`
	State       string `json:"state,omitempty"`
	Id          string `json:"id,omitempty"`
}

type manilaAllowAccessRequest struct {
	Access manilaAccess `json:"allow_access"`
}

type manilaAllowAccessResponse struct {
	Access manilaAccess `json:"access"`
}

type manilaExportLocation struct {
	Path            string `json:"path"`
	ShareInstanceID string `json:"share_instance_id"`
	IsAdminOnly     bool   `json:"is_admin_only"`
	Id              string `json:"id"`
	Preferred       bool   `json:"preferred"`
}

type manilaGetExportLocationsResponse struct {
	ExportLocations []manilaExportLocation `json:"export_locations"`
}

func (m *ManilaBackend) getManilaSession() (*openstack.Session, error) {

	resp, err := openstack.PostJSON(m.osConfig.AuthUrl+"/tokens", nil, nil, &m.osAuth, nil)
	if err != nil {
		m.logger.Printf("ManilaBackend: Auth error: %s", err.Error())
	}

	token := resp.Header.Get("X-Subject-Token")

	session, err := openstack.NewSession(nil, authRef{Token: token, Endpoint: m.osConfig.ManilaUrl, ProjectId: m.osConfig.ProjectId}, nil)
	if err != nil {
		return nil, fmt.Errorf("ManilaBackend: Could not create Openstack session: %s", err.Error())
	}
	session.Headers.Add("X-Openstack-Manila-Api-Version", "2.15")
	session.Headers.Add("User-Agent", "ibm-storage-broker")
	return session, nil
}

func (m *ManilaBackend) doManilaRequest(method, relativeUrl string, requestBody, responseContainer interface{}) (*http.Response, error) {
	session, err := m.getManilaSession()
	if err != nil {
		return nil, fmt.Errorf("ManilaBackend: Could not get manila session: %s", err.Error())
	}
	baseUrl, _ := session.AuthToken.GetEndpoint("", "")
	var resp *http.Response
	if method != "DELETE" {
		if method == "POST" {
			session.Headers.Add("Content-Type", "application/json")
		}
		resp, err = session.RequestJSON(method, baseUrl+relativeUrl, nil, &session.Headers, &requestBody, &responseContainer)
	} else {
		resp, err = session.Request(method, baseUrl+relativeUrl, nil, &session.Headers, &[]byte{})
	}
	if err != nil {
		return nil, fmt.Errorf("ManilaBackend: Could not retrieve %s %s: %s", method, baseUrl+relativeUrl, err)
	}
	return resp, nil
}

func (m *ManilaBackend) manilaCreateShare(name string, size int) (*manilaShare, error) {
	manilaShare := manilaShare{
		Name:       name,
		Size:       size,
		ShareProto: "nfs",
		Metadata:   map[string]string{},
	}
	manilaCreateShareRequest := manilaCreateShareRequest{Share: manilaShare}
	var shareDetailsContainer manilaCreateShareResponse
	_, err := m.doManilaRequest("POST", "/shares", &manilaCreateShareRequest, &shareDetailsContainer)
	if err != nil {
		return nil, fmt.Errorf("ManilaBackend: Could not create share with name %s: %s", name, err.Error())
	}

	// Block and delay here, otherwise allowing access to share will fail. Test setup required a minimum of around 900ms.
	// TODO: Figure out a way to check whether share has been successfully created and block only so long
	time.Sleep(2700 * time.Millisecond) // 900ms * 3 = 2700ms

	return &shareDetailsContainer.Share, err
}

func (m *ManilaBackend) manilaAllowAccess(shareId string) error {
	manilaAccess := manilaAccess{
		AccessLevel: "rw",
		AccessType:  "ip",
		AccessTo:    m.nfsClientCIDR,
	}
	manilaAllowAccessRequest := manilaAllowAccessRequest{Access: manilaAccess}
	var manilaAllowAccessResponse manilaAllowAccessResponse
	_, err := m.doManilaRequest("POST", "/shares/"+shareId+"/action", &manilaAllowAccessRequest, &manilaAllowAccessResponse)
	if err != nil {
		return fmt.Errorf("ManilaBackend: Could not allow access to share %s: %s", shareId, err.Error())
	}
	return err
}

func (m *ManilaBackend) manilaShowShare(shareId string) (*manilaShare, error) {
	var shareDetailsContainer manilaShowShareResponse
	_, err := m.doManilaRequest("GET", "/shares/"+shareId, nil, &shareDetailsContainer)
	if err != nil {
		return nil, fmt.Errorf("ManilaBackend: Could not show share %s: %s", shareId, err.Error())
	}
	return &shareDetailsContainer.Share, err
}

func (m *ManilaBackend) manilaGetShareExportLocations(shareId string) ([]manilaExportLocation, error) {
	var exportLocationsContainer manilaGetExportLocationsResponse
	_, err := m.doManilaRequest("GET", "/shares/"+shareId+"/export_locations", nil, &exportLocationsContainer)
	if err != nil {
		return nil, fmt.Errorf("ManilaBackend: Could not get share %s export locations: %s", shareId, err.Error())
	}
	return exportLocationsContainer.ExportLocations, err
}

func (m *ManilaBackend) manilaDeleteShare(shareId string) error {
	_, err := m.doManilaRequest("DELETE", "/shares/"+shareId, nil, nil)
	if err != nil {
		return fmt.Errorf("ManilaBackend: Could not delete share %s: %s", shareId, err.Error())
	}
	return err
}

func (m *ManilaBackend) manilaGetShares() ([]manilaShare, error) {
	var sharesListContainer manilaGetSharesResponse
	_, err := m.doManilaRequest("GET", "/shares/detail", nil, &sharesListContainer)
	if err != nil {
		return nil, fmt.Errorf("ManilaBackend: Could not get shares: %s", err.Error())
	}
	return sharesListContainer.Shares, nil
}

func getKeystoneV3Auth(osConfig OpenstackConfig) authContainer {
	return authContainer{
		Auth: authRequestContainer{
			Identity: identityContainer{
				Methods: []string{"password"},
				Password: passwordCreds{
					User: userCreds{
						Id:       osConfig.UserId,
						Password: osConfig.Password,
					},
				},
			},
			Scope: scopeContainer{
				Project: userProject{
					Id: osConfig.ProjectId,
				},
			},
		},
	}

}
