# GoZipCracker
使用Go语言暴力破解Zip加密。<br>
以输入的字符集合生成密码组合，可用正则匹配；<br>
通过打开文件的方式验证，效率极低。

依赖于 [https://github.com/yeka/zip](https://github.com/yeka/zip)
## 使用
在命令行中直接启动，
参数：
|			|		|	 |
|-			|-		|-	 |
|`-path`	|文件路径|必需|
|`-chars`	|字符集合|必需|
|`-pattern`	|匹配规则|
|`-min`		|最少位数|
|`-max`		|最多位数|

例：`ZipCracker -path test.zip -chars 不里许这u照r拍n -pattern '这里|run|许|拍照' -min 2 -max 9`
