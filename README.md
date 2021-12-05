# file-server
golang demo

### 编译
`go build`

### 配置文件
`config.yml` 放在可执行文件同一目录

```
# 数据库配置
db:
  name:     test
  user:     postgres
  password: postgres
  host:     localhost
  port:     5432
  # 是否使用数据库 默认false
  use:      false

server:
  port: 8084
  # 图片存放位置，默认 ./tmp
  baseFilePath: C:/Users/PC/Desktop/tmp/files
```

### 上传示例
```
curl --location --request POST 'http://localhost:8084/upload' \
--form 'file=@"C:\\Users\\PC\\Pictures\\1000.png"'
```
上传成功会返回图片链接，也可以用nginx自定义设置文件服务器
