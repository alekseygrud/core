package dwh

import (
	"crypto/ecdsa"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/blockchain"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type DWH struct {
	mu          sync.RWMutex
	ctx         context.Context
	cfg         *Config
	cancel      context.CancelFunc
	grpc        *grpc.Server
	logger      *zap.Logger
	db          *sql.DB
	creds       credentials.TransportCredentials
	certRotator util.HitlessCertRotator
	blockchain  blockchain.API
}

func NewDWH(ctx context.Context, cfg *Config, key *ecdsa.PrivateKey) (w *DWH, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	w = &DWH{
		ctx:    ctx,
		cfg:    cfg,
		logger: log.GetLogger(ctx),
	}

	var TLSConfig *tls.Config
	w.certRotator, TLSConfig, err = util.NewHitlessCertRotator(ctx, key)
	if err != nil {
		return nil, err
	}

	w.creds = util.NewTLS(TLSConfig)
	server := xgrpc.NewServer(w.logger,
		xgrpc.Credentials(w.creds),
		xgrpc.DefaultTraceInterceptor(),
	)
	w.grpc = server
	pb.RegisterDWHServer(w.grpc, w)
	grpc_prometheus.Register(w.grpc)

	bAPI, err := blockchain.NewBlockchainAPI(&cfg.Blockchain.EthEndpoint, &cfg.Blockchain.GasPrice)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init blockchain client")
	}

	w.blockchain = bAPI

	return
}

func (w *DWH) Serve() error {
	if err := w.setupDB(); err != nil {
		return err
	}

	lis, err := net.Listen("tcp", w.cfg.ListenAddr)
	if err != nil {
		return err
	}

	return w.grpc.Serve(lis)
}

func (w *DWH) GetDealsList(ctx context.Context, request *pb.DealsListRequest) (*pb.DealsListReply, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var (
		query      = "SELECT * FROM deals"
		conditions []string
		values     []interface{}
	)
	// Prepare WHERE clause.
	if len(request.SupplierID) > 0 {
		conditions = append(conditions, "supplier_id=?")
		values = append(values, request.SupplierID)
	}
	if len(request.ConsumerID) > 0 {
		conditions = append(conditions, "consumer_id=?")
		values = append(values, request.ConsumerID)
	}
	if len(request.ConsumerID) > 0 {
		conditions = append(conditions, "master_id=?")
		values = append(values, request.ConsumerID)
	}
	if len(request.ConsumerID) > 0 {
		conditions = append(conditions, "ask_id=?")
		values = append(values, request.ConsumerID)
	}
	if len(request.ConsumerID) > 0 {
		conditions = append(conditions, "bid_id=?")
		values = append(values, request.ConsumerID)
	}
	if request.DurationSeconds > 0 {
		conditions = append(conditions, fmt.Sprintf("duration%s?", getOperator(request.DurationOperator)))
		values = append(values, request.DurationSeconds)
	}
	if request.Price > 0 {
		conditions = append(conditions, fmt.Sprintf("price%s?", getOperator(request.PriceOperator)))
		values = append(values, request.Price)
	}
	if request.Status != pb.MarketDealStatus_MARKET_STATUS_UNKNOWN {
		conditions = append(conditions, "status=?")
		values = append(values, request.Status)
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY price ASC"
	if request.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", request.Limit)
	}
	if request.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", request.Offset)
	}
	query += ";"

	rows, err := w.db.Query(query, values...)
	if err != nil {
		return nil, errors.Wrapf(err, "query `%s` failed", query)
	}
	defer rows.Close()

	var deals []*pb.MarketDeal
	for rows.Next() {
		deal, err := w.decodeDeal(rows)
		if err != nil {
			return nil, err
		}
		deals = append(deals, deal)
	}

	return &pb.DealsListReply{Deals: deals}, nil
}

func (w *DWH) GetDealDetails(ctx context.Context, request *pb.ID) (*pb.MarketDeal, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getDealDetails(ctx, request)
}

