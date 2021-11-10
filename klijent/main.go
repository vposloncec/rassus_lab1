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
	logger   *zap.SugaredLogger
}

func (s *grpcServer) GetReading(context.Context, *pb.GetReadingRequest) (*pb.Reading, error) {
	reading := makeReading(s.logger, s.currTime, s.readings)
	s.logger.Infow("Got GRPC call for getReading, responding with ", "reading", reading)
	return reading, nil
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

const SERVER = "http://127.0.0.1:3000"

func main() {
	// Initialization
	start := time.Now()
	rawLog, _ := zap.NewDevelopment()
	readings := readCsvFile()
	log := rawLog.Sugar()
	client := resty.New()
	calibrationEnabled := true
	client.SetBaseURL(SERVER)
	rand.Seed(start.UnixNano())

	SensorInfo := RegisterRequest{
		Latitude:  15.87 + rand.Float64()*(16-15.86),
		Longitude: 45.75 + rand.Float64()*(45.85-45.75),
		Ip:        "127.0.0.1",
		Port:      10000 + rand.Intn(10000),
	}

	jsonReq, err := json.Marshal(SensorInfo)
	log.Infof("Sensor initialized with: %v", string(jsonReq))
	if err != nil {
		log.Fatal("could not marshall register req")
	}

	// Register on server
	log.Debugf("Using server %v for registration", SERVER)
	r, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(jsonReq).
		SetResult(&RegisterResponse{}).
		Post("/register")
	if err != nil {
		log.Fatal("could not connect to centralized server: ", err)
	}
	// log.Info("Got response: ", string(r.Body()))
	if r.IsSuccess() {
		log.Infof("Successfully registered with status %v; server response: %v", r.Status(),
			string(r.Body()))
	} else {
		log.Errorf("Could not register, response status: %v", r.Status())
		os.Exit(1)
	}

	// Start accepting grpc reading requests
	go startGrpcServer(log, start, strconv.Itoa(SensorInfo.Port))

	// Get neighbour
	log.Debugf("Fetching nearest neighbour from server %v", SERVER)
	myId := strconv.Itoa(r.Result().(*RegisterResponse).Id)
	r, err = client.R().
		SetResult(&RegisterResponse{}).
		Get("/sensors/" + myId + "/nearest")

	neighbour := r.Result().(*RegisterResponse)
	if r.IsError() || neighbour == nil {
		log.Error("nearest neighbour request error: ", r.Error())
		log.Infof("No neighbour! Grpc calls will be skipped, no calibration.")
		calibrationEnabled = false
	} else {
		log.Infof("My neighbour has id %v, calling him on ip: %v, port: %v", neighbour.Id, neighbour.Ip, neighbour.Port)
	}

	// Initialize grpc client
	neighbourAddr := neighbour.Ip + ":" + strconv.Itoa(neighbour.Port)
	conn, err := grpc.Dial(neighbourAddr, grpc.WithInsecure())
	if err != nil {
		log.Errorf("grpc could not connect: %v", err)
		calibrationEnabled = false
	} else {
		defer conn.Close()
	}
	c := pb.NewSensorClient(conn)

	// Reading loop
	for {
		reading := makeReading(log, start, readings)
		log.Info("Original reading: ", reading)

		if calibrationEnabled {
			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			neighbourReading, err := c.GetReading(ctx, &pb.GetReadingRequest{})
			if err != nil {
				log.Errorf("error while contacting neighbour: %v, skipping calibration", err)
				calibrationEnabled = false
			} else {
				log.Infof("Got reading from neighbour %v", neighbourReading)
				reading = &pb.Reading{
					Temperature: mergeReadings(reading.Temperature, neighbourReading.Temperature),
					Pressure:    mergeReadings(reading.Pressure, neighbourReading.Pressure),
					Humidity:    mergeReadings(reading.Humidity, neighbourReading.Humidity),
					Co:          mergeReadings(reading.Co, neighbourReading.Co),
					No2:         mergeReadings(reading.No2, neighbourReading.No2),
					So2:         mergeReadings(reading.So2, neighbourReading.So2),
				}
			}
		}

		serialized, _ := json.Marshal(reading)
		log.Debugf("sending reading %v to server", string(serialized))

		// Send reading to server
		r, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(serialized).
			Post("/sensors/" + myId + "/readings")
		if r.IsSuccess() {
			log.Infof("Successfully sent reading, return status %v; server response: %v", r.Status(),
				string(r.Body()))
		} else {
			log.Errorf("Error sending reading, response status: %v", r.Status())
			log.Error(err)
		}

		// Sleep for 3 to 8 seconds
		time.Sleep(time.Duration(3+rand.Intn(5)) * time.Second)
	}
}

func startGrpcServer(log *zap.SugaredLogger, startTime time.Time, port string) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSensorServer(s, &grpcServer{
		currTime: startTime,
		readings: readCsvFile(),
		logger:   log.Named("GrpcServer"),
	})
	log.Infof("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func makeReading(log *zap.SugaredLogger, current time.Time, readings []*pb.Reading) *pb.Reading {
	elapsed := time.Since(current)
	pickedReading := int(elapsed.Seconds()) % 100
	// log.Debugf("Time running: %v, picked reading number %v", elapsed, pickedReading)
	return readings[pickedReading]
}

func readCsvFile() []*pb.Reading {
	csvFile, _ := os.Open("readings.csv")
	csvLines, _ := csv.NewReader(csvFile).ReadAll()
	var readings []*pb.Reading
	for _, line := range csvLines[1:] {
		temp, _ := strconv.ParseFloat(line[0], 32)
		pressure, _ := strconv.ParseFloat(line[1], 32)
		h, _ := strconv.ParseFloat(line[2], 32)
		co, _ := strconv.ParseFloat(line[3], 32)
		no2, _ := strconv.ParseFloat(line[4], 32)
		so2, _ := strconv.ParseFloat(line[5], 32)
		reading := &pb.Reading{
			Temperature: float32(temp),
			Pressure:    float32(pressure),
			Humidity:    float32(h),
			Co:          float32(co),
			No2:         float32(no2),
			So2:         float32(so2),
		}
		readings = append(readings, reading)
	}
	return readings
}

func mergeReadings(r1, r2 float32) float32 {
	var defaultValue float32
	if r2 == defaultValue {
		return r1
	} else if r1 == defaultValue {
		return r2
	}

	return (r1 + r2) / 2
}
