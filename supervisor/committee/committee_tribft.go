package committee

import (
	"blockEmulator/core"
	"blockEmulator/message"
	"blockEmulator/networks"
	"blockEmulator/params"
	"blockEmulator/supervisor/signal"
	"blockEmulator/supervisor/supervisor_log"
	"blockEmulator/utils"
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"math/big"
	"os"
	"time"
)

// TriBFTCommitteeModule TriBFT 委员会模块（简单的交易注入）
type TriBFTCommitteeModule struct {
	csvPath      string
	dataTotalNum int
	nowDataNum   int
	batchDataNum int
	IpNodeTable  map[uint64]map[uint64]string
	sl           *supervisor_log.SupervisorLog
	Ss           *signal.StopSignal // to control the stop message sending
}

// NewTriBFTCommitteeModule 创建 TriBFT 委员会模块
func NewTriBFTCommitteeModule(Ip_nodeTable map[uint64]map[uint64]string, Ss *signal.StopSignal, slog *supervisor_log.SupervisorLog, csvFilePath string, dataNum, batchNum int) *TriBFTCommitteeModule {
	return &TriBFTCommitteeModule{
		csvPath:      csvFilePath,
		dataTotalNum: dataNum,
		batchDataNum: batchNum,
		nowDataNum:   0,
		IpNodeTable:  Ip_nodeTable,
		Ss:           Ss,
		sl:           slog,
	}
}

func (tcm *TriBFTCommitteeModule) HandleOtherMessage([]byte) {}

// txSending 发送交易到各个分片
func (tcm *TriBFTCommitteeModule) txSending(txlist []*core.Transaction) {
	sendToShard := make(map[uint64][]*core.Transaction)

	for idx := 0; idx <= len(txlist); idx++ {
		if idx > 0 && (idx%params.InjectSpeed == 0 || idx == len(txlist)) {
			// send to shard
			for sid := uint64(0); sid < uint64(params.ShardNum); sid++ {
				it := message.InjectTxs{
					Txs:       sendToShard[sid],
					ToShardID: sid,
				}
				itByte, err := json.Marshal(it)
				if err != nil {
					log.Panic(err)
				}
				send_msg := message.MergeMessage(message.CInject, itByte)
				go networks.TcpDial(send_msg, tcm.IpNodeTable[sid][0])
			}
			sendToShard = make(map[uint64][]*core.Transaction)
			time.Sleep(time.Second)
		}
		if idx == len(txlist) {
			break
		}
		tx := txlist[idx]
		sendersid := uint64(utils.Addr2Shard(tx.Sender))
		sendToShard[sendersid] = append(sendToShard[sendersid], tx)
	}
}

// MsgSendingControl 读取交易并发送
func (tcm *TriBFTCommitteeModule) MsgSendingControl() {
	tcm.sl.Slog.Println("TriBFT Committee: Starting transaction injection...")

	txfile, err := os.Open(tcm.csvPath)
	if err != nil {
		log.Panic(err)
	}
	defer txfile.Close()
	reader := csv.NewReader(txfile)
	txlist := make([]*core.Transaction, 0)

	for {
		data, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Panic(err)
		}

		// 将CSV数据转换为交易
		if tx, ok := tcm.parseTransaction(data, uint64(tcm.nowDataNum)); ok {
			txlist = append(txlist, tx)
			tcm.nowDataNum++
		} else {
			continue
		}

		// batch sending condition
		if len(txlist) == int(tcm.batchDataNum) || tcm.nowDataNum == tcm.dataTotalNum {
			tcm.sl.Slog.Printf("TriBFT Committee: Injecting batch of %d transactions (progress: %d/%d)\n",
				len(txlist), tcm.nowDataNum, tcm.dataTotalNum)
			tcm.txSending(txlist)
			txlist = make([]*core.Transaction, 0)
			tcm.Ss.StopGap_Reset()
		}

		if tcm.nowDataNum == tcm.dataTotalNum {
			break
		}
	}

	tcm.sl.Slog.Println("TriBFT Committee: Transaction injection complete")
}

// parseTransaction 将CSV数据转换为交易
func (tcm *TriBFTCommitteeModule) parseTransaction(data []string, nonce uint64) (*core.Transaction, bool) {
	if data[6] == "0" && data[7] == "0" && len(data[3]) > 16 && len(data[4]) > 16 && data[3] != data[4] {
		val, ok := new(big.Int).SetString(data[8], 10)
		if !ok {
			log.Panic("new int failed\n")
		}
		tx := core.NewTransaction(data[3][2:], data[4][2:], val, nonce, time.Now())
		return tx, true
	}
	return &core.Transaction{}, false
}

// HandleBlockInfo 处理区块信息（TriBFT 不需要特殊处理）
func (tcm *TriBFTCommitteeModule) HandleBlockInfo(b *message.BlockInfoMsg) {
	tcm.sl.Slog.Printf("TriBFT Committee: Received block info from shard %d in epoch %d.\n", b.SenderShardID, b.Epoch)
}