func (w *DWH) GetOrdersList(ctx context.Context, request *pb.OrdersListRequest) (*pb.OrdersListReply, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var filters []*Filter
	if request.Type > 0 {
		filters = append(filters, NewFilter("Type", pb.ComparisonOperator_EQ, request.Type))
	}
	if request.Status > 0 {
		filters = append(filters, NewFilter("Status", pb.ComparisonOperator_EQ, request.Status))
	}
	if len(request.AuthorID) > 0 {
		filters = append(filters, NewFilter("Author_id", pb.ComparisonOperator_EQ, request.AuthorID))
	}
	if len(request.CounterpartyID) > 0 {
		filters = append(filters, NewFilter("CounterpartyID", pb.ComparisonOperator_EQ, request.CounterpartyID))
	}
	if request.DurationSeconds > 0 {
		filters = append(filters, NewFilter("Duration", pb.ComparisonOperator_EQ, request.DurationSeconds))
	}
	if request.Price > 0 {
		filters = append(filters, NewFilter("Price", pb.ComparisonOperator_EQ, request.Price))
	}
	if request.Type > 0 {
		filters = append(filters, NewFilter("Type", pb.ComparisonOperator_EQ, request.Type))
	}
	rows, query, err := RunQuery(w.db, "orders", request.Offset, request.Limit, filters...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*pb.MarketOrder
	for rows.Next() {
		order, err := w.decodeOrder(rows)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode order, query `%s`", query)
		}
		orders = append(orders, order)
	}

	return &pb.OrdersListReply{Orders: orders}, nil
}

func (w *DWH) GetOrderDetails(ctx context.Context, request *pb.ID) (*pb.MarketOrder, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getOrderDetails(ctx, request)
}

func (w *DWH) GetDealChangeRequests(ctx context.Context, request *pb.ID) (*pb.DealChangeRequestsReply, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	rows, err := w.db.Query("SELECT * FROM change_requests WHERE deal=?", request.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changeRequests []*pb.DealChangeRequest
	for rows.Next() {
		var (
			duration uint64
			price    string
			dealID   string
		)
		if err := rows.Scan(&duration, &price, &dealID); err != nil {
			return nil, err
		}

		bigPrice := new(big.Int)
		bigPrice.SetString(price, 10)
		changeRequests = append(changeRequests, &pb.DealChangeRequest{
			DurationSeconds: duration,
			Price:           pb.NewBigInt(bigPrice),
		})
	}

	return &pb.DealChangeRequestsReply{ChangeRequests: changeRequests}, nil
}

func (w *DWH) getDealDetails(ctx context.Context, request *pb.ID) (*pb.MarketDeal, error) {
	rows, err := w.db.Query("SELECT * FROM deals WHERE id=?", request.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if ok := rows.Next(); !ok {
		return nil, errors.Errorf("deal `%s` not found", request.Id)
	}

	return w.decodeDeal(rows)
}

func (w *DWH) getOrderDetails(ctx context.Context, request *pb.ID) (*pb.MarketOrder, error) {
	rows, err := w.db.Query("SELECT * FROM orders WHERE id=?", request.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if ok := rows.Next(); !ok {
		return nil, errors.Errorf("order `%s` not found", request.Id)
	}

	return w.decodeOrder(rows)
}

func (w *DWH) setupDB() (err error) {
	db, err := w.setupSQLite()
	if err != nil {
		return err
	}

	w.db = db

	return nil
}

func (w *DWH) setupSQLite() (*sql.DB, error) {
	db, err := sql.Open(w.cfg.Storage.Backend, w.cfg.Storage.Endpoint)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			db.Close()
		}
	}()

	_, err = db.Exec(createTableDealsSQLite)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create deals table (%s)", w.cfg.Storage.Backend)
	}

	_, err = db.Exec(createTableOrdersSQLite)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create orders table (%s)", w.cfg.Storage.Backend)
	}

	_, err = db.Exec(createTableChangesSQLite)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create requests table (%s)", w.cfg.Storage.Backend)
	}

	_, err = db.Exec(createTableMiscSQLite)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create misc table (%s)", w.cfg.Storage.Backend)
	}

	return db, nil
}

func (w *DWH) monitor() error {
	w.logger.Info("starting monitoring")

	retryTime := time.Millisecond * 100
	for {
		if err := w.watchMarketEvents(); err != nil {
			w.logger.Error("failed to watch market events, retrying",
				zap.Error(err), zap.Duration("retry_time", retryTime))
			retryTime *= 2
			if retryTime > time.Second*10 {
				retryTime = time.Second * 10
			}
			time.Sleep(retryTime)
		}
	}
}

