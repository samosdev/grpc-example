package main

import (
	"context"
	"net/http"
	"os"
	"time"

	pb "github.com/epa-datos/grpc-example/protos"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type handlerService struct {
	client pb.GenshinClient
}

func NewHandlerServer() (*handlerService, error) {

	creds, err := credentials.NewClientTLSFromFile(os.Getenv("CERT_FILE"), "")
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial("localhost:8083", grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}

	return &handlerService{
		client: pb.NewGenshinClient(conn),
	}, nil
}

func (h *handlerService) getCharacterInfoThroughGRPC(name string) (*pb.CharacterReply, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := h.client.GetCharacterInfo(ctx, &pb.CharacterRequest{Name: name})
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (h *handlerService) getCharacterHandler(ctx *gin.Context) {
	name := ctx.Param("name")
	response, err := h.getCharacterInfoThroughGRPC(name)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "Genshin character " + name + " not found",
			"detail":  err,
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *handlerService) getAllCharactersHandler(ctx *gin.Context) {

	gRPCCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	response, err := h.client.GetAllElementsFromType(gRPCCtx, &pb.TypeRequest{Type: "characters"})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Could not list Genshin characters",
		})
		return
	}

	ctx.JSON(http.StatusOK, response.Elements)

}

func main() {

	handlerServ, err := NewHandlerServer()
	if err != nil {
		logrus.Fatal("Could not connect with server", err)
	}
	r := gin.Default()
	genshinRouter := r.Group("/genshin/characters")
	genshinRouter.GET("", handlerServ.getAllCharactersHandler)
	genshinRouter.GET(":name", handlerServ.getCharacterHandler)

	r.Run(":8084")
}
