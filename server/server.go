package server

import (
	"context"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"mytest/pkg/ui/data/swagger"
	"mytest/protoc/order"
	"net"
	"net/http"
	"path"
	"strings"
)

const (
	port = ":6666"
	httpPort =":9999"
)

type server struct {
	order.UnimplementedOrderServer
}

func (s *server) GetOrderInfo(ctx context.Context, req *order.GetOrderReq) (*order.GetOrderRsp, error) {
	log.Printf("req == %+v", req)

	return &order.GetOrderRsp{
		OrderName:   "dingdan1111",
		OrderInfo:   "buy a bag",
		Description: "test demo",
	}, nil
}

// grpc server
func RunGrpcSvr() {
	// 要监听的协议和端口
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 实例化gRPC server结构体
	s := grpc.NewServer()

	// 服务注册
	order.RegisterOrderServer(s, &server{})

	log.Println("Serving ...,grpc svr 0.0.0.0:6666")
	s.Serve(lis)

	return
}

// grpcgateway
func RunGrpcGw() {
	conn, err := grpc.DialContext(
		context.Background(),
		port,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}
	defer conn.Close()
	gwmux := runtime.NewServeMux()
	// Register order handler
	err = order.RegisterOrderHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	gwServer := &http.Server{
		Addr:    httpPort,
		Handler: gwmux,
	}

	log.Println("Serving gRPC-Gateway on http://0.0.0.0:9999")
	if err := gwServer.ListenAndServe(); err != nil {
		log.Fatalf("http.ListenAndServe err %v:", err)
	}

	return
}

// grpcgateway with swagger
func RunGrpcGwWithSwagger() {
	conn, err := grpc.DialContext(
		context.Background(),
		port,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}
	defer conn.Close()
	gwmux := runtime.NewServeMux()
	// Register order handler
	err = order.RegisterOrderHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", gwmux)
	mux.HandleFunc("/swagger/", svrSwaggerFile)
	svrSwaggerUI(mux)

	log.Println("grpc-gateway listen on localhost:9999")
	if err := http.ListenAndServe(httpPort, mux); err != nil {
		log.Fatalf("http.ListenAndServe err %v:", err)
	}

	return
}

// swagger handlefunc
func svrSwaggerFile(w http.ResponseWriter, r *http.Request) {
	log.Println("start svrSwaggerFile")

	// 校验后缀是否是 "swagger.json"
	if !strings.HasSuffix(r.URL.Path, "swagger.json") {
		log.Printf("Not Found: %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	// 去掉前缀 "/swagger/"
	p := strings.TrimPrefix(r.URL.Path, "/swagger/")
	// 获取 json 文件 p，并加上相对路径 ./order.swagger.json
	p = path.Join("./protoc/order/", p)

	log.Printf("Serving swagger-file path : %s", p)

	http.ServeFile(w, r, p)

	return
}

// swagger ui handler
func svrSwaggerUI(mux *http.ServeMux) {
	fileServer := http.FileServer(&assetfs.AssetFS{
		Asset:    swagger.Asset,
		AssetDir: swagger.AssetDir,
		Prefix:   "third_party/swagger-ui",
	})
	prefix := "/swagger-ui/"
	mux.Handle(prefix, http.StripPrefix(prefix, fileServer))

	return
}
