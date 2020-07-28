package balance

import (
	"github.com/social-network/subscan-plugin/router"
	"github.com/social-network/subscan-plugin/storage"
	"github.com/social-network/subscan/plugins/balance/http"
	"github.com/social-network/subscan/plugins/balance/model"
	"github.com/social-network/subscan/plugins/balance/service"
	"github.com/shopspring/decimal"
)

var srv *service.Service

type Account struct {
	d storage.Dao
}

func New() *Account {
	return &Account{}
}

func (a *Account) InitDao(d storage.Dao) {
	srv = service.New(d)
	a.d = d
	a.Migrate()
}

func (a *Account) InitHttp() []router.Http {
	return http.Router(srv)
}

func (a *Account) ProcessExtrinsic(block *storage.Block, extrinsic *storage.Extrinsic, events []storage.Event) error {
	return nil
}

func (a *Account) ProcessEvent(block *storage.Block, event *storage.Event, fee decimal.Decimal) error {
	return nil
}

func (a *Account) Version() string {
	return "0.1"
}

func (a *Account) Migrate() {
	db := a.d.DB()
	db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(
		&model.Account{},
	)
	db.Model(model.Account{}).AddUniqueIndex("address", "address")
}
