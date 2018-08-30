package test

import (
	"os"

	"github.com/OpenBazaar/openbazaar-go/repo/db"
	"github.com/OpenBazaar/openbazaar-go/schema"
	"github.com/OpenBazaar/wallet-interface"
)

// Repository represents a test (temporary/volitile) repository
type Repository struct {
	Path     string
	Password string
	DB       *db.SQLiteDatastore
}

// NewRepository creates and initializes a new temporary repository for tests
func NewRepository() (*Repository, error) {
	r := &Repository{
		Path:     getNewRepoPath(),
		Password: getMnemonic(),
	}

	appSchema := schema.MustNewCustomSchemaManager(schema.SchemaContext{
		DataPath:        r.Path,
		Mnemonic:        getMnemonic(),
		TestModeEnabled: true,
	})
	if err := appSchema.BuildSchemaDirectories(); err != nil {
		return nil, err
	}
	if err := appSchema.InitializeDatabase(); err != nil {
		return nil, err
	}
	if err := appSchema.InitializeIPFSRepo(); err != nil {
		return nil, err
	}

	var err error
	r.DB, err = db.Create(r.Path, "", true, wallet.Bitcoin)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Delete removes the test repository
func (r *Repository) Delete() error {
	return deleteDirectory(r.Path)
}

func deleteDirectory(path string) error {
	err := os.RemoveAll(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
