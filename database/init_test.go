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
	"fmt"
	"github.com/IBM/ubiquity/database"
	"github.com/IBM/ubiquity/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"strings"
)

var _ = Describe("Init", func() {
	var (
		hostname       string = "my-hostname"
		defaultUser    string = "postgres"
		defaultDbName  string = "postgres"
		newUser        string = "my-user"
		newDbName      string = "my-dbname"
		newPassword    string = "my-password"
		newPort        string = "my-port"
		newTimeout     string = "my-timeout"
		defaultSslMode string = resources.DefaultDbSslMode
		newSslMode     string = "my-sslmode"
		newSslRootCert string = "my-sslrootcert"
	)
	BeforeEach(func() {
	})

	Context(".Postgres connection", func() {
		It("only hostname", func() {
			res := database.GetPsqlConnectionParams(hostname)
			Expect(res).To(Equal(fmt.Sprintf("host=%s user=%s dbname=%s", hostname, defaultUser, defaultDbName)))
		})
		It("hostname user", func() {
			os.Setenv(database.KeyPsqlUser, newUser)
			res := database.GetPsqlConnectionParams(hostname)
			os.Unsetenv(database.KeyPsqlUser)
			Expect(res).To(Equal(fmt.Sprintf("host=%s user=%s dbname=%s", hostname, newUser, defaultDbName)))
		})
		It("hostname user dbname", func() {
			os.Setenv(database.KeyPsqlUser, newUser)
			os.Setenv(database.KeyPsqlDbName, newDbName)
			res := database.GetPsqlConnectionParams(hostname)
			os.Unsetenv(database.KeyPsqlUser)
			os.Unsetenv(database.KeyPsqlDbName)
			Expect(res).To(Equal(fmt.Sprintf("host=%s user=%s dbname=%s", hostname, newUser, newDbName)))
		})
		It("hostname user dbname password", func() {
			os.Setenv(database.KeyPsqlUser, newUser)
			os.Setenv(database.KeyPsqlDbName, newDbName)
			os.Setenv(database.KeyPsqlPassword, newPassword)
			res := database.GetPsqlConnectionParams(hostname)
			os.Unsetenv(database.KeyPsqlUser)
			os.Unsetenv(database.KeyPsqlDbName)
			os.Unsetenv(database.KeyPsqlPassword)
			Expect(res).To(Equal(fmt.Sprintf("host=%s user=%s dbname=%s password=%s", hostname, newUser, newDbName, newPassword)))
		})
		It("hostname user dbname password port", func() {
			os.Setenv(database.KeyPsqlUser, newUser)
			os.Setenv(database.KeyPsqlDbName, newDbName)
			os.Setenv(database.KeyPsqlPassword, newPassword)
			os.Setenv(database.KeyPsqlPort, newPort)
			res := database.GetPsqlConnectionParams(hostname)
			os.Unsetenv(database.KeyPsqlUser)
			os.Unsetenv(database.KeyPsqlDbName)
			os.Unsetenv(database.KeyPsqlPassword)
			os.Unsetenv(database.KeyPsqlPort)
			Expect(res).To(Equal(fmt.Sprintf("host=%s user=%s dbname=%s password=%s port=%s", hostname, newUser, newDbName, newPassword, newPort)))
		})
		It("hostname timeout", func() {
			os.Setenv(database.KeyPsqlTimeout, newTimeout)
			res := database.GetPsqlConnectionParams(hostname)
			os.Unsetenv(database.KeyPsqlTimeout)
			Expect(res).To(Equal(fmt.Sprintf("host=%s user=%s dbname=%s connect_timeout=%s", hostname, defaultUser, defaultDbName, newTimeout)))
		})
		It("hostname user dbname password port", func() {
			os.Setenv(database.KeyPsqlUser, newUser)
			os.Setenv(database.KeyPsqlDbName, newDbName)
			os.Setenv(database.KeyPsqlPassword, newPassword)
			os.Setenv(database.KeyPsqlPort, newPort)
			res := database.GetPsqlConnectionParams(hostname)
			os.Unsetenv(database.KeyPsqlUser)
			os.Unsetenv(database.KeyPsqlDbName)
			os.Unsetenv(database.KeyPsqlPassword)
			os.Unsetenv(database.KeyPsqlPort)
			newres := database.GetPsqlWithPassowrdStarred(res)
			Expect(strings.Contains(newres, newPassword)).To(BeFalse())
		})
	})
	Context(".Postgres ssl", func() {
		It("empty", func() {
			res := database.GetPsqlSslParams()
			Expect(res).To(Equal(fmt.Sprintf("sslmode=%s", defaultSslMode)))
		})
		It("sslmode", func() {
			os.Setenv(database.KeyPsqlSslMode, newSslMode)
			res := database.GetPsqlSslParams()
			os.Unsetenv(database.KeyPsqlSslMode)
			Expect(res).To(Equal(fmt.Sprintf("sslmode=%s", newSslMode)))
		})
		It("sslmode sslrootcert", func() {
			os.Setenv(database.KeyPsqlSslMode, newSslMode)
			os.Setenv(database.KeyPsqlSslRootCert, newSslRootCert)
			res := database.GetPsqlSslParams()
			os.Unsetenv(database.KeyPsqlSslMode)
			os.Unsetenv(database.KeyPsqlSslRootCert)
			Expect(res).To(Equal(fmt.Sprintf("sslmode=%s sslrootcert=%s", newSslMode, newSslRootCert)))
		})
	})
})
