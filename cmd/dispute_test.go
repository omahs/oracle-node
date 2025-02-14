package cmd

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/mock"
	"io/fs"
	"math/big"
	"razor/cmd/mocks"
	"razor/core/types"
	"razor/path"
	pathMocks "razor/path/mocks"
	"razor/pkg/bindings"
	"razor/utils"
	mocks2 "razor/utils/mocks"
	"reflect"
	"testing"
)

func TestDispute(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	txnOpts, _ := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(31337))

	var (
		client        *ethclient.Client
		config        types.Configurations
		account       types.Account
		epoch         uint32
		blockIndex    uint8
		proposedBlock bindings.StructsBlock
		leafId        uint16
		sortedValues  []*big.Int
		blockManager  *bindings.BlockManager
	)

	type args struct {
		containsStatus              bool
		positionOfCollectionInBlock *big.Int
		finalizeDisputeTxn          *Types.Transaction
		finalizeDisputeErr          error
		hash                        common.Hash
		storeBountyIdErr            error
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "Test 1: When Dispute function executes successfully",
			args: args{
				containsStatus:     false,
				finalizeDisputeTxn: &Types.Transaction{},
				hash:               common.BigToHash(big.NewInt(1)),
			},
			want: nil,
		},
		{
			name: "Test 2: When Dispute function executes successfully without executing giveSorted",
			args: args{
				containsStatus:     true,
				finalizeDisputeTxn: &Types.Transaction{},
				hash:               common.BigToHash(big.NewInt(1)),
			},
			want: nil,
		},
		{
			name: "Test 3: When FinalizeDispute transaction fails",
			args: args{
				containsStatus:     false,
				finalizeDisputeErr: errors.New("finalizeDispute error"),
			},
			want: errors.New("finalizeDispute error"),
		},
		{
			name: "Test 4: When Dispute function executes successfully but there is an error in storing bountyId",
			args: args{
				containsStatus:     false,
				finalizeDisputeTxn: &Types.Transaction{},
				hash:               common.BigToHash(big.NewInt(1)),
				storeBountyIdErr:   errors.New("storeBountyId error"),
			},
			want: errors.New("storeBountyId error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			utilsMock := new(mocks.UtilsInterface)
			cmdUtilsMock := new(mocks.UtilsCmdInterface)
			blockManagerUtilsMock := new(mocks.BlockManagerInterface)
			transactionUtilsMock := new(mocks.TransactionInterface)

			razorUtils = utilsMock
			cmdUtils = cmdUtilsMock
			blockManagerUtils = blockManagerUtilsMock
			transactionUtils = transactionUtilsMock

			utilsMock.On("GetBlockManager", mock.AnythingOfType("*ethclient.Client")).Return(blockManager)
			utilsMock.On("GetTxnOpts", mock.AnythingOfType("types.TransactionOptions")).Return(txnOpts)
			cmdUtilsMock.On("GiveSorted", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
			cmdUtilsMock.On("GetCollectionIdPositionInBlock", mock.Anything, mock.Anything, mock.Anything).Return(tt.args.positionOfCollectionInBlock)
			blockManagerUtilsMock.On("FinalizeDispute", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.args.finalizeDisputeTxn, tt.args.finalizeDisputeErr)
			transactionUtilsMock.On("Hash", mock.Anything).Return(tt.args.hash)
			cmdUtilsMock.On("StoreBountyId", mock.Anything, mock.Anything).Return(tt.args.storeBountyIdErr)
			utilsMock.On("WaitForBlockCompletion", mock.AnythingOfType("*ethclient.Client"), mock.AnythingOfType("string")).Return(nil)

			utils := &UtilsStruct{}

			err := utils.Dispute(client, config, account, epoch, blockIndex, proposedBlock, leafId, sortedValues)
			if err == nil || tt.want == nil {
				if err != tt.want {
					t.Errorf("Error for Dispute function, got = %v, want = %v", err, tt.want)
				}
			} else {
				if err.Error() != tt.want.Error() {
					t.Errorf("Error for Dispute function, got = %v, want = %v", err, tt.want)
				}
			}
		})
	}
}

