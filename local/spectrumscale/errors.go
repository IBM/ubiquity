/**
 * Copyright 2018 IBM Corp.
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

package spectrumscale

import (
	"fmt"
)

type SpectrumScaleConfigError struct {
    ConfigParam  string
}

func (e *SpectrumScaleConfigError) Error() string {
    return fmt.Sprintf("missing config %s", e.ConfigParam)
}

const SpectrumScaleGetClusterIdErrorStr = "Failed to get Spectrum Scale ClusterID"

type SpectrumScaleGetClusterIdError struct  {
	ErrorMsg string
}

func (e *SpectrumScaleGetClusterIdError) Error() string {
		return fmt.Sprintf(SpectrumScaleGetClusterIdErrorStr)
}

type SpectrumScaleFileSystemNotPresent struct {
	Filesystem string
}

func (e *SpectrumScaleFileSystemNotPresent) Error() string {
	return fmt.Sprintf("SpectrumScale filesystem [%s] does not exist", e.Filesystem)
}

type SpectrumScaleFileSystemNotMounted struct {
	Filesystem string
}

func (e *SpectrumScaleFileSystemNotMounted) Error() string {
	return fmt.Sprintf("SpectrumScale filesystem [%s] is not mounted", e.Filesystem)
}

type SpectrumScaleQuotaNotEnabledError struct {
	Filesystem string
}

func (e *SpectrumScaleQuotaNotEnabledError) Error() string {
	return fmt.Sprintf("Quota not enabled for Filesystem [%s]", e.Filesystem)
}

type SpectrumScaleFileSystemMountError struct {
	Filesystem string
}

func (e *SpectrumScaleFileSystemMountError) Error() string {
	return fmt.Sprintf("Failed to check if Filesystem [%s] is mounted", e.Filesystem)
}
