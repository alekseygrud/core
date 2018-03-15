package dwh

import (
	"context"
	"fmt"
	"os"
	"testing"

	"encoding/json"

	log "github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

const (
	testDBPath = "/tmp/sonm/dwh.db"
)

var (
	w *DWH
)

func TestMain(m *testing.M) {
	var err error
	w, err = getTestDWH()
	if err != nil {
		fmt.Println(err)
		os.Remove(testDBPath)
		os.Exit(1)
	}

	retCode := m.Run()
	w.db.Close()
	os.Remove(testDBPath)
	os.Exit(retCode)
}

func TestDWH_GetDealDetails(t *testing.T) {
	reply, err := w.GetDealDetails(context.Background(), &pb.ID{Id: "id_2"})

	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, int64(30012), reply.StartTime.Seconds)
}

func TestDWH_GetOrdersList(t *testing.T) {
	reply, err := w.GetOrdersList(context.Background(), &pb.OrdersListRequest{
		Type:             pb.MarketOrderType_MARKET_ANY,
		DurationSeconds:  20015,
		DurationOperator: pb.ComparisonOperator_EQ,
		Limit:            1,
	})

	if err != nil {
		t.Error(err)
		return
	}

	if len(reply.Orders) != 1 {
		t.Errorf("Expected 1 order in reply, got %d", len(reply.Orders))
		return
	}

	assert.Equal(t, uint64(20015), reply.Orders[0].Duration)

	reply, err = w.GetOrdersList(context.Background(), &pb.OrdersListRequest{
		Type:   pb.MarketOrderType_MARKET_ANY,
		Limit:  1,
		Offset: 1,
	})

	if err != nil {
		t.Error(err)
		return
	}

	if len(reply.Orders) != 1 {
		t.Errorf("Expected 1 order in reply, got %d", len(reply.Orders))
		return
	}

	assert.Equal(t, uint64(20011), reply.Orders[0].Duration)
}

func TestDWH_GetOrderDetails(t *testing.T) {
	reply, err := w.GetOrderDetails(context.Background(), &pb.ID{Id: "id_2"})

	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, uint64(20012), reply.Duration)
}

func getTestDWH() (*DWH, error) {
	var (
		ctx = context.Background()
		cfg = &Config{
			Storage: &storageConfig{
				Backend:  "sqlite3",
				Endpoint: testDBPath,
			},
		}
		w = &DWH{
			ctx:    ctx,
			cfg:    cfg,
			logger: log.GetLogger(ctx),
		}
	)

	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.setupDB(); err != nil {
		return nil, err
	}

	netflagsBytes, _ := json.Marshal([]bool{true, false})
	benchmarksBytes, _ := json.Marshal([]uint64{1, 2})

	for i := 0; i < 10; i++ {
		_, err := w.db.Exec(
			insertDealSQLite,
			fmt.Sprintf("id_%d", i),
			benchmarksBytes,
			fmt.Sprintf("supplier_%d", i),
			fmt.Sprintf("consumer_%d", i),
			"master",
			fmt.Sprintf("ask_id_%d", i),
			fmt.Sprintf("bid_id_%d", i),
			"10010", // Price
			20010+i, // Duration
			30010+i, // StartTime
			30020+i, // EndTime
			uint64(pb.MarketDealStatus_MARKET_STATUS_ACCEPTED),
			"10020", // BlockedBalance
			"10030", // TotalPayout
			30030+i, // LastBillTS
		)
		if err != nil {
			return nil, err
		}
	}

	for i := 0; i < 10; i++ {
		_, err := w.db.Exec(
			insertOrderSQLite,
			fmt.Sprintf("id_%d", i),
			uint64(pb.MarketOrderType_MARKET_ASK),
			uint64(pb.MarketOrderStatus_MARKET_ORDER_ACTIVE),
			fmt.Sprintf("author_%d", i),
			fmt.Sprintf("counterparty_%d", i),
			"10010", // Price
			20010+i,
			netflagsBytes,
			uint64(pb.MarketIdentityLevel_MARKET_IDENTRIFIED),
			fmt.Sprintf("blacklist_%d", i),
			[]byte{1, 2, 3}, // Tag
			benchmarksBytes,
			"10020", // FrozenSum
		)
		if err != nil {
			return nil, err
		}
	}

	return w, nil
}
