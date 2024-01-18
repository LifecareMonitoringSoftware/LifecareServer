// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This file contains functions for retrieving Google Cloud Run metadata.
package cloud_run

import (
	"context"
	"log"

	"cloud.google.com/go/compute/metadata"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

// defaultProjectID,
// This is our default project Id.
// When we are running on cloud run, this will be asserted to match the cloud run project ID.
// When we are not running on cloud run, this will be used as the project ID.
const defaultProjectID string = "lifecare-server"

// projectIdCache,
// The stores our project ID for repeated function calls.
// "" If not initialized, otherwise contains a non-empty project ID.
var projectIdCache string = ""

// projectIDCache,
// The stores our running on cloud run information, for repeated function calls.
// "0" If not initialized.
// 1 If running on cloud run.
// 2 If not running on cloud run.
var isRunningOnCloudRunCache int = 0

// ProjectID,
// This returns our Project ID. This is guaranteed to not be empty.
// This caches the result for quick execution.
// We are not running on Cloud Run, this will be the default Project ID.
// If we are running on Cloud Run, this will be obtained from the Cloud Run environment.
// If we are running on Cloud Run, this will guarantee that the default Project ID
// string matches the Cloud Run project ID.
func ProjectID() string {
	if projectIdCache == "" {
		// This line gets the Project ID from the Cloud Run service.
		// It fetches the information from the GCP metadata server.
		receivedProjectId, projectIdError := metadata.ProjectID()
		var didReceiveProjectID = (projectIdError == nil)
		if didReceiveProjectID {
			projectIdCache = receivedProjectId
			if receivedProjectId != defaultProjectID {
				log.Fatalf("the Default Project ID does not match the Cloud Run Project ID.\n"+
					"Default: '%v', CloudRun: '%v'.", defaultProjectID, receivedProjectId)
			}
		} else {
			projectIdCache = defaultProjectID
		}
	}
	return projectIdCache
}

// IsRunningOnCloudRun,
// This returns true if we are running on cloud run, otherwise false.
// This caches the result for quick execution.
func IsRunningOnCloudRun() bool {
	if isRunningOnCloudRunCache == 0 {
		_, projectIdError := metadata.ProjectID()
		var isRunningOnCloudRun = (projectIdError == nil)
		if isRunningOnCloudRun {
			isRunningOnCloudRunCache = 1
		} else {
			isRunningOnCloudRunCache = 2
		}
	}
	return (isRunningOnCloudRunCache == 1)
}

// Region returns the region of the Cloud Run service. It fetches this
// information from the GCP metadata server. The returned value is in the format
// of: projects/PROJECT_NUMBER/regions/REGION.
func Region() (string, error) {
	resp, err := metadata.Get("instance/region")
	if err != nil {
		return "", nil
	}

	return resp, nil
}

// IDToken returns a TokenSource that yields ID tokens. These tokens can be used
// to authenticate requests with the Token.SetAuthHeader method.
func IDToken(ctx context.Context, aud string) (oauth2.TokenSource, error) {
	idtoken.NewTokenSource(ctx, aud)
	return idtoken.NewTokenSource(ctx, aud)
}
