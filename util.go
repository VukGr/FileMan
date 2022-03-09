package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"math/rand"
	"mime/multipart"
)

func stringHash(data []byte) []byte {
    hash := sha256.Sum256(data)
    return hash[:]
}

func fileHash(fileHeader *multipart.FileHeader) (string, error) {
  fileData, err := fileHeader.Open()
  if err != nil {
    return "", err
  }
  defer fileData.Close()

  sha := sha256.New()
  if _, err := io.Copy(sha, fileData); err != nil {
    return "", err
  }

  return hex.EncodeToString(sha.Sum(nil)), nil
}


func randID(length int) string {
  b := make([]byte, length)
  rand.Read(b)
  return base64.URLEncoding.EncodeToString(b)
}

