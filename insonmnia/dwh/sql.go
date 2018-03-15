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
	CREATE TABLE IF NOT EXISTS deals (
		Id					TEXT PRIMARY KEY,
		Benchmarks			BLOB NOT NULL,
		SupplierID			TEXT NOT NULL,
		ConsumerID			TEXT NOT NULL,
		MasterID			TEXT NOT NULL,
		AskID				TEXT NOT NULL,
		BidID				TEXT NOT NULL,
		Duration 			UNSIGNED INTEGER NOT NULL,
		Price				TEXT NOT NULL,
		Start_time			UNSIGNED INTEGER NOT NULL,
		End_time			UNSIGNED INTEGER NOT NULL,
		Status				UNSIGNED INTEGER NOT NULL,
		Blocked_balance		TEXT NOT NULL,
		Total_payout		TEXT NOT NULL,
		Last_bill_ts		UNSIGNED INTEGER NOT NULL
	);`
	createTableOrdersSQLite = `
	CREATE TABLE IF NOT EXISTS orders (
		Id					TEXT PRIMARY KEY,
		Type				UNSIGNED INTEGER NOT NULL,
		Status				UNSIGNED INTEGER NOT NULL,
		AuthorID			TEXT NOT NULL,
		CounterpartyID		TEXT NOT NULL,
		Price				TEXT NOT NULL,
		Duration 			UNSIGNED INTEGER NOT NULL,
		Netflags			BLOB NOT NULL,
		IdentityLevel		UNSIGNED INTEGER NOT NULL,
		Blacklist			TEXT NOT NULL,
		Tag					BLOB NOT NULL,
		Benchmarks			BLOB NOT NULL,
		Frozen_sum			TEXT NOT NULL
	);`
	createTableChangesSQLite = `
	CREATE TABLE IF NOT EXISTS change_requests (
		Duration 			UNSIGNED INTEGER NOT NULL,
		Price				TEXT NOT NULL,
		Deal				TEXT NOT NULL,
		FOREIGN KEY (deal)	REFERENCES deals(Id) ON DELETE CASCADE
	);`
	createTableMiscSQLite = `
	CREATE TABLE IF NOT EXISTS misc (
		Last_known_block	INTEGER NOT NULL
	);`
	insertDealSQLite           = `INSERT OR REPLACE INTO deals VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	insertOrderSQLite          = `INSERT OR REPLACE INTO orders VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	deleteOrderSQLite          = `DELETE FROM orders WHERE id=?`
	updateLastKnownBlockSQLite = "UPDATE misc SET last_known_block=?"
	selectLastKnownBlock       = `SELECT * from misc;`
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

	rows, err := db.Query(query, values...)
	if err != nil {
		return nil, query, errors.Wrapf(err, "query `%s` failed", query)
	}

	return rows, query, nil
}
