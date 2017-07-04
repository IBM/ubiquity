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

package utils_test

import (
	"log"
	"os"
	"reflect"

	"fmt"

	"github.com/IBM/ubiquity/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("utils - locker", func() {
	var (
		locker         utils.Locker
		logger         *log.Logger
		sharedResource []string
		c1             chan int
		c2             chan int
		c3             chan int
		c4             chan int
		err            error
	)

	BeforeEach(func() {
		logger = log.New(os.Stdout, "locker: ", log.Lshortfile|log.LstdFlags)
		locker = utils.NewLocker()
		c1 = make(chan int)
		c2 = make(chan int)
		c3 = make(chan int)
		c4 = make(chan int)
		sharedResource = make([]string, 0)
		Expect(err).ToNot(HaveOccurred())
	})

	//Context(".ReadLock", func() {
	//	It("should succeed", func() {
	//		//expectedResource1 := []string{"a", "a", "a", "b", "b", "b"}
	//		//expectedResource2 := []string{"b", "b", "b", "a", "a", "a"}
	//		expectedResource1 := []string{"a", "a", "b", "b"}
	//		expectedResource2 := []string{"b", "b", "a", "a"}
	//		go readLockTest(locker, c1, &sharedResource, "a")
	//		go readLockTest(locker, c2, &sharedResource, "b")
	//		_, _ = <-c1, <-c2
	//		compare := (reflect.DeepEqual(sharedResource, expectedResource1) || reflect.DeepEqual(sharedResource, expectedResource2))
	//		Expect(compare).To(Equal(true))
	//	})
	//})
	Context(".ReadWriteLock", func() {
		It("should succeed", func() {
			//expectedResource1 := []string{"c", "c", "c", "d", "d", "d"}
			//expectedResource2 := []string{"d", "d", "d", "c", "c", "c"}
			expectedResource1 := []string{"c", "c", "d", "d"}
			expectedResource2 := []string{"d", "d", "c", "c"}
			go readWriteLockTest(locker, c3, &sharedResource, "c")
			go readWriteLockTest(locker, c4, &sharedResource, "d")
			_, _ = <-c3, <-c4
			compare := (reflect.DeepEqual(sharedResource, expectedResource1) || reflect.DeepEqual(sharedResource, expectedResource2))
			Expect(compare).To(Equal(true))

		})
	})
	Context(".WriteLock", func() {
		It("should succeed", func() {
			//expectedResource1 := []string{"c", "c", "c", "d", "d", "d"}
			//expectedResource2 := []string{"d", "d", "d", "c", "c", "c"}
			expectedResource1 := []string{"c", "c", "d", "d"}
			expectedResource2 := []string{"d", "d", "c", "c"}
			go writeLockTest(locker, c3, &sharedResource, "c")
			go writeLockTest(locker, c4, &sharedResource, "d")
			_, _ = <-c3, <-c4
			compare := (reflect.DeepEqual(sharedResource, expectedResource1) || reflect.DeepEqual(sharedResource, expectedResource2))
			Expect(compare).To(Equal(true))

		})
	})
})

//func readLockTest(locker utils.Locker, c chan int, sharedResource *[]string, letter string) {
//	locker.ReadLock("readLockTest")
//	for i := 0; i < 2; i++ {
//		*sharedResource = append(*sharedResource, letter)
//	}
//	locker.ReadUnlock("readLockTest")
//
//	c <- 1
//}

func writeLockTest(locker utils.Locker, c chan int, sharedResource *[]string, letter string) {
	locker.WriteLock("writeLockTest")

	for i := 0; i < 2; i++ {
		*sharedResource = append(*sharedResource, letter)
	}
	locker.WriteUnlock("writeLockTest")
	c <- 1
}

func readWriteLockTest(locker utils.Locker, c chan int, sharedResource *[]string, letter string) {
	locker.ReadLock("readWriteLockTest")
	fmt.Printf("%#v", sharedResource)
	locker.ReadUnlock("readWriteLockTest")
	locker.WriteLock("readWriteLockTest")
	for i := 0; i < 2; i++ {
		*sharedResource = append(*sharedResource, letter)
	}
	locker.WriteUnlock("readWriteLockTest")
	locker.ReadLock("readWriteLockTest")
	fmt.Printf("%#v", sharedResource)
	locker.ReadUnlock("readWriteLockTest")
	c <- 1
}
