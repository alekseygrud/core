package dwh

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto"
)

const (
	MaxLimit = 50
)

// SQL statements for SQLite backend.
var (
	createTableDealsSQLite = `
	CREATE TABLE IF NOT EXISTS Deals (
		Id						TEXT PRIMARY KEY,
		SupplierID				TEXT NOT NULL,
		ConsumerID				TEXT NOT NULL,
		MasterID				TEXT NOT NULL,
		AskID					TEXT NOT NULL,
		BidID					TEXT NOT NULL,
		Duration 				INTEGER NOT NULL,
		Price					TEXT NOT NULL,
		StartTime				INTEGER NOT NULL,
		EndTime					INTEGER NOT NULL,
		Status					INTEGER NOT NULL,
		BlockedBalance			TEXT NOT NULL,
		TotalPayout				TEXT NOT NULL,
		LastBillTS				INTEGER NOT NULL,
		Benchmarks				BLOB NOT NULL
	);`
	createTableOrdersSQLite = `
	CREATE TABLE IF NOT EXISTS Orders (
		Id						TEXT PRIMARY KEY,
		DealID					TEXT NOT NULL,
		Type					INTEGER NOT NULL,
		Status					INTEGER NOT NULL,
		AuthorID				TEXT NOT NULL,
		CounterpartyID			TEXT NOT NULL,
		Duration 				INTEGER NOT NULL,
		Price					TEXT NOT NULL,
		Netflags				INTEGER NULL,
		IdentityLevel			INTEGER NOT NULL,
		Blacklist				TEXT NOT NULL,
		Tag						TEXT NOT NULL,
		FrozenSum				TEXT NOT NULL,
		Benchmarks				BLOB NOT NULL
	);`
	createTableBenchmarksSQLite = `
	CREATE TABLE IF NOT EXISTS Benchmarks (
		BenchmarkID					INTEGER NOT NULL,
		Value						INTEGER NOT NULL,
		OrderID						TEXT NOT NULL,
		FOREIGN KEY(OrderID)		REFERENCES Orders(Id) ON DELETE CASCADE
	);`
	createTableConditionsSQLite = `
	CREATE TABLE IF NOT EXISTS Conditions (
		SupplierID					TEXT NOT NULL,
		ConsumerID					TEXT NOT NULL,
		MasterID					TEXT NOT NULL,
		Duration 					INTEGER NOT NULL,
		Price						TEXT NOT NULL,
		StartTime					INTEGER NOT NULL,
		EndTime						INTEGER NOT NULL,
		TotalPayout					TEXT NOT NULL,
		DealID						TEXT NOT NULL,
		FOREIGN KEY (DealID)		REFERENCES Deals(Id) ON DELETE CASCADE
	);`
	createTableChangeRequestsSQLite = `
	CREATE TABLE IF NOT EXISTS Orders (
		Id 							TEXT PRIMARY KEY,
		CreatedTS					INTEGER NOT NULL,
		AcceptedTS					INTEGER NOT NULL,
		RequestType					TEXT NOT NULL,
		Duration 					INTEGER NOT NULL,
		Price						TEXT NOT NULL,
		Status						INTEGER NOT NULL,
		DealID						TEXT NOT NULL,
		FOREIGN KEY (DealID)		REFERENCES Deals(Id) ON DELETE CASCADE
	);`
	createTableValidatorsSQLite = `
	CREATE TABLE IF NOT EXISTS Validators (
		Id							TEXT NOT NULL,
		Status						INTEGER NOT NULL,
		Level						INTEGER NOT NULL
	);`
	createTableCertificatesSQLite = `
	CREATE TABLE IF NOT EXISTS Certificates (
		Id							TEXT PRIMARY KEY,
		Address						TEXT NOT NULL,
		Attribute					TEXT NOT NULL,
		AttributeLevel				INTEGER NOT NULL,
		Value						TEXT NOT NULL,
		ValidatorID					TEXT NOT NULL,
		FOREIGN KEY (ValidatorID)	REFERENCES Validators(Id) ON DELETE CASCADE
	);`
	createTableBlacklistsSQLite = `
	CREATE TABLE IF NOT EXISTS Blacklists (
		OwnerAddress				TEXT NOT NULL,
		Address						TEXT NOT NULL
	);`
	createTableWorkersSQLite = `
	CREATE TABLE IF NOT EXISTS Workers (
		MasterID					TEXT NOT NULL,
		WorkerID					TEXT NOT NULL,
		Confirmed					INTEGER NOT NULL
	);`
	createTableMiscSQLite = `
	CREATE TABLE IF NOT EXISTS misc (
		LastKnownBlock				INTEGER NOT NULL
	);`
	insertDealSQLite              = `INSERT OR REPLACE INTO deals VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	insertOrderSQLite             = `INSERT OR REPLACE INTO orders VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	insertDealChangeRequestSQLite = `INSERT OR REPLACE INTO orders VALUES (?, ?, ?, ?, ?, ?, ?)`
	deleteOrderSQLite             = `DELETE FROM Orders WHERE Id=?`
	updateLastKnownBlockSQLite    = "UPDATE misc SET LastKnownBlock=?"
	selectLastKnownBlock          = `SELECT * from misc;`
)

type Filter struct {
	Field    string
	Operator string
	Value    interface{}
}

func NewFilter(field string, operator pb.ComparisonOperator, value interface{}) *Filter {
	var operatorStr string
	switch operator {
	case pb.ComparisonOperator_EQ:
		operatorStr = "="
	case pb.ComparisonOperator_GTE:
		operatorStr = ">"
	case pb.ComparisonOperator_LTE:
		operatorStr = "<"
	}

	fmt.Println(field, value)

	return &Filter{
		Field:    field,
		Operator: operatorStr,
		Value:    value,
	}
}

func RunQuery(db *sql.DB, table string, offset, limit uint64, filters ...*Filter) (*sql.Rows, string, error) {
	var (
		query      = fmt.Sprintf("SELECT * FROM %s", table)
		conditions []string
		values     []interface{}
	)
	for _, filter := range filters {
		conditions = append(conditions, fmt.Sprintf("%s%s?", filter.Field, filter.Operator))
		values = append(values, filter.Value)
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if limit > MaxLimit {
		limit = MaxLimit
	}
	query += " ORDER BY Id ASC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}
	query += ";"

	fmt.Println(query, values)

	rows, err := db.Query(query, values...)
	if err != nil {
		return nil, query, errors.Wrapf(err, "query `%s` failed", query)
	}

	return rows, query, nil
}