func (w *DWH) watchMarketEvents() error {
	rows, err := w.db.Query(selectLastKnownBlock)
	if err != nil {
		return err
	}
	defer rows.Close()

	if ok := rows.Next(); !ok {
		return errors.Wrapf(err, "failed to get last known block number")
	}

	var lastKnownBlock int64
	if err := rows.Scan(&lastKnownBlock); err != nil {
		return errors.Wrapf(err, "failed to parse last known block number")
	}

	dealEvents, err := w.blockchain.GetMarketEvents(context.Background(), big.NewInt(lastKnownBlock))
	if err != nil {
		return err
	}

	for event := range dealEvents {
		switch value := event.Data.(type) {
		case *blockchain.DealOpenedData:
			if err := w.onDealOpened(value.ID); err != nil {
				w.logger.Error("failed to process DealOpened event",
					zap.Error(err), zap.String("deal_id", value.ID.String()))
			}
		case *blockchain.DealUpdatedData:
			if err := w.onDealUpdated(value.ID); err != nil {
				w.logger.Error("failed to process DealUpdated event",
					zap.Error(err), zap.String("deal_id", value.ID.String()))
			}
		case *blockchain.OrderPlacedData:
			if err := w.onOrderPlaced(value.ID); err != nil {
				w.logger.Error("failed to process OrderPlaced event",
					zap.Error(err), zap.String("deal_id", value.ID.String()))
			}
		case *blockchain.OrderUpdatedData:
			if err := w.onOrderCancelled(value.ID); err != nil {
				w.logger.Error("failed to process OrderCancelled event",
					zap.Error(err), zap.String("deal_id", value.ID.String()))
			}
		case *blockchain.ErrorData:
			w.logger.Error("received error from events channel", zap.Error(err))
		}

		if _, err := w.db.Exec(updateLastKnownBlockSQLite, event.BlockNumber); err != nil {
			w.logger.Error("failed to save order", zap.Error(err),
				zap.Uint64("block_number", event.BlockNumber))
		}
	}

	return errors.New("events channel closed")
}

func (w *DWH) onDealOpened(dealID *big.Int) error {
	deal, err := w.blockchain.GetDealInfo(context.Background(), dealID)
	if err != nil {
		return errors.Wrapf(err, "failed to get deal info")
	}

	benchmarkBytes, err := json.Marshal(deal.Benchmarks)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal benchmarks")
	}

	_, err = w.db.Exec(
		insertDealSQLite,
		deal.Id,
		benchmarkBytes,
		deal.SupplierID,
		deal.ConsumerID,
		deal.MasterID,
		deal.AskID,
		deal.BidID,
		deal.Price.Unwrap().String(),
		deal.Duration,
		deal.StartTime.Seconds,
		deal.EndTime.Seconds,
		uint64(deal.Status),
		deal.BlockedBalance.Unwrap().String(),
		deal.TotalPayout.String(),
		deal.LastBillTS.Seconds,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to insert deal into table")
	}

	return nil
}

func (w *DWH) onDealUpdated(dealID *big.Int) error {
	return w.onDealOpened(dealID)
}

func (w *DWH) onDealClosed(dealID *big.Int) error {
	return w.onDealOpened(dealID)
}

func (w *DWH) onOrderPlaced(orderID *big.Int) error {
	order, err := w.blockchain.GetOrderInfo(context.Background(), orderID)
	if err != nil {
		return errors.Wrapf(err, "failed to get order info")
	}

	benchmarkBytes, err := json.Marshal(order.Benchmarks)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal benchmarks")
	}

	netflagsBytes, err := json.Marshal(order.Netflags)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal netFlagss")
	}

	_, err = w.db.Exec(
		insertOrderSQLite,
		order.Id,
		uint64(order.OrderType),
		uint64(order.OrderStatus),
		order.Author,
		order.Counterparty,
		order.Price.Unwrap().String(),
		order.Duration,
		netflagsBytes,
		uint64(order.IdentityLevel),
		order.Blacklist,
		order.Tag,
		benchmarkBytes,
		order.FrozenSum.Unwrap().String(),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to insert order into table")
	}

	return nil
}

