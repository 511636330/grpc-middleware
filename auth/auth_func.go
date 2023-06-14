package middleware_auth

import (
	"context"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/spf13/cast"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func CheckAuth(ctx context.Context, method string, permissions map[string]GroupAuth) error {
	if env := os.Getenv("APP_ENV"); env == "development" {
		return nil
	}
	var (
		group         string
		action        string
		consumer      string
		scope         string
		requestScopes []string
	)
	method = strings.Replace(method, "/proto.", "", 1)
	path := strings.Split(method, "/")
	if len(path) == 1 {
		group = path[0]
	} else if len(path) >= 2 {
		group = path[0]
		action = path[1]
	} else {
		return status.Errorf(codes.InvalidArgument, "invalidRequest")
	}

	md, _ := metadata.FromIncomingContext(ctx)
	consumers := md.Get("x-consumer-username")
	scopes := md.Get("X-Authenticated-Scope")

	if len(consumers) > 0 {
		consumer = consumers[0]
	}
	if len(scopes) > 0 {
		scope = scopes[0]
		spaceRe, _ := regexp.Compile(`\s+`)
		requestScopes = spaceRe.Split(scope, -1)
	}
	unauthenticatedException := status.Errorf(codes.Unauthenticated, "Unauthenticated")

	// check group consumer
	if requredConsumer := permissions[group].Consumer; len(requredConsumer) > 0 && requredConsumer != consumer {
		return unauthenticatedException
	}

	// checkt group scopes
	if requiredScopes := permissions[group].Scopes; len(requiredScopes) > 0 {
		if intersectedScopes := Intersect(requestScopes, requiredScopes); len(intersectedScopes) == 0 {
			return unauthenticatedException
		}
	}

	// check action consumer
	if requredConsumer := permissions[group].Actions[action].Consumer; len(requredConsumer) > 0 && requredConsumer != consumer {
		return unauthenticatedException
	}

	// check action scopes
	if requiredScopes := permissions[group].Actions[action].Scopes; len(requiredScopes) > 0 {
		if intersectedScopes := Intersect(requestScopes, requiredScopes); len(intersectedScopes) == 0 {
			return unauthenticatedException
		}
	}

	return nil
}

func Intersect(a, b interface{}) []interface{} {
	set := make([]interface{}, 0)
	hash := make(map[interface{}]bool)
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	for i := 0; i < av.Len(); i++ {
		el := av.Index(i).Interface()
		hash[el] = true
	}

	for i := 0; i < bv.Len(); i++ {
		el := bv.Index(i).Interface()
		if _, found := hash[el]; found {
			set = append(set, el)
		}
	}

	return set
}

func GetRequestUser(ctx context.Context) (userId, parentId int) {
	var requestUser string
	md, _ := metadata.FromIncomingContext(ctx)
	requestUsers := md.Get("X-Authenticated-Userid")

	if len(requestUsers) > 0 {
		requestUser = requestUsers[0]
	}
	reg := regexp.MustCompile(":")

	userInfo := reg.Split(requestUser, -1)
	userId = cast.ToInt(userInfo[0])

	if len(userInfo) > 1 {
		parentId = cast.ToInt(userInfo[1])
	} else {
		parentId = userId
	}

	return

}
