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

type apiResponse struct {
  Recipient struct {
    ID string `json:"id"`
  } `json:"recipient"`
  SenderAction string `json:"sender_action"`
  Message struct {
    Text string `json:"text"`
  } `json:"message"`
}

type apiRequest struct {
  Object string `json:"object"`
  Entries []struct {
    ID string `json:"id"`
    Time int64 `json:"time"`
    Messaging []struct{
      Message struct {
        MID string `json:"mid"`
        Seq int64 `json:"seq"`
        Text string `json:"text"`
        Attachments []struct {
          Type string `json:"type"`
          Payload struct {
            URL string `json:"url,omitempty"`
            Coords struct {
              Lat float64 `json:"lat"`
              Long float64 `json:"long"`
            } `json:"coordinates,omitempty"`
          } `json:"payload"`
        } `json:"attachments,omitempty"`
        QuickReply struct {
          Payload string `json:"payload"`
        } `json:"quick_reply,omitempty"`
      }  `json:"message,omitempty"`
    } `json:"messaging"`
  } `json:"entry"`
}
