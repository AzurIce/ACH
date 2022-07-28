package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"
)

// GetTimeStamp ...
func GetTimeStamp() string {
	return time.Now().Format("2006-01-02 15-04-05")
}

// CopyFile ...
func CopyFile(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// DeletePath
func DeletePath(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	return nil
}

// CopyDir ...
func CopyDir(srcDir string, dstDir string) error {
	// fmt.Print([]byte(dstDir))
	err := os.Mkdir(dstDir, 0666)
	if err != nil {
		log.Println(err)
	}
	fileInfoList, _ := ioutil.ReadDir(srcDir)
	for i := 0; i < len(fileInfoList); i++ {
		// fmt.Println("Copying: ", fileInfoList[i].Name(), fileInfoList[i].IsDir(), "...")
		if fileInfoList[i].IsDir() {
			CopyDir(path.Join(srcDir, fileInfoList[i].Name()), path.Join(dstDir, fileInfoList[i].Name()))
		} else {
			CopyFile(path.Join(srcDir, fileInfoList[i].Name()), path.Join(dstDir, fileInfoList[i].Name()))
		}
	}
	return nil
}

func ForwardStd(f io.ReadCloser, c chan string) {
	// for {
	// 	log.Print("1")
	// }
	defer func() {
		recover()
	}()
	cache := ""
	buf := make([]byte, 1024)
	for {
		num, err := f.Read(buf)
		if err != nil && err != io.EOF { // 非EOF错误
			log.Panicln(err)
		}
		if num > 0 {
			str := cache + string(buf[:num])
			lines := strings.SplitAfter(str, "\n") // 按行分割
			for i := 0; i < len(lines)-1; i++ {
				c <- lines[i]
			}
			cache = lines[len(lines)-1] // 最后一行下次循环处理
		}
	}
}

func RandStringRunes(n int) string {
	rand.Seed(time.Now().Unix())
	var letterRunes = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func EncodePassword(password string, salt string) string {
	//计算 Salt 和密码组合的SHA1摘要
	hash := sha1.New()
	hash.Write([]byte(password + salt))
	bs := hex.EncodeToString(hash.Sum(nil))

	//存储 Salt 值和摘要， ":"分割
	return salt + ":" + string(bs)
}