package scbe

import (
	"fmt"
	"github.com/IBM/ubiquity/resources"
)

type serviceDoesntExistError struct {
	volName     string
	serviceName string
	scbeName    string
}

func (e *serviceDoesntExistError) Error() string {
	return fmt.Sprintf("Cannot create volume [%s] on service [%s]. Reason : Service does not exist or not delegated to the Ubiquity interface in [%s]",
		e.volName, e.serviceName, e.scbeName)
}

type mappingResponseError struct {
	mapping ScbeResponseMappings
}

func (e *mappingResponseError) Error() string {
	return fmt.Sprintf("Mapping operation succeed but response is missing the mapping details. %#v", e.mapping)
}

type volumeNotFoundError struct {
	volName string
}

func (e *volumeNotFoundError) Error() string {
	return fmt.Sprintf("Volume [%s] was not found", e.volName)
}

type hostNotFoundvolumeNotFoundError struct {
	volName  string
	arrName  string
	hostName string
}

func (e *hostNotFoundvolumeNotFoundError) Error() string {
	return fmt.Sprintf("Host name [%s] was not found on the storage system [%s] that related to volume with WWN [%s]",
		e.hostName, e.arrName, e.volName)
}

type activateDefaultServiceError struct {
	serviceName string
	scbeName    string
}

func (e *activateDefaultServiceError) Error() string {
	return fmt.Sprintf("Error in activate SCBE backend. The default service %s does not exist in SCBE %s",
		e.serviceName, e.scbeName)
}

type volAlreadyExistsError struct {
	volName string
}

func (e *volAlreadyExistsError) Error() string {
	return fmt.Sprintf("Volume [%s] already exists.", e.volName)
}

type provisionParamMissingError struct {
	volName string
	param   string
}

func (e *provisionParamMissingError) Error() string {
	return fmt.Sprintf("Fail to provision a volume [%s] because the [%s] option is missing", e.volName, e.param)
}

type provisionParamIsNotNumberError struct {
	volName string
	param   string
}

func (e *provisionParamIsNotNumberError) Error() string {
	return fmt.Sprintf("Fail to provision a volume [%s] because the [%s] option is not a number", e.volName, e.param)
}

type volAlreadyAttachedError struct {
	volName  string
	hostName string
}

func (e *volAlreadyAttachedError) Error() string {
	return fmt.Sprintf("Volume [%s] already attached to [%s]", e.volName, e.hostName)
}

type volNotAttachedError struct {
	volName string
}

func (e *volNotAttachedError) Error() string {
	return fmt.Sprintf("Volume [%s] not attached", e.volName)
}

type ConfigDefaultSizeNotNumError struct {
	size string
}

func (e *ConfigDefaultSizeNotNumError) Error() string {
	return fmt.Sprintf("Error in config file. The parameter [%s] must be a number",
		"ScbeConfig.DefaultVolumeSize")
}

type ConfigScbeUbiquityInstanceNameWrongSize struct {
}

func (e *ConfigScbeUbiquityInstanceNameWrongSize) Error() string {
	return fmt.Sprintf("Error in config file. The parameter [%s] max length is [%d]",
		"ScbeConfig.UbiquityInstanceName", resources.UbiquityInstanceNameMaxSize)
}