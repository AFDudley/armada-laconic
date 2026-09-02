package gql

import (
	"context"
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	types "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"git.vdb.to/cerc-io/laconicd/gql/schema"
	"git.vdb.to/cerc-io/laconicd/utils"
	auctiontypes "git.vdb.to/cerc-io/laconicd/x/auction"
	bondtypes "git.vdb.to/cerc-io/laconicd/x/bond"
	onboardingTypes "git.vdb.to/cerc-io/laconicd/x/onboarding"
	registrytypes "git.vdb.to/cerc-io/laconicd/x/registry"
)

// DefaultLogNumLines is the number of log lines to tail by default.
const DefaultLogNumLines = 50

// MaxLogNumLines is the max number of log lines that can be tailed.
const MaxLogNumLines = 1000

// Whether to use default page limit when pagination args are not passed.
const UseDefaultPagination = false

type Resolver struct {
	ctx     client.Context
	logFile string
}

// Query is the entry point to query execution.
func (r *Resolver) Query() schema.QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (q queryResolver) GetAuthorities(ctx context.Context, owner *string) ([]*schema.Authority, error) {
	nsQueryClient := registrytypes.NewQueryClient(q.ctx)
	auctionQueryClient := auctiontypes.NewQueryClient(q.ctx)

	authoritiesReq := &registrytypes.QueryAuthoritiesRequest{}
	if owner != nil {
		authoritiesReq.Owner = *owner
	}

	authoritiesResp, err := nsQueryClient.Authorities(ctx, authoritiesReq)
	if err != nil {
		return nil, err
	}

	authorities := make([]*schema.Authority, len(authoritiesResp.GetAuthorities()))
	for i, a := range authoritiesResp.Authorities {
		entry, err := getAuthorityRecord(*a.Entry, auctionQueryClient)
		if err != nil {
			return nil, err
		}

		authorities[i] = &schema.Authority{
			Name:  a.Name,
			Entry: entry,
		}
	}

	return authorities, nil
}

func (q queryResolver) LookupAuthorities(ctx context.Context, names []string) ([]*schema.AuthorityRecord, error) {
	nsQueryClient := registrytypes.NewQueryClient(q.ctx)
	auctionQueryClient := auctiontypes.NewQueryClient(q.ctx)
	gqlResponse := []*schema.AuthorityRecord{}

	for _, name := range names {
		res, err := nsQueryClient.Whois(context.Background(), &registrytypes.QueryWhoisRequest{Name: name})
		if err != nil {
			if strings.Contains(err.Error(), "Name authority not found") {
				gqlResponse = append(gqlResponse, nil)
				continue
			}

			return nil, err
		}

		nameAuthority := res.GetNameAuthority()
		gqlNameAuthorityRecord, err := getAuthorityRecord(nameAuthority, auctionQueryClient)
		if err != nil {
			return nil, err
		}

		gqlResponse = append(gqlResponse, gqlNameAuthorityRecord)
	}

	return gqlResponse, nil
}

func (q queryResolver) ResolveNames(ctx context.Context, names []string) ([]*schema.Record, error) {
	nsQueryClient := registrytypes.NewQueryClient(q.ctx)
	var gqlResponse []*schema.Record
	for _, name := range names {
		res, err := nsQueryClient.ResolveLrn(context.Background(), &registrytypes.QueryResolveLrnRequest{Lrn: name})
		if err != nil {
			// Return nil for record not found.
			gqlResponse = append(gqlResponse, nil)
		} else {
			gqlRecord, err := getGQLRecord(context.Background(), q, *res.GetRecord())
			if err != nil {
				return nil, err
			}

			gqlResponse = append(gqlResponse, gqlRecord)
		}
	}

	return gqlResponse, nil
}

func (q queryResolver) LookupNames(ctx context.Context, names []string) ([]*schema.NameRecord, error) {
	nsQueryClient := registrytypes.NewQueryClient(q.ctx)
	var gqlResponse []*schema.NameRecord

	for _, name := range names {
		res, err := nsQueryClient.LookupLrn(context.Background(), &registrytypes.QueryLookupLrnRequest{Lrn: name})
		if err != nil {
			// Return nil for name not found.
			gqlResponse = append(gqlResponse, nil)
		} else {
			gqlRecord, err := getGQLNameRecord(res.GetName())
			if err != nil {
				return nil, err
			}

			gqlResponse = append(gqlResponse, gqlRecord)
		}
	}

	return gqlResponse, nil
}

