package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	constants "money-module/internal/config"
	"money-module/internal/service"
	pb "warchest/protos/mm_pb"
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
	err := h.Svc.CreateAccount(ctx, uuid, req.BudgetName, req.OwnerType.String(), constants.ACCOUNT)
	if err != nil {
		return nil, status.Error(codes.Internal, "User already exists")
	}

	out := &pb.WalletDetailsResponse{
		OwnerType:  pb.OwnerType_USER,
		OwnerId:    req.OwnerId,
		WalletType: pb.WalletType_ACCOUNT,
		WalletName: req.BudgetName,
	}
	return out, nil
}

func (h *MoneyModuleHandler) ReadAccount(
	ctx context.Context,
	req *pb.GetWalletDetailsRequest,
) (*pb.WalletDetailsResponse, error) {
	//TODO: Do not know what to do with req.state, unknownFields or cache size

	if req.GetId().String() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id cannot be empty")
	}
	uuid := [16]byte(req.GetId().Value)
	// forward to service layer

	ownerType, ownerId, walletType, walletName, isArchived, err := h.Svc.GetDetails(ctx, uuid, constants.ACCOUNT)

	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}

	//TODO: What about state, unknown fields, size cache??
	out := &pb.WalletDetailsResponse{
		OwnerType:  ownerType,
		WalletType: walletType,
		WalletName: walletName,
		Id:         req.Id,
		IsArchived: isArchived,
		OwnerId:    &ownerId,
	}
	return out, nil
}

func (h *MoneyModuleHandler) UpdateAccount(
	ctx context.Context,
	req *pb.UpdateAccountRequest,
) (*pb.WalletDetailsResponse, error) {

	if req.GetId().String() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id cannot be empty")
	}
	uuid := [16]byte(req.GetId().Value)

	err := h.Svc.UpdateDetails(ctx, uuid, req.OwnerType, req.OwnerId, req.WalletName, constants.ACCOUNT)
	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}
	ownerType, ownerId, walletType, walletName, isArchived, err := h.Svc.GetDetails(ctx, uuid, constants.ACCOUNT)

	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}

	//TODO: What to do with req.state, unknown fields and sizeCache

	//TODO: What about state, unknown fields, size cache??
	out := &pb.WalletDetailsResponse{
		OwnerType:  ownerType,
		WalletType: walletType,
		WalletName: walletName,
		Id:         req.Id,
		IsArchived: isArchived,
		OwnerId:    &ownerId,
	}
	return out, nil
}

func (h *MoneyModuleHandler) ArchiveAccount(
	ctx context.Context,
	req *pb.ArchiveWalletRequest,
) (*pb.WalletDetailsResponse, error) {
	if req.GetId().String() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id cannot be empty")
	}
	uuid := [16]byte(req.GetId().Value)

	err := h.Svc.ArchiveAccount(ctx, constants.ACCOUNT)
	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}
	ownerType, ownerId, walletType, walletName, isArchived, err := h.Svc.GetDetails(ctx, uuid, constants.ACCOUNT)

	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}

	//TODO: What to do with req.state, unknown fields and sizeCache

	//TODO: What about state, unknown fields, size cache??
	out := &pb.WalletDetailsResponse{
		OwnerType:  ownerType,
		WalletType: walletType,
		WalletName: walletName,
		Id:         req.Id,
		IsArchived: isArchived,
		OwnerId:    &ownerId,
	}
	return out, nil
}

func (h *MoneyModuleHandler) CreateBudget(
	ctx context.Context,
	req *pb.CreateBudgetRequest,
) (*pb.WalletDetailsResponse, error) {
	if req.GetOwnerId().String() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id cannot be empty")
	}
	uuid := [16]byte(req.GetOwnerId().Value)
	// forward to service layer
	err := h.Svc.CreateAccount(ctx, uuid, req.BudgetName, req.OwnerType.String(), constants.BUDGET)
	if err != nil {
		return nil, status.Error(codes.Internal, "User already exists")
	}

	out := &pb.WalletDetailsResponse{
		OwnerType:  pb.OwnerType_USER,
		OwnerId:    req.OwnerId,
		WalletType: pb.WalletType_ACCOUNT,
		WalletName: req.BudgetName,
	}
	return out, nil
}

func (h *MoneyModuleHandler) ReadBudget(
	ctx context.Context,
	req *pb.GetWalletDetailsRequest,
) (*pb.WalletDetailsResponse, error) {
	//TODO: Do not know what to do with req.state, unknownFields or cache size

	if req.GetId().String() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id cannot be empty")
	}
	uuid := [16]byte(req.GetId().Value)
	// forward to service layer

	ownerType, ownerId, walletType, walletName, isArchived, err := h.Svc.ReadDetails(ctx, uuid, constants.BUDGET)

	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}

	//TODO: What about state, unknown fields, size cache??
	out := &pb.WalletDetailsResponse{
		OwnerType:  ownerType,
		WalletType: walletType,
		WalletName: walletName,
		Id:         req.Id,
		IsArchived: isArchived,
		OwnerId:    &ownerId,
	}
	return out, nil
}

