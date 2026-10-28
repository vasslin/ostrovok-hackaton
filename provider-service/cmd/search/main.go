package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	pb "github.com/godnota/pkg/generated/provider/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func readObject(path string) (*structpb.Struct, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var object map[string]any
	if err := json.Unmarshal(data, &object); err != nil || object == nil {
		return nil, fmt.Errorf("invalid JSON object in %s", path)
	}
	return structpb.NewStruct(object)
}

func run() error {
	address := flag.String("address", "127.0.0.1:50058", "provider gRPC address")
	kind := flag.String("kind", "transport", "transport, hotel, activity or parts")
	format := flag.String("format", "trip_request", "trip_request, ai_config or frontend_criteria")
	tripPath := flag.String("trip", "configs/examples/trip_request.kazan_weekend.json", "input JSON path")
	resolvedPath := flag.String("resolved", "", "resolved JSON path")
	flag.Parse()
	input, err := readObject(*tripPath)
	if err != nil {
		return err
	}
	var resolved *structpb.Struct
	if *resolvedPath != "" {
		resolved, err = readObject(*resolvedPath)
		if err != nil {
			return err
		}
	}
	conn, err := grpc.NewClient(*address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pb.NewProviderServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	search := &pb.SearchInput{TripRequest: input, Resolved: resolved, InputFormat: *format}
	var result proto.Message
	switch *kind {
	case "parts":
		if *format != "trip_request" {
			return fmt.Errorf("parts requires trip_request input")
		}
		result, err = client.SearchParts(ctx, &pb.SearchPartsRequest{Config: input, Resolved: resolved})
	case "transport":
		result, err = client.SearchTransport(ctx, &pb.SearchTransportRequest{Input: search})
	case "hotel":
		result, err = client.SearchHotels(ctx, &pb.SearchHotelsRequest{Input: search})
	case "activity":
		result, err = client.SearchActivities(ctx, &pb.SearchActivitiesRequest{Input: search})
	default:
		return fmt.Errorf("unknown kind %q", *kind)
	}
	if err != nil {
		return err
	}
	output, err := (protojson.MarshalOptions{Multiline: true, Indent: "  "}).Marshal(result)
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
