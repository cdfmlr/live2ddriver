FROM golang:1.19.7-alpine3.16 AS builder

WORKDIR /app

# dependencies

COPY go.mod go.sum ./

RUN go env -w GOPROXY=https://goproxy.cn,direct && \
    go mod download

# build

COPY . .

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o live2ddriver.out .

# scratch 似乎每次都会遇到问题，还是 alpine 吧。
FROM alpine:3.16 AS runner

COPY --chown=0:0 --from=builder /app/live2ddriver.out /app/live2ddriver.out

CMD ["/app/live2ddriver.out", \
     "-wsAddr", "0.0.0.0:9001", \
     "-httpAddr", "0.0.0.0:9002", \
     "-shizuku", "0.0.0.0:9004", \
     "-verbose"]

