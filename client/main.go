package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/Akash-private/Cloudbees_code/proto"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewTicketReservationClient(conn)

	var option int

	for option != 4 {
		fmt.Println("\nWelcome! Train Ticket Reservation")
		fmt.Println("1. Reserve Ticket")
		fmt.Println("2. Modify Seat Allotment")
		fmt.Println("3. Cancel Ticket")
		fmt.Println("4. Close")
		fmt.Print("Choose option: ")
		fmt.Scan(&option)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

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
				fmt.Println("Passenger", i+1)
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

			resp, err := client.ReserveTicket(ctx, req)
			cancel()

			if err != nil {
				log.Println("gRPC error:", err)
				break
			}

			fmt.Println("Receipt:", resp)

		case 2:
			var ticketNo uint64
			var count int

			fmt.Print("Ticket No: ")
			fmt.Scan(&ticketNo)
			fmt.Print("Passenger Count: ")
			fmt.Scan(&count)

			passengers := []*pb.UserDetails{}
			for i := 0; i < count; i++ {
				p := &pb.UserDetails{}
				fmt.Println("Passenger", i+1)
				fmt.Print("Section: ")
				fmt.Scan(&p.Section)
				fmt.Print("Seat: ")
				fmt.Scan(&p.Seat)
				passengers = append(passengers, p)
			}

			req := &pb.ReservationRequest{
				TicketNo:       &ticketNo,
				PassengerCount: uint64(count),
				Passengers:     passengers,
			}

			resp, err := client.ModifyTicket(ctx, req)
			cancel()

			if err != nil {
				log.Println("gRPC error:", err)
				break
			}
			fmt.Println("Receipt:", resp)

		case 3:
			var ticketNo uint64
			fmt.Print("Ticket No: ")
			fmt.Scan(&ticketNo)

			resp, err := client.CancelTicket(ctx, &pb.ReservationRequest{
				TicketNo: &ticketNo,
			})
			cancel()

			if err != nil {
				log.Println("gRPC error:", err)
				break
			}
			fmt.Println("Receipt:", resp)

		}
	}
}
