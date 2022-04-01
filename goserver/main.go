package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	pb "github.com/epa-datos/grpc-example/protos"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	GenshinAPI = "https://api.genshin.dev"
)

type server struct {
	pb.UnimplementedGenshinServer
}

func (s *server) GetCharacterInfo(ctx context.Context, input *pb.CharacterRequest) (*pb.CharacterReply, error) {

	response, err := http.Get(GenshinAPI + "/characters/" + input.Name)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	characterInfo := &pb.CharacterReply{}
	if err := json.Unmarshal(body, characterInfo); err != nil {
		logrus.Error(err)
		return nil, err
	}

	return characterInfo, nil
}

func (s *server) GetAllElementsFromType(ctx context.Context, input *pb.TypeRequest) (*pb.TypeListReply, error) {

	response, err := http.Get(GenshinAPI + "/" + input.Type)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	elements := []string{}
	if err := json.Unmarshal(body, &elements); err != nil {
		return nil, err
	}

	return &pb.TypeListReply{
		Elements: elements,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":8083")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterGenshinServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
