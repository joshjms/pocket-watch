package rpc

import (
	"context"
	"fmt"
	"net"

	pb "github.com/joshjms/pocket-watch/internal/watch"
	"github.com/joshjms/pocket-watch/src/consts"
	"github.com/joshjms/pocket-watch/src/isolate"
	"github.com/joshjms/pocket-watch/src/models"
	"google.golang.org/grpc"
)

type watchServer struct {
	pb.UnimplementedWatchServer
}

func newServer() *watchServer {
	return &watchServer{}
}

func (*watchServer) Run(ctx context.Context, req *pb.WatchRequest) (*pb.WatchResponse, error) {
	request := &models.Request{
		Code:     req.Code,
		Language: req.Language,
		Stdin:    req.Input,
	}

	instance, err := isolate.CreateInstance(*request)
	if err != nil {
		return nil, err
	}

	var resp pb.WatchResponse

	var int32Time []int32 = []int32{}
	var int32Memory []int32 = []int32{}

	for _, time := range instance.Response.Time {
		int32Memory = append(int32Time, int32(time))
	}

	for _, memory := range instance.Response.Memory {
		int32Time = append(int32Memory, int32(memory))
	}

	resp.Verdict = instance.Response.Verdict
	resp.Stdout = instance.Response.Stdout
	resp.Stderr = instance.Response.Stderr
	resp.Time = int32Time
	resp.Memory = int32Memory

	return &resp, nil
}

func StartServer() error {
	config := consts.GetConsts().RPCConfig
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		return err
	}

	var opts []grpc.ServerOption
	// allows inse

	grpcServer := grpc.NewServer(opts...)

	pb.RegisterWatchServer(grpcServer, newServer())

	go grpcServer.Serve(lis)

	return nil
}
