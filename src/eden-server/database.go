/*-
 * Copyright (C) 2017, Vi Grey
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 *
 * 1. Redistributions of source code must retain the above copyright
 *    notice, this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY AUTHOR AND CONTRIBUTORS ``AS IS'' AND
 * ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL AUTHOR OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS
 * OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
 * HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
 * LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
 * OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
 * SUCH DAMAGE.
 */

package main

import (
  _ "github.com/go-sql-driver/mysql"
  "database/sql"
  "encoding/base64"
  "fmt"
)

type auth struct {
  dbKey string
  iterations int
  saltVal string
}

type client struct {
  id string
  name string
  tlsCert string
}

type clientDecrypted struct {
  id []byte
  name string
  tlsCert []byte
}

func getDatabase() {
  dbVal, err := sql.Open("mysql", mysqlUser + ":" + mysqlPass + "@" +
                        mysqlConnType + "(" + mysqlConnPath + ")/")
  if err != nil {
    panic(err)
  }
  if _, err := dbVal.Exec("CREATE DATABASE IF NOT EXISTS eden"); err != nil {
    panic(err)
  }
  dbVal.Close()
  db, err = sql.Open("mysql", mysqlUser + ":" + mysqlPass + "@" +
                     mysqlConnType + "(" + mysqlConnPath + ")/eden")
  if err != nil {
    panic(err)
  }
  if _, err := db.Exec("CREATE TABLE IF NOT EXISTS iv (iteration " +
                       "BIGINT NOT NULL, PRIMARY KEY " +
                       "(iteration))"); err != nil {
    panic(err)
  }
  initialIV := getRandNumber(9223372036854775807)
  if _, err := db.Exec("INSERT INTO iv (iteration) SELECT ? WHERE NOT " +
                       "EXISTS (SELECT * FROM iv)", initialIV); err != nil {
    panic(err)
  }
  if _, err := db.Exec("CREATE TABLE IF NOT EXISTS auth (dbKey " +
                       "VARBINARY(80) NOT NULL, iterations INT, " +
                       "saltVal VARBINARY(64), PRIMARY KEY " +
                       "(dbKey))"); err != nil {
    panic(err)
  }
  if _, err := db.Exec("CREATE TABLE IF NOT EXISTS clients " +
                       "(id VARBINARY(80) NOT NULL, name VARBINARY(544) " +
                       "NOT NULL, tlsCert VARBINARY(1000) NOT NULL, PRIMARY " +
                       "KEY (id))"); err != nil {
    panic(err)
  }
  getAESGCMIV()
}

func setupDBKey() {
  var exists bool
  rows, err := db.Query("SELECT * FROM auth")
  if err != nil {
    panic(err)
  }
  defer rows.Close()
  if rows.Next() {
    exists = true
  }
  if exists {
    enterPassphrase()
  } else {
    newPassphrase()
  }
}

func addClient(name string, cert []byte) {
  var idCT string
  var idRepeated, clientRepeated bool
  nameCT := encrypt([]byte(name), dbKey)
  certCT := encrypt(cert, dbKey)
  for idRepeated || len(idCT) == 0 {
    id := randByteArray(32)
    idCT := encrypt(id, dbKey)
    for _, c := range clients {
      if string(c.id) == string(id) {
        idRepeated = true
      }
      if c.name == name {
        fmt.Println("\x1b[91mClient " + name + " Already Exists!\x1b[0m")
        clientRepeated = true
        break
      }
      if string(c.tlsCert) == string(cert) {
        fmt.Println("\x1b[91mClient With That TLS Certificate Already Exists!\x1b[0m")
        clientRepeated = true
        break
      }
      if len(name) == 0 {
        fmt.Println("\x1b[91mClient Must Include Name!\x1b[0m")
        clientRepeated = true
        break
      }
    }
    if !idRepeated && !clientRepeated {
      fmt.Println("\x1b[92mClient " + name + " Added!\x1b[0m")
      if _, err := db.Exec("INSERT INTO clients (id, name, tlsCert) VALUES " +
                           "(?, ?, ?)", idCT, nameCT, certCT); err != nil {
        panic(err)
      }
      break
    } else if clientRepeated {
      break
    }
  }
}

func addDBKey(passphrase []byte) {
  saltVal := randByteArray(32)
  saltValBase64 := base64.StdEncoding.EncodeToString(saltVal)
  key := getPBKDF2Key(passphrase, saltVal, iterations)
  dbKey = randByteArray(32)
  dbKeyEncrypted := encrypt(dbKey, key)
  _, err := db.Exec("INSERT INTO auth (dbKey, iterations, saltVal) VALUES " +
                    "(?, ?, ?)", dbKeyEncrypted, iterations, saltValBase64)
  if err != nil {
    panic(err)
  }
}

func checkDBKey(passphrase string) bool {
  rows, err := db.Query("SELECT * FROM auth")
  if err != nil {
    panic(err)
  }
  defer rows.Close()
  for rows.Next() {
    authRow := new(auth)
    rows.Scan(&authRow.dbKey, &authRow.iterations, &authRow.saltVal)
    saltVal, err := base64.StdEncoding.DecodeString(authRow.saltVal)
    if err != nil {
      panic(err)
    }
    key := getPBKDF2Key([]byte(passphrase), saltVal, authRow.iterations)
    if dbKeyPT, ok := decrypt(authRow.dbKey, key); ok {
      dbKey = dbKeyPT
      return true
    }
  }
  return false
}

func getClients() {
  rows, err := db.Query("SELECT * FROM clients")
  if err != nil {
    panic(err)
  }
  defer rows.Close()
  for rows.Next() {
    clientRow := new(client)
    rows.Scan(&clientRow.id, &clientRow.name, &clientRow.tlsCert)
    if idPT, ok := decrypt(clientRow.id, dbKey); ok {
      if namePT, ok := decrypt(clientRow.name, dbKey); ok {
        if tlsCertPT, ok := decrypt(clientRow.tlsCert, dbKey); ok {
          var newClient clientDecrypted
          newClient.id = idPT
          newClient.name = string(namePT)
          newClient.tlsCert = tlsCertPT
          clients = append(clients, newClient)
        }
      }
    }
  }
}

func getAESGCMIV() {
  rows, err := db.Query("SELECT * FROM iv")
  if err != nil {
    panic(err)
  }
  defer rows.Close()
  if rows.Next() {
    rows.Scan(&aesGCMIV)
  }
}
