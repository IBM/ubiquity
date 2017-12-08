/**
 * Copyright 2017 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	return fmt.Sprintf("Cannot create volume [%s] on SCBE service [%s]. The service does not exist or it is not delegated to the %s interface in [%s]",
		e.volName, e.serviceName, resources.ScbeInterfaceName, e.scbeName)
}

type mappingResponseError struct {
	mapping ScbeResponseMappings
}

func (e *mappingResponseError) Error() string {
	return fmt.Sprintf("Mapping operation succeeded, but the mapping details are missing from the response. %#v", e.mapping)
}

type volumeNotFoundError struct {
	volName string
}

func (e *volumeNotFoundError) Error() string {
	return fmt.Sprintf("Volume [%s] was not found", e.volName)
}

type hostNotFoundvolumeNotFoundError struct {
	volName   string
	arrayName string
	hostName  string
}

func (e *hostNotFoundvolumeNotFoundError) Error() string {
	return fmt.Sprintf("Host name [%s], related to the volume with WWN [%s] was not found on the storage system [%s], according to SCBE caching data",
		e.hostName, e.volName, e.arrayName)
}

type activateDefaultServiceError struct {
	serviceName string
	scbeName    string
}

func (e *activateDefaultServiceError) Error() string {
	return fmt.Sprintf("SCBE backend activation error. The default service [%s] does not exist on SCBE [%s]",
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
	return fmt.Sprintf("Volume [%s] provisioning failure due to the missing [%s] option", e.volName, e.param)
}

type FsTypeNotSupportedError struct {
	volName        string
	wrongFStype    string
	supportedTypes string
}

func (e *FsTypeNotSupportedError) Error() string {
	return fmt.Sprintf("Volume [%s] provisioning failure due to mismatch in file system type [%s] [%s]",
		e.volName, e.supportedTypes, e.wrongFStype)
}

type provisionParamIsNotNumberError struct {
	volName string
	param   string
}

func (e *provisionParamIsNotNumberError) Error() string {
	return fmt.Sprintf("Volume [%s] provisioning failure due to a non-numeric value set for option [%s]",
		e.volName, e.param)
}

type volAlreadyAttachedError struct {
	volName  string
	hostName string
}

func (e *volAlreadyAttachedError) Error() string {
	return fmt.Sprintf("Volume [%s] already attached to [%s]", e.volName, e.hostName)
}

type CannotDeleteVolWhichAttachedToHostError struct {
	volName  string
	hostName string
}

func (e *CannotDeleteVolWhichAttachedToHostError) Error() string {
	return fmt.Sprintf("Volume [%s] deletion failure. Volume is currently attached to host [%s]",
		e.volName, e.hostName)
}

type volNotAttachedError struct {
	volName string
}

func (e *volNotAttachedError) Error() string {
	return fmt.Sprintf("Volume [%s] is not attached to any host", e.volName)
}

type ConfigDefaultSizeNotNumError struct {
	size string
}

func (e *ConfigDefaultSizeNotNumError) Error() string {
	return fmt.Sprintf("Configuration file error. The [%s] parameter value must be a number",
		"ScbeConfig.DefaultVolumeSize")
}

type ConfigDefaultFilesystemTypeNotSupported struct {
	wrongFStype    string
	supportedTypes string
}

func (e *ConfigDefaultFilesystemTypeNotSupported) Error() string {
	return fmt.Sprintf("Configuration file error. The allowed values for parameter [%s] must be [%s] (given [%s])",
		"ScbeConfig.DefaultFileSystemType", e.supportedTypes, e.wrongFStype)
}

type ConfigScbeUbiquityInstanceNameWrongSize struct {
}

func (e *ConfigScbeUbiquityInstanceNameWrongSize) Error() string {
	return fmt.Sprintf("Configuration file error. The max length of the parameter [%s] value is [%d]",
		"ScbeConfig.UbiquityInstanceName", resources.UbiquityInstanceNameMaxSize)
}

type VolumeNameExceededMaxLengthError struct {
	volName         string
	maxVolumeLength int
}

func (e *VolumeNameExceededMaxLengthError) Error() string {
	return fmt.Sprintf("The volume name [%s] length [%d] is out of range. The max name length is [%d] characters",
		e.volName, len(e.volName), e.maxVolumeLength)
}

type InValidRequestError struct {
	requestType       string
	badParam          string
	paramCurrentValue string
	paramExpectedToBe string
}

func (e *InValidRequestError) Error() string {
	return fmt.Sprintf(
		"Request type [%s] is invalid. The expected value of parameter [%s] [%s] is [%s]",
		e.requestType,
		e.badParam,
		e.paramCurrentValue,
		e.paramExpectedToBe)
}

type SslModeValueInvalid struct {
	sslModeInValid string
}

func (e *SslModeValueInvalid) Error() string {
	return fmt.Sprintf("Illegal SSL mode value [%s]. The allowed values are [%s, %s]",
		e.sslModeInValid, resources.SslModeRequire, resources.SslModeVerifyFull)
}

type SslModeFullVerifyWithoutCAfile struct {
	VerifyCaEnvName string
}

func (e *SslModeFullVerifyWithoutCAfile) Error() string {
	return fmt.Sprintf("Environment variable [%s] must be set for the SSL mode [%s]",
		e.VerifyCaEnvName, resources.SslModeVerifyFull)
}

type InvalidMappingsForVolume struct {
	volWwn     string
}

func (e *InvalidMappingsForVolume) Error() string {
	return fmt.Sprintf("Invalid mappings for volume [%s]", e.volWwn)
}