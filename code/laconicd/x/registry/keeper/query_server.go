package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	registrytypes "git.vdb.to/cerc-io/laconicd/x/registry"
)

var _ registrytypes.QueryServer = queryServer{}

type queryServer struct {
	k Keeper
}

// NewQueryServerImpl returns an implementation of the module QueryServer.
func NewQueryServerImpl(k Keeper) registrytypes.QueryServer {
	return queryServer{k}
}

func (qs queryServer) Params(ctx context.Context, _ *registrytypes.QueryParamsRequest) (*registrytypes.QueryParamsResponse, error) {
	params, err := qs.k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return &registrytypes.QueryParamsResponse{Params: params}, nil
}

func (qs queryServer) Records(ctx context.Context, req *registrytypes.QueryRecordsRequest) (*registrytypes.QueryRecordsResponse, error) {
	attributes := req.GetAttributes()
	all := req.GetAll()

	var records []registrytypes.Record
	var pageResp *query.PageResponse
	var err error
	if len(attributes) > 0 {
		records, pageResp, err = qs.k.PaginatedRecordsFromAttributes(ctx, attributes, all, req.Pagination)
		if err != nil {
			return nil, err
		}
	} else {
		records, pageResp, err = qs.k.PaginatedListRecords(ctx, req.Pagination)
		if err != nil {
			return nil, err
		}
	}

	return &registrytypes.QueryRecordsResponse{Records: records, Pagination: pageResp}, nil
}

func (qs queryServer) GetRecord(ctx context.Context, req *registrytypes.QueryGetRecordRequest) (*registrytypes.QueryGetRecordResponse, error) {
	id := req.GetId()

	if has, err := qs.k.HasRecord(ctx, req.Id); !has {
		if err != nil {
			return nil, err
		}

		return nil, errorsmod.Wrap(sdkerrors.ErrNotFound, "record not found")
	}

	record, err := qs.k.GetRecordById(ctx, id)
	if err != nil {
		return nil, err
	}

	return &registrytypes.QueryGetRecordResponse{Record: record}, nil
}

func (qs queryServer) GetRecordsByBondId(
	ctx context.Context,
	req *registrytypes.QueryGetRecordsByBondIdRequest,
) (*registrytypes.QueryGetRecordsByBondIdResponse, error) {
	records, err := qs.k.GetRecordsByBondId(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &registrytypes.QueryGetRecordsByBondIdResponse{Records: records}, nil
}

func (qs queryServer) GetRegistryModuleBalance(ctx context.Context,
	_ *registrytypes.QueryGetRegistryModuleBalanceRequest,
) (*registrytypes.QueryGetRegistryModuleBalanceResponse, error) {
	balances := qs.k.GetModuleBalances(ctx)

	return &registrytypes.QueryGetRegistryModuleBalanceResponse{
		Balances: balances,
	}, nil
}

func (qs queryServer) NameRecords(ctx context.Context, _ *registrytypes.QueryNameRecordsRequest) (*registrytypes.QueryNameRecordsResponse, error) {
	nameRecords, err := qs.k.ListNameRecords(ctx)
	if err != nil {
		return nil, err
	}

	return &registrytypes.QueryNameRecordsResponse{Names: nameRecords}, nil
}

func (qs queryServer) Whois(ctx context.Context, req *registrytypes.QueryWhoisRequest) (*registrytypes.QueryWhoisResponse, error) {
	nameAuthority, err := qs.k.GetNameAuthority(ctx, req.GetName())
	if err != nil {
		return nil, err
	}

	return &registrytypes.QueryWhoisResponse{NameAuthority: nameAuthority}, nil
}

func (qs queryServer) Authorities(ctx context.Context, req *registrytypes.QueryAuthoritiesRequest) (*registrytypes.QueryAuthoritiesResponse, error) {
	authorityEntries, err := qs.k.ListNameAuthorityRecords(ctx, req.GetOwner())
	if err != nil {
		return nil, err
	}

	return &registrytypes.QueryAuthoritiesResponse{Authorities: authorityEntries}, nil
}

func (qs queryServer) LookupLrn(ctx context.Context, req *registrytypes.QueryLookupLrnRequest) (*registrytypes.QueryLookupLrnResponse, error) {
	lrn := req.GetLrn()

	lrnExists, err := qs.k.HasNameRecord(ctx, lrn)
	if err != nil {
		return nil, err
	}
	if !lrnExists {
		return nil, errorsmod.Wrap(sdkerrors.ErrNotFound, "LRN not found")
	}

	nameRecord, err := qs.k.LookupNameRecord(ctx, lrn)
	if nameRecord == nil {
		if err != nil {
			return nil, err
		}

		return nil, errorsmod.Wrap(sdkerrors.ErrNotFound, "name record not found")
	}

	return &registrytypes.QueryLookupLrnResponse{Name: nameRecord}, nil
}

func (qs queryServer) ResolveLrn(ctx context.Context, req *registrytypes.QueryResolveLrnRequest) (*registrytypes.QueryResolveLrnResponse, error) {
	lrn := req.GetLrn()
	record, err := qs.k.ResolveLRN(ctx, lrn)
	if record == nil {
		if err != nil {
			return nil, err
		}

		return nil, errorsmod.Wrap(sdkerrors.ErrNotFound, "record not found")
	}

	return &registrytypes.QueryResolveLrnResponse{Record: record}, nil
}