func (q queryResolver) QueryRecords(ctx context.Context, attributes []*schema.KeyValueInput, all *bool, limit *int, offset *int) ([]*schema.Record, error) {
	nsQueryClient := registrytypes.NewQueryClient(q.ctx)

	var pagination *query.PageRequest

	// Use defaults only if limit and offset not provided
	// and UseDefaultPagination is true
	if limit == nil && offset == nil {
		if UseDefaultPagination {
			pagination = &query.PageRequest{}
		}
	} else {
		pagination = &query.PageRequest{}

		if limit != nil {
			pagination.Limit = uint64(*limit)
		}
		if offset != nil {
			pagination.Offset = uint64(*offset)
		}
	}

	res, err := nsQueryClient.Records(
		context.Background(),
		&registrytypes.QueryRecordsRequest{
			Attributes: toRPCAttributes(attributes),
			All:        (all != nil && *all),
			Pagination: pagination,
		},
	)
	if err != nil {
		return nil, err
	}

	records := res.GetRecords()
	gqlResponse := make([]*schema.Record, len(records))

	for i, record := range records {
		gqlRecord, err := getGQLRecord(context.Background(), q, record)
		if err != nil {
			return nil, err
		}
		gqlResponse[i] = gqlRecord
	}

	return gqlResponse, nil
}

func (q queryResolver) GetRecordsByIds(ctx context.Context, ids []string) ([]*schema.Record, error) {
	nsQueryClient := registrytypes.NewQueryClient(q.ctx)
	gqlResponse := make([]*schema.Record, len(ids))

	for i, id := range ids {
		res, err := nsQueryClient.GetRecord(context.Background(), &registrytypes.QueryGetRecordRequest{Id: id})
		if err != nil {
			// Return nil for record not found.
			gqlResponse[i] = nil
		} else {
			record, err := getGQLRecord(context.Background(), q, res.GetRecord())
			if err != nil {
				return nil, err
			}
			gqlResponse[i] = record
		}
	}

	return gqlResponse, nil
}

func (q queryResolver) GetStatus(ctx context.Context) (*schema.Status, error) {
	nodeInfo, syncInfo, validatorInfo, err := getStatusInfo(q.ctx)
	if err != nil {
		return nil, err
	}

	numPeers, peers, err := getNetInfo(q.ctx)
	if err != nil {
		return nil, err
	}

	validatorSet, err := getValidatorSet(q.ctx)
	if err != nil {
		return nil, err
	}

	diskUsage, err := GetDiskUsage(NodeDataPath)
	if err != nil {
		return nil, err
	}

	return &schema.Status{
		Version:    RegistryVersion,
		Node:       nodeInfo,
		Sync:       syncInfo,
		Validator:  validatorInfo,
		Validators: validatorSet,
		NumPeers:   numPeers,
		Peers:      peers,
		DiskUsage:  diskUsage,
	}, nil
}

func (q queryResolver) GetAccounts(ctx context.Context, addresses []string) ([]*schema.Account, error) {
	accounts := make([]*schema.Account, len(addresses))
	for index, address := range addresses {
		account, err := q.GetAccount(ctx, address)
		if err != nil {
			return nil, err
		}
		accounts[index] = account
	}
	return accounts, nil
}

func (q queryResolver) GetAccount(ctx context.Context, address string) (*schema.Account, error) {
	authQueryClient := authtypes.NewQueryClient(q.ctx)
	accountResponse, err := authQueryClient.Account(ctx, &authtypes.QueryAccountRequest{Address: address})
	if err != nil {
		return nil, err
	}
	var account types.AccountI
	err = q.ctx.Codec.UnpackAny(accountResponse.GetAccount(), &account)
	if err != nil {
		return nil, err
	}
	var pubKey *string
	if account.GetPubKey() != nil {
		pubKeyStr := base64.StdEncoding.EncodeToString(account.GetPubKey().Bytes())
		pubKey = &pubKeyStr
	}

	// Get the account balance
	bankQueryClient := banktypes.NewQueryClient(q.ctx)
	balance, err := bankQueryClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: address})
	if err != nil {
		return nil, err
	}
	accNum := strconv.FormatUint(account.GetAccountNumber(), 10)
	seq := strconv.FormatUint(account.GetSequence(), 10)

	return &schema.Account{
		Address:  address,
		Number:   accNum,
		Sequence: seq,
		PubKey:   pubKey,
		Balance:  getGQLCoins(balance.GetBalances()),
	}, nil
}

func (q queryResolver) GetBondsByIds(ctx context.Context, ids []string) ([]*schema.Bond, error) {
	bonds := make([]*schema.Bond, len(ids))
	for index, id := range ids {
		bondObj, err := q.GetBond(ctx, id)
		if err != nil {
			return nil, err
		}
		bonds[index] = bondObj
	}

	return bonds, nil
}

func (q *queryResolver) GetBond(ctx context.Context, id string) (*schema.Bond, error) {
	bondQueryClient := bondtypes.NewQueryClient(q.ctx)
	bondResp, err := bondQueryClient.GetBondById(context.Background(), &bondtypes.QueryGetBondByIdRequest{Id: id})
	if err != nil {
		if strings.Contains(err.Error(), "Bond not found") {
			return nil, nil
		}

		return nil, err
	}

	bond := bondResp.GetBond()
	if bond == nil {
		return nil, nil
	}
	return getGQLBond(bondResp.GetBond())
}

