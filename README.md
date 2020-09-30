## 简介
该项目是一个公共模块，可以快速集成solace、mysql等组件，并提供一些常用工具库。

## Quick Start
1. 初始化一个go mod项目。执行：
```gotemplate
go mod init <project name>
```
2. append go replace into your `go.mod`, 注意：修改地址为你的本地项目位置
```shell script
replace (
	git.forms.io/universe/comm-agent => /Users/inori/go/src/git.forms.io/universe/comm-agent
	git.forms.io/universe/common => /Users/inori/go/src/git.forms.io/universe/common
	git.forms.io/universe/dts => /Users/inori/go/src/git.forms.io/universe/dts
	git.forms.io/universe/solapp-sdk => /Users/inori/go/src/git.forms.io/universe/solapp-sdk
)
```

2. 下载该包
```shell
go get github.com/yinjk/go-utils

```
3. 在main方法中编写
```go
func main() {
    engine := solace.Default()
 	engine.Register(handler.NewStatisticsHandler(config))
 	engine.Register(handler.NewHealthHandler(config))
 	engine.Register(handler.NewZipkinHandler(config.Zipkin.BaseUrl))
 	engine.ListenAndStartUp()
}
```
4. 目录结构

![项目目录结构](.README_images/8a34c0c5.png)