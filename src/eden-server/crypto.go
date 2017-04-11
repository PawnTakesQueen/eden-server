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
  "golang.org/x/crypto/pbkdf2"
  "golang.org/x/crypto/ssh/terminal"
  "crypto/aes"
  "crypto/cipher"
  "crypto/rand"
  "crypto/sha256"
  "crypto/x509"
  "encoding/base64"
  "encoding/pem"
  "fmt"
  "io/ioutil"
  "math/big"
  "syscall"
)

// Get a random int between 0 and number
func getRandNumber(number int64) int64 {
  randNumber, _ := rand.Int(rand.Reader, big.NewInt(number))
  return randNumber.Int64()
}

// Generate a random byte array of size length
func randByteArray(size int) []byte {
  randValue := make([]byte, size)
  if _, err := rand.Read(randValue); err != nil {
    panic(err)
  }
  return randValue
}

/*
 * Encrypts plaintext using key with AES256-GCM and returns the ciphertext.
 * plaintext can be []byte or *string
 */
func encrypt(plaintext interface{}, key []byte) string {
  var byteArray []byte
  var err error
  switch v := plaintext.(type) {
  case []byte:
    byteArray = make([]byte, len(v))
    copy(byteArray, v)
  case *string:
    byteArray = make([]byte, len(*v))
    for i := range(*v) {
      byteArray[i] = (*v)[i]
    }
  }
  block, err := aes.NewCipher(key)
  if err != nil {
    panic(err)
  }
  /*
   * Add one to aesGCMIV, roll value over to 0 if the uint64 value hits
   * 2 ** 64 - 1
   */
  aesGCMIVOld := aesGCMIV
  aesGCMIV = int64(uint64(aesGCMIV) + 1 % 18446744073709551615)
  if _, err := db.Exec("UPDATE iv SET iteration = ? WHERE iteration = ?",
                       aesGCMIV, aesGCMIVOld); err != nil {
    panic(err)
  }
  // Pad aesGCMIV value with 4 random bytes to create the iv value
  iv := append(numToBytes(aesGCMIV, 8), randByteArray(4)...)
  mode, err := cipher.NewGCM(block)
  if err != nil {
    panic(err)
  }
  ciphertext := mode.Seal(nil, iv, byteArray, nil)
  b64 := base64.StdEncoding.EncodeToString(append(iv, ciphertext...))
  return b64
}

// Decrypts the byte slice ciphertext and unpads it.
func decrypt(ciphertextBase64 string, key []byte) ([]byte, bool) {
  ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
  if err != nil {
    panic(err)
  }
  if len(ciphertext) < 12 {
    return nil, false
  }
  block, err := aes.NewCipher(key)
  if err != nil {
    panic(err)
  }
  iv := ciphertext[:12]
  aesGCMIV = bytesToNum(iv[:8])
  ct := ciphertext[12:]
  mode, err := cipher.NewGCM(block)
  if err != nil {
    panic(err)
  }
  plaintext, err := mode.Open(nil, iv, ct, nil)
  if err != nil {
    return nil, false
  }
  return plaintext, true
}

func makeCACertPool() {
  caCertPool = x509.NewCertPool()
  for _, c := range clients {
    if !caCertPool.AppendCertsFromPEM(c.tlsCert) {
      panic("Unable to read client TLS certificates")
    }
  }
}

func validCert(newCert string) ([]byte, bool) {
  newCertContent, err := ioutil.ReadFile(newCert)
  if err != nil {
    panic(err)
  }
  newCertBlock, _ := pem.Decode(newCertContent)
  if newCertBlock != nil {
    if _, err := x509.ParseCertificate(newCertBlock.Bytes); err == nil {
      return newCertContent, true
    }
  }
  return []byte{}, false
}

func enterPassphrase() string {
  var bytePassword []byte
  for {
    fmt.Printf("\x1b[1mEnter Passphrase:\x1b[0m ")
    bytePass, err := terminal.ReadPassword(int(syscall.Stdin))
    if err != nil {
      panic(err)
    }
    bytePassword = bytePass
    if checkDBKey(string(bytePassword)) {
      fmt.Println("\x1b[92mSUCCESS\x1b[0m")
      break
    } else {
      fmt.Println("\x1b[91mFAIL\x1b[0m")
    }
  }
  return string(bytePassword)
}

func newPassphrase() {
  var bytePassword, bytePasswordConfirm []byte
  for (string(bytePassword) != string(bytePasswordConfirm) ||
       len(bytePassword) == 0) {
    fmt.Printf("\x1b[1mEnter New Passphrase:\x1b[0m ")
    bytePass, err := terminal.ReadPassword(int(syscall.Stdin))
    if err != nil {
      panic(err)
    }
    bytePassword = bytePass
    fmt.Printf("\n\x1b[1mConfirm New Passphrase:\x1b[0m ")
    bytePassConfirm, err := terminal.ReadPassword(int(syscall.Stdin))
    if err != nil {
      panic(err)
    }
    bytePasswordConfirm = bytePassConfirm
    if string(bytePassword) != string(bytePasswordConfirm) {
      fmt.Println("\x1b[91mFAIL - PASSWORDS MUST MATCH\x1b[0m\n")
    } else if len(bytePassword) == 0 {
      fmt.Println("\x1b[91mFAIL - MUST ENTER PASSWORD\x1b[0m\n")
    }
  }
  addDBKey(bytePassword)
  fmt.Println("\x1b[92mSUCCESS\x1b[0m\n")
}

func getPBKDF2Key(pass, salt []byte, iter int) []byte {
  return pbkdf2.Key(pass, salt, iter, 32, sha256.New)
}