func (h *MoneyModuleHandler) UpdateBudget(
	ctx context.Context,
	req *pb.UpdateBudgetRequest,
) (*pb.WalletDetailsResponse, error) {
	if req.GetId().String() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id cannot be empty")
	}
	uuid := [16]byte(req.GetId().Value)

	err := h.Svc.UpdateDetails(ctx, uuid, req.OwnerType, req.OwnerId, req.WalletName, constants.BUDGET)
	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}
	ownerType, ownerId, walletType, walletName, isArchived, err := h.Svc.ReadDetails(ctx, uuid, constants.BUDGET)

	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}

	//TODO: What to do with req.state, unknown fields and sizeCache

	//TODO: What about state, unknown fields, size cache??
	out := &pb.WalletDetailsResponse{
		OwnerType:  ownerType,
		WalletType: walletType,
		WalletName: walletName,
		Id:         req.Id,
		IsArchived: isArchived,
		OwnerId:    &ownerId,
	}
	return out, nil
}

func (h *MoneyModuleHandler) ArchiveBudget(
	ctx context.Context,
	req *pb.ArchiveWalletRequest,
) (*pb.WalletDetailsResponse, error) {
	if req.GetId().String() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id cannot be empty")
	}
	uuid := [16]byte(req.GetId().Value)

	err := h.Svc.ArchiveAccount(ctx, uuid, constants.BUDGET)
	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}
	ownerType, ownerId, walletType, walletName, isArchived, err := h.Svc.GetDetails(ctx, uuid, constants.BUDGET)

	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}

	//TODO: What to do with req.state, unknown fields and sizeCache

	//TODO: What about state, unknown fields, size cache??
	out := &pb.WalletDetailsResponse{
		OwnerType:  ownerType,
		WalletType: walletType,
		WalletName: walletName,
		Id:         req.Id,
		IsArchived: isArchived,
		OwnerId:    &ownerId,
	}
	return out, nil
}

// Money routes
func (h *MoneyModuleHandler) PayAccountFromBudget(
	ctx context.Context,
	req *pb.PaymentRequest,
) (*pb.PaymentResponse, error) {

	/*	state         protoimpl.MessageState `protogen:"open.v1"`
		TransactionId *UUID                  `protobuf:"bytes,1,opt,name=transaction_id,json=transactionId,proto3" json:"transaction_id,omitempty"`
		Status        Status                 `protobuf:"varint,2,opt,name=status,proto3,enum=money_module.Status" json:"status,omitempty"`
		unknownFields protoimpl.UnknownFields
		sizeCache     protoimpl.SizeCache
	*/

	transaction_id, status, err := h.Svc.Transaction(req.PayerId, req.PayeeId, req.Amount, constants.ACCOUNT, constants.BUDGET)

	if err != nil {
		return err
	}

	return &pb.PaymentResponse{
		TransactionId: transaction_id,
		Status:        status,
	}, nil
}

func (h *MoneyModuleHandler) PayBudgetFromAccount(
	ctx context.Context,
	req *pb.PaymentRequest,
) (*pb.PaymentResponse, error) {
	transaction_id, status, err := h.Svc.Transaction(req.PayerId, req.PayeeId, req.Amount, constants.BUDGET, constants.ACCOUNT)

	if err != nil {
		return err
	}

	return &pb.PaymentResponse{
		TransactionId: transaction_id,
		Status:        status,
	}, nil
}

func (h *MoneyModuleHandler) PayRootFromAccount(
	ctx context.Context,
	req *pb.RootPaymentRequest,
) (*pb.PaymentResponse, error) {
	transaction_id, status, err := h.Svc.Transaction(req.PayerId, req.PayeeId, req.Amount, constants.ROOT, constants.ACCOUNT)

	if err != nil {
		return err
	}

	return &pb.PaymentResponse{
		TransactionId: transaction_id,
		Status:        status,
	}, nil
}

func (h *MoneyModuleHandler) PayAccountFromRoot(
	ctx context.Context,
	req *pb.RootPaymentRequest,
) (*pb.PaymentResponse, error) {
	transaction_id, status, err := h.Svc.Transaction(req.PayerId, req.PayeeId, req.Amount, constants.ACCOUNT, constants.ROOT)

	if err != nil {
		return err
	}

	return &pb.PaymentResponse{
		TransactionId: transaction_id,
		Status:        status,
	}, nil
}

func (h *MoneyModuleHandler) SendBetweenBudgets(
	ctx context.Context,
	req *pb.SendRequest,
) (*pb.SendResponse, error) {
	transaction_id, status, err := h.Svc.Transaction(req.PayerId, req.PayeeId, req.Amount, constants.BUDGET, constants.BUDGET)

	if err != nil {
		return err
	}

	return &pb.SendResponse{
		TransactionId: transaction_id,
		Status:        status,
	}, nil
}

func (h *MoneyModuleHandler) GetAccountTransfers(
	ctx context.Context,
	req *pb.TransfersRequest,
) (*pb.TransfersResponse, error) {
	transfers, err := h.Svc.getTransferrs(req.Id, constants.ACCOUNT)
	if err != nil {
		return err
	}
	return &pb.TransfersResponse{Transfers: transfers}, nil
}

func (h *MoneyModuleHandler) GetBudgetTransfers(
	ctx context.Context,
	req *pb.TransfersRequest,
) (*pb.TransfersResponse, error) {
	return nil, status.Error(
		codes.Unimplemented, "method GetBudgetTransfers not implemented")
}

//TODO: Check this function doesn't have to be implemented
// func (h *MoneyModuleHandler) GetSocietyTransfers(
// 	ctx context.Context,
// 	req *emptypb.Empty,
// ) (*pb.TransfersResponse, error) {
// 	return nil, status.Error(
// 		codes.Unimplemented, "method GetSocietyTransfers not implemented")
// }
