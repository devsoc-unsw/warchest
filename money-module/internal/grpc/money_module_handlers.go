package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	constants "money-module/internal/config"
	"money-module/internal/service"
	pb "warchest/protos/mm_pb"

	tb "github.com/tigerbeetle/tigerbeetle-go"
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

	if req.GetId().String() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id cannot be empty")
	}
	uuid := [16]byte(req.GetId().Value)
	// forward to service layer

	ownerType, ownerId, walletType, walletName, isArchived, err := h.Svc.ReadAccount(ctx, uuid, constants.ACCOUNT)

	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}

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
	owner_uuid := [16]byte(req.GetOwnerId().Value)

	err := h.Svc.UpdateAccount(ctx, uuid, req.OwnerType.String(), owner_uuid, req.WalletName, constants.ACCOUNT)
	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}
	ownerType, ownerId, walletType, walletName, isArchived, err := h.Svc.ReadAccount(ctx, uuid, constants.ACCOUNT)

	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}

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

	err := h.Svc.ArchiveAccount(ctx, uuid)
	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}
	ownerType, ownerId, walletType, walletName, isArchived, err := h.Svc.ReadAccount(ctx, uuid, constants.ACCOUNT)

	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}

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

	ownerType, ownerId, walletType, walletName, isArchived, err := h.Svc.ReadAccount(ctx, uuid, constants.BUDGET)

	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}

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
	owner_uuid := [16]byte(req.GetOwnerId().Value)

	err := h.Svc.UpdateAccount(ctx, uuid, req.OwnerType.String(), owner_uuid, req.WalletName, constants.BUDGET)
	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}
	ownerType, ownerId, walletType, walletName, isArchived, err := h.Svc.ReadAccount(ctx, uuid, constants.BUDGET)

	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}

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

	err := h.Svc.ArchiveAccount(ctx, uuid)
	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}
	ownerType, ownerId, walletType, walletName, isArchived, err := h.Svc.ReadAccount(ctx, uuid, constants.BUDGET)

	if err != nil {
		//TODO: Determine what the error code should be
		return nil, status.Error(codes.Internal, "User already exists")
	}

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

func parseUUID(u *pb.UUID, field string) ([16]byte, error) {
	if u == nil || len(u.GetValue()) != 16 {
		return [16]byte{}, status.Errorf(codes.InvalidArgument, "%s must be a 16-byte UUID", field)
	}
	return [16]byte(u.GetValue()), nil
}

func toProtoStatus(txStatus constants.TransactionStatus) pb.Status {
	switch txStatus {
	case constants.TransactionSuccess:
		return pb.Status_PASSED
	default:
		return pb.Status_FAILED
	}
}

