package rpc

import (
	"context"
	"fmt"
	"net"

	pb "github.com/joshjms/pocket-watch/internal/judge"
	"github.com/joshjms/pocket-watch/src/consts"
	"github.com/joshjms/pocket-watch/src/isolate"
	"github.com/joshjms/pocket-watch/src/models"
	"google.golang.org/grpc"
)

type judgeServer struct {
	pb.UnimplementedJudgeServer
}

func newServer() *judgeServer {
	return &judgeServer{}
}

func (*judgeServer) Judge(ctx context.Context, req *pb.JudgeRequest) (*pb.JudgeResponse, error) {
	request := &models.Request{
		Code:     req.Code,
		Language: req.Language,
		Stdin:    req.Input,
	}

	instance, err := isolate.CreateInstance(*request)
	if err != nil {
		return nil, err
	}

	var resp pb.JudgeResponse

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

	pb.RegisterJudgeServer(grpcServer, newServer())

	go grpcServer.Serve(lis)

	return nil
}
