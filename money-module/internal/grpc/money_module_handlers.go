package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "warchest/protos/mm_pb"
	"money-module/internal/service"
)

type MoneyModuleHandler struct {
	pb.UnimplementedMoneyModuleServer
	Svc *service.MoneyModuleService
}

func NewMoneyModuleHandler() *MoneyModuleHandler {
	return &MoneyModuleHandler{}
}

func (h *MoneyModuleHandler) CreateAccount(
	ctx context.Context,
	req *pb.CreateAccountRequest,
) (*pb.WalletDetailsResponse, error) {
	if req.GetOwnerId().String() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id cannot be empty")
	}
	uuid := [16]byte(req.GetOwnerId().Value)	
	// forward to service layer
	err := h.Svc.CreateAccount(ctx, uuid, req.BudgetName, req.OwnerType.String())
	if err != nil {
		return nil, status.Error(codes.Internal, "User already exists")
	}

	out := &pb.WalletDetailsResponse {
		OwnerType: pb.OwnerType_USER,
		OwnerId: req.OwnerId,
		WalletType: pb.WalletType_ACCOUNT,
		WalletName: req.BudgetName,
	}
	return out, nil 
}

func (h *MoneyModuleHandler) ReadAccount(
	ctx context.Context,
	req *pb.GetWalletDetailsRequest,
) (*pb.WalletDetailsResponse, error) {
	return nil, status.Error(
		codes.Unimplemented, "method ReadAccount not implemented")
}

func (h *MoneyModuleHandler) UpdateAccount(
	ctx context.Context,
	req *pb.UpdateAccountRequest,
) (*pb.WalletDetailsResponse, error) {
	return nil, status.Error(
		codes.Unimplemented, "method UpdateAccount not implemented")
}

func (h *MoneyModuleHandler) ArchiveAccount(
	ctx context.Context,
	req *pb.ArchiveWalletRequest,
) (*pb.WalletDetailsResponse, error) {
	return nil, status.Error(
		codes.Unimplemented, "method ArchiveAccount not implemented")
}

func (h *MoneyModuleHandler) CreateBudget(
	ctx context.Context,
	req *pb.CreateBudgetRequest,
) (*pb.WalletDetailsResponse, error) {
	return nil, status.Error(
		codes.Unimplemented, "method CreateBudget not implemented")
}

func (h *MoneyModuleHandler) ReadBudget(
	ctx context.Context,
	req *pb.GetWalletDetailsRequest,
) (*pb.WalletDetailsResponse, error) {
	return nil, status.Error(
		codes.Unimplemented, "method ReadBudget not implemented")
}

func (h *MoneyModuleHandler) UpdateBudget(
	ctx context.Context,
	req *pb.UpdateBudgetRequest,
) (*pb.WalletDetailsResponse, error) {
	return nil, status.Error(
		codes.Unimplemented, "method UpdateBudget not implemented")
}

func (h *MoneyModuleHandler) ArchiveBudget(
	ctx context.Context,
	req *pb.ArchiveWalletRequest,
) (*pb.WalletDetailsResponse, error) {
	return nil, status.Error(
		codes.Unimplemented, "method ArchiveBudget not implemented")
}

// Money routes
func (h *MoneyModuleHandler) PayAccountFromBudget(
	ctx context.Context,
	req *pb.PaymentRequest,
) (*pb.PaymentResponse, error) {
	return nil, status.Error(
		codes.Unimplemented, "method PayAccountFromBudget not implemented")
}

func (h *MoneyModuleHandler) PayBudgetFromAccount(
	ctx context.Context,
	req *pb.PaymentRequest,
) (*pb.PaymentResponse, error) {
	return nil, status.Error(
		codes.Unimplemented, "method PayBudgetFromAccount not implemented")
}

func (h *MoneyModuleHandler) PayRootFromAccount(
	ctx context.Context,
	req *pb.RootPaymentRequest,
) (*pb.PaymentResponse, error) {
	return nil, status.Error(
		codes.Unimplemented, "method PayRootFromAccount not implemented")
}

func (h *MoneyModuleHandler) PayAccountFromRoot(
	ctx context.Context,
	req *pb.RootPaymentRequest,
) (*pb.PaymentResponse, error) {
	return nil, status.Error(
		codes.Unimplemented, "method PayAccountFromRoot not implemented")
}

func (h *MoneyModuleHandler) SendBetweenBudgets(
	ctx context.Context,
	req *pb.SendRequest,
) (*pb.SendResponse, error) {
	return nil, status.Error(
		codes.Unimplemented, "method SendBetweenBudgets not implemented")
}

func (h *MoneyModuleHandler) GetAccountTransfers(
	ctx context.Context,
	req *pb.TransfersRequest,
) (*pb.TransfersResponse, error) {
	return nil, status.Error(
		codes.Unimplemented, "method GetAccountTransfers not implemented")
}

func (h *MoneyModuleHandler) GetBudgetTransfers(
	ctx context.Context,
	req *pb.TransfersRequest,
) (*pb.TransfersResponse, error) {
	return nil, status.Error(
		codes.Unimplemented, "method GetBudgetTransfers not implemented")
}

func (h *MoneyModuleHandler) GetSocietyTransfers(
	ctx context.Context,
	req *emptypb.Empty,
) (*pb.TransfersResponse, error) {
	return nil, status.Error(
		codes.Unimplemented, "method GetSocietyTransfers not implemented")
}
