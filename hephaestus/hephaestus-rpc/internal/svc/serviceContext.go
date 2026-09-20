package svc

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/zrpc"
	"jian-unified-system/hephaestus/hephaestus-rpc/internal/application"
	"jian-unified-system/hephaestus/hephaestus-rpc/internal/config"
	"jian-unified-system/hephaestus/hephaestus-rpc/internal/infrastructure/files"
	"jian-unified-system/hephaestus/hephaestus-rpc/internal/infrastructure/mysql"
	"jian-unified-system/hephaestus/internal/apollosso"
	"os"
	"path/filepath"
	"time"
)

type ServiceContext struct {
	Config config.Config
	App    *application.Service
	Store  *mysql.Store
}

func NewServiceContext(c config.Config) *ServiceContext {
	zrpc.DontLogContentForMethod("/income.Imports/GoogleStage")
	if len(c.ServiceKey) < 32 {
		panic("ServiceKey must contain at least 32 characters")
	}
	if !filepath.IsAbs(c.Storage.Root) {
		panic("Storage.Root must be an absolute private directory")
	}
	if e := os.MkdirAll(c.Storage.Root, 0700); e != nil {
		panic("income private storage unavailable")
	}
	probe, e := os.CreateTemp(c.Storage.Root, ".readiness-")
	if e != nil {
		panic("income private storage is not writable")
	}
	probe.Close()
	os.Remove(probe.Name())
	store, e := mysql.Open(c.DB.DataSource)
	if e != nil {
		panic("cannot configure income database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if e = store.Health(ctx); e != nil {
		store.DB.Close()
		panic(fmt.Sprintf("income schema/readiness check failed: %v", e))
	}
	authority, e := apollosso.New(c.Apollo)
	if e != nil {
		store.DB.Close()
		panic(e)
	}
	return &ServiceContext{c, &application.Service{Authority: authority, Repo: store, Parser: &files.Parser{Root: c.Storage.Root}, Now: time.Now}, store}
}
