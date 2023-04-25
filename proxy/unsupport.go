package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/filecoin-project/go-address"
	datatransfer "github.com/filecoin-project/go-data-transfer"
	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/builtin/v8/paych"
	"github.com/filecoin-project/go-state-types/crypto"
	api1 "github.com/filecoin-project/lotus/api"
	apitypes "github.com/filecoin-project/lotus/api/types"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/journal/alerting"
	marketevents "github.com/filecoin-project/lotus/markets/loggers"
	"github.com/filecoin-project/lotus/node/repo/imports"
	"github.com/google/uuid"
	"github.com/ipfs-force-community/chain-co/api"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-libipfs/blocks"
	"github.com/libp2p/go-libp2p/core/metrics"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

var _ UnSupportAPI = (*UnSupport)(nil)

type UnSupportAPI interface {
	api.UnSupport
}

type UnSupport struct {
	Select func(types.TipSetKey) (UnSupportAPI, error)
}

// impl api.UnSupport
func (p *UnSupport) AuthNew(in0 context.Context, in1 []string) (out0 []uint8, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api AuthNew %v", err)
		return
	}
	return cli.AuthNew(in0, in1)
}

func (p *UnSupport) AuthVerify(in0 context.Context, in1 string) (out0 []string, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api AuthVerify %v", err)
		return
	}
	return cli.AuthVerify(in0, in1)
}

func (p *UnSupport) ChainBlockstoreInfo(in0 context.Context) (out0 map[string]interface{}, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ChainBlockstoreInfo %v", err)
		return
	}
	return cli.ChainBlockstoreInfo(in0)
}

func (p *UnSupport) ChainCheckBlockstore(in0 context.Context) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ChainCheckBlockstore %v", err)
		return
	}
	return cli.ChainCheckBlockstore(in0)
}

func (p *UnSupport) ChainDeleteObj(in0 context.Context, in1 cid.Cid) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ChainDeleteObj %v", err)
		return
	}
	return cli.ChainDeleteObj(in0, in1)
}

func (p *UnSupport) ChainExport(in0 context.Context, in1 abi.ChainEpoch, in2 bool, in3 types.TipSetKey) (out0 <-chan []uint8, err error) {
	cli, err := p.Select(in3)
	if err != nil {
		err = fmt.Errorf("api ChainExport %v", err)
		return
	}
	return cli.ChainExport(in0, in1, in2, in3)
}

func (p *UnSupport) ChainGetNode(in0 context.Context, in1 string) (out0 *api1.IpldObject, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ChainGetNode %v", err)
		return
	}
	return cli.ChainGetNode(in0, in1)
}

func (p *UnSupport) ChainPrune(in0 context.Context, in1 api1.PruneOpts) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ChainPrune %v", err)
		return
	}
	return cli.ChainPrune(in0, in1)
}

func (p *UnSupport) ChainPutObj(in0 context.Context, in1 blocks.Block) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ChainPutObj %v", err)
		return
	}
	return cli.ChainPutObj(in0, in1)
}

func (p *UnSupport) ChainSetHead(in0 context.Context, in1 types.TipSetKey) (err error) {
	cli, err := p.Select(in1)
	if err != nil {
		err = fmt.Errorf("api ChainSetHead %v", err)
		return
	}
	return cli.ChainSetHead(in0, in1)
}

func (p *UnSupport) ClientCalcCommP(in0 context.Context, in1 string) (out0 *api1.CommPRet, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientCalcCommP %v", err)
		return
	}
	return cli.ClientCalcCommP(in0, in1)
}

func (p *UnSupport) ClientCancelDataTransfer(in0 context.Context, in1 datatransfer.TransferID, in2 peer.ID, in3 bool) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientCancelDataTransfer %v", err)
		return
	}
	return cli.ClientCancelDataTransfer(in0, in1, in2, in3)
}

func (p *UnSupport) ClientCancelRetrievalDeal(in0 context.Context, in1 retrievalmarket.DealID) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientCancelRetrievalDeal %v", err)
		return
	}
	return cli.ClientCancelRetrievalDeal(in0, in1)
}

func (p *UnSupport) ClientDataTransferUpdates(in0 context.Context) (out0 <-chan api1.DataTransferChannel, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientDataTransferUpdates %v", err)
		return
	}
	return cli.ClientDataTransferUpdates(in0)
}

func (p *UnSupport) ClientDealPieceCID(in0 context.Context, in1 cid.Cid) (out0 api1.DataCIDSize, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientDealPieceCID %v", err)
		return
	}
	return cli.ClientDealPieceCID(in0, in1)
}

func (p *UnSupport) ClientDealSize(in0 context.Context, in1 cid.Cid) (out0 api1.DataSize, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientDealSize %v", err)
		return
	}
	return cli.ClientDealSize(in0, in1)
}