func TestHandleDispute(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	txnOpts, _ := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(31337))

	var client *ethclient.Client
	var config types.Configurations
	var account types.Account
	var epoch uint32
	var blockNumber *big.Int
	var rogueData types.Rogue
	var blockManager *bindings.BlockManager

	type args struct {
		sortedProposedBlockIds       []uint32
		sortedProposedBlockIdsErr    error
		biggestStake                 *big.Int
		biggestStakeId               uint32
		biggestStakeErr              error
		randomSortedProposedBlockIds []uint32
		medians                      []*big.Int
		revealedCollectionIds        []uint16
		revealedDataMaps             *types.RevealedDataMaps
		mediansErr                   error
		proposedBlock                bindings.StructsBlock
		proposedBlockErr             error
		disputeBiggestStakeTxn       *Types.Transaction
		disputeBiggestStakeErr       error
		Hash                         common.Hash
		idDisputeTxn                 *Types.Transaction
		idDisputeTxnErr              error
		misMatchIndex                int
		leafId                       uint16
		leafIdErr                    error
		disputeErr                   error
		storeBountyIdErr             error
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "Test 1: When HandleDispute function executes successfully",
			args: args{
				sortedProposedBlockIds:       []uint32{3, 1, 2, 0, 4},
				randomSortedProposedBlockIds: []uint32{2, 3, 1, 4, 0},
				biggestStake:                 big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				biggestStakeId:               2,
				medians:                      []*big.Int{big.NewInt(6901548), big.NewInt(498307)},
				revealedCollectionIds:        []uint16{1},
				revealedDataMaps: &types.RevealedDataMaps{
					SortedRevealedValues: nil,
					VoteWeights:          nil,
					InfluenceSum:         nil,
				},
				proposedBlock: bindings.StructsBlock{
					Medians:      []*big.Int{big.NewInt(6901548), big.NewInt(498307)},
					Valid:        true,
					BiggestStake: big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				},
				disputeErr: nil,
			},
			want: nil,
		},
		{
			name: "Test 2: When HandleDispute function executes successfully when there is no dispute case",
			args: args{
				sortedProposedBlockIds:       []uint32{45, 65, 23, 64, 12},
				randomSortedProposedBlockIds: []uint32{23, 64, 12, 65, 23},
				biggestStake:                 big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				biggestStakeId:               2,
				proposedBlock: bindings.StructsBlock{
					Medians:      []*big.Int{big.NewInt(6701548), big.NewInt(478307)},
					BiggestStake: big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				},
				disputeErr: nil,
			},
			want: nil,
		},
		{
			name: "Test 3: When there is an error in getting sortedProposedBlockIds",
			args: args{
				sortedProposedBlockIdsErr: errors.New("sortedProposedBlockIds error"),
				proposedBlock: bindings.StructsBlock{
					Medians: []*big.Int{big.NewInt(6701548), big.NewInt(478307)},
				},
				disputeErr: nil,
			},
			want: errors.New("sortedProposedBlockIds error"),
		},
		{
			name: "Test 4: When there is an error in getting proposedBlock",
			args: args{
				sortedProposedBlockIds:       []uint32{45, 65, 23, 64, 12},
				randomSortedProposedBlockIds: []uint32{23, 64, 12, 65, 23},
				proposedBlockErr:             errors.New("proposedBlock error"),
				disputeErr:                   nil,
			},
			want: nil,
		},
		{
			name: "Test 5: When there is a case of Dispute but block is already disputed",
			args: args{
				sortedProposedBlockIds:       []uint32{45, 65, 23, 64, 12},
				randomSortedProposedBlockIds: []uint32{23, 64, 12, 65, 23},
				biggestStake:                 big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				biggestStakeId:               2,
				proposedBlock: bindings.StructsBlock{
					Medians:      []*big.Int{big.NewInt(6701548), big.NewInt(478307)},
					BiggestStake: big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				},
			},
			want: nil,
		},
		{
			name: "Test 6: When HandleDispute function executes successfully when there is a biggest influence dispute case",
			args: args{
				sortedProposedBlockIds:       []uint32{45, 65, 23, 64, 12},
				randomSortedProposedBlockIds: []uint32{23, 64, 12, 65, 23},
				biggestStake:                 big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				biggestStakeId:               2,
				proposedBlock: bindings.StructsBlock{
					Medians:      []*big.Int{big.NewInt(6701548), big.NewInt(478307)},
					Valid:        true,
					BiggestStake: big.NewInt(1).Mul(big.NewInt(4356), big.NewInt(1e18)),
				},
				disputeBiggestStakeTxn: &Types.Transaction{},
				Hash:                   common.BigToHash(big.NewInt(1)),
				disputeErr:             nil,
			},
			want: nil,
		},
		{
			name: "Test 7: When there is an error in getting biggestInfluenceAndId",
			args: args{
				sortedProposedBlockIds:       []uint32{45, 65, 23, 64, 12},
				randomSortedProposedBlockIds: []uint32{23, 64, 12, 65, 23},
				biggestStakeErr:              errors.New("biggestInfluenceAndIdErr"),
				proposedBlock: bindings.StructsBlock{
					Medians:      []*big.Int{big.NewInt(6701548), big.NewInt(478307)},
					Valid:        true,
					BiggestStake: big.NewInt(1).Mul(big.NewInt(4356), big.NewInt(1e18)),
				},
				disputeBiggestStakeTxn: &Types.Transaction{},
				Hash:                   common.BigToHash(big.NewInt(1)),
				disputeErr:             nil,
			},
			want: errors.New("biggestInfluenceAndIdErr"),
		},

		{
			name: "Test 8: When DisputeBiggestStakeProposed transaction fails",
			args: args{
				sortedProposedBlockIds:       []uint32{45, 65, 23, 64, 12},
				randomSortedProposedBlockIds: []uint32{23, 64, 12, 65, 23},
				biggestStake:                 big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				biggestStakeId:               2,
				proposedBlock: bindings.StructsBlock{
					Medians:      []*big.Int{big.NewInt(6701548), big.NewInt(478307)},
					Valid:        true,
					BiggestStake: big.NewInt(1).Mul(big.NewInt(4356), big.NewInt(1e18)),
				},
				disputeBiggestStakeErr: errors.New("disputeBiggestStake error"),
				Hash:                   common.BigToHash(big.NewInt(1)),
				disputeErr:             nil,
			},
			want: nil,
		},
		{
			name: "Test 9: When there is an error in getting medians",
			args: args{
				sortedProposedBlockIds:       []uint32{45, 65, 23, 64, 12},
				randomSortedProposedBlockIds: []uint32{23, 64, 12, 65, 23},
				biggestStake:                 big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				biggestStakeId:               2,
				mediansErr:                   errors.New("error in getting medians"),
			},
			want: errors.New("error in getting medians"),
		},
		{
			name: "Test 10: When there is an error in fetching Ids from CheckDisputeForIds",
			args: args{
				sortedProposedBlockIds:       []uint32{45, 65, 23, 64, 12},
				randomSortedProposedBlockIds: []uint32{23, 64, 12, 65, 23},
				biggestStake:                 big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				biggestStakeId:               2,
				medians:                      []*big.Int{big.NewInt(6901548), big.NewInt(498307)},
				revealedCollectionIds:        []uint16{1},
				revealedDataMaps: &types.RevealedDataMaps{
					SortedRevealedValues: nil,
					VoteWeights:          nil,
					InfluenceSum:         nil,
				},
				proposedBlock: bindings.StructsBlock{
					Medians:      []*big.Int{big.NewInt(6901548), big.NewInt(498307)},
					Valid:        true,
					BiggestStake: big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				},
				idDisputeTxnErr: errors.New("error in fetching Ids from CheckDisputeForIds"),
				disputeErr:      nil,
			},
			want: nil,
		},
		{
			name: "Test 11: When idDisputeTxn is not nil",
			args: args{
				sortedProposedBlockIds:       []uint32{45, 65, 23, 64, 12},
				randomSortedProposedBlockIds: []uint32{23, 64, 12, 65, 23},
				biggestStake:                 big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				biggestStakeId:               2,
				medians:                      []*big.Int{big.NewInt(6901548), big.NewInt(498307)},
				revealedCollectionIds:        []uint16{1},
				revealedDataMaps: &types.RevealedDataMaps{
					SortedRevealedValues: nil,
					VoteWeights:          nil,
					InfluenceSum:         nil,
				},
				proposedBlock: bindings.StructsBlock{
					Medians:      []*big.Int{big.NewInt(6901548), big.NewInt(498307)},
					Valid:        true,
					BiggestStake: big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				},
				idDisputeTxn: &Types.Transaction{},
				disputeErr:   nil,
			},
			want: nil,
		},
		{
			name: "Test 12: When it is a median dispute case and error in getting leafId",
			args: args{
				sortedProposedBlockIds:       []uint32{45, 65, 23, 64, 12},
				randomSortedProposedBlockIds: []uint32{23, 64, 12, 65, 23},
				biggestStake:                 big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				biggestStakeId:               2,
				medians:                      []*big.Int{big.NewInt(6901548), big.NewInt(498307)},
				revealedCollectionIds:        []uint16{1},
				revealedDataMaps: &types.RevealedDataMaps{
					SortedRevealedValues: nil,
					VoteWeights:          nil,
					InfluenceSum:         nil,
				},
				proposedBlock: bindings.StructsBlock{
					Ids:          []uint16{1, 2},
					Medians:      []*big.Int{big.NewInt(6901548), big.NewInt(498307)},
					Valid:        true,
					BiggestStake: big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				},
				misMatchIndex: 0,
				leafIdErr:     errors.New("error in getting leafId"),
				disputeErr:    nil,
			},
			want: nil,
		},
		{
			name: "Test 13: When there is an error in dispute",
			args: args{
				sortedProposedBlockIds:       []uint32{45, 65, 23, 64, 12},
				randomSortedProposedBlockIds: []uint32{23, 64, 12, 65, 23},
				biggestStake:                 big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				biggestStakeId:               2,
				medians:                      []*big.Int{big.NewInt(6901548), big.NewInt(498307)},
				revealedCollectionIds:        []uint16{1},
				revealedDataMaps: &types.RevealedDataMaps{
					SortedRevealedValues: nil,
					VoteWeights:          nil,
					InfluenceSum:         nil,
				},
				proposedBlock: bindings.StructsBlock{
					Ids:          []uint16{1, 2},
					Medians:      []*big.Int{big.NewInt(6901548), big.NewInt(498307)},
					Valid:        true,
					BiggestStake: big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				},
				misMatchIndex: 0,
				leafId:        1,
				disputeErr:    errors.New("error in dispute"),
			},
			want: nil,
		},
		{
			name: "Test 14: When there is a biggest influence dispute case but there is an error in storing bountyId",
			args: args{
				sortedProposedBlockIds:       []uint32{45, 65, 23, 64, 12},
				randomSortedProposedBlockIds: []uint32{23, 64, 12, 65, 23},
				biggestStake:                 big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				biggestStakeId:               2,
				proposedBlock: bindings.StructsBlock{
					Medians:      []*big.Int{big.NewInt(6701548), big.NewInt(478307)},
					Valid:        true,
					BiggestStake: big.NewInt(1).Mul(big.NewInt(4356), big.NewInt(1e18)),
				},
				disputeBiggestStakeTxn: &Types.Transaction{},
				Hash:                   common.BigToHash(big.NewInt(1)),
				disputeErr:             nil,
				storeBountyIdErr:       errors.New("storeBountyId error"),
			},
			want: nil,
		},
		{
			name: "Test 15: When there is a idsDispute case but there is an error in storing bountyId",
			args: args{
				sortedProposedBlockIds:       []uint32{45, 65, 23, 64, 12},
				randomSortedProposedBlockIds: []uint32{23, 64, 12, 65, 23},
				biggestStake:                 big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				biggestStakeId:               2,
				medians:                      []*big.Int{big.NewInt(6901548), big.NewInt(498307)},
				revealedCollectionIds:        []uint16{1},
				revealedDataMaps: &types.RevealedDataMaps{
					SortedRevealedValues: nil,
					VoteWeights:          nil,
					InfluenceSum:         nil,
				},
				proposedBlock: bindings.StructsBlock{
					Medians:      []*big.Int{big.NewInt(6901548), big.NewInt(498307)},
					Valid:        true,
					BiggestStake: big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				},
				idDisputeTxn:     &Types.Transaction{},
				storeBountyIdErr: errors.New("storeBountyId error"),
				disputeErr:       nil,
			},
			want: nil,
		},
		{
			name: "Test 16: When HandleDispute function executes successfully and medians proposed are empty",
			args: args{
				sortedProposedBlockIds:       []uint32{45, 65, 23, 64, 12},
				randomSortedProposedBlockIds: []uint32{23, 64, 12, 65, 23},
				biggestStake:                 big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				biggestStakeId:               2,
				medians:                      []*big.Int{big.NewInt(6901548), big.NewInt(498307)},
				revealedCollectionIds:        []uint16{1},
				revealedDataMaps: &types.RevealedDataMaps{
					SortedRevealedValues: nil,
					VoteWeights:          nil,
					InfluenceSum:         nil,
				},
				proposedBlock: bindings.StructsBlock{
					Medians:      []*big.Int{},
					Valid:        true,
					BiggestStake: big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				},
				disputeErr: nil,
			},
			want: nil,
		},
		{
			name: "Test 17: When there is a case of blockIndex = -1",
			args: args{
				sortedProposedBlockIds:       []uint32{45, 65, 23, 64, 12},
				randomSortedProposedBlockIds: []uint32{23, 64, 12, 65, 23},
				biggestStake:                 big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				biggestStakeId:               2,
				medians:                      []*big.Int{big.NewInt(6901548), big.NewInt(498307)},
				revealedCollectionIds:        []uint16{1},
				revealedDataMaps: &types.RevealedDataMaps{
					SortedRevealedValues: nil,
					VoteWeights:          nil,
					InfluenceSum:         nil,
				},
				proposedBlock: bindings.StructsBlock{
					Medians:      []*big.Int{big.NewInt(6901548), big.NewInt(498307)},
					Valid:        true,
					BiggestStake: big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				},
				disputeErr: nil,
			},
			want: nil,
		},
		{
			name: "Test 18: When HandleDispute function executes successfully and contains different values in sortedProposedBlockIds",
			args: args{
				sortedProposedBlockIds:       []uint32{45, 65, 23, 64, 12},
				randomSortedProposedBlockIds: []uint32{23, 64, 12, 65, 23},
				biggestStake:                 big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				biggestStakeId:               2,
				medians:                      []*big.Int{big.NewInt(6901548), big.NewInt(498307)},
				revealedCollectionIds:        []uint16{1},
				revealedDataMaps: &types.RevealedDataMaps{
					SortedRevealedValues: nil,
					VoteWeights:          nil,
					InfluenceSum:         nil,
				},
				proposedBlock: bindings.StructsBlock{
					Medians:      []*big.Int{big.NewInt(6901548), big.NewInt(498307)},
					Valid:        true,
					BiggestStake: big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)),
				},
				disputeErr: nil,
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			utilsMock := new(mocks.UtilsInterface)
			cmdUtilsMock := new(mocks.UtilsCmdInterface)
			blockManagerUtilsMock := new(mocks.BlockManagerInterface)
			transactionUtilsMock := new(mocks.TransactionInterface)
			utilsPkgMock := new(mocks2.Utils)

			razorUtils = utilsMock
			cmdUtils = cmdUtilsMock
			blockManagerUtils = blockManagerUtilsMock
			transactionUtils = transactionUtilsMock

			utils.UtilsInterface = utilsPkgMock

			utilsMock.On("GetSortedProposedBlockIds", mock.AnythingOfType("*ethclient.Client"), mock.AnythingOfType("uint32")).Return(tt.args.sortedProposedBlockIds, tt.args.sortedProposedBlockIdsErr)
			cmdUtilsMock.On("GetBiggestStakeAndId", mock.AnythingOfType("*ethclient.Client"), mock.AnythingOfType("string"), mock.AnythingOfType("uint32")).Return(tt.args.biggestStake, tt.args.biggestStakeId, tt.args.biggestStakeErr)
			cmdUtilsMock.On("GetLocalMediansData", mock.AnythingOfType("*ethclient.Client"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.args.medians, tt.args.revealedCollectionIds, tt.args.revealedDataMaps, tt.args.mediansErr)
			utilsPkgMock.On("Shuffle", mock.Anything).Return(tt.args.randomSortedProposedBlockIds)
			utilsMock.On("GetProposedBlock", mock.AnythingOfType("*ethclient.Client"), mock.AnythingOfType("uint32"), mock.AnythingOfType("uint32")).Return(tt.args.proposedBlock, tt.args.proposedBlockErr)
			utilsMock.On("GetTxnOpts", mock.AnythingOfType("types.TransactionOptions")).Return(txnOpts)
			blockManagerUtilsMock.On("DisputeBiggestStakeProposed", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.args.disputeBiggestStakeTxn, tt.args.disputeBiggestStakeErr)
			transactionUtilsMock.On("Hash", mock.Anything).Return(tt.args.Hash)
			utilsMock.On("WaitForBlockCompletion", mock.AnythingOfType("*ethclient.Client"), mock.AnythingOfType("string")).Return(nil)
			cmdUtilsMock.On("CheckDisputeForIds", mock.AnythingOfType("*ethclient.Client"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.args.idDisputeTxn, tt.args.idDisputeTxnErr)
			utilsPkgMock.On("GetLeafIdOfACollection", mock.AnythingOfType("*ethclient.Client"), mock.Anything).Return(tt.args.leafId, tt.args.leafIdErr)
			cmdUtilsMock.On("Dispute", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.args.disputeErr)
			utilsMock.On("GetBlockManager", mock.AnythingOfType("*ethclient.Client")).Return(blockManager)
			cmdUtilsMock.On("StoreBountyId", mock.Anything, mock.Anything).Return(tt.args.storeBountyIdErr)
			cmdUtilsMock.On("ResetDispute", mock.AnythingOfType("*ethclient.Client"), mock.Anything, mock.Anything, mock.Anything)

			utils := &UtilsStruct{}
			err := utils.HandleDispute(client, config, account, epoch, blockNumber, rogueData)
			if err == nil || tt.want == nil {
				if err != tt.want {
					t.Errorf("Error for HandleDispute function, got = %v, want = %v", err, tt.want)
				}
			} else {
				if err.Error() != tt.want.Error() {
					t.Errorf("Error for HandleDispute function, got = %v, want = %v", err, tt.want)
				}
			}
		})
	}
}

