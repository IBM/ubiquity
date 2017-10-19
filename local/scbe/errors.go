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
	volName   string
	arrayName string
	hostName  string
}

func (e *hostNotFoundvolumeNotFoundError) Error() string {
	return fmt.Sprintf("Host name [%s] was not found on the storage system [%s](according to SCBE caching) that related to volume with WWN [%s]",
		e.hostName, e.arrayName, e.volName)
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

type FsTypeNotSupportedError struct {
	volName        string
	wrongFStype    string
	supportedTypes string
}

func (e *FsTypeNotSupportedError) Error() string {
	return fmt.Sprintf("Fail to provision a volume [%s]. Supported filesystem types are [%s] (but given [%s])",
		e.volName, e.supportedTypes, e.wrongFStype)
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

type CannotDeleteVolWhichAttachedToHostError struct {
	volName  string
	hostName string
}

func (e *CannotDeleteVolWhichAttachedToHostError) Error() string {
	return fmt.Sprintf("Cannot delete a volume that is attached to a host. The volume [%s] currently attached to host [%s]", e.volName, e.hostName)
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

type ConfigDefaultFilesystemTypeNotSupported struct {
	wrongFStype    string
	supportedTypes string
}

func (e *ConfigDefaultFilesystemTypeNotSupported) Error() string {
	return fmt.Sprintf("Error in config file. The parameter [%s] can be the following values [%s] (given [%s])",
		"ScbeConfig.DefaultFileSystemType", e.supportedTypes, e.wrongFStype)
}

type ConfigScbeUbiquityInstanceNameWrongSize struct {
}

func (e *ConfigScbeUbiquityInstanceNameWrongSize) Error() string {
	return fmt.Sprintf("Error in config file. The parameter [%s] max length is [%d]",
		"ScbeConfig.UbiquityInstanceName", resources.UbiquityInstanceNameMaxSize)
}

type VolumeNameExceededMaxLengthError struct {
	volName         string
	maxVolumeLength int
}

func (e *VolumeNameExceededMaxLengthError) Error() string {
	return fmt.Sprintf("Volume name [%s] is too long(len=%d). Max length should be [%d]",
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
		"Request type [%s] not valid. Param [%s]=[%s] but expacted to be [%s]",
		e.requestType,
		e.badParam,
		e.paramCurrentValue,
		e.paramExpectedToBe)
}

type SslModeValueInvalid struct {
	sslModeInValid string
}

func (e *SslModeValueInvalid) Error() string {
	return fmt.Sprintf("SSL Mode [%s] is invalid. Available values are [%s, %s]",
		e.sslModeInValid, resources.SslModeRequire, resources.SslModeVerifyFull)
}

type SslModeFullVerifyWithoutCAfile struct {
	VerifyCaEnvName string
}

func (e *SslModeFullVerifyWithoutCAfile) Error() string {
	return fmt.Sprintf("ENV [%s] is missing. Must set in case of SSL mode [%s]",
		e.VerifyCaEnvName, resources.SslModeVerifyFull)
}
