package keeper

import (
	"context"
	"time"

	"git.vdb.to/cerc-io/laconicd/x/registry"
)

// InitGenesis initializes the module state from a genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, data *registry.GenesisState) error {
	if err := k.Params.Set(ctx, data.Params); err != nil {
		return err
	}
	headerInfo := k.headerService.GetHeaderInfo(ctx)

	for _, record := range data.Records {
		if err := k.SaveRecord(ctx, record); err != nil {
			return err
		}
		// Add to record expiry queue if expiry time is in the future.
		expiryTime, err := time.Parse(time.RFC3339, record.ExpiryTime)
		if err != nil {
			return err
		}
		if expiryTime.After(headerInfo.Time) {
			if err := k.insertRecordExpiryQueue(ctx, record); err != nil {
				return err
			}
		}

		readableRecord := record.ToReadableRecord()
		if err := k.processAttributes(ctx, readableRecord.Attributes, record.Id); err != nil {
			return err
		}
	}

	for _, authority := range data.Authorities {
		// Only import authorities that are marked active.
		if authority.Entry.Status == registry.AuthorityActive {
			// Reset authority height
			authority.Entry.Height = uint64(headerInfo.Height)
			if err := k.SaveNameAuthority(ctx, authority.Name, authority.Entry); err != nil {
				return err
			}
			// Add authority name to expiry queue.
			if err := k.insertAuthorityExpiryQueue(ctx, authority.Name, authority.Entry.ExpiryTime); err != nil {
				return err
			}
		}
	}

	for _, nameEntry := range data.Names {
		if err := k.SaveNameRecord(ctx, nameEntry.Name, nameEntry.Entry.Latest.Id); err != nil {
			return err
		}
	}
	return nil
}

// ExportGenesis exports the module state to a genesis state.
func (k *Keeper) ExportGenesis(ctx context.Context) (*registry.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	records, _, err := k.PaginatedListRecords(ctx, nil)
	if err != nil {
		return nil, err
	}
	authorityEntries, err := k.ListNameAuthorityRecords(ctx, "")
	if err != nil {
		return nil, err
	}
	names, err := k.ListNameRecords(ctx)
	if err != nil {
		return nil, err
	}

	return &registry.GenesisState{
		Params:      params,
		Records:     records,
		Authorities: authorityEntries,
		Names:       names,
	}, nil
}
