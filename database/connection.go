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
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/postgres"
    _ "github.com/jinzhu/gorm/dialects/sqlite"
    "github.com/IBM/ubiquity/utils/logs"
    "errors"
)

var globalConnectionFactory ConnectionFactory = nil

func initConnectionFactory(connectionFactory ConnectionFactory) func() {
    if globalConnectionFactory != nil {
        panic("globalConnectionFactory already initialized")
    }
    globalConnectionFactory = connectionFactory
    return func() { globalConnectionFactory = nil }
}

type ConnectionFactory interface {
    newConnection() (*gorm.DB, error)
}

type postgresFactory struct {
    psql     string
    psqlLog  string
}

type sqliteFactory struct {
    path     string
}

type testErrorFactory struct {
}

func (f *postgresFactory) newConnection() (*gorm.DB, error) {
    logger := logs.GetLogger()
    logger.Debug("", logs.Args{{"psql", f.psqlLog}})
    return gorm.Open("postgres", f.psql)
}

func (f *sqliteFactory) newConnection() (*gorm.DB, error) {
    return gorm.Open("sqlite3", f.path)
}

func (f *testErrorFactory) newConnection() (*gorm.DB, error) {
    return nil, errors.New("testErrorFactory")
}

type Connection struct {
    factory  ConnectionFactory
    logger   logs.Logger
    db       *gorm.DB
}

func NewConnection() Connection {
    return Connection{logger: logs.GetLogger(), factory: globalConnectionFactory}
}

func (c *Connection) Open() (error) {
    defer c.logger.Trace(logs.DEBUG)()
    var err error

    // sanity
    if c.db != nil {
        return c.logger.ErrorRet(errors.New("Connection already open"), "failed")
    }

    // open db connection
    if c.db, err = c.factory.newConnection(); err != nil {
        return c.logger.ErrorRet(err, "failed")
    }

    // do migrations
    if err = doMigrations(*c); err != nil {
        defer c.Close()
        return c.logger.ErrorRet(err, "doMigrations failed")
    }

    return nil
}

func (c *Connection) Close() (error) {
    defer c.logger.Trace(logs.DEBUG)()
    var err error

    // sanity
    if c.db == nil {
        return c.logger.ErrorRet(errors.New("Connection already closed"), "failed")
    }

    // close db connection
    err = c.db.Close()
    c.db = nil
    if err != nil {
        return c.logger.ErrorRet(err, "failed")
    }

    return nil
}

func (c *Connection) GetDb() (*gorm.DB) {
    defer c.logger.Trace(logs.DEBUG)()

    return c.db
}
