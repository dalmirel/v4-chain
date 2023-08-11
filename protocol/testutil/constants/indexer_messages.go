package constants

import (
	"fmt"

	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/dydxprotocol/v4/indexer/msgsender"
	"github.com/dydxprotocol/v4/lib"
	clobtypes "github.com/dydxprotocol/v4/x/clob/types"
)

// Used to construct the constants below.
var (
	testMessage = msgsender.Message{
		Key:   []byte("key"),
		Value: []byte("value"),
	}
	testOrderId = Order_Alice_Num0_Id7_Clob0_Sell25_Price15_GTB20.OrderId
)

var (
	TestTxBytes      = []byte{0x1, 0x2, 0x3}
	TestTxHashBytes  = tmhash.Sum(TestTxBytes)
	TestTxHashString = lib.TxHash(fmt.Sprintf("%X", TestTxHashBytes))
	TestTxHashHeader = msgsender.MessageHeader{
		Key:   msgsender.TransactionHashHeaderKey,
		Value: TestTxHashBytes,
	}
	TestTxBytes1        = []byte{0x4, 0x5, 0x6}
	TestTxHashBytes1    = tmhash.Sum(TestTxBytes1)
	TestTxHashString1   = lib.TxHash(fmt.Sprintf("%X", TestTxHashBytes1))
	TestOffchainUpdates = &clobtypes.OffchainUpdates{
		PlaceMessages: map[clobtypes.OrderId]msgsender.Message{
			testOrderId: testMessage,
		},
		UpdateMessages: map[clobtypes.OrderId]msgsender.Message{
			testOrderId: testMessage,
		},
		RemoveMessages: map[clobtypes.OrderId]msgsender.Message{
			testOrderId: testMessage,
		},
	}
	TestOffchainMessages = TestOffchainUpdates.GetMessages()
)