func TestGiveSorted(t *testing.T) {
	var client *ethclient.Client
	var blockManager *bindings.BlockManager
	var txnOpts *bind.TransactOpts
	var epoch uint32
	var assetId uint16
	type args struct {
		sortedValues  []*big.Int
		giveSorted    *Types.Transaction
		giveSortedErr error
		hash          common.Hash
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test 1: When Give Sorted executes successfully",
			args: args{
				sortedValues: []*big.Int{big.NewInt(2), big.NewInt(1), big.NewInt(3), big.NewInt(5)},
				giveSorted:   &Types.Transaction{},
				hash:         common.BigToHash(big.NewInt(1)),
			},
		},
		{
			name: "Test 2: When there is an error from GiveSorted",
			args: args{
				sortedValues:  []*big.Int{big.NewInt(2), big.NewInt(1), big.NewInt(3), big.NewInt(5)},
				giveSortedErr: errors.New("giveSorted error"),
			},
		},
		{
			name: "Test 3: When sortedStakers is nil",
			args: args{
				sortedValues: nil,
			},
		},
		{
			name: "Test 4: When error is gas limit reached",
			args: args{
				sortedValues:  []*big.Int{big.NewInt(2), big.NewInt(1), big.NewInt(3), big.NewInt(5)},
				giveSortedErr: errors.New("gas limit reached"),
				giveSorted:    &Types.Transaction{},
				hash:          common.BigToHash(big.NewInt(1)),
			},
		},
		{
			name: "Test 5: When error is gas limit reached with higher number of stakers",
			args: args{
				sortedValues:  []*big.Int{big.NewInt(2), big.NewInt(1), big.NewInt(3), big.NewInt(5), big.NewInt(7), big.NewInt(8), big.NewInt(9), big.NewInt(10), big.NewInt(6), big.NewInt(11), big.NewInt(13), big.NewInt(12), big.NewInt(14), big.NewInt(15), big.NewInt(4), big.NewInt(20), big.NewInt(19), big.NewInt(18), big.NewInt(17), big.NewInt(16)},
				giveSortedErr: errors.New("gas limit reached"),
				giveSorted:    &Types.Transaction{},
				hash:          common.BigToHash(big.NewInt(1)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utilsMock := new(mocks.UtilsInterface)
			blockManagerUtilsMock := new(mocks.BlockManagerInterface)
			transactionUtilsMock := new(mocks.TransactionInterface)

			razorUtils = utilsMock
			blockManagerUtils = blockManagerUtilsMock
			transactionUtils = transactionUtilsMock

			blockManagerUtilsMock.On("GiveSorted", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.args.giveSorted, tt.args.giveSortedErr).Once()
			transactionUtilsMock.On("Hash", mock.Anything).Return(tt.args.hash)
			utilsMock.On("WaitForBlockCompletion", mock.AnythingOfType("*ethclient.Client"), mock.AnythingOfType("string")).Return(nil)
			blockManagerUtilsMock.On("GiveSorted", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.args.giveSorted, nil)

			GiveSorted(client, blockManager, txnOpts, epoch, assetId, tt.args.sortedValues)
		})
	}
}

