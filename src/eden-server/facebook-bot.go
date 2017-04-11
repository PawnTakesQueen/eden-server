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
  "bytes"
  "encoding/json"
  "fmt"
  "net/http"
)

func handleFacebookMessage(message messageRequest) {
  return
}

func handleFacebook(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
  case "GET":
    r.ParseForm()
    rVal := r.Form
    if rVal["hub.mode"][0] == "subscribe" {
      if rVal["hub.verify_token"][0] == facebookVerifyToken {
        fmt.Fprintf(w, rVal["hub.challenge"][0])
      }
    }
  case "POST":
    var req apiRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
      panic(err)
    }
    for _, x := range req.Entries {
      for _, y := range x.Messaging {
        handleFacebookMessage(y.Message)
        if y.Sender.ID == facebookID {
          markSeen(y.Sender.ID)
        }
      }
    }
  }
}

func markSeen(id string) {
  var messageJSON apiResponse
  messageJSON.Recipient.ID = id
  messageJSON.SenderAction = "mark_seen"
  jsonValue, _ := json.Marshal(messageJSON)
  http.Post("https://graph.facebook.com/v2.8/me/messages?access_token=" +
            facebookPageAccessToken, "application/json",
            bytes.NewBuffer(jsonValue))
}