func (p *UnSupport) ClientExport(in0 context.Context, in1 api1.ExportRef, in2 api1.FileRef) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientExport %v", err)
		return
	}
	return cli.ClientExport(in0, in1, in2)
}

func (p *UnSupport) ClientFindData(in0 context.Context, in1 cid.Cid, in2 *cid.Cid) (out0 []api1.QueryOffer, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientFindData %v", err)
		return
	}
	return cli.ClientFindData(in0, in1, in2)
}

func (p *UnSupport) ClientGenCar(in0 context.Context, in1 api1.FileRef, in2 string) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientGenCar %v", err)
		return
	}
	return cli.ClientGenCar(in0, in1, in2)
}

func (p *UnSupport) ClientGetDealInfo(in0 context.Context, in1 cid.Cid) (out0 *api1.DealInfo, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientGetDealInfo %v", err)
		return
	}
	return cli.ClientGetDealInfo(in0, in1)
}

func (p *UnSupport) ClientGetDealStatus(in0 context.Context, in1 uint64) (out0 string, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientGetDealStatus %v", err)
		return
	}
	return cli.ClientGetDealStatus(in0, in1)
}

func (p *UnSupport) ClientGetDealUpdates(in0 context.Context) (out0 <-chan api1.DealInfo, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientGetDealUpdates %v", err)
		return
	}
	return cli.ClientGetDealUpdates(in0)
}

func (p *UnSupport) ClientGetRetrievalUpdates(in0 context.Context) (out0 <-chan api1.RetrievalInfo, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientGetRetrievalUpdates %v", err)
		return
	}
	return cli.ClientGetRetrievalUpdates(in0)
}

func (p *UnSupport) ClientHasLocal(in0 context.Context, in1 cid.Cid) (out0 bool, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientHasLocal %v", err)
		return
	}
	return cli.ClientHasLocal(in0, in1)
}

func (p *UnSupport) ClientImport(in0 context.Context, in1 api1.FileRef) (out0 *api1.ImportRes, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientImport %v", err)
		return
	}
	return cli.ClientImport(in0, in1)
}

func (p *UnSupport) ClientListDataTransfers(in0 context.Context) (out0 []api1.DataTransferChannel, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientListDataTransfers %v", err)
		return
	}
	return cli.ClientListDataTransfers(in0)
}

func (p *UnSupport) ClientListDeals(in0 context.Context) (out0 []api1.DealInfo, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientListDeals %v", err)
		return
	}
	return cli.ClientListDeals(in0)
}

func (p *UnSupport) ClientListImports(in0 context.Context) (out0 []api1.Import, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientListImports %v", err)
		return
	}
	return cli.ClientListImports(in0)
}

func (p *UnSupport) ClientListRetrievals(in0 context.Context) (out0 []api1.RetrievalInfo, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientListRetrievals %v", err)
		return
	}
	return cli.ClientListRetrievals(in0)
}

func (p *UnSupport) ClientMinerQueryOffer(in0 context.Context, in1 address.Address, in2 cid.Cid, in3 *cid.Cid) (out0 api1.QueryOffer, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientMinerQueryOffer %v", err)
		return
	}
	return cli.ClientMinerQueryOffer(in0, in1, in2, in3)
}

func (p *UnSupport) ClientQueryAsk(in0 context.Context, in1 peer.ID, in2 address.Address) (out0 *api1.StorageAsk, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientQueryAsk %v", err)
		return
	}
	return cli.ClientQueryAsk(in0, in1, in2)
}

func (p *UnSupport) ClientRemoveImport(in0 context.Context, in1 imports.ID) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientRemoveImport %v", err)
		return
	}
	return cli.ClientRemoveImport(in0, in1)
}

func (p *UnSupport) ClientRestartDataTransfer(in0 context.Context, in1 datatransfer.TransferID, in2 peer.ID, in3 bool) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientRestartDataTransfer %v", err)
		return
	}
	return cli.ClientRestartDataTransfer(in0, in1, in2, in3)
}

func (p *UnSupport) ClientRetrieve(in0 context.Context, in1 api1.RetrievalOrder) (out0 *api1.RestrievalRes, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientRetrieve %v", err)
		return
	}
	return cli.ClientRetrieve(in0, in1)
}

func (p *UnSupport) ClientRetrieveTryRestartInsufficientFunds(in0 context.Context, in1 address.Address) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientRetrieveTryRestartInsufficientFunds %v", err)
		return
	}
	return cli.ClientRetrieveTryRestartInsufficientFunds(in0, in1)
}

func (p *UnSupport) ClientRetrieveWait(in0 context.Context, in1 retrievalmarket.DealID) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientRetrieveWait %v", err)
		return
	}
	return cli.ClientRetrieveWait(in0, in1)
}

