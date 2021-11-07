package main

import (
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"math/rand"
)

type RegisterRequest struct {
	Latitude float64
	Longitude float64
	Ip string
	Port int
}
type RegisterResponse struct {
	Latitude float64
	Longitude float64
	Ip string
	Port int
	Id int
}

func main() {
	rawLog, _ := zap.NewDevelopment()
	log := rawLog.Sugar()
	client := resty.New()
	// client.Debug = true
	// Set up a connection to the server.
	registerReq := RegisterRequest{
		Latitude:  15.87 + rand.Float64() * (16 - 15.86),
		Longitude: 45.75 + rand.Float64() * (45.85 - 45.75),
		Ip:        "127.0.0.1",
		Port:      10000 + rand.Intn(10000),
	}

	jsonReq, err := json.Marshal(registerReq)
	log.Infof("%v", string(jsonReq))
	if err != nil {
		log.Fatal("could not marshall register req")
	}
	_, err = client.R().
		SetBody(jsonReq).
		SetResult(&RegisterResponse{}).
		Post("http://127.0.0.1:3000/register")
	if err != nil {
		log.Fatal("could not connect to centralized server: ", err)
	}
	// log.Info("Got response: ", string(r.Body()))
	// if r.IsSuccess() {
	// 	log.Info("Successfully registered; server response: %v", r.Body())
	// }

	// Get neighbour
	// myId := r.Result().(*RegisterResponse).Id
	// r, err = client.R().
	// 	SetResult(&RegisterResponse{}).
	// 	Get("http://127.0.0.1:3000/sensors/" + strconv.Itoa(myId) + "/nearest")

	// conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	//if err != nil {
	//	log.Fatalf("grpc did not connect: %v", err)
	//}
	//defer conn.Close()
	//c := pb.NewGreeterClient(conn)
	//
	//// Contact the server and print out its response.
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	//r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
	//if err != nil {
	//	log.Fatalf("could not greet: %v", err)
	//}
	//log.Printf("Greeting: %s", r.GetMessage())
}
