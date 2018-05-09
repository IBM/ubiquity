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
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils/logs"
	"os"
	"strings"
)

const (
	KeyPsqlHost        = "UBIQUITY_DB_PSQL_HOST"
	KeySqlitePath      = "UBIQUITY_DB_SQLITE_PATH"
	KeyPsqlUser        = "UBIQUITY_DB_USERNAME"
	KeyPsqlPassword    = "UBIQUITY_DB_PASSWORD"
	KeyPsqlDbName      = "UBIQUITY_DB_NAME"
	KeyPsqlPort        = "UBIQUITY_DB_PSQL_PORT"
	KeyPsqlTimeout     = "UBIQUITY_DB_CONNECT_TIMEOUT"
	KeyPsqlSslMode     = "UBIQUITY_DB_SSL_MODE"
	KeyPsqlSslRootCert = "UBIQUITY_DB_SSL_ROOT_CERT"
	keyPsqlDbPVName    = "IBM_UBIQUITY_DB_PV_NAME"
)

func GetPsqlConnectionParams(hostname string) string {
	str := ""
	// add host
	str += "host=" + hostname
	// add user
	psqlUser := os.Getenv(KeyPsqlUser)
	if psqlUser == "" {
		psqlUser = "postgres"
	}
	str += " user=" + psqlUser
	// add dbname
	psqlDbName := os.Getenv(KeyPsqlDbName)
	if psqlDbName == "" {
		psqlDbName = "postgres"
	}
	str += " dbname=" + psqlDbName
	// add password
	psqlPassword := os.Getenv(KeyPsqlPassword)
	if psqlPassword != "" {
		str += " password=" + psqlPassword
	}
	// add port
	psqlPort := os.Getenv(KeyPsqlPort)
	if psqlPort != "" {
		str += " port=" + psqlPort
	}
	// add connect_timeout
	psqlTimeout := os.Getenv(KeyPsqlTimeout)
	if psqlTimeout != "" {
		str += " connect_timeout=" + psqlTimeout
	}
	return str
}

func GetPsqlSslParams() string {
	str := ""
	// add sslmode
	psqlSslMode := os.Getenv(KeyPsqlSslMode)
	if psqlSslMode == "" {
		psqlSslMode = resources.DefaultDbSslMode
	}
	str += "sslmode=" + psqlSslMode
	// add sslrootcert
	psqlSslRootCert := os.Getenv(KeyPsqlSslRootCert)
	if psqlSslRootCert != "" {
		// if sslmode is verify-full then postgres will verify the certificate, if its require then it will automatically skip verify
		str += " sslrootcert=" + psqlSslRootCert
	}
	return str
}

func GetPsqlWithPassowrdStarred(psql string) string {
	slc := strings.Split(psql, " ")
	newslc := slc[:0]
	for _, x := range slc {
		if !strings.HasPrefix(x, "password") {
			newslc = append(newslc, x)
		} else {
			newslc = append(newslc, "password=****")
		}
	}
	return strings.Join(newslc, " ")
}

func InitPostgres(hostname string) func() {
	defer logs.GetLogger().Trace(logs.DEBUG)()
	psqlStr := GetPsqlConnectionParams(hostname) + " " + GetPsqlSslParams()
	return initConnectionFactory(&postgresFactory{psql: psqlStr, psqlLog: GetPsqlWithPassowrdStarred(psqlStr)})
}

func InitSqlite(filepath string) func() {
	defer logs.GetLogger().Trace(logs.DEBUG)()
	return initConnectionFactory(&sqliteFactory{path: filepath})
}

func InitTestError() func() {
	defer logs.GetLogger().Trace(logs.DEBUG)()
	return initConnectionFactory(&testErrorFactory{})
}

func Initialize() func() {
	defer logs.GetLogger().Trace(logs.DEBUG)()

	psqlHost := os.Getenv(KeyPsqlHost)
	if psqlHost != "" {
		return InitPostgres(psqlHost)
	}
	return InitSqlite(os.Getenv(KeySqlitePath))
}
