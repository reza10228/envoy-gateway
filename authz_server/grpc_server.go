package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"main/build"
	"main/database"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	rcontext "context"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/gogo/googleapis/google/rpc"
)

var (
	grpcport = flag.String("grpcport", ":50051", "grpcport")
	conn     *grpc.ClientConn
	hs       *health.Server
)

const (
	address string = ":50051"
)

var rctx = rcontext.Background()

type healthServer struct{}

func (s *healthServer) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	log.Printf("Handling grpc Check request")

	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (s *healthServer) Watch(in *healthpb.HealthCheckRequest, srv healthpb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

type AuthorizationServer struct{}

func (a *AuthorizationServer) Check(ctx context.Context, req *auth.CheckRequest) (*auth.CheckResponse, error) {

	splitPath := strings.Split(req.Attributes.Request.Http.Path, "/")
	re := regexp.MustCompile(`\/([0-9\+]+)`)
	replaceString := "/{" + splitPath[3] + "}"
	method := req.Attributes.Request.Http.Method
	qrPath := StripQueryString(req.Attributes.Request.Http.Path)
	path := strings.Split(re.ReplaceAllString(qrPath, replaceString), "api/v1/")
	rPath := fmt.Sprintf("%s/%s", method, path[1])
	redis, rerr := RedisConnection()

	if rerr != nil {

	}

	b, err := json.MarshalIndent(req.Attributes.Request.Http.Headers, "", "  ")
	if err == nil {
		log.Println("Inbound Headers: ")
		log.Println((string(b)))
	}

	ct, err := json.MarshalIndent(req.Attributes.ContextExtensions, "", "  ")
	if err == nil {
		log.Println("Context Extensions: ")
		log.Println((string(ct)))
	}

	authHeaderAuthorization, okAuthorization := req.Attributes.Request.Http.Headers["authorization"]
	authHeaderApikey, okApikey := req.Attributes.Request.Http.Headers["apikey"]

	var splitToken []string
	ids := []uint64{205868,369586,392722}
	if okAuthorization {
		splitToken = strings.Split(authHeaderAuthorization, "Bearer ")
	}
	if len(splitToken) == 2 {
		var userData map[string]interface{}
		token := splitToken[1]
		res := redis.Get(rctx, token)
		user, _ := res.Result()
		uErr := json.Unmarshal([]byte(user), &userData)
		if uErr != nil {
			return unauthorizedResponse("Invalid Authorization Token"), nil
		}
		//_, cerr := jwt.TokenValidate(token)
		if res.Err() == nil && uErr == nil {
			userId := fmt.Sprintf("%v", userData["old_user_id"])
			parentId := fmt.Sprintf("%v", userData["parent_id"])
			aclId := fmt.Sprintf("%v", userData["old_acl_id"])
			roleName := fmt.Sprintf("role%v", aclId)
			b := redis.HExists(rctx, roleName, rPath)
			if !b.Val() {
				return unauthorizedResponse("You cant Access this route"), nil
			}
			return successResponse(userId, parentId, aclId), nil
		} else {
			return unauthorizedResponse("PERMISSION_DENIED"), nil
		}

	}
	if okApikey {
		key := redis.HGetAll(rctx, authHeaderApikey)
		keyCheck, _ := key.Result()
	
		if len(keyCheck) == 0 {
			result, err := database.DB.CheckKey(authHeaderApikey)
			if err == nil && !result.Revoked {
				roleName := fmt.Sprintf("role%d", result.User.AclRoleId)
				b := redis.HExists(rctx, roleName, rPath)
				if contains(ids,result.UserId){
					return successResponse(strconv.FormatUint(result.UserId, 10), strconv.FormatUint(result.User.Parent, 10), strconv.FormatUint(result.User.AclRoleId, 10)), nil
				}
				if !b.Val() {
					return unauthorizedResponse("You cant Access this route"), nil
				}	
				return successResponse(strconv.FormatUint(result.UserId, 10), strconv.FormatUint(result.User.Parent, 10), strconv.FormatUint(result.User.AclRoleId, 10)), nil
			} else {
				return unauthorizedResponse("PERMISSION_DENIED"), nil

			}
		} else {
			key, _ := key.Result()
			roleName := fmt.Sprintf("role%s", key["acl_id"])
			b := redis.HExists(rctx, roleName, rPath)
			i, _ := strconv.ParseUint(key["user_id"], 10, 64)

			if contains(ids,i){
				return successResponse(key["user_id"], key["parent_id"], key["acl_id"]), nil
			}
			if !b.Val() {
				return unauthorizedResponse("You cant Access this route"), nil
			}
			return successResponse(key["user_id"], key["parent_id"], key["acl_id"]), nil

		}

	}
	return unauthorizedResponse("Authorization Header malformed or not provided"), nil
}

func main() {

	flag.Parse()
	build.LoadConfig(".")
	database.ConnectDb()

	if *grpcport == "" {
		fmt.Fprintln(os.Stderr, "missing -grpcport flag (:50051)")
		flag.Usage()
		os.Exit(2)
	}

	lis, err := net.Listen("tcp", *grpcport)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{grpc.MaxConcurrentStreams(10)}
	opts = append(opts)

	s := grpc.NewServer(opts...)

	auth.RegisterAuthorizationServer(s, &AuthorizationServer{})
	healthpb.RegisterHealthServer(s, &healthServer{})

	log.Printf("Starting gRPC Server at %s", *grpcport)
	s.Serve(lis)

}

func successResponse(userId string, parentId string, roleId string) *auth.CheckResponse {
	return &auth.CheckResponse{
		Status: &rpcstatus.Status{
			Code: int32(rpc.OK),
		},
		HttpResponse: &auth.CheckResponse_OkResponse{
			OkResponse: &auth.OkHttpResponse{
				Headers: []*core.HeaderValueOption{
					{
						Header: &core.HeaderValue{
							Key:   "x-consumer-custom-id",
							Value: userId,
						},
					},
					{
						Header: &core.HeaderValue{
							Key:   "x-parent-id",
							Value: parentId,
						},
					},
				},
			},
		},
	}
}

func unauthorizedResponse(msg string) *auth.CheckResponse {
	return &auth.CheckResponse{
		Status: &rpcstatus.Status{
			Code: int32(rpc.UNAUTHENTICATED),
		},
		HttpResponse: &auth.CheckResponse_DeniedResponse{
			DeniedResponse: &auth.DeniedHttpResponse{
				Status: &envoy_type.HttpStatus{
					Code: envoy_type.StatusCode_Unauthorized,
				},
				Body: msg,
			},
		},
	}
}

func RedisConnection() (conn *redis.Client, error error) {
	address := fmt.Sprintf("%s:%s", build.Config.RedisHost, build.Config.RedisPort)

	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: build.Config.RedisPassword,
		DB:       0,
	})
	err := rdb.Set(rctx, "key", "value", 0).Err()
	return rdb, err
}
func StripQueryString(inputUrl string) string {
	u, err := url.Parse(inputUrl)
	if err != nil {
		panic(err)
	}
	u.RawQuery = ""
	return u.String()
}
func contains(ids []uint64, id uint64) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}

	return false
}
