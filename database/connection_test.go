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
    "github.com/IBM/ubiquity/database"
)

var _ = Describe("Connection", func() {
    var (
        err error
    )
    BeforeEach(func() {
    })

    Context(".Open", func() {
        It("open success", func() {
			defer database.InitTestCorrect()()
            dbConnection := database.NewConnection()
            Expect(dbConnection.GetDb()).To(BeNil())
            err = dbConnection.Open()
            Expect(err).To(Not(HaveOccurred()))
        })
        It("close fail", func() {
			defer database.InitTestCorrect()()
            dbConnection := database.NewConnection()
            Expect(dbConnection.GetDb()).To(BeNil())
            err = dbConnection.Close()
            Expect(err).To(HaveOccurred())
        })
    })
})