func toGRPCError(txStatus constants.TransactionStatus, err error) error {
	switch txStatus {
	case constants.TransactionInvalidAmount, constants.TransactionSameAccount, constants.TransactionInvalidAccountType:
		return status.Error(codes.InvalidArgument, err.Error())
	case constants.TransactionAccountNotFound, constants.TransactionPayerNotFound, constants.TransactionPayeeNotFound:
		return status.Error(codes.NotFound, err.Error())
	case constants.TransactionAccountArchived, constants.TransactionPayerArchived, constants.TransactionPayeeArchived, constants.TransactionInsufficientFunds:
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

// Money routes
func (h *MoneyModuleHandler) PayAccountFromBudget(
	ctx context.Context,
	req *pb.PaymentRequest,
) (*pb.PaymentResponse, error) {
	budgetUUID, err := parseUUID(req.GetPayerId(), "payer_id")
	if err != nil {
		return nil, err
	}
	accountUUID, err := parseUUID(req.GetPayeeId(), "payee_id")
	if err != nil {
		return nil, err
	}

	transactionID, txStatus, err := h.Svc.BudgetToAccount(ctx, accountUUID, budgetUUID, req.GetAmount())
	if err != nil {
		return &pb.PaymentResponse{Status: toProtoStatus(txStatus)}, toGRPCError(txStatus, err)
	}

	return &pb.PaymentResponse{
		TransactionId: &pb.UUID{Value: transactionID[:]},
		Status:        toProtoStatus(txStatus),
	}, nil
}

func (h *MoneyModuleHandler) PayBudgetFromAccount(
	ctx context.Context,
	req *pb.PaymentRequest,
) (*pb.PaymentResponse, error) {
	accountUUID, err := parseUUID(req.GetPayerId(), "payer_id")
	if err != nil {
		return nil, err
	}
	budgetUUID, err := parseUUID(req.GetPayeeId(), "payee_id")
	if err != nil {
		return nil, err
	}

	transactionID, txStatus, err := h.Svc.AccountToBudget(ctx, accountUUID, budgetUUID, req.GetAmount())
	if err != nil {
		return &pb.PaymentResponse{Status: toProtoStatus(txStatus)}, toGRPCError(txStatus, err)
	}

	return &pb.PaymentResponse{
		TransactionId: &pb.UUID{Value: transactionID[:]},
		Status:        toProtoStatus(txStatus),
	}, nil
}

func (h *MoneyModuleHandler) PayRootFromAccount(
	ctx context.Context,
	req *pb.RootPaymentRequest,
) (*pb.PaymentResponse, error) {
	accountUUID, err := parseUUID(req.GetNonRootId(), "non_root_id")
	if err != nil {
		return nil, err
	}

	transactionID, txStatus, err := h.Svc.AccountToRootAccount(ctx, accountUUID, req.GetAmount())
	if err != nil {
		return &pb.PaymentResponse{Status: toProtoStatus(txStatus)}, toGRPCError(txStatus, err)
	}

	return &pb.PaymentResponse{
		TransactionId: &pb.UUID{Value: transactionID[:]},
		Status:        toProtoStatus(txStatus),
	}, nil
}

func (h *MoneyModuleHandler) PayAccountFromRoot(
	ctx context.Context,
	req *pb.RootPaymentRequest,
) (*pb.PaymentResponse, error) {
	accountUUID, err := parseUUID(req.GetNonRootId(), "non_root_id")
	if err != nil {
		return nil, err
	}

	transactionID, txStatus, err := h.Svc.RootAccountToAccount(ctx, accountUUID, req.GetAmount())
	if err != nil {
		return &pb.PaymentResponse{Status: toProtoStatus(txStatus)}, toGRPCError(txStatus, err)
	}

	return &pb.PaymentResponse{
		TransactionId: &pb.UUID{Value: transactionID[:]},
		Status:        toProtoStatus(txStatus),
	}, nil
}

func (h *MoneyModuleHandler) SendBetweenBudgets(
	ctx context.Context,
	req *pb.SendRequest,
) (*pb.SendResponse, error) {
	payerUUID, err := parseUUID(req.GetPayerId(), "payer_id")
	if err != nil {
		return nil, err
	}
	payeeUUID, err := parseUUID(req.GetPayeeId(), "payee_id")
	if err != nil {
		return nil, err
	}

	transactionID, txStatus, err := h.Svc.BudgetToBudget(ctx, payerUUID, payeeUUID, req.GetAmount())
	if err != nil {
		return &pb.SendResponse{Status: toProtoStatus(txStatus)}, toGRPCError(txStatus, err)
	}

	return &pb.SendResponse{
		TransactionId: &pb.UUID{Value: transactionID[:]},
		Status:        toProtoStatus(txStatus),
	}, nil
}

func toProtoTransfers(transfers []tb.Transfer) []*pb.Transfer {
	out := make([]*pb.Transfer, 0, len(transfers))
	for _, transfer := range transfers {
		id := transfer.ID.Bytes()
		debit := transfer.DebitAccountID.Bytes()
		credit := transfer.CreditAccountID.Bytes()
		out = append(out, &pb.Transfer{
			TransferId: &pb.UUID{Value: id[:]},
			Sender:     &pb.UUID{Value: debit[:]},
			Receiver:   &pb.UUID{Value: credit[:]},
		})
	}
	return out
}

func (h *MoneyModuleHandler) GetAccountTransfers(
	ctx context.Context,
	req *pb.TransfersRequest,
) (*pb.TransfersResponse, error) {
	accountUUID, err := parseUUID(req.GetId(), "id")
	if err != nil {
		return nil, err
	}

	transfers, txStatus, err := h.Svc.GetAllAccountTransfers(ctx, accountUUID)
	if err != nil {
		return nil, toGRPCError(txStatus, err)
	}

	return &pb.TransfersResponse{Transfers: toProtoTransfers(transfers)}, nil
}

func (h *MoneyModuleHandler) GetBudgetTransfers(
	ctx context.Context,
	req *pb.TransfersRequest,
) (*pb.TransfersResponse, error) {
	budgetUUID, err := parseUUID(req.GetId(), "id")
	if err != nil {
		return nil, err
	}

	transfers, txStatus, err := h.Svc.GetAllBudgetTransfers(ctx, budgetUUID)
	if err != nil {
		return nil, toGRPCError(txStatus, err)
	}

	return &pb.TransfersResponse{Transfers: toProtoTransfers(transfers)}, nil
}

//TODO: Check this function doesn't have to be implemented
// func (h *MoneyModuleHandler) GetSocietyTransfers(
// 	ctx context.Context,
// 	req *emptypb.Empty,
// ) (*pb.TransfersResponse, error) {
// 	return nil, status.Error(
// 		codes.Unimplemented, "method GetSocietyTransfers not implemented")
// }
