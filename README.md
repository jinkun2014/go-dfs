# 基于Golang写的文件上传下载工具

## 配置
```
# 查看帮助
go run main.go -h

Godfs version: Godfs/1.0.0
Usage: Godfs [-h 帮助] [-d 设置域名] [-D 设置数据存储路径] [-p 设置端口] [-g 设置组名]
启动成功后,打开浏览器访问端口即可，例如:http://127.0.0.1:8080
Options:
  -D string
        设置数据存储路径 (default "data")
  -d string
        设置域名 (default "http://127.0.0.1")
  -g string
        设置组名
  -h    帮助
  -p int
        设置端口 (default 8080)

# 启动
go run main.go -p 9090 -D "D:\\temp" -d "http://127.0.0.1:9090"
```

## 打包
```
# 打包Windows(含压缩)
SET GOOS=windows
SET GOARCH=amd64
go build -o Godfs.exe -ldflags "-s -w" main.go

# 打包Linux版本(含压缩)
SET GOOS=linux
SET GOARCH=amd64
go build -o Godfs -ldflags "-s -w" main.go


# 压缩,其中  -ldflags 里的  -s 去掉符号信息， -w 去掉DWARF调试信息，得到的程序就不能用gdb调试了
go build -o Godfs.exe -ldflags "-s -w" main.go

# upx压缩 
upx -9 Godfs.exe
```

## API
```
# 文件上传
http://127.0.0.1:5050/upload
# 参数
file = 文件。。。

#返回
{
    "code": 200,
    "message": "Upload success",
    "data": {
        "remotePath": "http://127.0.0.1:9090/2019/12/04/1575447317445-r756s5.mp3",
        "domain": "http://127.0.0.1:9090",
        "path": "/2019/12/04/1575447317445-r756s5.mp3",
        "newName": "1575447317445-r756s5.mp3",
        "originName": "test.mp3",
        "size": 3089554
    }
}
```