func (w *DWH) onOrderCancelled(orderID *big.Int) error {
	_, err := w.db.Exec(deleteOrderSQLite, orderID.String())
	return err
}

func (w *DWH) decodeDeal(rows *sql.Rows) (*pb.MarketDeal, error) {
	var (
		id              string
		benchmarksBytes []byte
		supplierID      string
		consumerID      string
		masterID        string
		askID           string
		bidID           string
		price           string
		duration        uint64
		startTime       int64
		endTime         int64
		status          uint64
		blockedBalance  string
		totalPayout     string
		lastBillTS      int64
	)
	if err := rows.Scan(
		&id,
		&benchmarksBytes,
		&supplierID,
		&consumerID,
		&masterID,
		&askID,
		&bidID,
		&price,
		&duration,
		&startTime,
		&endTime,
		&status,
		&blockedBalance,
		&totalPayout,
		&lastBillTS,
	); err != nil {
		w.logger.Error("failed to scan deal row", zap.Error(err))
		return nil, err
	}

	bigPrice := new(big.Int)
	bigPrice.SetString(price, 10)
	bigBlockedBalance := new(big.Int)
	bigBlockedBalance.SetString(blockedBalance, 10)
	bigTotalPayout := new(big.Int)
	bigTotalPayout.SetString(totalPayout, 10)
	var benchmarks []uint64
	if err := json.Unmarshal(benchmarksBytes, &benchmarksBytes); err != nil {
		return nil, err
	}

	return &pb.MarketDeal{
		Id:             id,
		Benchmarks:     benchmarks,
		SupplierID:     supplierID,
		ConsumerID:     consumerID,
		MasterID:       masterID,
		Price:          pb.NewBigInt(bigPrice),
		Duration:       duration,
		StartTime:      &pb.Timestamp{Seconds: startTime},
		EndTime:        &pb.Timestamp{Seconds: endTime},
		Status:         pb.MarketDealStatus(status),
		BlockedBalance: pb.NewBigInt(bigBlockedBalance),
		TotalPayout:    pb.NewBigInt(bigTotalPayout),
		LastBillTS:     &pb.Timestamp{Seconds: lastBillTS},
	}, nil
}

func (w *DWH) decodeOrder(rows *sql.Rows) (*pb.MarketOrder, error) {
	var (
		id              string
		orderType       uint64
		status          uint64
		author          string
		counterAgent    string
		price           string
		duration        uint64
		netflagsBytes   []byte
		identityLevel   uint64
		blackList       string
		tag             []byte
		benchmarksBytes []byte
		frozenSum       string
	)
	if err := rows.Scan(
		&id,
		&orderType,
		&status,
		&author,
		&counterAgent,
		&price,
		&duration,
		&netflagsBytes,
		&identityLevel,
		&blackList,
		&tag,
		&benchmarksBytes,
		&frozenSum,
	); err != nil {
		w.logger.Error("failed to scan order row", zap.Error(err))
		return nil, err
	}

	bigPrice := new(big.Int)
	bigPrice.SetString(price, 10)
	bigFrozenSum := new(big.Int)
	bigFrozenSum.SetString(frozenSum, 10)
	var netFlags []bool
	if err := json.Unmarshal(netflagsBytes, &netFlags); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal netflags")
	}
	var benchmarks []uint64
	if err := json.Unmarshal(benchmarksBytes, &benchmarksBytes); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal benchmarks")
	}

	return &pb.MarketOrder{
		Id:            id,
		OrderType:     pb.MarketOrderType(orderType),
		OrderStatus:   pb.MarketOrderStatus(status),
		Author:        author,
		Counterparty:  counterAgent,
		Price:         pb.NewBigInt(bigPrice),
		Duration:      duration,
		Netflags:      netFlags,
		IdentityLevel: pb.MarketIdentityLevel(identityLevel),
		Blacklist:     blackList,
		Tag:           tag,
		Benchmarks:    benchmarks,
		FrozenSum:     pb.NewBigInt(bigFrozenSum),
	}, nil
}
