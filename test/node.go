package test

import (
	"gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"

	"github.com/OpenBazaar/openbazaar-go/core"
	"github.com/OpenBazaar/openbazaar-go/net"
	"github.com/OpenBazaar/openbazaar-go/net/service"
	"github.com/OpenBazaar/spvwallet"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ipfs/go-ipfs/core/mock"
)

// NewNode creates a new *core.OpenBazaarNode prepared for testing
func NewNode(repository *Repository) (*core.OpenBazaarNode, error) {
	// Create test ipfs node
	ipfsNode, err := coremock.NewMockNode()
	if err != nil {
		return nil, err
	}

	spvwalletConfig := &spvwallet.Config{
		Mnemonic:  getMnemonic(),
		Params:    &chaincfg.RegressionNetParams,
		MaxFee:    50000,
		LowFee:    8000,
		MediumFee: 16000,
		HighFee:   24000,
		RepoPath:  repository.Path,
		DB:        repository.DB,
		UserAgent: "OpenBazaar",
		Proxy:     nil,
		Logger:    NewLogger(),
	}

	spvwallet.LOOKAHEADWINDOW = 1
	wallet, err := spvwallet.NewSPVWallet(spvwalletConfig)
	if err != nil {
		return nil, err
	}

	// Put it all together in an OpenBazaarNode
	node := &core.OpenBazaarNode{
		RepoPath:   repository.Path,
		IpfsNode:   ipfsNode,
		Datastore:  repository.DB,
		Wallet:     wallet,
		BanManager: net.NewBanManager([]peer.ID{}),
	}

	node.Service = service.New(node, repository.DB)

	return node, nil
}
