package vm

import (
	"bytes"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	ChainMetricsPrecompileAddress = common.HexToAddress("0x1000000000000000000000000000000000000004")

	chainTallyAddr    = common.HexToAddress("0x1000000000000000000000000000000000000002")
	slotCurrentEpoch  = common.Hash{0}
	slotLastTallyBase = common.Hash{3}

	selectorCurrentEpoch        = crypto.Keccak256([]byte("currentEpoch()"))[:4]
	selectorLastEpochBlockTally = crypto.Keccak256([]byte("lastEpochBlockTally(address)"))[:4]

	errStatefulPrecompileRequiresState = errors.New("stateful precompile requires state")
)

// chainMetricsPrecompile exposes chain epoch and mining tallies.
type chainMetricsPrecompile struct{}

func (p *chainMetricsPrecompile) RequiredGas(input []byte) uint64 {
	if len(input) < 4 {
		return 0
	}
	switch {
	case bytes.Equal(input[:4], selectorCurrentEpoch):
		return 500
	case bytes.Equal(input[:4], selectorLastEpochBlockTally):
		return 2000
	default:
		return 0
	}
}

func (p *chainMetricsPrecompile) Run(input []byte) ([]byte, error) {
	return nil, errStatefulPrecompileRequiresState
}

func (p *chainMetricsPrecompile) RunStateful(statedb StateDB, input []byte) ([]byte, error) {
	if statedb == nil {
		return nil, errStatefulPrecompileRequiresState
	}
	if len(input) < 4 {
		return nil, nil
	}
	switch {
	case bytes.Equal(input[:4], selectorCurrentEpoch):
		currentEpoch := statedb.GetState(chainTallyAddr, slotCurrentEpoch).Big()
		return common.LeftPadBytes(currentEpoch.Bytes(), 32), nil
	case bytes.Equal(input[:4], selectorLastEpochBlockTally):
		if len(input) < 36 {
			return nil, nil
		}
		miner := common.BytesToAddress(input[16:36])
		baseSlot := statedb.GetState(chainTallyAddr, slotLastTallyBase)
		if baseSlot == (common.Hash{}) {
			return common.LeftPadBytes([]byte{}, 32), nil
		}
		tallySlot := calculateMappingSlot(miner, baseSlot)
		tally := statedb.GetState(chainTallyAddr, tallySlot).Big()
		return common.LeftPadBytes(tally.Bytes(), 32), nil
	default:
		return nil, nil
	}
}

func calculateMappingSlot(addr common.Address, baseSlot common.Hash) common.Hash {
	b := make([]byte, 64)
	copy(b[12:32], addr.Bytes())
	copy(b[32:], baseSlot.Bytes())
	return crypto.Keccak256Hash(b)
}
