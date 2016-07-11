package model


type ServiceBinding struct {
	Id                string                 `json:"id"`
	ServiceId         string                 `json:"service_id"`
	AppId             string                 `json:"app_id"`
	ServicePlanId     string                 `json:"service_plan_id"`
	PrivateKey        string                 `json:"private_key"`
	ServiceInstanceId string                 `json:"service_instance_id"`
	Parameters        map[string]interface{} `json:"parameters"`
}
type DummyCredentials struct {
	URI string `json:"uri"`
}

type CreateServiceBindingResponse struct {
	VolumeMounts []VolumeMount    `json:"volume_mounts"`
	Credentials  DummyCredentials `json:"credentials"`
}

type VolumeMount struct {
	ContainerPath string                    `json:"container_path"`
	Mode          string                    `json:"mode"`
	Private       VolumeMountPrivateDetails `json:"private"`
}

type VolumeMountPrivateDetails struct {
	Driver  string                `json:"driver"`
	GroupId string                `json:"group_id"`
	Config  string 		      `json:"config"`
}
