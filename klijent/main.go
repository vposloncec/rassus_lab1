package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	pb "github.com/vposloncec/rassus_lab1/klijent/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

type grpcServer struct {
	pb.UnimplementedSensorServer
	currTime time.Time
	readings []*pb.Reading
}

func (s *grpcServer) GetReading(context.Context, *pb.GetReadingRequest) (*pb.Reading, error) {
	return makeReading(s.currTime, s.readings), nil
}

type RegisterRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Ip        string  `json:"ip"`
	Port      int     `json:"port"`
}
type RegisterResponse struct {
	Latitude  float64
	Longitude float64
	Ip        string
	Port      int
	Id        int
}

func main() {
	start := time.Now()
	rawLog, _ := zap.NewDevelopment()
	readings := readCsvFile()
	log := rawLog.Sugar()
	client := resty.New()
	rand.Seed(start.UnixNano())
	//client.Debug = true
	// Set up a connection to the server.
	registerReq := RegisterRequest{
		Latitude:  15.87 + rand.Float64()*(16-15.86),
		Longitude: 45.75 + rand.Float64()*(45.85-45.75),
		Ip:        "127.0.0.1",
		Port:      10000 + rand.Intn(10000),
	}

	jsonReq, err := json.Marshal(registerReq)
	log.Infof("%v", string(jsonReq))
	if err != nil {
		log.Fatal("could not marshall register req")
	}
	r, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(jsonReq).
		SetResult(&RegisterResponse{}).
		Post("http://127.0.0.1:3000/register")
	if err != nil {
		log.Fatal("could not connect to centralized server: ", err)
	}
	// log.Info("Got response: ", string(r.Body()))
	if r.IsSuccess() {
		log.Infof("Successfully registered with status %v; server response: %v", r.Status(),
			string(r.Body()))
	} else {
		log.Errorf("Could not register, response status: %v", r.Status())
	}

	go startGrpcServer(start, strconv.Itoa(registerReq.Port))

	//Get neighbour
	myId := r.Result().(*RegisterResponse).Id
	r, err = client.R().
		SetResult(&RegisterResponse{}).
		Get("http://127.0.0.1:3000/sensors/" + strconv.Itoa(myId) + "/nearest")

	neighbour := r.Result().(*RegisterResponse)
	if neighbour == nil {
		log.Infof("No neighbour! skipping grpc calls...")
	}
	log.Infof("My neighbour has id %v, calling him on ip: %v, port: %v", neighbour.Id, neighbour.Ip, neighbour.Port)

	neighbourAddr := neighbour.Ip + ":" + strconv.Itoa(neighbour.Port)
	currentReading := makeReading(start, readings)
	log.Info("Current reading: ", currentReading)

	// Initialize grpc client
	conn, err := grpc.Dial(neighbourAddr, grpc.WithInsecure())
	c := pb.NewSensorClient(conn)
	if err != nil {
		log.Fatalf("grpc did not connect: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	neighbourReading, err := c.GetReading(ctx, &pb.GetReadingRequest{})
	if err != nil {
		log.Errorf("error while contacting neighbour: %v", err)
	} else {
		log.Infof("Got reading from neighbour %v", neighbourReading)
	}
	//if err != nil {
	defer conn.Close()

	//// Contact the server and print out its response.
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	//r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
	//if err != nil {
	//	log.Fatalf("could not greet: %v", err)
	//}
	//log.Printf("Greeting: %s", r.GetMessage())
}

func startGrpcServer(startTime time.Time, port string) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSensorServer(s, &grpcServer{
		currTime: startTime,
		readings: readCsvFile(),
	})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func makeReading(current time.Time, readings []*pb.Reading) *pb.Reading {
	elapsed := time.Since(current)
	return readings[(int(elapsed.Seconds()) % 100)]
}

func readCsvFile() []*pb.Reading {
	csvFile, _ := os.Open("readings.csv")
	csvLines, _ := csv.NewReader(csvFile).ReadAll()
	var readings []*pb.Reading
	for _, line := range csvLines {
		temp, _ := strconv.ParseFloat(line[0], 32)
		pressure, _ := strconv.ParseFloat(line[1], 32)
		h, _ := strconv.ParseFloat(line[2], 32)
		co, _ := strconv.ParseFloat(line[3], 32)
		so2, _ := strconv.ParseFloat(line[4], 32)
		reading := &pb.Reading{
			Temperature: float32(temp),
			Pressure:    float32(pressure),
			Humidity:    float32(h),
			Co:          float32(co),
			So2:         float32(so2),
		}
		readings = append(readings, reading)
	}
	return readings
}
