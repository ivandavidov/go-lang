package main

import "fmt"
import "os"
import "crypto/rand"
import "strings"

const BUFFER_SIZE int = 1024
const EXT_ENC string = ".enc"
const EXT_PAD string = ".pad"

func main() {
  if len(os.Args) <= 2 {
    printHelp()
    os.Exit(1)
  }
  
  operation := os.Args[1]
  
  if operation == "encrypt" && len(os.Args) >= 3 {
    encrypt(os.Args[2])
  } else if operation == "decrypt" && len(os.Args) >= 4 {
    decrypt(os.Args[2], os.Args[3])
  } else {
    printHelp()
    os.Exit(1)
  }
}

func encrypt(normalFileName string) {
  if !checkFileExists(normalFileName) {
    fmt.Println("The file '" + normalFileName + "' does not exist.")
    os.Exit(1)
  }
  
  normalFile, err := os.Open(normalFileName)
  check(err)
  defer normalFile.Close()
  
  normalBuf := make([]byte, BUFFER_SIZE)
  padBuf := make([]byte, BUFFER_SIZE)
  encryptedBuf := make([]byte, BUFFER_SIZE)
  
  encryptedFileName := normalFileName + EXT_ENC
  encryptedFile, err := os.Create(encryptedFileName)
  check(err)
  defer encryptedFile.Close()
  
  padFileName := normalFileName + EXT_PAD
  padFile, err := os.Create(padFileName)
  check(err)
  defer padFile.Close()
  
  numBytes, err := normalFile.Read(normalBuf);
  for numBytes > 0 {
    check(err)
    normalSlice := normalBuf[0:numBytes]
    
    _, err = rand.Read(padBuf)
    check(err)
    padSlice := padBuf[0:numBytes]
    
    encryptedSlice := encryptedBuf[0:numBytes]
    
    for i, normalByte := range normalSlice {
      encryptedSlice[i] = normalByte ^ padSlice[i]
    }
    
    _, err = padFile.Write(padSlice)
    check(err)
    
    _, err = encryptedFile.Write(encryptedSlice)
    check(err)
    
    numBytes, err = normalFile.Read(normalBuf);
  }
}

func decrypt(encryptedFileName string, padFileName string) {
  if encryptedFileName == padFileName {
    fmt.Println("Cannot decrypt file by using it as pad.")
    os.Exit(1)
  }
  
  if !checkFileExists(encryptedFileName) {
    fmt.Println("The file '" + encryptedFileName + "' does not exist.")
    os.Exit(1)
  }

  if !checkFileExists(padFileName) {
    fmt.Println("The pad file '" + padFileName + "' does not exist.")
    os.Exit(1)
  }
  
  encryptedFileNameExt := encryptedFileName[len(encryptedFileName) - 4:len(encryptedFileName)]
  if encryptedFileNameExt != EXT_ENC {
    fmt.Println("The encrypted file '" + encryptedFileName + "' does not have extension '" + EXT_ENC + "'.")
    os.Exit(1)
  }
  
  encryptedFile, err := os.Open(encryptedFileName)
  check(err)
  defer encryptedFile.Close()
  
  encryptedFileInfo, err := encryptedFile.Stat()
  check(err)
  
  encryptedBuf := make([]byte, BUFFER_SIZE)
  padBuf := make([]byte, BUFFER_SIZE)
  normalBuf := make([]byte, BUFFER_SIZE)
  
  normalFileName := encryptedFileName[0:len(encryptedFileName) - 4]
  normalFile, err := os.Create(normalFileName)
  check(err)
  defer normalFile.Close()
  
  padFile, err := os.Open(padFileName)
  check(err)
  defer padFile.Close()
  
  padFileInfo, err := padFile.Stat()
  check(err)
  
  if encryptedFileInfo.Size() != padFileInfo.Size() {
    fmt.Println("Size mismatch between encrypted file and pad file.")
    os.Exit(1)
  }
  
  numBytes, err := encryptedFile.Read(encryptedBuf);
  for numBytes > 0 {
    check(err)
    encryptedSlice := encryptedBuf[0:numBytes]
    
    numBytes, err = padFile.Read(padBuf);
    check(err)
    padSlice := padBuf[0:numBytes]
    
    normalSlice := normalBuf[0:numBytes]
    
    for i, encryptedByte := range encryptedSlice {
      normalSlice[i] = encryptedByte ^ padSlice[i]
    }
    
    _, err = normalFile.Write(normalSlice)
    check(err)
    
    numBytes, err = encryptedFile.Read(encryptedBuf);
  }
}

func check (e error) {
  if e != nil {
    panic(e)
  }
}

func checkFileExists(file string) bool {
  fi, err := os.Stat(file)
  if err != nil {
    fmt.Println(err)
    return false
  }
  
  return fi.Mode().IsRegular()
}

func printHelp() {
  helpStr := `
  Tool for one-time pad encryption/decryption.
  
    http://en.wikipedia.org/wiki/One-time_pad
  
  Usage:
  
  {progname} encrypt <file>
    
    The original file content will be encrypted by using the XOR one-time
    pad algorithm which is mathematically unbreakable. The only way to
    decrypt the file is to use the same algorithm and provide the generated
    pad file which contains random bytes. You need both the encrypted file
    and the pad file in order to reverse the encryption. If you lose one of
    the files, you can never restore the original content of the file.
    
    Example:
    
      {progname} encrypt important.txt
      
    The text file 'important.txt' will be encrypted and the following files
    will be present in the current directory:
    
      important.txt     - the original file
      important.txt.enc - the encrypted file
      important.txt.pad - the pad file      
  
  {progname} decrypt <file> <pad>
  
    The encrypted file will be decrypted by performing XOR on the file with
    the pad file. The decrypted file name is derived from the encrypted file
    name by removing the '.enc' extension.
    
    Example:
    
      {progname} decrypt important.txt.enc important.txt.pad
      
    The encrypted file 'important.txt.enc' will be decrypted and the following
    files will be present in the current directory:
    
      important.txt.enc - the encrypted file
      important.txt.pad - the pad file      
      important.txt     - the decrypted file    
`

  replacer := strings.NewReplacer("{progname}", os.Args[0])
  fmt.Println(replacer.Replace(helpStr))
}
