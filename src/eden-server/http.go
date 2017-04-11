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
  "github.com/gorilla/mux"
  "crypto/tls"
  "fmt"
  "net/http"
  "time"
)

var (
  indexContent, _ = Asset("static/index.html")
  cssContent, _ = Asset("static/css/layout.css")
  raleway400FontContent, _ = Asset("static/fonts/raleway-400.woff")
  faviconContent, _ = Asset("static/layout/favicon.png")
  edenGifContent, _ = Asset("static/images/eden.gif")
)

func handleLayoutCSS(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Content-Type", "text/css")
  w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
  fmt.Fprintf(w, regExpReplace(string(cssContent), "%", "%%"))
}

func handleRaleway400Font(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Content-Type", "applicatoin/font-woff")
  w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
  fmt.Fprintf(w, regExpReplace(string(raleway400FontContent), "%", "%%"))
}

func handleFavicon(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Content-Type", "image/png")
  w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
  fmt.Fprintf(w, regExpReplace(string(faviconContent), "%", "%%"))
}

func handleEdenGif(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Content-Type", "image/gif")
  w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
  fmt.Fprintf(w, regExpReplace(string(edenGifContent), "%", "%%"))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
  fmt.Fprintf(w, string(indexContent))
}

func startHTTPListen() {
  router := mux.NewRouter()
  router.HandleFunc("/", handleIndex)
  router.HandleFunc("/css/layout.css", handleLayoutCSS)
  router.HandleFunc("/fonts/raleway-400.woff", handleRaleway400Font)
  router.HandleFunc("/layout/favicon.png", handleFavicon)
  router.HandleFunc("/images/eden.gif", handleEdenGif)
  router.HandleFunc("/facebook", handleFacebook)
  cfg := &tls.Config{
    MinVersion: tls.VersionTLS12,
    CurvePreferences: []tls.CurveID{tls.X25519},
    PreferServerCipherSuites: true,
    CipherSuites: []uint16{
      tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
      tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
      tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
      tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
    },
  }
  srv := &http.Server {
    Addr: "localhost:" + httpPort,
    Handler: router,
    TLSConfig: cfg,
    TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
    ReadTimeout: 1 * time.Minute,
    WriteTimeout: 1 * time.Minute,
  }
  if err := srv.ListenAndServeTLS(tlsCert, tlsKey); err != nil {
    panic(err)
  }
}
