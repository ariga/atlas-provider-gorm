// package recordriver provides a driver for database/sql which records queries and statements
// and allows you to set responses for queries. It is used for testing or providing a runtime replacement
// for a real database in cases where you want to learn the queries and statements that are executed.

package recordriver

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
	"sync"
)

func init() {
	sql.Register("recordriver", &drv{})
}

var (
	sessions = map[string]*Session{}
	mu       sync.Mutex
)

type (
	// Session is a session of recordriver which records queries and statements.
	Session struct {
		Queries    []string
		Statements []string
		responses  map[string]*Response
	}
	// Response is a response to a query.
	Response struct {
		Cols []string
		Data [][]driver.Value
	}
	drv  struct{}
	conn struct {
		session string
	}
	stmt struct {
		query   string
		session string
	}
	tx          struct{}
	emptyResult struct{}
)

// StrStmts returns the statements as a string, separated by semicolons and newlines.
func (s *Session) Stmts() string {
	var sb strings.Builder
	for _, stmt := range s.Statements {
		sb.WriteString(stmt)
		sb.WriteString(";\n")
	}
	return sb.String()
}

// Get returns the session with the given name and reports whether it exists.
func Session(name string) (*Session, bool) {
	mu.Lock()
	defer mu.Unlock()
	h, ok := sessions[name]
	return h, ok
}

// SetResponse sets the response for the given session and query.
func SetResponse(session string, query string, resp *Response) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := sessions[session]; !ok {
		sessions[session] = &Session{
			responses: make(map[string]*Response),
		}
	}
	sessions[session].responses[query] = resp
}

// Open returns a new connection to the database.
func (d *drv) Open(name string) (driver.Conn, error) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := sessions[name]; !ok {
		sessions[name] = &Session{
			responses: make(map[string]*Response),
		}
	}
	return &conn{session: name}, nil
}

// Prepare returns a prepared statement, bound to this connection.
func (mc *conn) Prepare(query string) (driver.Stmt, error) {
	return &stmt{query: query, session: mc.session}, nil
}

// Close closes the connection.
func (mc *conn) Close() error {
	mu.Lock()
	defer mu.Unlock()
	delete(sessions, mc.session)
	return nil
}

// Begin starts and returns a new transaction.
func (mc *conn) Begin() (driver.Tx, error) {
	return &tx{}, nil
}

// Commit commits the transaction. It is a noop.
func (*tx) Commit() error {
	return nil
}

// Rollback rolls back the transaction. It is a noop.
func (*tx) Rollback() error {
	return nil
}

// Close closes the statement.
func (*stmt) Close() error {
	return nil
}

// NumInput returns the number of placeholder parameters. Reporting -1 does not know the
// number of parameters.
func (*stmt) NumInput() int {
	return -1
}

// Exec executes a query that doesn't return rows, such as an CREATE or ALTER TABLE.
func (ms *stmt) Exec(_ []driver.Value) (driver.Result, error) {
	mu.Lock()
	defer mu.Unlock()
	sessions[ms.session].Statements = append(sessions[ms.session].Statements, ms.query)
	return emptyResult{}, nil
}

// Query executes a query that may return rows, such as an SELECT.
func (ms *stmt) Query(_ []driver.Value) (driver.Rows, error) {
	mu.Lock()
	defer mu.Unlock()
	s := ms.session
	sessions[s].Queries = append(sessions[s].Queries, ms.query)
	if resp, ok := sessions[s].responses[ms.query]; ok {
		return resp, nil
	}
	return &Response{}, nil
}

// Columns returns the names of the columns in the result set.
func (*Response) Columns() []string {
	return mr.Cols
}

// Close closes the rows iterator. It is a noop.
func (*Response) Close() error {
	return nil
}

// Next is called to populate the next row of data into the provided slice.
func (mr *Response) Next(dest []driver.Value) error {
	if len(mr.Data) == 0 {
		return io.EOF
	}
	copy(dest, mr.Data[0])
	mr.Data = mr.Data[1:]
	return nil
}

// LastInsertId returns the integer generated by the database in response to a command. LastInsertId
// always returns a value of 0.
func (emptyResult) LastInsertId() (int64, error) {
	return 0, nil
}

// RowsAffected returns the number of rows affected by the query. RowsAffected always returns a
// value of 0.
func (emptyResult) RowsAffected() (int64, error) {
	return 0, nil
}
