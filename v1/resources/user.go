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

package resources

import (
  "github.com/gin-gonic/gin"
  "github.com/satori/go.uuid"
  "golang.org/x/crypto/bcrypt"
  "gopkg.in/redis.v3"
  "encoding/base64"

  log "github.com/Sirupsen/logrus"
)

type UserResource struct {
  Redis *redis.Client
  Engine *gin.RouterGroup
}

func (u UserResource) Routes() {
  u.Engine.GET("/users/:uuid", u.Get)
  u.Engine.POST("/users", u.Create)
}

func (u UserResource) Get(c *gin.Context) {
  val, err := u.Redis.Get("user::id::"+c.Param("uuid")).Result()
  if err == redis.Nil {
    NotFound("User", c)
  } else if err != nil {
    log.Error(err)
    InternalError(c)
  } else {
    c.JSON(200, gin.H{
      "message": "User fetched.",
      "value": val,
    })
  }
}

func (u UserResource) Create(c *gin.Context) {
  user := base64.StdEncoding.EncodeToString([]byte(c.PostForm("username")))
  _, err := u.Redis.Get("user::name::"+user).Result()
  if err != redis.Nil {
    c.JSON(405, gin.H{
      "message": "User already available",
    })
  } else {
    new_user := NewUser(c.PostForm("username"),c.PostForm("password"), &u)
    if new_user.Save() {
      c.JSON(201, gin.H{
        "user": new_user.Uuid,
      })
    } else { 
      c.JSON(422, gin.H{
        "message": "Createing new user failed.",
      })
    }
  }
}

type User struct {
  UserResource *UserResource
  Uuid string
  Username string
  Password string
  PasswordDigest string
}

func NewUser(username string, password string, ur *UserResource) User {
  new_user := User{
    UserResource: ur,
    Username: username,
    Password: password,
    Uuid: uuid.NewV4().String(),
  }
  v, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
  new_user.PasswordDigest = string(v)
  return new_user
}

func (u *User) Save() bool {
  err := u.UserResource.Redis.Set("user::id::"+u.GetUuid(), u.EncodedUsername(), 0).Err()
  if err != nil {
    log.Error(err, "Error on saving uuid <> username mapping")
    return false
  }
  err = u.UserResource.Redis.Set("user::name::"+u.EncodedUsername(), u.GetUuid(), 0).Err()
  if err != nil {
    log.Error(err, "Error on saving username <> uuid mapping")
    return false
  }
  err = u.UserResource.Redis.Set("user::pass::"+u.EncodedUsername(), u.GetPasswordDigest(), 0).Err()
  if err != nil {
    log.Error(err, "Error on saving username <> password_digest")
    return false
  }
  return true
}

func (u *User) EncodedUsername() string {
  return base64.StdEncoding.EncodeToString([]byte(u.GetUsername()))
}

func (u *User) GetUuid() string {
  if u.Uuid == "" {
    val, err := u.UserResource.Redis.Get("user::name::"+u.EncodedUsername()).Result()
    if err != nil {
      log.Error(err, "Fetching Uuid failed")
    } else {
      u.Uuid = val
    }
  }
  return u.Uuid
}

func (u *User) ResetUuid() string {
  id := uuid.NewV4().String()
  err := u.UserResource.Redis.Set("user::id::"+id, u.EncodedUsername(), 0).Err()
  if err != nil {
    log.Error(err, "Error on setting new id.")
  }
  err = u.UserResource.Redis.Set("user::name::"+u.EncodedUsername(), id, 0).Err()
  if err != nil {
    log.Error(err, "Error on setting id to username.")
  }
  err = u.UserResource.Redis.Del("user::id::"+u.GetUuid()).Err()
  if err != nil {
    log.Error(err, "Error on deleting old id.")
  } else {
    u.Uuid = id
  }
  return id
}

func (u *User) GetUsername() string {
  if u.Username == "" {
    val, err := u.UserResource.Redis.Get("user::id::"+u.Uuid).Result()
    if err != nil {
      log.Error(err, "Fetching Username failed")
    } else {
      str, _ := base64.StdEncoding.DecodeString(val)
      u.Username = string(str)
    }
  }
  return u.Username
}

func (u *User) GetPasswordDigest() string {
  if u.PasswordDigest == "" {
    val, err := u.UserResource.Redis.Get("user::pass::"+u.EncodedUsername()).Result()
    if err != nil {
      log.Error(err, "Fetching PasswordDigest failed")
    } else {
      u.PasswordDigest = val
    }
  }
  return u.PasswordDigest
}