func (p *UnSupport) ClientRetrieveWithEvents(in0 context.Context, in1 api1.RetrievalOrder, in2 *api1.FileRef) (out0 <-chan marketevents.RetrievalEvent, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientRetrieveWithEvents %v", err)
		return
	}
	return cli.ClientRetrieveWithEvents(in0, in1, in2)
}

func (p *UnSupport) ClientStartDeal(in0 context.Context, in1 *api1.StartDealParams) (out0 *cid.Cid, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientStartDeal %v", err)
		return
	}
	return cli.ClientStartDeal(in0, in1)
}

func (p *UnSupport) ClientStatelessDeal(in0 context.Context, in1 *api1.StartDealParams) (out0 *cid.Cid, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ClientStatelessDeal %v", err)
		return
	}
	return cli.ClientStatelessDeal(in0, in1)
}

func (p *UnSupport) Closing(in0 context.Context) (out0 <-chan struct{}, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api Closing %v", err)
		return
	}
	return cli.Closing(in0)
}

func (p *UnSupport) CreateBackup(in0 context.Context, in1 string) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api CreateBackup %v", err)
		return
	}
	return cli.CreateBackup(in0, in1)
}

func (p *UnSupport) Discover(in0 context.Context) (out0 apitypes.OpenRPCDocument, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api Discover %v", err)
		return
	}
	return cli.Discover(in0)
}

func (p *UnSupport) ID(in0 context.Context) (out0 peer.ID, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api ID %v", err)
		return
	}
	return cli.ID(in0)
}

func (p *UnSupport) LogAlerts(in0 context.Context) (out0 []alerting.Alert, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api LogAlerts %v", err)
		return
	}
	return cli.LogAlerts(in0)
}

func (p *UnSupport) LogList(in0 context.Context) (out0 []string, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api LogList %v", err)
		return
	}
	return cli.LogList(in0)
}

func (p *UnSupport) LogSetLevel(in0 context.Context, in1 string, in2 string) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api LogSetLevel %v", err)
		return
	}
	return cli.LogSetLevel(in0, in1, in2)
}

func (p *UnSupport) MarketAddBalance(in0 context.Context, in1 address.Address, in2 address.Address, in3 big.Int) (out0 cid.Cid, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MarketAddBalance %v", err)
		return
	}
	return cli.MarketAddBalance(in0, in1, in2, in3)
}

func (p *UnSupport) MarketGetReserved(in0 context.Context, in1 address.Address) (out0 big.Int, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MarketGetReserved %v", err)
		return
	}
	return cli.MarketGetReserved(in0, in1)
}

func (p *UnSupport) MarketReleaseFunds(in0 context.Context, in1 address.Address, in2 big.Int) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MarketReleaseFunds %v", err)
		return
	}
	return cli.MarketReleaseFunds(in0, in1, in2)
}

func (p *UnSupport) MarketReserveFunds(in0 context.Context, in1 address.Address, in2 address.Address, in3 big.Int) (out0 cid.Cid, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MarketReserveFunds %v", err)
		return
	}
	return cli.MarketReserveFunds(in0, in1, in2, in3)
}

func (p *UnSupport) MarketWithdraw(in0 context.Context, in1 address.Address, in2 address.Address, in3 big.Int) (out0 cid.Cid, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MarketWithdraw %v", err)
		return
	}
	return cli.MarketWithdraw(in0, in1, in2, in3)
}

func (p *UnSupport) MpoolBatchPushMessage(in0 context.Context, in1 []*types.Message, in2 *api1.MessageSendSpec) (out0 []*types.SignedMessage, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MpoolBatchPushMessage %v", err)
		return
	}
	return cli.MpoolBatchPushMessage(in0, in1, in2)
}

func (p *UnSupport) MpoolCheckMessages(in0 context.Context, in1 []*api1.MessagePrototype) (out0 [][]api1.MessageCheckStatus, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MpoolCheckMessages %v", err)
		return
	}
	return cli.MpoolCheckMessages(in0, in1)
}

func (p *UnSupport) MpoolCheckPendingMessages(in0 context.Context, in1 address.Address) (out0 [][]api1.MessageCheckStatus, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MpoolCheckPendingMessages %v", err)
		return
	}
	return cli.MpoolCheckPendingMessages(in0, in1)
}

func (p *UnSupport) MpoolCheckReplaceMessages(in0 context.Context, in1 []*types.Message) (out0 [][]api1.MessageCheckStatus, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MpoolCheckReplaceMessages %v", err)
		return
	}
	return cli.MpoolCheckReplaceMessages(in0, in1)
}

func (p *UnSupport) MpoolClear(in0 context.Context, in1 bool) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MpoolClear %v", err)
		return
	}
	return cli.MpoolClear(in0, in1)
}

