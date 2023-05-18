package main

import (
	"flag"
	"fmt"
	"math"
	"runtime"
	"time"
)

func main() {
	// 获得命令行参数
	path, chars, pattern, minLen, maxLen, err := getFlags()
	if err != nil {
		fmt.Printf("\033[1;31m应用参数失败: %s\033[0m\n", err)
		flag.PrintDefaults()
		return
	}
	listLen := countListLen(minLen, maxLen, len([]rune(chars))) // 计算字符组合数量
	cpuNum := runtime.NumCPU()                                  // 获取CPU线程数

	// 创建破解器
	cracker, err := NewCracker(path, chars, pattern, minLen, maxLen)
	if err != nil {
		fmt.Printf("\033[1;31m%s\033[0m\n", err)
		return
	}

	// 控制台输出准备信息
	fmt.Printf("CPU线程数: %d\n", cpuNum)
	fmt.Printf("字符集合: \033[4m%s\033[0m\n", chars)
	if pattern != "" {
		fmt.Printf("匹配正则: %s\n", cracker.regex)
	}
	fmt.Printf("\n位数范围 %d~%d\n", minLen, maxLen)
	fmt.Printf("组合总共 %d 种\n", listLen)

	// 生成所有字符组合并校检每个密码
	startTime := time.Now() // 记录启动时间
	go cracker.generatePasswordList()
	go cracker.testPassword(cpuNum)

	// 从结果通道读取结果
	password := <-cracker.resChan
	fmt.Printf("\n\n耗时 %s\n", time.Since(startTime)) // 输出所耗时间
	if password != "" {
		// 取得密码
		fmt.Printf("\033[1;36m密码是: \033[4m%s\033[0m\n", password)
	} else {
		// 未得密码
		fmt.Println("\033[1;31m未找到密码\033[0m")
	}
}

// 获取命令行参数
func getFlags() (path string, chars string, pattern string, minLen int, maxLen int, err error) {
	flag.StringVar(&path, "path", "", "(必需)Zip文件路径")
	flag.StringVar(&chars, "chars", "", "(必需)可能出现的字符集合, 例 -chars abc")
	flag.StringVar(&pattern, "pattern", "", "(可选)正则表达式, 例 -pattern \"aa|bb|cc\"")
	flag.IntVar(&minLen, "min", 1, "(可选)最小密码位数")
	flag.IntVar(&maxLen, "max", 0, "(可选)最大密码位数")
	flag.Parse()

	// 整理命令行参数
	if path == "" || chars == "" {
		err = fmt.Errorf("缺少参数[path]或[chars]")
		return
	}
	chars = uniqueChars(chars)
	if minLen < 1 {
		minLen = 1
	}
	if maxLen == 0 {
		maxLen = len([]rune(chars))
	}
	if minLen > maxLen {
		maxLen = minLen
	}
	return
}

// 去除重复字符
func uniqueChars(charSet string) (uniqueCharSet string) {
	charMap := make(map[rune]bool)
	for _, char := range charSet {
		charMap[char] = true
	}
	for char := range charMap {
		uniqueCharSet += string(char)
	}
	return
}

// 计算组合总数
func countListLen(minLen, maxLen, charCount int) int {
	count := 0
	for length := minLen; length <= maxLen; length++ {
		count += int(math.Pow(float64(charCount), float64(length)))
	}
	return count
}
