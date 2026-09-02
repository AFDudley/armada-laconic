package gql

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"git.vdb.to/cerc-io/laconicd/gql/schema"
	"github.com/cosmos/cosmos-sdk/client"
)

// NodeDataPath is the path to the laconicd data folder.
var NodeDataPath = os.ExpandEnv("$HOME/.laconicd/data")

func getStatusInfo(client client.Context) (*schema.NodeInfo, *schema.SyncInfo, *schema.ValidatorInfo, error) {
	nodeClient, err := client.GetNode()
	if err != nil {
		return nil, nil, nil, err
	}
	nodeStatus, err := nodeClient.Status(context.Background())
	if err != nil {
		return nil, nil, nil, err
	}

	return &schema.NodeInfo{
			ID:      string(nodeStatus.NodeInfo.ID()),
			Network: nodeStatus.NodeInfo.Network,
			Moniker: nodeStatus.NodeInfo.Moniker,
		}, &schema.SyncInfo{
			LatestBlockHash:   nodeStatus.SyncInfo.LatestBlockHash.String(),
			LatestBlockHeight: strconv.FormatInt(nodeStatus.SyncInfo.LatestBlockHeight, 10),
			LatestBlockTime:   nodeStatus.SyncInfo.LatestBlockTime.String(),
			CatchingUp:        nodeStatus.SyncInfo.CatchingUp,
		}, &schema.ValidatorInfo{
			Address:          nodeStatus.ValidatorInfo.Address.String(),
			VotingPower:      strconv.FormatInt(nodeStatus.ValidatorInfo.VotingPower, 10),
			ProposerPriority: nil,
		}, nil
}

// nolint: all
func getNetInfo(_ client.Context) (string, []*schema.PeerInfo, error) {
	// TODO: Implement

	// nodeClient, err := client.GetNode()
	// if err != nil {
	// 	return "", nil, err
	// }
	// netInfo, err := nodeClient.NetInfo(context.Background())
	// if err != nil {
	// 	return "", nil, err
	// }

	// peersInfo := make([]*PeerInfo, netInfo.NPeers)
	// // TODO: find a way to get the peer information from nodeClient
	// for index, peer := range netInfo.Peers {
	// 	peersInfo[index] = &PeerInfo{
	// 		Node: &NodeInfo{
	// 			ID: string(peer.NodeInfo.ID()),
	// 			// Moniker: peer.Node.Moniker,
	// 			// Network: peer.Node.Network,
	// 		},
	// 		// IsOutbound: peer.IsOutbound,
	// 		// RemoteIP:   peer.RemoteIP,
	// 	}
	// }

	// return strconv.FormatInt(int64(netInfo.NPeers), 10), peersInfo, nil

	return strconv.FormatInt(int64(0), 10), []*schema.PeerInfo{}, nil
}

func getValidatorSet(client client.Context) ([]*schema.ValidatorInfo, error) {
	nodeClient, err := client.GetNode()
	if err != nil {
		return nil, err
	}
	res, err := nodeClient.Validators(context.Background(), nil, nil, nil)
	if err != nil {
		return nil, err
	}

	validatorSet := make([]*schema.ValidatorInfo, len(res.Validators))
	for index, validator := range res.Validators {
		proposerPriority := strconv.FormatInt(validator.ProposerPriority, 10)
		validatorSet[index] = &schema.ValidatorInfo{
			Address:          validator.Address.String(),
			VotingPower:      strconv.FormatInt(validator.VotingPower, 10),
			ProposerPriority: &proposerPriority,
		}
	}

	return validatorSet, nil
}

// GetDiskUsage returns disk usage for the given path.
func GetDiskUsage(dirPath string) (string, error) {
	out, err := exec.Command("du", "-sh", dirPath).Output() // #nosec G204
	if err != nil {
		return "", err
	}

	return strings.Fields(string(out))[0], nil
}