func (p *UnSupport) MpoolPushUntrusted(in0 context.Context, in1 *types.SignedMessage) (out0 cid.Cid, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MpoolPushUntrusted %v", err)
		return
	}
	return cli.MpoolPushUntrusted(in0, in1)
}

func (p *UnSupport) MpoolSetConfig(in0 context.Context, in1 *types.MpoolConfig) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MpoolSetConfig %v", err)
		return
	}
	return cli.MpoolSetConfig(in0, in1)
}

func (p *UnSupport) MpoolSub(in0 context.Context) (out0 <-chan api1.MpoolUpdate, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MpoolSub %v", err)
		return
	}
	return cli.MpoolSub(in0)
}

func (p *UnSupport) MsigAddApprove(in0 context.Context, in1 address.Address, in2 address.Address, in3 uint64, in4 address.Address, in5 address.Address, in6 bool) (out0 *api1.MessagePrototype, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MsigAddApprove %v", err)
		return
	}
	return cli.MsigAddApprove(in0, in1, in2, in3, in4, in5, in6)
}

func (p *UnSupport) MsigAddCancel(in0 context.Context, in1 address.Address, in2 address.Address, in3 uint64, in4 address.Address, in5 bool) (out0 *api1.MessagePrototype, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MsigAddCancel %v", err)
		return
	}
	return cli.MsigAddCancel(in0, in1, in2, in3, in4, in5)
}

func (p *UnSupport) MsigAddPropose(in0 context.Context, in1 address.Address, in2 address.Address, in3 address.Address, in4 bool) (out0 *api1.MessagePrototype, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MsigAddPropose %v", err)
		return
	}
	return cli.MsigAddPropose(in0, in1, in2, in3, in4)
}

func (p *UnSupport) MsigApprove(in0 context.Context, in1 address.Address, in2 uint64, in3 address.Address) (out0 *api1.MessagePrototype, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MsigApprove %v", err)
		return
	}
	return cli.MsigApprove(in0, in1, in2, in3)
}

func (p *UnSupport) MsigApproveTxnHash(in0 context.Context, in1 address.Address, in2 uint64, in3 address.Address, in4 address.Address, in5 big.Int, in6 address.Address, in7 uint64, in8 []uint8) (out0 *api1.MessagePrototype, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MsigApproveTxnHash %v", err)
		return
	}
	return cli.MsigApproveTxnHash(in0, in1, in2, in3, in4, in5, in6, in7, in8)
}

func (p *UnSupport) MsigCancel(in0 context.Context, in1 address.Address, in2 uint64, in3 address.Address) (out0 *api1.MessagePrototype, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MsigCancel %v", err)
		return
	}
	return cli.MsigCancel(in0, in1, in2, in3)
}

func (p *UnSupport) MsigCancelTxnHash(in0 context.Context, in1 address.Address, in2 uint64, in3 address.Address, in4 big.Int, in5 address.Address, in6 uint64, in7 []uint8) (out0 *api1.MessagePrototype, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MsigCancelTxnHash %v", err)
		return
	}
	return cli.MsigCancelTxnHash(in0, in1, in2, in3, in4, in5, in6, in7)
}

func (p *UnSupport) MsigCreate(in0 context.Context, in1 uint64, in2 []address.Address, in3 abi.ChainEpoch, in4 big.Int, in5 address.Address, in6 big.Int) (out0 *api1.MessagePrototype, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MsigCreate %v", err)
		return
	}
	return cli.MsigCreate(in0, in1, in2, in3, in4, in5, in6)
}

func (p *UnSupport) MsigGetAvailableBalance(in0 context.Context, in1 address.Address, in2 types.TipSetKey) (out0 big.Int, err error) {
	cli, err := p.Select(in2)
	if err != nil {
		err = fmt.Errorf("api MsigGetAvailableBalance %v", err)
		return
	}
	return cli.MsigGetAvailableBalance(in0, in1, in2)
}

func (p *UnSupport) MsigGetPending(in0 context.Context, in1 address.Address, in2 types.TipSetKey) (out0 []*api1.MsigTransaction, err error) {
	cli, err := p.Select(in2)
	if err != nil {
		err = fmt.Errorf("api MsigGetPending %v", err)
		return
	}
	return cli.MsigGetPending(in0, in1, in2)
}

func (p *UnSupport) MsigGetVested(in0 context.Context, in1 address.Address, in2 types.TipSetKey, in3 types.TipSetKey) (out0 big.Int, err error) {
	cli, err := p.Select(in3)
	if err != nil {
		err = fmt.Errorf("api MsigGetVested %v", err)
		return
	}
	return cli.MsigGetVested(in0, in1, in2, in3)
}

