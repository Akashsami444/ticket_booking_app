package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	pb "github.com/Akash-private/Cloudbees_code/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var client pb.TicketReservationClient

func main() {
	// Connect to gRPC server using the Docker service name
	conn, err := grpc.Dial("grpc-server:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("gRPC connection failed: %v", err)
	}
	defer conn.Close()
	client = pb.NewTicketReservationClient(conn)

	// Routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/book", handleBook)
	http.HandleFunc("/modify", handleModify)
	http.HandleFunc("/cancel", handleCancel)

	fmt.Println("🌐 Web UI starting on http://localhost:8888")
	log.Fatal(http.ListenAndServe(":8888", nil))
}

func handleBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.ReserveTicket(ctx, &pb.ReservationRequest{
		FromCode: "London",
		ToCode:   "Paris",
		Passengers: []*pb.UserDetails{
			{
				FirstName: r.FormValue("first_name"),
				Email:     r.FormValue("email"),
			},
		},
	})

	renderResult(w, "Booking Result", resp, err)
}

func handleModify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	tNo, _ := strconv.ParseUint(r.FormValue("ticket_no"), 10, 64)
	seat, _ := strconv.ParseUint(r.FormValue("seat"), 10, 32)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.ModifyTicket(ctx, &pb.ReservationRequest{
		TicketNo: &tNo,
		Passengers: []*pb.UserDetails{
			{Section: r.FormValue("section"), Seat: uint32(seat)},
		},
	})

	renderResult(w, "Modification Result", resp, err)
}

func handleCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	tNo, _ := strconv.ParseUint(r.FormValue("ticket_no"), 10, 64)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.CancelTicket(ctx, &pb.ReservationRequest{
		TicketNo: &tNo,
	})

	renderResult(w, "Cancellation Result", resp, err)
}

func renderResult(w http.ResponseWriter, title string, resp *pb.ReservationResponse, err error) {
	if err != nil {
		fmt.Fprintf(w, "<h2>Error</h2><p>%v</p><a href='/'>Go Back</a>", err)
		return
	}
	fmt.Fprintf(w, "<h2>%s</h2><p>Ticket No: %d</p><p>Status: %s</p><a href='/'>Go Back</a>", title, resp.TicketNo, resp.Status)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Fetch all tickets from gRPC server to display in the table
	resp, err := client.GetAllTickets(ctx, &pb.EmptyRequest{})
	if err != nil {
		// If the server is down or DB is empty, we handle it gracefully
		log.Printf("Could not fetch tickets: %v", err)
		// We can still show the page even if the list is empty
		resp = &pb.AllTicketsResponse{}
	}

	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, "Template index.html not found", 500)
		return
	}

	// 2. Pass the list of tickets to the HTML template
	tmpl.Execute(w, resp.Tickets)
}
