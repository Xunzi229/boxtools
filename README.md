# boxtools

个人工具集

## bmarks

#### Usage

1. 获取浏览器的书签, 支持  chrome edge yandex
```shell
$: go install ./tools/bmarks
$: bmarks # bmarks -b edge
```

## ssb

#### Usage

主要用于管理多ssh key的问题

* 安装

```shell
$: go install ./tools/ssb
```

* 生成的新的KEY
```shell
$: ssb g 
# 或者
$: ssb gen
```

* 备份当前的 key
```shell
$: ssb backup tagName
# 或者
$: ssb b tagName

#   UniqueId    TagName
#* 6fed5f86d8     home
#  6fed5f86d8     work
```

tagName: 用户恢复 或者切换配置的时候使用的

* 切换的备份

```shell
$: ssb switch tagName # ssb switch UniqueId
# 或者
$: ssb s tagName      # ssb switch TagName
```

* 导出备份文件

> 默认备份在主目录 $HOME

```shell
$: ssb p # 默认备份在主目录
$: ssb p ~/Desktop/ # 备份文件存在桌面
$: ssb export .     # 备份文件存在当前目录
```

* 恢复备份文件

```shell
$: ssb load ~/Desktop/backup.zip
```

## search and delete (sdl)

史上最快查找重复文件，删除文件

```shell
go install tools/sdl/sdl.go

sdl  
  -d   /home/xz/path 选择去重的目录, 绝对路劲
  --dl Y             删除重复的文件
  -f   .jpg          去重的文件后缀, 多个文件后缀选择使用逗号隔开(.jpg,.png), 不区分大小写
```