func (p *UnSupport) MsigGetVestingSchedule(in0 context.Context, in1 address.Address, in2 types.TipSetKey) (out0 api1.MsigVesting, err error) {
	cli, err := p.Select(in2)
	if err != nil {
		err = fmt.Errorf("api MsigGetVestingSchedule %v", err)
		return
	}
	return cli.MsigGetVestingSchedule(in0, in1, in2)
}

func (p *UnSupport) MsigPropose(in0 context.Context, in1 address.Address, in2 address.Address, in3 big.Int, in4 address.Address, in5 uint64, in6 []uint8) (out0 *api1.MessagePrototype, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MsigPropose %v", err)
		return
	}
	return cli.MsigPropose(in0, in1, in2, in3, in4, in5, in6)
}

func (p *UnSupport) MsigRemoveSigner(in0 context.Context, in1 address.Address, in2 address.Address, in3 address.Address, in4 bool) (out0 *api1.MessagePrototype, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MsigRemoveSigner %v", err)
		return
	}
	return cli.MsigRemoveSigner(in0, in1, in2, in3, in4)
}

func (p *UnSupport) MsigSwapApprove(in0 context.Context, in1 address.Address, in2 address.Address, in3 uint64, in4 address.Address, in5 address.Address, in6 address.Address) (out0 *api1.MessagePrototype, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MsigSwapApprove %v", err)
		return
	}
	return cli.MsigSwapApprove(in0, in1, in2, in3, in4, in5, in6)
}

func (p *UnSupport) MsigSwapCancel(in0 context.Context, in1 address.Address, in2 address.Address, in3 uint64, in4 address.Address, in5 address.Address) (out0 *api1.MessagePrototype, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MsigSwapCancel %v", err)
		return
	}
	return cli.MsigSwapCancel(in0, in1, in2, in3, in4, in5)
}

func (p *UnSupport) MsigSwapPropose(in0 context.Context, in1 address.Address, in2 address.Address, in3 address.Address, in4 address.Address) (out0 *api1.MessagePrototype, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api MsigSwapPropose %v", err)
		return
	}
	return cli.MsigSwapPropose(in0, in1, in2, in3, in4)
}

func (p *UnSupport) NetAddrsListen(in0 context.Context) (out0 peer.AddrInfo, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetAddrsListen %v", err)
		return
	}
	return cli.NetAddrsListen(in0)
}

func (p *UnSupport) NetAgentVersion(in0 context.Context, in1 peer.ID) (out0 string, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetAgentVersion %v", err)
		return
	}
	return cli.NetAgentVersion(in0, in1)
}

func (p *UnSupport) NetAutoNatStatus(in0 context.Context) (out0 api1.NatInfo, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetAutoNatStatus %v", err)
		return
	}
	return cli.NetAutoNatStatus(in0)
}

func (p *UnSupport) NetBandwidthStats(in0 context.Context) (out0 metrics.Stats, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetBandwidthStats %v", err)
		return
	}
	return cli.NetBandwidthStats(in0)
}

func (p *UnSupport) NetBandwidthStatsByPeer(in0 context.Context) (out0 map[string]metrics.Stats, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetBandwidthStatsByPeer %v", err)
		return
	}
	return cli.NetBandwidthStatsByPeer(in0)
}

func (p *UnSupport) NetBandwidthStatsByProtocol(in0 context.Context) (out0 map[protocol.ID]metrics.Stats, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetBandwidthStatsByProtocol %v", err)
		return
	}
	return cli.NetBandwidthStatsByProtocol(in0)
}

func (p *UnSupport) NetBlockAdd(in0 context.Context, in1 api1.NetBlockList) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetBlockAdd %v", err)
		return
	}
	return cli.NetBlockAdd(in0, in1)
}

func (p *UnSupport) NetBlockList(in0 context.Context) (out0 api1.NetBlockList, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetBlockList %v", err)
		return
	}
	return cli.NetBlockList(in0)
}

func (p *UnSupport) NetBlockRemove(in0 context.Context, in1 api1.NetBlockList) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetBlockRemove %v", err)
		return
	}
	return cli.NetBlockRemove(in0, in1)
}

func (p *UnSupport) NetConnect(in0 context.Context, in1 peer.AddrInfo) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetConnect %v", err)
		return
	}
	return cli.NetConnect(in0, in1)
}

func (p *UnSupport) NetConnectedness(in0 context.Context, in1 peer.ID) (out0 network.Connectedness, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetConnectedness %v", err)
		return
	}
	return cli.NetConnectedness(in0, in1)
}

func (p *UnSupport) NetDisconnect(in0 context.Context, in1 peer.ID) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetDisconnect %v", err)
		return
	}
	return cli.NetDisconnect(in0, in1)
}

