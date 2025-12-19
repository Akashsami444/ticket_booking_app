package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"sync"

	pb "github.com/Akash-private/Cloudbees_code/proto"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TicketReservationServer struct {
	pb.UnimplementedTicketReservationServer
	mu sync.Mutex
	db *sql.DB
}

func main() {
	// Connect to DB via Environment Variable
	dbURL := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Could not connect to DB: %v", err)
	}
	defer db.Close()

	// Setup Database Table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tickets (
		id SERIAL PRIMARY KEY,
		passenger_name TEXT,
		email TEXT,
		section TEXT,
		seat INT,
		status TEXT
	)`)
	if err != nil {
		log.Fatalf("Table creation failed: %v", err)
	}

	// Start Listener
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTicketReservationServer(s, &TicketReservationServer{db: db})

	log.Println("🚆 gRPC Server with Postgres running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *TicketReservationServer) ReserveTicket(ctx context.Context, req *pb.ReservationRequest) (*pb.ReservationResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var id uint64
	p := req.Passengers[0]

	// Default allocation for demo: Section A, Seat 1 (You can add logic to find next free seat)
	err := s.db.QueryRow(
		"INSERT INTO tickets (passenger_name, email, section, seat, status) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		p.FirstName, p.Email, "A", 1, "Confirmed",
	).Scan(&id)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "DB Insert Error: %v", err)
	}

	return &pb.ReservationResponse{TicketNo: id, Status: "Booked Successfully", Passengers: req.Passengers}, nil
}

func (s *TicketReservationServer) ModifyTicket(ctx context.Context, req *pb.ReservationRequest) (*pb.ReservationResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.TicketNo == nil {
		return nil, status.Error(codes.InvalidArgument, "ID required")
	}

	p := req.Passengers[0]
	_, err := s.db.Exec("UPDATE tickets SET section = $1, seat = $2, status = $3 WHERE id = $4",
		p.Section, p.Seat, "Modified", *req.TicketNo)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "DB Update Error: %v", err)
	}

	return &pb.ReservationResponse{TicketNo: *req.TicketNo, Status: "Modification Saved"}, nil
}

func (s *TicketReservationServer) CancelTicket(ctx context.Context, req *pb.ReservationRequest) (*pb.ReservationResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.TicketNo == nil {
		return nil, status.Error(codes.InvalidArgument, "ID required")
	}

	_, err := s.db.Exec("DELETE FROM tickets WHERE id = $1", *req.TicketNo)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "DB Delete Error: %v", err)
	}

	return &pb.ReservationResponse{TicketNo: *req.TicketNo, Status: "Ticket Cancelled/Deleted"}, nil
}

func (s *TicketReservationServer) GetAllTickets(ctx context.Context, req *pb.EmptyRequest) (*pb.AllTicketsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query("SELECT id, passenger_name, email, section, seat, status FROM tickets ORDER BY id DESC")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "DB Query Error: %v", err)
	}
	defer rows.Close()

	var tickets []*pb.ReservationResponse
	for rows.Next() {
		var t pb.ReservationResponse
		var p pb.UserDetails
		err := rows.Scan(&t.TicketNo, &p.FirstName, &p.Email, &p.Section, &p.Seat, &t.Status)
		if err != nil {
			continue
		}
		t.Passengers = []*pb.UserDetails{&p}
		tickets = append(tickets, &t)
	}

	return &pb.AllTicketsResponse{Tickets: tickets}, nil
}