func TestGetLocalMediansData(t *testing.T) {
	var (
		client      *ethclient.Client
		account     types.Account
		blockNumber *big.Int
		rogueData   types.Rogue
	)
	type args struct {
		epoch                 uint32
		fileName              string
		fileNameErr           error
		proposedData          types.ProposeFileData
		proposeDataErr        error
		medians               []*big.Int
		revealedCollectionIds []uint16
		revealedDataMaps      *types.RevealedDataMaps
		mediansErr            error
		stakerId              uint32
		stakerIdErr           error
		lastProposedEpoch     uint32
		lastProposedEpochErr  error
	}
	tests := []struct {
		name    string
		args    args
		want    []*big.Int
		want1   []uint16
		want2   *types.RevealedDataMaps
		wantErr bool
	}{
		{
			name: "Test 1: When there is an error in getting fileName",
			args: args{
				fileNameErr: errors.New("error in getting fileName"),
			},
			want:    nil,
			want1:   nil,
			want2:   nil,
			wantErr: false,
		},
		{
			name: "Test 2: When there is an error in getting proposedData",
			args: args{
				fileName:       "",
				proposeDataErr: errors.New("error in getting proposedData"),
			},
			want:    nil,
			want1:   nil,
			want2:   nil,
			wantErr: false,
		},
		{
			name: "Test 3: When file does not contain latest data",
			args: args{
				fileName:     "",
				proposedData: types.ProposeFileData{Epoch: 3},
				epoch:        5,
			},
			want:    nil,
			want1:   nil,
			want2:   nil,
			wantErr: false,
		},
		{
			name: "Test 4: When there is an error in getting medians",
			args: args{
				mediansErr: errors.New("error in fetching medians"),
			},
			want:    nil,
			want1:   nil,
			want2:   nil,
			wantErr: true,
		},
		{
			name: "Test 5: When GetLocalMediansData executes successfully",
			args: args{
				medians:               []*big.Int{big.NewInt(100), big.NewInt(200), big.NewInt(300)},
				revealedCollectionIds: []uint16{1, 2, 3},
				revealedDataMaps:      &types.RevealedDataMaps{},
			},
			want:    []*big.Int{big.NewInt(100), big.NewInt(200), big.NewInt(300)},
			want1:   []uint16{1, 2, 3},
			want2:   &types.RevealedDataMaps{},
			wantErr: false,
		},
		{
			name: "Test 6: When there is an error in getting stakerId",
			args: args{
				medians:               []*big.Int{big.NewInt(100), big.NewInt(200), big.NewInt(300)},
				revealedCollectionIds: []uint16{1, 2, 3},
				revealedDataMaps:      &types.RevealedDataMaps{},
				stakerIdErr:           errors.New("stakerId error"),
			},
			want:    nil,
			want1:   nil,
			want2:   nil,
			wantErr: true,
		},
		{
			name: "Test 7: When there is an error in getting last proposed epoch",
			args: args{
				medians:               []*big.Int{big.NewInt(100), big.NewInt(200), big.NewInt(300)},
				revealedCollectionIds: []uint16{1, 2, 3},
				revealedDataMaps:      &types.RevealedDataMaps{},
				stakerId:              2,
				epoch:                 5,
				lastProposedEpochErr:  errors.New("lastProposedEpoch error"),
			},
			want:    nil,
			want1:   nil,
			want2:   nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utilsMock := new(mocks.UtilsInterface)
			cmdUtilsMock := new(mocks.UtilsCmdInterface)

			razorUtils = utilsMock
			cmdUtils = cmdUtilsMock

			utilsMock.On("GetProposeDataFileName", mock.AnythingOfType("string")).Return(tt.args.fileName, tt.args.fileNameErr)
			utilsMock.On("ReadFromProposeJsonFile", mock.Anything).Return(tt.args.proposedData, tt.args.proposeDataErr)
			cmdUtilsMock.On("MakeBlock", mock.AnythingOfType("*ethclient.Client"), mock.Anything, mock.Anything, mock.Anything).Return(tt.args.medians, tt.args.revealedCollectionIds, tt.args.revealedDataMaps, tt.args.mediansErr)
			utilsMock.On("GetStakerId", mock.AnythingOfType("*ethclient.Client"), mock.AnythingOfType("string")).Return(tt.args.stakerId, tt.args.stakerIdErr)
			cmdUtilsMock.On("GetLastProposedEpoch", mock.AnythingOfType("*ethclient.Client"), mock.AnythingOfType("*big.Int"), mock.AnythingOfType("uint32")).Return(tt.args.lastProposedEpoch, tt.args.lastProposedEpochErr)
			ut := &UtilsStruct{}
			got, got1, got2, err := ut.GetLocalMediansData(client, account, tt.args.epoch, blockNumber, rogueData)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLocalMediansData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLocalMediansData() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("GetLocalMediansData() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("GetLocalMediansData() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestGetCollectionIdPositionInBlock(t *testing.T) {
	var client *ethclient.Client
	var leafId uint16
	type args struct {
		proposedBlock     bindings.StructsBlock
		idToBeDisputed    uint16
		idToBeDisputedErr error
	}
	tests := []struct {
		name string
		args args
		want *big.Int
	}{
		{
			name: "Test 1: When GetCollectionIdPositionInBlock() executes successfully",
			args: args{
				proposedBlock: bindings.StructsBlock{
					Ids: []uint16{1, 2, 3},
				},
				idToBeDisputed: 1,
			},
			want: big.NewInt(0),
		},
		{
			name: "Test 2: When there is an error in getting idToBeDisputed",
			args: args{
				proposedBlock: bindings.StructsBlock{
					Ids: []uint16{1, 2, 3},
				},
				idToBeDisputedErr: errors.New("error in fetching idToBeDisputed"),
			},
			want: nil,
		},
		{
			name: "Test 3: When idToBeDisputes is not present in proposedBlock.Ids",
			args: args{
				proposedBlock: bindings.StructsBlock{
					Ids: []uint16{1, 2, 3},
				},
				idToBeDisputed: 4,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utilsPkgMock := new(mocks2.Utils)

			utils.UtilsInterface = utilsPkgMock

			utilsPkgMock.On("GetCollectionIdFromLeafId", mock.AnythingOfType("*ethclient.Client"), mock.Anything).Return(tt.args.idToBeDisputed, tt.args.idToBeDisputedErr)
			ut := &UtilsStruct{}
			if got := ut.GetCollectionIdPositionInBlock(client, leafId, tt.args.proposedBlock); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCollectionIdPositionInBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckDisputeForIds(t *testing.T) {
	var (
		client          *ethclient.Client
		transactionOpts types.TransactionOptions
		epoch           uint32
		blockIndex      uint8
	)

	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	txnOpts, _ := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(31337))

	type args struct {
		idsInProposedBlock                    []uint16
		revealedCollectionIds                 []uint16
		DisputeOnOrderOfIds                   *Types.Transaction
		DisputeOnOrderOfIdsErr                error
		incrementedGasLimit                   uint64
		incrementedGasLimitErr                error
		DisputeCollectionIdShouldBePresent    *Types.Transaction
		DisputeCollectionIdShouldBePresentErr error
		DisputeCollectionIdShouldBeAbsent     *Types.Transaction
		DisputeCollectionIdShouldBeAbsentErr  error
	}
	tests := []struct {
		name    string
		args    args
		want    *Types.Transaction
		wantErr bool
	}{
		{
			name: "Test 1: When CheckDisputeForIds executes successfully with sorted Ids dispute case",
			args: args{
				idsInProposedBlock:    []uint16{1, 3, 2},
				revealedCollectionIds: []uint16{1, 2, 3},
				DisputeOnOrderOfIds:   &Types.Transaction{},
			},
			want:    &Types.Transaction{},
			wantErr: false,
		},
		{
			name: "Test 2: When CheckDisputeForIds executes successfully with collectionIdShouldBePresent dispute case",
			args: args{
				idsInProposedBlock:                 []uint16{1, 2, 4},
				revealedCollectionIds:              []uint16{1, 2, 3},
				DisputeCollectionIdShouldBePresent: &Types.Transaction{},
			},
			want:    &Types.Transaction{},
			wantErr: false,
		},
		{
			name: "Test 3: When its a collectionIdShouldBePresent dispute case and their is an error in incrementalGasLimit",
			args: args{
				idsInProposedBlock:     []uint16{1, 2, 4},
				revealedCollectionIds:  []uint16{1, 2, 3},
				incrementedGasLimitErr: errors.New("error in incremented gas limit"),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test 4: When CheckDisputeForIds executes successfully with collectionIdShouldBeAbsent case",
			args: args{
				idsInProposedBlock:                []uint16{1, 2, 3, 4},
				revealedCollectionIds:             []uint16{1, 2, 3},
				DisputeCollectionIdShouldBeAbsent: &Types.Transaction{},
			},
			want:    &Types.Transaction{},
			wantErr: false,
		},
		{
			name: "Test 5: When its a collectionIdShouldBeAbsent dispute case and their is an error in incrementalGasLimit",
			args: args{
				idsInProposedBlock:                []uint16{1, 2, 3, 4},
				revealedCollectionIds:             []uint16{1, 2, 3},
				DisputeCollectionIdShouldBeAbsent: &Types.Transaction{},
				incrementedGasLimitErr:            errors.New("error in incremented gas limit"),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test 5: When there is no dispute",
			args: args{
				idsInProposedBlock:    []uint16{1, 2, 3},
				revealedCollectionIds: []uint16{1, 2, 3},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utilsMock := new(mocks.UtilsInterface)
			blockManagerUtilsMock := new(mocks.BlockManagerInterface)
			utilsPkgMock := new(mocks2.Utils)

			utils.UtilsInterface = utilsPkgMock
			razorUtils = utilsMock
			blockManagerUtils = blockManagerUtilsMock
			utilsInterface = utilsPkgMock

			utilsMock.On("GetTxnOpts", mock.AnythingOfType("types.TransactionOptions")).Return(txnOpts)
			blockManagerUtilsMock.On("DisputeOnOrderOfIds", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.args.DisputeOnOrderOfIds, tt.args.DisputeOnOrderOfIdsErr)
			utilsPkgMock.On("IncreaseGasLimitValue", mock.Anything, mock.Anything, mock.Anything).Return(tt.args.incrementedGasLimit, tt.args.incrementedGasLimitErr)
			blockManagerUtilsMock.On("DisputeCollectionIdShouldBePresent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.args.DisputeCollectionIdShouldBePresent, tt.args.DisputeCollectionIdShouldBePresentErr)
			blockManagerUtilsMock.On("DisputeCollectionIdShouldBeAbsent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.args.DisputeCollectionIdShouldBeAbsent, tt.args.DisputeCollectionIdShouldBeAbsentErr)
			utilsPkgMock.On("IncreaseGasLimitValue", mock.Anything, mock.Anything, mock.Anything).Return(uint64(2000), nil)
			ut := &UtilsStruct{}
			got, err := ut.CheckDisputeForIds(client, transactionOpts, epoch, blockIndex, tt.args.idsInProposedBlock, tt.args.revealedCollectionIds)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckDisputeForIds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckDisputeForIds() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBountyIdFromEvents(t *testing.T) {
	var (
		client       *ethclient.Client
		blockNumber  *big.Int
		bountyHunter string
	)
	type args struct {
		fromBlock      *big.Int
		fromBlockErr   error
		logs           []Types.Log
		logsErr        error
		contractABI    abi.ABI
		contractABIErr error
		data           []interface{}
		unpackErr      error
	}
	tests := []struct {
		name    string
		args    args
		want    uint32
		wantErr bool
	}{
		{
			name: "Test 1: When GetBountyIdFromEvents() executes successfully",
			args: args{
				fromBlock: big.NewInt(0),
				logs: []Types.Log{
					{
						Data:   []byte{2},
						Topics: []common.Hash{common.BigToHash(big.NewInt(1)), common.HexToHash("0x000000000000000000000000000000000000dead")},
					},
				},
				contractABI: abi.ABI{},
				data:        convertToSliceOfInterface([]uint32{4, 2}),
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "Test 2: When there is an error in getting blockNumber",
			args: args{
				fromBlockErr: errors.New("error in getting blockNumber"),
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "Test 3: When there is an error in getting logs",
			args: args{
				fromBlock: big.NewInt(0),
				logsErr:   errors.New("error in getting logs"),
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "Test 4: When there is an error in getting contractABI",
			args: args{
				fromBlock: big.NewInt(0),
				logs: []Types.Log{
					{
						Data: []byte{4, 2},
					},
				},
				contractABIErr: errors.New("error in contractABI"),
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "Test 5: When there is an error in unpacking",
			args: args{
				fromBlock: big.NewInt(0),
				logs: []Types.Log{
					{
						Data: []byte{4, 2},
					},
				},
				contractABI: abi.ABI{},
				unpackErr:   errors.New("error in unpacking"),
			},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abiMock := new(mocks.AbiInterface)
			utilsPkgMock := new(mocks2.Utils)
			abiUtilsMock := new(mocks2.ABIUtils)
			utilsPkgMock2 := new(mocks2.Utils)

			utilsInterface = utilsPkgMock2
			abiUtils = abiMock
			utils.UtilsInterface = utilsPkgMock
			utils.ABIInterface = abiUtilsMock

			utilsPkgMock.On("CalculateBlockNumberAtEpochBeginning", mock.AnythingOfType("*ethclient.Client"), mock.Anything, mock.Anything).Return(tt.args.fromBlock, tt.args.fromBlockErr)
			abiUtilsMock.On("Parse", mock.Anything).Return(tt.args.contractABI, tt.args.contractABIErr)
			utilsPkgMock.On("FilterLogsWithRetry", mock.AnythingOfType("*ethclient.Client"), mock.AnythingOfType("ethereum.FilterQuery")).Return(tt.args.logs, tt.args.logsErr)
			abiMock.On("Unpack", mock.Anything, mock.Anything, mock.Anything).Return(tt.args.data, tt.args.unpackErr)
			ut := &UtilsStruct{}
			got, err := ut.GetBountyIdFromEvents(client, blockNumber, bountyHunter)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBountyIdFromEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetBountyIdFromEvents() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResetDispute(t *testing.T) {
	var (
		client       *ethclient.Client
		blockManager *bindings.BlockManager
		txnOpts      *bind.TransactOpts
		epoch        uint32
	)
	type args struct {
		ResetDisputeTxn    *Types.Transaction
		ResetDisputeTxnErr error
		hash               common.Hash
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test 1: When ResetDispute() executes successfully",
			args: args{
				ResetDisputeTxn: &Types.Transaction{},
				hash:            common.Hash{1},
			},
		},
		{
			name: "Test 2: When there is an error in executing ResetDispute()",
			args: args{
				ResetDisputeTxnErr: errors.New("error in resetting dispute"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utilsMock := new(mocks.UtilsInterface)
			blockManagerMock := new(mocks.BlockManagerInterface)
			transactionUtilsMock := new(mocks.TransactionInterface)

			razorUtils = utilsMock
			blockManagerUtils = blockManagerMock
			transactionUtils = transactionUtilsMock

			blockManagerMock.On("ResetDispute", mock.Anything, mock.Anything, mock.Anything).Return(tt.args.ResetDisputeTxn, tt.args.ResetDisputeTxnErr)
			transactionUtilsMock.On("Hash", mock.AnythingOfType("*types.Transaction")).Return(tt.args.hash)
			utilsMock.On("WaitForBlockCompletion", mock.AnythingOfType("*ethclient.Client"), mock.AnythingOfType("string")).Return(nil)

			ut := &UtilsStruct{}
			ut.ResetDispute(client, blockManager, txnOpts, epoch)
		})
	}
}

func BenchmarkGetCollectionIdPositionInBlock(b *testing.B) {
	var client *ethclient.Client
	var leafId uint16
	var table = []struct {
		numOfIds       uint16
		idToBeDisputed uint16
	}{
		{numOfIds: 10, idToBeDisputed: 9},
		{numOfIds: 100, idToBeDisputed: 99},
		{numOfIds: 1000, idToBeDisputed: 999},
		{numOfIds: 10000, idToBeDisputed: 9999},
	}
	for _, v := range table {
		b.Run(fmt.Sprintf("Number_Of_Ids_In_ProposedBlock_%d", v.numOfIds), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				utilsPkgMock := new(mocks2.Utils)

				utils.UtilsInterface = utilsPkgMock

				utilsPkgMock.On("GetCollectionIdFromLeafId", mock.AnythingOfType("*ethclient.Client"), mock.Anything).Return(v.idToBeDisputed, nil)
				ut := &UtilsStruct{}
				ut.GetCollectionIdPositionInBlock(client, leafId, bindings.StructsBlock{Ids: getDummyIds(v.numOfIds)})
			}
		})
	}
}

func BenchmarkHandleDispute(b *testing.B) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	txnOpts, _ := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(31337))

	var client *ethclient.Client
	var config types.Configurations
	var account types.Account
	var epoch uint32
	var blockNumber *big.Int
	var rogueData types.Rogue
	var blockManager *bindings.BlockManager
	var randomSortedPorposedBlockIds []uint32

	table := []struct {
		numOfSortedBlocks uint32
	}{
		{numOfSortedBlocks: 5},
		{numOfSortedBlocks: 50},
		{numOfSortedBlocks: 500},
		{numOfSortedBlocks: 5000},
	}
	for _, v := range table {
		b.Run(fmt.Sprintf("Number_Of_Sorted_Blocks_Proposed%d", v.numOfSortedBlocks), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				utilsMock := new(mocks.UtilsInterface)
				cmdUtilsMock := new(mocks.UtilsCmdInterface)
				blockManagerUtilsMock := new(mocks.BlockManagerInterface)
				transactionUtilsMock := new(mocks.TransactionInterface)
				utilsPkgMock := new(mocks2.Utils)

				razorUtils = utilsMock
				cmdUtils = cmdUtilsMock
				blockManagerUtils = blockManagerUtilsMock
				transactionUtils = transactionUtilsMock

				utils.UtilsInterface = utilsPkgMock

				medians := []*big.Int{big.NewInt(6901548), big.NewInt(498307)}
				revealedCollectionIds := []uint16{1}
				revealedDataMaps := &types.RevealedDataMaps{
					SortedRevealedValues: nil,
					VoteWeights:          nil,
					InfluenceSum:         nil}
				proposedBlock := bindings.StructsBlock{
					Medians:      []*big.Int{big.NewInt(6901548), big.NewInt(498307)},
					Valid:        true,
					BiggestStake: big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18))}

				utilsMock.On("GetSortedProposedBlockIds", mock.AnythingOfType("*ethclient.Client"), mock.AnythingOfType("uint32")).Return(getUint32DummyIds(v.numOfSortedBlocks), nil)
				cmdUtilsMock.On("GetBiggestStakeAndId", mock.AnythingOfType("*ethclient.Client"), mock.AnythingOfType("string"), mock.AnythingOfType("uint32")).Return(big.NewInt(1).Mul(big.NewInt(5356), big.NewInt(1e18)), uint32(2), nil)
				cmdUtilsMock.On("GetLocalMediansData", mock.AnythingOfType("*ethclient.Client"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(medians, revealedCollectionIds, revealedDataMaps, nil)
				utilsPkgMock.On("Shuffle", mock.Anything).Return(randomSortedPorposedBlockIds)
				utilsMock.On("GetProposedBlock", mock.AnythingOfType("*ethclient.Client"), mock.AnythingOfType("uint32"), mock.AnythingOfType("uint32")).Return(proposedBlock, nil)
				utilsMock.On("GetTxnOpts", mock.AnythingOfType("types.TransactionOptions")).Return(txnOpts)
				blockManagerUtilsMock.On("DisputeBiggestStakeProposed", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&Types.Transaction{}, nil)
				transactionUtilsMock.On("Hash", mock.Anything).Return(common.BigToHash(big.NewInt(1)))
				utilsMock.On("WaitForBlockCompletion", mock.AnythingOfType("*ethclient.Client"), mock.AnythingOfType("string")).Return(nil)
				cmdUtilsMock.On("CheckDisputeForIds", mock.AnythingOfType("*ethclient.Client"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&Types.Transaction{}, nil)
				utilsPkgMock.On("IsEqualUint32", mock.Anything, mock.Anything).Return(true, 0)
				utilsPkgMock.On("GetLeafIdOfACollection", mock.AnythingOfType("*ethclient.Client"), mock.Anything).Return(0, nil)
				cmdUtilsMock.On("Dispute", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				utilsMock.On("GetBlockManager", mock.AnythingOfType("*ethclient.Client")).Return(blockManager)
				cmdUtilsMock.On("StoreBountyId", mock.Anything, mock.Anything).Return(nil)
				cmdUtilsMock.On("ResetDispute", mock.AnythingOfType("*ethclient.Client"), mock.Anything, mock.Anything, mock.Anything)

				utils := &UtilsStruct{}
				err := utils.HandleDispute(client, config, account, epoch, blockNumber, rogueData)
				if err != nil {
					log.Fatal(err)
				}
			}
		})
	}

}

func TestStoreBountyId(t *testing.T) {
	var (
		client   *ethclient.Client
		account  types.Account
		fileInfo fs.FileInfo
	)
	type args struct {
		disputeFilePath    string
		disputeFilePathErr error
		disputedFlag       bool
		latestHeader       *Types.Header
		latestHeaderErr    error
		latestBountyId     uint32
		latestBountyIdErr  error
		statErr            error
		disputeData        types.DisputeFileData
		disputeDataErr     error
		saveDataErr        error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test 1: When StoreBountyId() executes successfully",
			args: args{
				disputeFilePath: "",
				disputedFlag:    true,
				latestHeader:    &Types.Header{Number: big.NewInt(1)},
				latestBountyId:  1,
				statErr:         nil,
				disputeData:     types.DisputeFileData{BountyIdQueue: []uint32{1}},
				saveDataErr:     nil,
			},
			wantErr: false,
		},
		{
			name: "Test 2: When StoreBountyId() executes successfully and there are more than one bountyId in queue",
			args: args{
				disputeFilePath: "",
				disputedFlag:    true,
				latestHeader:    &Types.Header{Number: big.NewInt(1)},
				latestBountyId:  1,
				statErr:         nil,
				disputeData:     types.DisputeFileData{BountyIdQueue: []uint32{1, 2}},
				saveDataErr:     nil,
			},
			wantErr: false,
		},
		{
			name: "Test 3: When there is an error in getting disputeFilePath",
			args: args{
				disputeFilePathErr: errors.New("error in getting disputeFilePath"),
			},
			wantErr: true,
		},
		{
			name: "Test 4: When there is an error in getting latest header",
			args: args{
				disputeFilePath: "",
				disputedFlag:    true,
				latestHeaderErr: errors.New("error in getting latest header"),
			},
			wantErr: true,
		},
		{
			name: "Test 5: When there is an error in not getting latest bountyId",
			args: args{
				disputeFilePath:   "",
				disputedFlag:      true,
				latestHeader:      &Types.Header{Number: big.NewInt(1)},
				latestBountyIdErr: errors.New("error in getting latest bountyId"),
			},
			wantErr: true,
		},
		{
			name: "Test 6: When there is an error in getting disputeData",
			args: args{
				disputeFilePath: "",
				disputedFlag:    true,
				latestHeader:    &Types.Header{Number: big.NewInt(1)},
				latestBountyId:  1,
				statErr:         nil,
				disputeDataErr:  errors.New("error in getting diapute data"),
			},
			wantErr: true,
		},
		{
			name: "When there is an error in saving data to file",
			args: args{
				disputeFilePath: "",
				disputedFlag:    true,
				latestHeader:    &Types.Header{Number: big.NewInt(1)},
				latestBountyId:  1,
				statErr:         nil,
				disputeData:     types.DisputeFileData{BountyIdQueue: []uint32{1}},
				saveDataErr:     errors.New("error in saving data to file"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utilsMock := new(mocks.UtilsInterface)
			cmdUtilsMock := new(mocks.UtilsCmdInterface)
			utilsPkgMock := new(mocks2.Utils)
			osUtilsMock := new(pathMocks.OSInterface)

			razorUtils = utilsMock
			cmdUtils = cmdUtilsMock
			utils.UtilsInterface = utilsPkgMock
			utilsInterface = utilsPkgMock
			path.OSUtilsInterface = osUtilsMock

			utilsMock.On("GetDisputeDataFileName", mock.AnythingOfType("string")).Return(tt.args.disputeFilePath, tt.args.disputeFilePathErr)
			utilsPkgMock.On("GetLatestBlockWithRetry", mock.AnythingOfType("*ethclient.Client")).Return(tt.args.latestHeader, tt.args.latestHeaderErr)
			cmdUtilsMock.On("GetBountyIdFromEvents", mock.Anything, mock.Anything, mock.Anything).Return(tt.args.latestBountyId, tt.args.latestBountyIdErr)
			osUtilsMock.On("Stat", mock.Anything).Return(fileInfo, tt.args.statErr)
			utilsMock.On("ReadFromDisputeJsonFile", mock.Anything).Return(tt.args.disputeData, tt.args.disputeDataErr)
			utilsMock.On("SaveDataToDisputeJsonFile", mock.Anything, mock.Anything, mock.Anything).Return(tt.args.saveDataErr)

			ut := &UtilsStruct{}
			if err := ut.StoreBountyId(client, account); (err != nil) != tt.wantErr {
				t.Errorf("AutoClaimBounty() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func getDummyIds(numOfIds uint16) []uint16 {
	var result []uint16
	for i := uint16(1); i <= numOfIds; i++ {
		result = append(result, i)
	}
	return result
}

func getUint32DummyIds(numOfIds uint32) []uint32 {
	var result []uint32
	for i := uint32(1); i < numOfIds; i++ {
		result = append(result, i)
	}
	return result
}