func (p *UnSupport) NetFindPeer(in0 context.Context, in1 peer.ID) (out0 peer.AddrInfo, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetFindPeer %v", err)
		return
	}
	return cli.NetFindPeer(in0, in1)
}

func (p *UnSupport) NetLimit(in0 context.Context, in1 string) (out0 api1.NetLimit, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetLimit %v", err)
		return
	}
	return cli.NetLimit(in0, in1)
}

func (p *UnSupport) NetPeerInfo(in0 context.Context, in1 peer.ID) (out0 *api1.ExtendedPeerInfo, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetPeerInfo %v", err)
		return
	}
	return cli.NetPeerInfo(in0, in1)
}

func (p *UnSupport) NetPeers(in0 context.Context) (out0 []peer.AddrInfo, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetPeers %v", err)
		return
	}
	return cli.NetPeers(in0)
}

func (p *UnSupport) NetPing(in0 context.Context, in1 peer.ID) (out0 time.Duration, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetPing %v", err)
		return
	}
	return cli.NetPing(in0, in1)
}

func (p *UnSupport) NetProtectAdd(in0 context.Context, in1 []peer.ID) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetProtectAdd %v", err)
		return
	}
	return cli.NetProtectAdd(in0, in1)
}

func (p *UnSupport) NetProtectList(in0 context.Context) (out0 []peer.ID, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetProtectList %v", err)
		return
	}
	return cli.NetProtectList(in0)
}

func (p *UnSupport) NetProtectRemove(in0 context.Context, in1 []peer.ID) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetProtectRemove %v", err)
		return
	}
	return cli.NetProtectRemove(in0, in1)
}

func (p *UnSupport) NetPubsubScores(in0 context.Context) (out0 []api1.PubsubScore, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetPubsubScores %v", err)
		return
	}
	return cli.NetPubsubScores(in0)
}

func (p *UnSupport) NetSetLimit(in0 context.Context, in1 string, in2 api1.NetLimit) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetSetLimit %v", err)
		return
	}
	return cli.NetSetLimit(in0, in1, in2)
}

func (p *UnSupport) NetStat(in0 context.Context, in1 string) (out0 api1.NetStat, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NetStat %v", err)
		return
	}
	return cli.NetStat(in0, in1)
}

func (p *UnSupport) NodeStatus(in0 context.Context, in1 bool) (out0 api1.NodeStatus, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api NodeStatus %v", err)
		return
	}
	return cli.NodeStatus(in0, in1)
}

func (p *UnSupport) PaychAllocateLane(in0 context.Context, in1 address.Address) (out0 uint64, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api PaychAllocateLane %v", err)
		return
	}
	return cli.PaychAllocateLane(in0, in1)
}

func (p *UnSupport) PaychAvailableFunds(in0 context.Context, in1 address.Address) (out0 *api1.ChannelAvailableFunds, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api PaychAvailableFunds %v", err)
		return
	}
	return cli.PaychAvailableFunds(in0, in1)
}

func (p *UnSupport) PaychAvailableFundsByFromTo(in0 context.Context, in1 address.Address, in2 address.Address) (out0 *api1.ChannelAvailableFunds, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api PaychAvailableFundsByFromTo %v", err)
		return
	}
	return cli.PaychAvailableFundsByFromTo(in0, in1, in2)
}

func (p *UnSupport) PaychCollect(in0 context.Context, in1 address.Address) (out0 cid.Cid, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api PaychCollect %v", err)
		return
	}
	return cli.PaychCollect(in0, in1)
}

func (p *UnSupport) PaychFund(in0 context.Context, in1 address.Address, in2 address.Address, in3 big.Int) (out0 *api1.ChannelInfo, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api PaychFund %v", err)
		return
	}
	return cli.PaychFund(in0, in1, in2, in3)
}

func (p *UnSupport) PaychGet(in0 context.Context, in1 address.Address, in2 address.Address, in3 big.Int, in4 api1.PaychGetOpts) (out0 *api1.ChannelInfo, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api PaychGet %v", err)
		return
	}
	return cli.PaychGet(in0, in1, in2, in3, in4)
}

func (p *UnSupport) PaychGetWaitReady(in0 context.Context, in1 cid.Cid) (out0 address.Address, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api PaychGetWaitReady %v", err)
		return
	}
	return cli.PaychGetWaitReady(in0, in1)
}

func (p *UnSupport) PaychList(in0 context.Context) (out0 []address.Address, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api PaychList %v", err)
		return
	}
	return cli.PaychList(in0)
}

func (p *UnSupport) PaychNewPayment(in0 context.Context, in1 address.Address, in2 address.Address, in3 []api1.VoucherSpec) (out0 *api1.PaymentInfo, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api PaychNewPayment %v", err)
		return
	}
	return cli.PaychNewPayment(in0, in1, in2, in3)
}

