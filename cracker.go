package main

import (
	"fmt"
	"io"
	"regexp"
	"sync"

	"github.com/yeka/zip"
)

type Cracker struct {
	pwdChan chan string
	resChan chan string // 结果通道
	doneNum int
	skipNum int
	charSet []rune
	regex   *regexp.Regexp
	minLen  int
	maxLen  int
	zipFile *zip.ReadCloser
}

func NewCracker(path string, chars string, pattern string, minLen int, maxLen int) (cracker Cracker, err error) {
	// 编译正则表达式
	regex, err := regexp.Compile(pattern)
	if err != nil {
		err = fmt.Errorf("正则编译失败: %s", err)
		return
	}
	// 打开zip文件
	zipFile, err := zip.OpenReader(path)
	if err != nil {
		err = fmt.Errorf("文件打开失败: %s", err)
		return
	}
	cracker = Cracker{
		pwdChan: make(chan string),
		resChan: make(chan string),
		doneNum: 0,
		skipNum: 0,
		charSet: []rune(chars),
		regex:   regex,
		minLen:  minLen,
		maxLen:  maxLen,
		zipFile: zipFile,
	}
	return
}

// 生成指定长度范围的所有字符组合
func (cracker *Cracker) generatePasswordList() {
	defer close(cracker.pwdChan)
	var wg sync.WaitGroup
	wg.Add(cracker.maxLen - cracker.minLen + 1)
	for i := cracker.minLen; i <= cracker.maxLen; i++ {
		go cracker.generatePasswords(i, &wg)
	}
	wg.Wait()
}

// 生成指定长度的所有字符组合
func (cracker *Cracker) generatePasswords(length int, wg *sync.WaitGroup) {
	defer wg.Done()
	passwordRunes := make([]rune, length)
	for i := range passwordRunes {
		passwordRunes[i] = cracker.charSet[0]
	}

	for {
		password := string(passwordRunes)
		if !cracker.regex.MatchString(password) {
			cracker.skipNum++
		} else {
			cracker.pwdChan <- password
		}
		i := length - 1
		for i >= 0 && passwordRunes[i] == cracker.charSet[len(cracker.charSet)-1] {
			passwordRunes[i] = cracker.charSet[0]
			i--
		}
		if i < 0 {
			break
		}
		passwordRunes[i] = cracker.charSet[getIndex(cracker.charSet, passwordRunes[i])+1]
	}
}
func getIndex(charSet []rune, char rune) int {
	for i, c := range charSet {
		if c == char {
			return i
		}
	}
	return -1
}

// 尝试用生成的密码解压文件
// 启动numWorkers个goroutine
func (cracker *Cracker) testPassword(numWorkers int) {
	defer cracker.zipFile.Close()
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for password := range cracker.pwdChan {
				// 检查通道是否已关闭
				select {
				case <-cracker.resChan:
					return
				default:
				}
				cracker.doneNum++
				cracker.printProgress()
				// 遍历文件并尝试用密码打开
				for _, f := range cracker.zipFile.File {
					f.SetPassword(password)
					rc, err := f.Open()
					if err != nil {
						// 打开失败，密码错误
						break
					}
					// 读取一些字节，检查解密数据
					testBuf := make([]byte, 10)
					_, err = io.ReadFull(rc, testBuf)
					rc.Close()
					if err != nil {
						// 读取失败，密码错误
						break
					}
					// 解压成功，密码正确，传出密码
					cracker.resChan <- password
					close(cracker.resChan)
					cracker.printProgress()
					return
				}
				// 所有文件皆无法解密，密码不正确
			}
		}()
	}
	wg.Wait()
	cracker.printProgress()
	// 确认关闭通道
	select {
	case <-cracker.resChan:
		return
	default:
	}
	close(cracker.resChan)
}

// 控制台输出进度
func (cracker *Cracker) printProgress() {
	if cracker.skipNum > 0 {
		fmt.Printf("\r已尝试 %d 个, 跳过 %d 个", cracker.doneNum, cracker.skipNum)
	} else {
		fmt.Printf("\r已尝试 %d 个", cracker.doneNum)
	}
}
