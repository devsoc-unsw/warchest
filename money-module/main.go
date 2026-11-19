package main

import (
	ctx "context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	constants "money-module/internal/config"
	handlers "money-module/internal/grpc"
	service "money-module/internal/service"

	pb "warchest/protos/mm_pb"

	"github.com/jackc/pgx/v5/pgxpool"
	tb "github.com/tigerbeetle/tigerbeetle-go"
	"google.golang.org/grpc"
)

func init_tb() (tb.Client, error) {
	tbAddress := os.Getenv("TB_ADDRESS")
	if len(tbAddress) == 0 {
		tbAddress = "3000"
	}
	client, err := tb.NewClient(tb.ToUint128(0),
		[]string{tbAddress})
	if err != nil {
		log.Printf("Error creating client: %s", err)
		return nil, err
	}

	// boot-strap system accounts
	var systemAccounts []tb.Account = []tb.Account{
		{
			ID:     constants.GetRealLedgerLiquidityBookId(),
			Ledger: constants.RealAccountLedger,
			Code:   constants.RealAccountCode,
		},
		{
			ID:     constants.GetRealLedgerWorldBookId(),
			Ledger: constants.RealAccountLedger,
			Code:   constants.RealAccountCode,
		},
		{
			ID:     constants.GetVirtualLedgerLiquidityBookId(),
			Ledger: constants.VirtualAccountLedger,
			Code:   constants.VirtualAccountCode,
		},
	}
	results, err := client.CreateAccounts(systemAccounts)

	if err != nil {
		log.Printf("Error creating accounts: %v", err)
		return nil, err
	}

	for _, result := range results {
		// If the account already exists or it has been created we are fine
		if result.Status != tb.AccountExists && result.Status != tb.AccountCreated {
			client.Close()
			return nil, fmt.Errorf("Issue with world account creating accounts: %v", result.Status.String())
		}
	}
	return client, nil
}

func main() {
	grpcServer := grpc.NewServer()

	initCtx, cancel := ctx.WithTimeout(ctx.Background(), 10*time.Second)
	defer cancel()

	dbURL := os.Getenv("MM_DATABASE_URL")
	if dbURL == "" {
		log.Fatal("MM_DATABASE_URL environment variable is required")
	}

	dbPool, err := pgxpool.New(initCtx, dbURL)
	if err != nil {
		log.Fatalf("Failed to initalize Postgres Pool: %v", err)
	}
	defer dbPool.Close()
	if err := dbPool.Ping(initCtx); err != nil {
		log.Fatalf("Failed to ping Postgres: %v", err)
	}

	tbClient, err := init_tb()
	if err != nil {
		log.Fatalf("Error initialising tb client: %v", err)
	}
	defer tbClient.Close()

	service := &service.MoneyModuleService{
		Tb: tbClient,
		Pg: dbPool,
	}

	handler := &handlers.MoneyModuleHandler{
		Svc: service,
	}
	pb.RegisterMoneyModuleServer(grpcServer, handler)

	addr := os.Getenv("TIGERBEETLE_ADDRESS")
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		log.Fatalf("%v", err)
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	log.Printf("MoneyModule gRPC server listening on :%s...", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server failed: %v", err)
	}
}
