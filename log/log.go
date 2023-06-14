package middleware_log

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UnaryServerInterceptor(logger *log.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, _ := metadata.FromIncomingContext(ctx)
		host := getStrignFromMetadata(md, "x-real-ip")
		if len(host) == 0 {
			host = getStrignFromMetadata(md, "x-forwarded-for")
		}
		if len(host) > 0 {
			strs := strings.Split(host, ",")
			host = strs[0]
		}
		remoteUser := ""
		timeLocal := time.Now().Format("02/Jan/2006:15:04:05 +0000")
		method := "GRPC"
		path := info.FullMethod
		protol := "GRPC"
		reqTime := float64(time.Now().UnixNano())
		reqBody, _ := json.Marshal(req)
		contentLength := getStrignFromMetadata(md, "content-length")
		respTime := float64(time.Now().UnixNano())
		refer := getStrignFromMetadata(md, "referer")
		agent := getStrignFromMetadata(md, "gateway-user-agent")
		if agent == "" {
			agent = getStrignFromMetadata(md, "user-agent")
		}
		userId := getStrignFromMetadata(md, "x-authenticated-userid")
		scope := getStrignFromMetadata(md, "x-authenticated-scope")

		resp, err = handler(ctx, req)
		httpStatus := 200
		if err != nil {
			httpStatus = 400
		}
		bodySent, _ := json.Marshal(resp)

		logger.Printf("\"%s\" - %s [%s] \"%s %s %s\" %f %s %d %d \"%s\" \"%s\" \"%s\" \"%s\" \"%s\"\n",
			host,
			remoteUser,
			timeLocal,
			method,
			path,
			protol,
			(respTime-reqTime)/1e9,
			contentLength,
			httpStatus,
			len(bodySent),
			refer,
			agent,
			userId,
			scope,
			strings.ReplaceAll(string(reqBody), "\"", ""),
		)
		return
	}
}

func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		err = handler(srv, ss)
		return err
	}
}

func getStrignFromMetadata(md metadata.MD, key string) string {
	strs := md.Get(key)
	if len(strs) > 0 {
		return strs[0]
	}
	return ""
}
