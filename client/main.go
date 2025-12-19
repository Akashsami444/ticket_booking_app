package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/Akash-private/Cloudbees_code/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Updated to use NewClientConn as WithInsecure is deprecated in newer versions
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewTicketReservationClient(conn)

	var option int
	for {
		fmt.Println("\n--- Train Ticket Reservation ---")
		fmt.Println("1. Reserve Ticket")
		fmt.Println("2. Modify Seat Allotment")
		fmt.Println("3. Cancel Ticket")
		fmt.Println("4. Close")
		fmt.Print("Choose option: ")
		fmt.Scan(&option)

		if option == 4 {
			break
		}

		switch option {
		case 1:
			var from, to string
			var count int

			fmt.Print("From: ")
			fmt.Scan(&from)
			fmt.Print("To: ")
			fmt.Scan(&to)
			fmt.Print("Passenger Count: ")
			fmt.Scan(&count)

			passengers := []*pb.UserDetails{}
			for i := 0; i < count; i++ {
				p := &pb.UserDetails{}
				fmt.Printf("\nPassenger %d\n", i+1)
				fmt.Print("First Name: ")
				fmt.Scan(&p.FirstName)
				fmt.Print("Last Name: ")
				fmt.Scan(&p.LastName)
				fmt.Print("Email: ")
				fmt.Scan(&p.Email)
				fmt.Print("Address: ")
				fmt.Scan(&p.Address)
				passengers = append(passengers, p)
			}

			req := &pb.ReservationRequest{
				FromCode:       from,
				ToCode:         to,
				PricePaid:      uint64(count * 20),
				PassengerCount: uint64(count),
				Passengers:     passengers,
			}

			// CREATE CONTEXT HERE: After input is finished
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			resp, err := client.ReserveTicket(ctx, req)
			cancel()

			if err != nil {
				log.Println("gRPC error:", err)
			} else {
				fmt.Println("\n✅ Reservation Successful!")
				fmt.Println(resp)
			}

		case 2:
			var ticketNo uint64
			var count int
			fmt.Print("Ticket No: ")
			fmt.Scan(&ticketNo)
			fmt.Print("How many seats to modify? ")
			fmt.Scan(&count)

			passengers := []*pb.UserDetails{}
			for i := 0; i < count; i++ {
				p := &pb.UserDetails{}
				fmt.Printf("Passenger %d - New Section (A/B): ", i+1)
				fmt.Scan(&p.Section)
				fmt.Print("New Seat No: ")
				fmt.Scan(&p.Seat)
				passengers = append(passengers, p)
			}

			req := &pb.ReservationRequest{
				TicketNo:       &ticketNo,
				PassengerCount: uint64(count),
				Passengers:     passengers,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			resp, err := client.ModifyTicket(ctx, req)
			cancel()

			if err != nil {
				log.Println("gRPC error:", err)
			} else {
				fmt.Println("\n🔄 Modification Result:", resp.Status)
				fmt.Println(resp)
			}

		case 3:
			var ticketNo uint64
			fmt.Print("Ticket No to Cancel: ")
			fmt.Scan(&ticketNo)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			resp, err := client.CancelTicket(ctx, &pb.ReservationRequest{
				TicketNo: &ticketNo,
			})
			cancel()

			if err != nil {
				log.Println("gRPC error:", err)
			} else {
				fmt.Println("\n❌ Ticket Cancelled:", resp.Status)
			}
		}
	}
}
