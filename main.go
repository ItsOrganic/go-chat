package main

import (
	"context"
	"log"
	"os"
	"sync"
    "net"

    "go-chat/proto"
	glog "google.golang.org/grpc/grpclog"
	grpc "google.golang.org/grpc"
)

var grpcLog glog.LoggerV2

func init() {
    grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

type Connection struct {
    stream proto.Broadcast_CreateStreamServer
    id string
    active bool
    error chan error
}

type Server struct {
    Connection []*Connection
}

func (s *Server) CreateStream(pconn *proto.Connect, stream proto.Broadcast_CreateStreamServer) error {
    conn := &Connection{
        stream: stream,
        id: pconn.User.Id,
        active: true,
        error: make(chan error),
    }
    s.Connection = append(s.Connection, conn)
    return <- conn.error
}

func (s *Server) BroadcastMessage(ctx context.Context, msg *proto.Message) (*proto.Close, error) {
    wait := sync.WaitGroup{}
    done := make(chan int)
    for _, conn := range s.Connection {
        wait.Add(1)
    go func(msg *proto.Message, conn *Connection) {
        if conn.active {
            err := conn.stream.Send(msg)
            grpcLog.Info("Sending message to:", conn.stream)

            if err != nil {
                grpcLog.Errorf("Error with stream %s", err)
                conn.active = false
                conn.error <- err
            }
        }
    }(msg,conn)
}
go func() {
    wait.Wait()
    close(done)
}()
    <-done
    return &proto.Close{},nil
}

func main() {

    var connections []*Connection

    server := &Server{connections}

    grpcServer := grpc.NewServer()
    listener, err := net.Listen("tcp", ":8080")

    if err != nil {
        log.Fatal("Error creating server", err)
    }

    grpcLog.Info("Starting server at prot :8080")
    proto.RegisterBroadcastServer(grpcServer, server)

    grpcServer.Serve(listener)

}
