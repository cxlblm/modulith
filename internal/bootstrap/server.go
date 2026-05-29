package bootstrap

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"gorm.io/gorm"

	"modular_monolith/internal/account"
	"modular_monolith/internal/entitlement"
	"modular_monolith/internal/fulfillment"
	"modular_monolith/internal/order"
	"modular_monolith/internal/order/adapters/accountclient"
	"modular_monolith/internal/order/adapters/pricingclient"
	"modular_monolith/internal/order/adapters/shopclient"
	"modular_monolith/internal/payment"
	"modular_monolith/internal/platform/config"
	"modular_monolith/internal/platform/eventbus"
	"modular_monolith/internal/platform/httpserver"
	"modular_monolith/internal/platform/logging"
	"modular_monolith/internal/platform/mysql"
	"modular_monolith/internal/pricing"
	pricingshopclient "modular_monolith/internal/pricing/adapters/shopclient"
	"modular_monolith/internal/quiz"
	"modular_monolith/internal/quiz/adapters/entitlementclient"
	"modular_monolith/internal/quiz/adapters/rewardclient"
	"modular_monolith/internal/reward"
	"modular_monolith/internal/shop"
)

func Run(ctx context.Context, cfg config.Config) error {
	return run(ctx, cfg, os.Stdout)
}

func run(ctx context.Context, cfg config.Config, logOutput io.Writer) error {
	logger, err := logging.New(cfg.Log, logOutput)
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}

	server := httpserver.New(cfg.HTTP, logger)
	var db *gorm.DB
	if cfg.MySQL.DSN != "" {
		var err error
		db, err = mysql.Open(ctx, cfg.MySQL)
		if err != nil {
			return fmt.Errorf("open mysql: %w", err)
		}
		defer func() {
			if err := mysql.Close(db); err != nil {
				logger.ErrorContext(ctx, "close mysql", "error", err)
			}
		}()
		if err := mountBusiness(ctx, server, db, cfg.MySQL.AutoMigrate, logger); err != nil {
			return err
		}
	}

	if err := server.Start(ctx); err != nil {
		return fmt.Errorf("start http server: %w", err)
	}
	return nil
}

func mountBusiness(ctx context.Context, server *httpserver.Server, db *gorm.DB, autoMigrate bool, logger *slog.Logger) error {
	if autoMigrate {
		if err := db.WithContext(ctx).AutoMigrate(allModels()...); err != nil {
			return fmt.Errorf("auto migrate mysql models: %w", err)
		}
	}

	bus := eventbus.New()

	accountMod, err := account.NewModule(account.Deps{DB: db, Logger: logger})
	if err != nil {
		return fmt.Errorf("create account module: %w", err)
	}
	shopMod, err := shop.NewModule(shop.Deps{DB: db, Logger: logger})
	if err != nil {
		return fmt.Errorf("create shop module: %w", err)
	}
	productCatalog := pricingshopclient.NewProductCatalogService(shopMod.PortsModule)
	pricingMod, err := pricing.NewModule(pricing.Deps{DB: db, Logger: logger, Products: productCatalog})
	if err != nil {
		return fmt.Errorf("create pricing module: %w", err)
	}
	rewardMod, err := reward.NewModule(reward.Deps{DB: db, Logger: logger})
	if err != nil {
		return fmt.Errorf("create reward module: %w", err)
	}
	entitlementMod, err := entitlement.NewModule(entitlement.Deps{DB: db, Logger: logger})
	if err != nil {
		return fmt.Errorf("create entitlement module: %w", err)
	}
	rewardService := rewardclient.NewRewardService(rewardMod.PortsModule)
	revivalCards := entitlementclient.NewRevivalCardsService(entitlementMod.PortsModule)
	quizMod, err := quiz.NewModule(quiz.Deps{DB: db, Logger: logger, Rewards: rewardService, Revivals: revivalCards})
	if err != nil {
		return fmt.Errorf("create quiz module: %w", err)
	}

	products := shopclient.NewProductsService(shopMod.PortsModule)
	addresses := accountclient.NewAddressService(accountMod.PortsModule)
	users := accountclient.NewUserEligibilityService(accountMod.PortsModule)
	pricingService := pricingclient.NewPricingService(pricingMod.PortsModule)
	orderMod, err := order.NewModule(order.Deps{
		DB:        db,
		Logger:    logger,
		EventBus:  bus,
		Products:  products,
		Addresses: addresses,
		Users:     users,
		Pricing:   pricingService,
	})
	if err != nil {
		return fmt.Errorf("create order module: %w", err)
	}
	paymentMod, err := payment.NewModule(payment.Deps{DB: db, Logger: logger, EventBus: bus})
	if err != nil {
		return fmt.Errorf("create payment module: %w", err)
	}
	fulfillmentMod, err := fulfillment.NewModule(fulfillment.Deps{DB: db, Logger: logger, EventBus: bus})
	if err != nil {
		return fmt.Errorf("create fulfillment module: %w", err)
	}

	accountMod.RegisterHTTP(server.Echo())
	shopMod.RegisterHTTP(server.Echo())
	orderMod.RegisterHTTP(server.Echo())
	paymentMod.RegisterHTTP(server.Echo())
	fulfillmentMod.RegisterHTTP(server.Echo())
	entitlementMod.RegisterHTTP(server.Echo())
	quizMod.RegisterHTTP(server.Echo())

	orderMod.RegisterEventSubs()
	paymentMod.RegisterEventSubs()
	fulfillmentMod.RegisterEventSubs()

	return nil
}

func allModels() []any {
	var models []any
	models = append(models, account.Models()...)
	models = append(models, shop.Models()...)
	models = append(models, pricing.Models()...)
	models = append(models, order.Models()...)
	models = append(models, payment.Models()...)
	models = append(models, fulfillment.Models()...)
	models = append(models, reward.Models()...)
	models = append(models, entitlement.Models()...)
	models = append(models, quiz.Models()...)
	return models
}
