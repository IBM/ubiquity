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

package database_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"

	"github.com/IBM/ubiquity/database"
	"github.com/IBM/ubiquity/utils/logs"
	"os"
)

const (
	dbPath = "/tmp/database_test"
)

func TestDb(t *testing.T) {
	RegisterFailHandler(Fail)
	defer logs.InitStdoutLogger(logs.DEBUG)()
	os.Remove(dbPath)
	defer database.InitSqlite(dbPath)()
	RunSpecs(t, "database Test Suite")
}
