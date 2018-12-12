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

package database

import (
	"fmt"
	"sync"
	"github.com/IBM/ubiquity/utils/logs"
)

var migrations = new(sync.Map)

func RegisterMigration(obj interface{}) {
	defer logs.GetLogger().Trace(logs.DEBUG)()
	migrations.Store(fmt.Sprintf("%v", obj), obj)
}

func UnregisterAllMigrations() {
	defer logs.GetLogger().Trace(logs.DEBUG)()
	migrations = new(sync.Map)
}

func doMigrations(connection Connection) error {
	defer logs.GetLogger().Trace(logs.DEBUG)()

	logger := logs.GetLogger()

	migrations.Range(func(k, v interface{}) bool {
		logger.Info("migrating", logs.Args{{"migration", k}})
		if err := connection.GetDb().AutoMigrate(v).Error; err != nil {
			logger.ErrorRet(err, "failed")
		}
		migrations.Delete(k)
		return true
	})

	return nil
}
