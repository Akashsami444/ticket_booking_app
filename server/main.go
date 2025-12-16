package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	pb "github.com/Akash-private/Cloudbees_code/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TicketReservationServer struct {
	pb.UnimplementedTicketReservationServer
	mu            sync.Mutex
	seatDB        map[string][]*uint64
	ticketDB      map[uint64]*pb.ReservationResponse
	ticketCounter uint64
}

func NewServer() *TicketReservationServer {
	return &TicketReservationServer{
		seatDB: map[string][]*uint64{
			"A": make([]*uint64, 20),
			"B": make([]*uint64, 20),
		},
		ticketDB: make(map[uint64]*pb.ReservationResponse),
	}
}

func (s *TicketReservationServer) ReserveTicket(
	ctx context.Context,
	req *pb.ReservationRequest,
) (*pb.ReservationResponse, error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	s.ticketCounter++
	ticketNo := s.ticketCounter

	fmt.Println(req)

	passengers := make([]*pb.UserDetails, 0)

	for i, p := range req.Passengers {
		isReserved := false
		section := p.Section
		seat := p.Seat

		newP := &pb.UserDetails{
			FirstName: p.FirstName,
			LastName:  p.LastName,
			Email:     p.Email,
			Address:   p.Address,
		}

		if section != "" && seat > 0 {
			if s.seatDB[section][seat-1] != nil {
				for j := range s.seatDB[section] {
					if s.seatDB[section][j] == nil {
						newP.Seat = uint32(j + 1)
						newP.Section = section
						isReserved = true
						break
					}
				}
			} else {
				newP.Seat = seat
				newP.Section = section
				isReserved = true
			}
		}

		if !isReserved {
			found := false
			for _, sec := range []string{"A", "B"} {
				for j := range s.seatDB[sec] {
					if s.seatDB[sec][j] == nil {
						newP.Seat = uint32(j + 1)
						newP.Section = sec
						found = true
						break
					}
				}
				if found {
					break
				}
			}
		}

		s.seatDB[newP.Section][newP.Seat-1] = &ticketNo
		passengers = append(passengers, newP)

		fmt.Println("Passenger", i+1, "=>", newP)
	}

	resp := &pb.ReservationResponse{
		TicketNo:       ticketNo,
		FromCode:       req.FromCode,
		ToCode:         req.ToCode,
		PricePaid:      req.PricePaid,
		PassengerCount: req.PassengerCount,
		Passengers:     passengers,
		Status:         "Confirmed",
	}

	s.ticketDB[ticketNo] = resp

	fmt.Println("Seat_DB:", s.seatDB)
	fmt.Println("Ticket_DB:", s.ticketDB)

	return resp, nil
}

func (s *TicketReservationServer) ModifyTicket(
	ctx context.Context,
	req *pb.ReservationRequest,
) (*pb.ReservationResponse, error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	if req.TicketNo == nil {
		return nil, status.Error(codes.InvalidArgument, "ticket_no is required")
	}
	ticketNo := *req.TicketNo

	resp := s.ticketDB[ticketNo]

	for i, p := range req.Passengers {
		if s.seatDB[p.Section][p.Seat-1] == nil ||
			*s.seatDB[p.Section][p.Seat-1] != ticketNo {

			if s.seatDB[p.Section][p.Seat-1] != nil {
				resp.Status = "Failed! The seat is already booked!"
			} else {
				old := resp.Passengers[i]
				s.seatDB[old.Section][old.Seat-1] = nil
				s.seatDB[p.Section][p.Seat-1] = &ticketNo
				resp.Status = "Success! The seat(s) are booked!"
			}
		}
		resp.Passengers[i].Section = p.Section
		resp.Passengers[i].Seat = p.Seat
	}

	if resp.Status == "Success! The seat(s) are booked!" {
		s.ticketDB[*req.TicketNo] = resp
	}

	fmt.Println("Seat_DB:", s.seatDB)
	fmt.Println("Ticket_DB:", s.ticketDB)

	return resp, nil
}

func (s *TicketReservationServer) CancelTicket(
	ctx context.Context,
	req *pb.ReservationRequest,
) (*pb.ReservationResponse, error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	if req.TicketNo == nil {
		return nil, status.Error(codes.InvalidArgument, "ticket_no is required")
	}
	ticketNo := *req.TicketNo

	resp := s.ticketDB[ticketNo]

	resp.Status = "Cancelled"

	for _, p := range resp.Passengers {
		s.seatDB[p.Section][p.Seat-1] = nil
		p.Section = ""
		p.Seat = 0
	}

	fmt.Println("Seat_DB:", s.seatDB)
	fmt.Println("Ticket_DB:", s.ticketDB)

	return resp, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterTicketReservationServer(grpcServer, NewServer())

	log.Println("🚆 gRPC server running on :50051")
	grpcServer.Serve(lis)
}
