FROM golang:alpine AS builder

# 为镜像设置必要的环境变量
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 移动到工作目录：/build
WORKDIR /build

# 下载依赖信息
COPY go.mod .
COPY go.sum .
RUN go mod download

# 将代码复制到容器中
COPY . .

# 将我们的代码编译成二进制可执行文件 islet_app
RUN go build -o islet_app .

###################
# 创建一个小镜像
###################
FROM scratch

# 从builder镜像中把静态文件拷贝到当前目录
COPY ./wait-for.sh /
COPY ./templates /templates
COPY ./static /static

# 从builder镜像中把配置文件拷贝到当前目录
COPY ./conf /conf

# 从builder镜像中把/dist/app 拷贝到当前目录
COPY --from=builder /build/islet_app /

# 需要运行的命令
#ENTRYPOINT ["/islet_app", "conf/config.yaml"]

RUN set -eux; \
	apt-get update; \
	apt-get install -y \
		--no-install-recommends \
		netcat; \
        chmod 755 wait-for.sh

# 声明服务端口（只是声明，起提示作用）
EXPOSE 8888

# 生成镜像
# docker build . -t islet_app

# 运行镜像
# docker run -p 8081:8888 islet_app
# [本地端口]:[docker 端口]

# 绑定端口
# docker run --link=mysql830:mysql830 -p 8888:8888 islet_app
