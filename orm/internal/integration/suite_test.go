//go:build e2e

package integration

import (
	"github.com/jackycsl/geektime-go-practical/orm"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
	driver string
	dsn    string

	db *orm.DB
}

func (s *Suite) SetupSuite() {
	db, err := orm.Open(s.driver, s.dsn)
	require.NoError(s.T(), err)
	db.Wait()
	s.db = db
}