func (p *UnSupport) PaychSettle(in0 context.Context, in1 address.Address) (out0 cid.Cid, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api PaychSettle %v", err)
		return
	}
	return cli.PaychSettle(in0, in1)
}

func (p *UnSupport) PaychStatus(in0 context.Context, in1 address.Address) (out0 *api1.PaychStatus, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api PaychStatus %v", err)
		return
	}
	return cli.PaychStatus(in0, in1)
}

func (p *UnSupport) PaychVoucherAdd(in0 context.Context, in1 address.Address, in2 *paych.SignedVoucher, in3 []uint8, in4 big.Int) (out0 big.Int, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api PaychVoucherAdd %v", err)
		return
	}
	return cli.PaychVoucherAdd(in0, in1, in2, in3, in4)
}

func (p *UnSupport) PaychVoucherCheckSpendable(in0 context.Context, in1 address.Address, in2 *paych.SignedVoucher, in3 []uint8, in4 []uint8) (out0 bool, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api PaychVoucherCheckSpendable %v", err)
		return
	}
	return cli.PaychVoucherCheckSpendable(in0, in1, in2, in3, in4)
}

func (p *UnSupport) PaychVoucherCheckValid(in0 context.Context, in1 address.Address, in2 *paych.SignedVoucher) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api PaychVoucherCheckValid %v", err)
		return
	}
	return cli.PaychVoucherCheckValid(in0, in1, in2)
}

func (p *UnSupport) PaychVoucherCreate(in0 context.Context, in1 address.Address, in2 big.Int, in3 uint64) (out0 *api1.VoucherCreateResult, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api PaychVoucherCreate %v", err)
		return
	}
	return cli.PaychVoucherCreate(in0, in1, in2, in3)
}

func (p *UnSupport) PaychVoucherList(in0 context.Context, in1 address.Address) (out0 []*paych.SignedVoucher, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api PaychVoucherList %v", err)
		return
	}
	return cli.PaychVoucherList(in0, in1)
}

func (p *UnSupport) PaychVoucherSubmit(in0 context.Context, in1 address.Address, in2 *paych.SignedVoucher, in3 []uint8, in4 []uint8) (out0 cid.Cid, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api PaychVoucherSubmit %v", err)
		return
	}
	return cli.PaychVoucherSubmit(in0, in1, in2, in3, in4)
}

func (p *UnSupport) RaftLeader(in0 context.Context) (out0 peer.ID, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api RaftLeader %v", err)
		return
	}
	return cli.RaftLeader(in0)
}

func (p *UnSupport) RaftState(in0 context.Context) (out0 *api1.RaftStateData, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api RaftState %v", err)
		return
	}
	return cli.RaftState(in0)
}

func (p *UnSupport) Session(in0 context.Context) (out0 uuid.UUID, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api Session %v", err)
		return
	}
	return cli.Session(in0)
}

func (p *UnSupport) Shutdown(in0 context.Context) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api Shutdown %v", err)
		return
	}
	return cli.Shutdown(in0)
}

func (p *UnSupport) StartTime(in0 context.Context) (out0 time.Time, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api StartTime %v", err)
		return
	}
	return cli.StartTime(in0)
}

func (p *UnSupport) StateCompute(in0 context.Context, in1 abi.ChainEpoch, in2 []*types.Message, in3 types.TipSetKey) (out0 *api1.ComputeStateOutput, err error) {
	cli, err := p.Select(in3)
	if err != nil {
		err = fmt.Errorf("api StateCompute %v", err)
		return
	}
	return cli.StateCompute(in0, in1, in2, in3)
}

func (p *UnSupport) StateDecodeParams(in0 context.Context, in1 address.Address, in2 abi.MethodNum, in3 []uint8, in4 types.TipSetKey) (out0 interface{}, err error) {
	cli, err := p.Select(in4)
	if err != nil {
		err = fmt.Errorf("api StateDecodeParams %v", err)
		return
	}
	return cli.StateDecodeParams(in0, in1, in2, in3, in4)
}

func (p *UnSupport) StateEncodeParams(in0 context.Context, in1 cid.Cid, in2 abi.MethodNum, in3 json.RawMessage) (out0 []uint8, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api StateEncodeParams %v", err)
		return
	}
	return cli.StateEncodeParams(in0, in1, in2, in3)
}

func (p *UnSupport) StateListMessages(in0 context.Context, in1 *api1.MessageMatch, in2 types.TipSetKey, in3 abi.ChainEpoch) (out0 []cid.Cid, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api StateListMessages %v", err)
		return
	}
	return cli.StateListMessages(in0, in1, in2, in3)
}

func (p *UnSupport) StateReplay(in0 context.Context, in1 types.TipSetKey, in2 cid.Cid) (out0 *api1.InvocResult, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api StateReplay %v", err)
		return
	}
	return cli.StateReplay(in0, in1, in2)
}