func (q queryResolver) QueryBonds(ctx context.Context) ([]*schema.Bond, error) {
	bondQueryClient := bondtypes.NewQueryClient(q.ctx)
	bonds, err := bondQueryClient.Bonds(context.Background(), &bondtypes.QueryBondsRequest{})
	if err != nil {
		return nil, err
	}

	gqlResponse := make([]*schema.Bond, len(bonds.GetBonds()))
	for i, bondObj := range bonds.GetBonds() {
		gqlBond, err := getGQLBond(bondObj)
		if err != nil {
			return nil, err
		}
		gqlResponse[i] = gqlBond
	}

	return gqlResponse, nil
}

// QueryBondsByOwner will return bonds by owner
func (q queryResolver) QueryBondsByOwner(ctx context.Context, ownerAddresses []string) ([]*schema.OwnerBonds, error) {
	ownerBonds := make([]*schema.OwnerBonds, len(ownerAddresses))
	for index, ownerAddress := range ownerAddresses {
		bondsObj, err := q.GetBondsByOwner(ctx, ownerAddress)
		if err != nil {
			return nil, err
		}
		ownerBonds[index] = bondsObj
	}

	return ownerBonds, nil
}

func (q queryResolver) GetBondsByOwner(ctx context.Context, address string) (*schema.OwnerBonds, error) {
	bondQueryClient := bondtypes.NewQueryClient(q.ctx)
	bondResp, err := bondQueryClient.GetBondsByOwner(context.Background(), &bondtypes.QueryGetBondsByOwnerRequest{Owner: address})
	if err != nil {
		return nil, err
	}

	ownerBonds := make([]*schema.Bond, len(bondResp.GetBonds()))
	for i, bond := range bondResp.GetBonds() {
		// #nosec G601
		bondObj, err := getGQLBond(&bond) //nolint: all
		if err != nil {
			return nil, err
		}
		ownerBonds[i] = bondObj
	}

	return &schema.OwnerBonds{Bonds: ownerBonds, Owner: address}, nil
}

func (q queryResolver) GetAuctionsByIds(ctx context.Context, ids []string) ([]*schema.Auction, error) {
	auctionQueryClient := auctiontypes.NewQueryClient(q.ctx)
	gqlAuctionResponse := make([]*schema.Auction, len(ids))
	for i, id := range ids {
		auctionObj, err := auctionQueryClient.GetAuction(context.Background(), &auctiontypes.QueryGetAuctionRequest{Id: id})
		if err != nil {
			return nil, err
		}
		bidsObj, err := auctionQueryClient.GetBids(context.Background(), &auctiontypes.QueryGetBidsRequest{AuctionId: id})
		if err != nil {
			return nil, err
		}

		gqlAuction, err := GetGQLAuction(auctionObj.GetAuction(), bidsObj.GetBids())
		if err != nil {
			return nil, err
		}

		gqlAuctionResponse[i] = gqlAuction
	}

	return gqlAuctionResponse, nil
}

func (q queryResolver) GetParticipants(ctx context.Context) ([]*schema.Participant, error) {
	onboardingQueryClient := onboardingTypes.NewQueryClient(q.ctx)
	participantResp, err := onboardingQueryClient.Participants(context.Background(), &onboardingTypes.QueryParticipantsRequest{})
	if err != nil {
		return nil, err
	}

	participants := make([]*schema.Participant, len(participantResp.GetParticipants()))
	for i, p := range participantResp.Participants {
		nitroAddress, err := utils.EthAddressFromPubKey(p.PublicKey)
		if err != nil {
			return nil, err
		}
		participants[i] = &schema.Participant{
			CosmosAddress: p.CosmosAddress,
			NitroAddress:  nitroAddress.String(),
			Role:          p.Role,
			KycID:         p.KycId,
		}
	}

	return participants, nil
}

func (q queryResolver) GetParticipantByAddress(ctx context.Context, address string) (*schema.Participant, error) {
	onboardingQueryClient := onboardingTypes.NewQueryClient(q.ctx)
	participantResp, err := onboardingQueryClient.GetParticipantByAddress(ctx, &onboardingTypes.QueryGetParticipantByAddressRequest{Address: address})
	if err != nil {
		return nil, err
	}

	p := participantResp.Participant
	nitroAddress, err := utils.EthAddressFromPubKey(p.PublicKey)
	if err != nil {
		return nil, err
	}
	participant := &schema.Participant{
		CosmosAddress: p.CosmosAddress,
		NitroAddress:  nitroAddress.String(),
		Role:          p.Role,
		KycID:         p.KycId,
	}

	return participant, nil
}

func (q queryResolver) GetParticipantByNitroAddress(ctx context.Context, nitroAddress string) (*schema.Participant, error) {
	onboardingQueryClient := onboardingTypes.NewQueryClient(q.ctx)
	participantResp, err := onboardingQueryClient.GetParticipantByNitroAddress(
		ctx,
		&onboardingTypes.QueryGetParticipantByNitroAddressRequest{
			NitroAddress: nitroAddress,
		},
	)
	if err != nil {
		return nil, err
	}

	p := participantResp.Participant
	participant := &schema.Participant{
		CosmosAddress: p.CosmosAddress,
		NitroAddress:  nitroAddress,
		Role:          p.Role,
		KycID:         p.KycId,
	}

	return participant, nil
}
