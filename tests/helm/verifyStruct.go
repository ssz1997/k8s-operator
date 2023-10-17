/*
 * The Alluxio Open Foundation licenses this work under the Apache License, version 2.0
 * (the "License"). You may not use this work except in compliance with the License, which is
 * available at www.apache.org/licenses/LICENSE-2.0
 *
 * This software is distributed on an "AS IS" basis, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied, as more fully set forth in the License.
 *
 * See the NOTICE file distributed with this work for information regarding copyright ownership.
 */

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"

	"sigs.k8s.io/yaml"

	alluxiov1alpha1 "github.com/alluxio/k8s-operator/api/v1alpha1"
)

func main() {
	verifyHelmStructIsUpToDate()
	verifyOperatorStructIsSubsetOfHelmStruct()
	os.Exit(0)
}

func verifyHelmStructIsUpToDate() {
	_, scriptPath, _, ok := runtime.Caller(1)
	if !ok {
		panic("not able to get script path")
	}
	testConfigBytes, err := os.ReadFile(filepath.Join(scriptPath, "../../../deploy/charts/alluxio/values.yaml"))
	if err != nil {
		panic(err)
	}
	helmStruct := alluxiov1alpha1.AlluxioClusterHelmChartSpec{}
	if err := yaml.Unmarshal(testConfigBytes, helmStruct); err != nil {
		panic(err)
	}
}

func verifyOperatorStructIsSubsetOfHelmStruct() {
	pass := IsSubset(alluxiov1alpha1.AlluxioClusterSpec{}, alluxiov1alpha1.AlluxioClusterHelmChartSpec{})
	if !pass {
		panic("operator struct is not a subset of helm struct. Please double check.")
	}
}

// IsSubset checks if struct 'a' is a subset of struct 'b'
func IsSubset(a, b interface{}) bool {
	valA := reflect.ValueOf(a)
	valB := reflect.ValueOf(b)

	for i := 0; i < valA.NumField(); i++ {
		fieldA := valA.Field(i)
		fieldB := valB.FieldByName(valA.Type().Field(i).Name)

		// Check if field exists in struct 'b'
		if !fieldB.IsValid() {
			fmt.Println("?")
			return false
		}

		// Recursively check nested structs
		if fieldA.Kind() == reflect.Struct {
			fmt.Println("recursive")
			if !IsSubset(fieldA.Interface(), fieldB.Interface()) {
				return false
			}
		}
	}
	return true
}
