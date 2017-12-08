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
    "github.com/IBM/ubiquity/utils/logs"
)


var migrations = make(map[string]interface{})

func RegisterMigration(obj interface{}) {
    defer logs.GetLogger().Trace(logs.DEBUG)()

    migrations[fmt.Sprintf("%v", obj)] = obj
}

func UnregisterAllMigrations() {
    defer logs.GetLogger().Trace(logs.DEBUG)()

    migrations = make(map[string]interface{})
}

func doMigrations(connection Connection) error {
    defer logs.GetLogger().Trace(logs.DEBUG)()

    logger := logs.GetLogger()
    for k, v := range migrations {
        logger.Info("migrating", logs.Args{{"migration", k}})
        if err := connection.GetDb().AutoMigrate(v).Error; err != nil {
            logger.ErrorRet(err, "failed")
        }
        delete(migrations, k)
    }
    return nil
}
