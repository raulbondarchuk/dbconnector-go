package dbconnector

import (
	"fmt"
	"log"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DBManagerMlt struct {
	db  *gorm.DB
	dsn string
}

func (d *DBManagerMlt) GetDB() *gorm.DB {
	return d.db
}

type ManagerRegistry struct {
	managers map[string]*DBManagerMlt
	sync.RWMutex
}

var (
	instanceMlt *ManagerRegistry
	onceMlt     sync.Once
)

// GetInstanceMlt returns the singleton instance of ManagerRegistry.
func GetInstanceMlt() *ManagerRegistry {
	onceMlt.Do(func() {
		instanceMlt = &ManagerRegistry{
			managers: make(map[string]*DBManagerMlt),
		}
	})
	return instanceMlt
}

// AddDBManager adds or updates a DBManager instance in the registry based on the name.
func (r *ManagerRegistry) AddDBManager(name string) {
	r.Lock()
	defer r.Unlock()

	if _, exists := r.managers[name]; exists {
		log.Printf("DBManager for '%s' already exists. Reusing the existing connection.", name)
		return
	}

	dsn, err := createDSN(name)
	if err != nil {
		log.Fatalf("Error creating DSN for '%s': %v", name, err)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error opening database for '%s': %v", name, err)
	}

	r.managers[name] = &DBManagerMlt{db: db, dsn: dsn}
}

// GetDBManager returns a DBManager instance by name from the registry.
func (r *ManagerRegistry) GetDBManager(name string) *DBManagerMlt {
	r.RLock()
	defer r.RUnlock()

	manager, exists := r.managers[name]
	if !exists {
		log.Fatalf("DBManager named '%s' does not exist.", name)
	}
	return manager
}

// createDSN creates a DSN string for the given database name.
func createDSN(name string) (string, error) {
	user := viper.GetString(fmt.Sprintf("%s.user", name))
	pass := viper.GetString(fmt.Sprintf("%s.pass", name))
	host := viper.GetString(fmt.Sprintf("%s.host", name))
	port := viper.GetString(fmt.Sprintf("%s.port", name))
	database := viper.GetString(fmt.Sprintf("%s.database", name))

	if user == "" || pass == "" || host == "" || port == "" || database == "" {
		return "", fmt.Errorf("missing database configuration for '%s'", name)
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, pass, host, port, database), nil
}
