package forks

import (
	"math/big"
	"sort"

	"github.com/scroll-tech/go-ethereum/params"
)

// CollectSortedForkHeights returns a sorted set of block numbers that one or more forks are activated on
func CollectSortedForkHeights(config *params.ChainConfig) ([]uint64, map[uint64]bool, map[string]uint64) {
	forkHeightsMap := make(map[uint64]bool)
	forkNameHeightMap := make(map[string]uint64)
	type nameFork struct {
		name  string
		block *big.Int
	}
	for _, fork := range []nameFork{
		{name: "homestead", block: config.HomesteadBlock},
		{name: "daoFork", block: config.DAOForkBlock},
		{name: "eip150", block: config.EIP150Block},
		{name: "eip155", block: config.EIP155Block},
		{name: "eip158", block: config.EIP158Block},
		{name: "byzantium", block: config.ByzantiumBlock},
		{name: "constantinople", block: config.ConstantinopleBlock},
		{name: "petersburg", block: config.PetersburgBlock},
		{name: "istanbul", block: config.IstanbulBlock},
		{name: "muirGlacier", block: config.MuirGlacierBlock},
		{name: "berlin", block: config.BerlinBlock},
		{name: "london", block: config.LondonBlock},
		{name: "arrowGlacier", block: config.ArrowGlacierBlock},
		{name: "archimedes", block: config.ArchimedesBlock},
		{name: "shanghai", block: config.ShanghaiBlock},
		{name: "banach", block: config.BanachBlock},
	} {
		if fork.block == nil {
			continue
		}

		height := fork.block.Uint64()
		if height == 0 {
			continue
		}

		if _, ok := forkHeightsMap[height]; ok {
			continue
		}

		forkHeightsMap[height] = true
		forkNameHeightMap[fork.name] = height
	}

	var forkHeights []uint64
	for height := range forkHeightsMap {
		forkHeights = append(forkHeights, height)
	}
	sort.Slice(forkHeights, func(i, j int) bool {
		return forkHeights[i] < forkHeights[j]
	})
	return forkHeights, forkHeightsMap, forkNameHeightMap
}

// BlocksUntilFork returns the number of blocks until the next fork
// returns 0 if there is no fork scheduled for the future
func BlocksUntilFork(blockHeight uint64, forkHeights []uint64) uint64 {
	for _, forkHeight := range forkHeights {
		if forkHeight > blockHeight {
			return forkHeight - blockHeight
		}
	}
	return 0
}
