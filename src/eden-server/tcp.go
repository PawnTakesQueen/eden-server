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
  "crypto/tls"
  "io"
  "net"
)

type connection struct {
  UserID []byte
  Conn net.Conn
}

func handleRequest(request []byte) int {
  return len(request)
}

// Handles incoming requests.
func handleConn(conn net.Conn) {
  sock := connection {}
  sock.Conn = conn
  socketList = append(socketList, sock)
  var request []byte
  for {
    // Read the incoming connection into the buffer.
    buf := make([]byte, 1024)
    reqLen, err := conn.Read(buf)
    if err != nil && err != io.EOF {
      break
    } else {
      request = append(request, buf[:reqLen]...)
      newIndex := handleRequest(request)
      request = request[newIndex:]
    }
    if err == io.EOF {
      break
    }
  }
  defer conn.Close()
  for x, sock := range socketList {
    if sock.Conn == conn {
      socketList = append(socketList[:x], socketList[x + 1:]...)
    }
  }
}

// Listens for TCP traffic
func startTCPListen() {
  makeCACertPool()
  cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
  if err != nil {
    panic(err)
  }
  cfg := &tls.Config{
    Certificates: []tls.Certificate{cert},
    ClientAuth: tls.RequireAndVerifyClientCert,
    ClientCAs: caCertPool,
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
  if l, err := tls.Listen("tcp", ":" + tcpPort, cfg); err == nil {
    defer l.Close()
    for {
      // Listen for an incoming connection.
      if conn, err := l.Accept(); err != nil {
        panic("TCP connection not accepted")
      } else if conn != nil {
        // Handle connections in a new goroutine.
        go handleConn(conn)
      }
    }
  }
}