func (p *UnSupport) StateSearchMsgLimited(in0 context.Context, in1 cid.Cid, in2 abi.ChainEpoch) (out0 *api1.MsgLookup, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api StateSearchMsgLimited %v", err)
		return
	}
	return cli.StateSearchMsgLimited(in0, in1, in2)
}

func (p *UnSupport) StateWaitMsgLimited(in0 context.Context, in1 cid.Cid, in2 uint64, in3 abi.ChainEpoch) (out0 *api1.MsgLookup, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api StateWaitMsgLimited %v", err)
		return
	}
	return cli.StateWaitMsgLimited(in0, in1, in2, in3)
}

func (p *UnSupport) SyncCheckBad(in0 context.Context, in1 cid.Cid) (out0 string, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api SyncCheckBad %v", err)
		return
	}
	return cli.SyncCheckBad(in0, in1)
}

func (p *UnSupport) SyncCheckpoint(in0 context.Context, in1 types.TipSetKey) (err error) {
	cli, err := p.Select(in1)
	if err != nil {
		err = fmt.Errorf("api SyncCheckpoint %v", err)
		return
	}
	return cli.SyncCheckpoint(in0, in1)
}

func (p *UnSupport) SyncIncomingBlocks(in0 context.Context) (out0 <-chan *types.BlockHeader, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api SyncIncomingBlocks %v", err)
		return
	}
	return cli.SyncIncomingBlocks(in0)
}

func (p *UnSupport) SyncMarkBad(in0 context.Context, in1 cid.Cid) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api SyncMarkBad %v", err)
		return
	}
	return cli.SyncMarkBad(in0, in1)
}

func (p *UnSupport) SyncUnmarkAllBad(in0 context.Context) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api SyncUnmarkAllBad %v", err)
		return
	}
	return cli.SyncUnmarkAllBad(in0)
}

func (p *UnSupport) SyncUnmarkBad(in0 context.Context, in1 cid.Cid) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api SyncUnmarkBad %v", err)
		return
	}
	return cli.SyncUnmarkBad(in0, in1)
}

func (p *UnSupport) SyncValidateTipset(in0 context.Context, in1 types.TipSetKey) (out0 bool, err error) {
	cli, err := p.Select(in1)
	if err != nil {
		err = fmt.Errorf("api SyncValidateTipset %v", err)
		return
	}
	return cli.SyncValidateTipset(in0, in1)
}

func (p *UnSupport) WalletDefaultAddress(in0 context.Context) (out0 address.Address, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api WalletDefaultAddress %v", err)
		return
	}
	return cli.WalletDefaultAddress(in0)
}

func (p *UnSupport) WalletDelete(in0 context.Context, in1 address.Address) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api WalletDelete %v", err)
		return
	}
	return cli.WalletDelete(in0, in1)
}

func (p *UnSupport) WalletExport(in0 context.Context, in1 address.Address) (out0 *types.KeyInfo, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api WalletExport %v", err)
		return
	}
	return cli.WalletExport(in0, in1)
}

func (p *UnSupport) WalletImport(in0 context.Context, in1 *types.KeyInfo) (out0 address.Address, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api WalletImport %v", err)
		return
	}
	return cli.WalletImport(in0, in1)
}

func (p *UnSupport) WalletList(in0 context.Context) (out0 []address.Address, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api WalletList %v", err)
		return
	}
	return cli.WalletList(in0)
}

func (p *UnSupport) WalletNew(in0 context.Context, in1 types.KeyType) (out0 address.Address, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api WalletNew %v", err)
		return
	}
	return cli.WalletNew(in0, in1)
}

func (p *UnSupport) WalletSetDefault(in0 context.Context, in1 address.Address) (err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api WalletSetDefault %v", err)
		return
	}
	return cli.WalletSetDefault(in0, in1)
}

func (p *UnSupport) WalletSignMessage(in0 context.Context, in1 address.Address, in2 *types.Message) (out0 *types.SignedMessage, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api WalletSignMessage %v", err)
		return
	}
	return cli.WalletSignMessage(in0, in1, in2)
}

func (p *UnSupport) WalletValidateAddress(in0 context.Context, in1 string) (out0 address.Address, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api WalletValidateAddress %v", err)
		return
	}
	return cli.WalletValidateAddress(in0, in1)
}

func (p *UnSupport) WalletVerify(in0 context.Context, in1 address.Address, in2 []uint8, in3 *crypto.Signature) (out0 bool, err error) {
	cli, err := p.Select(types.EmptyTSK)
	if err != nil {
		err = fmt.Errorf("api WalletVerify %v", err)
		return
	}
	return cli.WalletVerify(in0, in1, in2, in3)
}
