// Copyright 2016 Tim Foerster <github@mailserver.1n3t.de>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/timmyArch/muh-api/v1/resources"
	"gopkg.in/redis.v3"
	"os"
)

// RedisClient - Get new redis connection.
func RedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
		Network:  os.Getenv("REDIS_NETWORK"),
	})
}

// Routes - Register all routes for API version 1
func Routes(api *gin.Engine) {
	r := RedisClient()
	version := api.Group("/v1")
	version.Use(Ratelimit(r))
	version.GET("/ping", Ping)

	resources.UserResource{
		Redis:  r,
		Engine: version,
	}.Routes()

	resources.GistResource{
		Redis:  r,
		Engine: version,
	}.Routes()

}

// Ping - a generic ping / pong route
func Ping(c *gin.Context) {
	c.JSON(418, gin.H{
		"message": "pong",
	})
}
