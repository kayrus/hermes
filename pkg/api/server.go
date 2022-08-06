/*******************************************************************************
*
* Copyright 2022 SAP SE
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You should have received a copy of the License along with this
* program. If not, you may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*
*******************************************************************************/

package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/spf13/viper"

	"github.com/sapcc/hermes/pkg/identity"
	"github.com/sapcc/hermes/pkg/storage"
	"github.com/sapcc/hermes/pkg/util"
)

// Server Set up and start the API server, hooking it up to the API router
func Server(keystone identity.Identity, storage storage.Storage) error {
	fmt.Println("API")
	mainRouter := setupRouter(keystone, storage)

	http.Handle("/", mainRouter)

	//start HTTP server
	listenaddress := viper.GetString("API.ListenAddress")
	util.LogInfo("listening on %s", listenaddress)
	//enable cors support
	c := cors.New(cors.Options{
		AllowedHeaders: []string{"X-Auth-Token", "Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "HEAD"},
		MaxAge:         600,
	})
	handler := c.Handler(mainRouter)
	return http.ListenAndServe(listenaddress, handler)
}

func setupRouter(keystone identity.Identity, storage storage.Storage) http.Handler {
	mainRouter := mux.NewRouter()
	//hook up the v1 API (this code is structured so that a newer API version can
	//be added easily later)
	v1Router, v1VersionData := NewV1Handler(keystone, storage)
	mainRouter.PathPrefix("/v1/").Handler(v1Router)

	//add the version advertisement that lists all available API versions
	mainRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		allVersions := struct {
			Versions []VersionData `json:"versions"`
		}{[]VersionData{v1VersionData}}
		ReturnJSON(w, http.StatusMultipleChoices, allVersions)
	})

	// instrumentation
	mainRouter.Handle("/metrics", promhttp.Handler())

	return gaugeInflight(mainRouter)
}
