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

import(
  "flag"
  "fmt"
  "os"
)

type flagVals struct {
  adduser *bool
  c *string
  u *string
  v *bool
}

func handleStartingFlags() {
  if (*flags.v) {
    fmt.Printf("eden-server %s\n", version)
    os.Exit(0)
  }
}

func handleAfterPasswordFlags() {
  if *flags.adduser {
    if len(*flags.c) > 0 && len(*flags.u) > 0 {
      if cert, ok := validCert(*flags.c); ok {
        addClient(*flags.u, cert)
      } else {
        fmt.Println("\x1b[91mInvalid Client TLS Certificate!\x1b[0m")
      }
    } else if len(*flags.c) > 0 {
      fmt.Println("\x1b[91mClient TLS Certificate Path Must Be Specified!" +
                  "\x1b[0m")
    } else {
      fmt.Println("\x1b[91Client Username Must Be Specified!\x1b[0m")
    }
    os.Exit(0)
  } else if len(*flags.u) > 0 && len(*flags.c) > 0 {
    os.Exit(0)
  }
}

func getFlags() {
  flags.adduser = flag.Bool("adduser", false, "Add new client")
  flags.c = flag.String("c", "", "Path to client TLS certificate " +
                         "content")
  flags.u = flag.String("u", "", "New client username")
  flags.v = flag.Bool("v", false, "Print version information and exit")
  flag.Parse()
  handleStartingFlags()
}
