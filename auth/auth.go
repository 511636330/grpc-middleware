package middleware_auth

import (
	"context"

	middleware "gitlab.qkids.com/group-api-common/grpc-middleware.git"
	"google.golang.org/grpc"
)

type ActionAuth struct {
	Consumer string
	Scopes   []string
}

type GroupAuth struct {
	Consumer string
	Scopes   []string
	Actions  map[string]ActionAuth
}

func UnaryServerInterceptor(permissions map[string]GroupAuth) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		err = CheckAuth(ctx, info.FullMethod, permissions)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func StreamServerInterceptor(permissions map[string]GroupAuth) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		err := CheckAuth(ctx, info.FullMethod, permissions)
		if err != nil {
			return err
		}
		wrapped := middleware.WrapServerStream(ss)
		wrapped.WrappedContext = ctx
		return handler(srv, wrapped)
	}
}
