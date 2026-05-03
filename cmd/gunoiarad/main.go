package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/app"
	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/config"
	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/etl"
	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/metrics"
	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/store"
)

func main() {
	if err := run(); err != nil {
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	command := "serve"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	cfg := config.FromEnv()
	flags := flag.NewFlagSet(command, flag.ExitOnError)
	flags.StringVar(&cfg.HTTPAddr, "addr", cfg.HTTPAddr, "HTTP listen address")
	flags.StringVar(&cfg.DBPath, "db", cfg.DBPath, "SQLite database path")
	flags.StringVar(&cfg.RawDir, "raw-dir", cfg.RawDir, "raw snapshot directory")
	flags.StringVar(&cfg.PublicBaseURL, "public-base-url", cfg.PublicBaseURL, "public base URL for share and ICS links")
	flags.DurationVar(&cfg.RefreshInterval, "refresh-interval", cfg.RefreshInterval, "background ETL refresh interval")
	flags.BoolVar(&cfg.BootstrapETL, "bootstrap-etl", cfg.BootstrapETL, "run ETL before serving when database is empty")
	if len(os.Args) > 1 {
		if err := flags.Parse(os.Args[2:]); err != nil {
			return err
		}
	} else if err := flags.Parse(nil); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	switch command {
	case "serve":
		return serve(ctx, cfg)
	case "etl":
		return runETL(ctx, cfg)
	case "help", "-h", "--help":
		fmt.Println("usage: gunoiarad [serve|etl] [flags]")
		flags.PrintDefaults()
		return nil
	default:
		return fmt.Errorf("unknown command %q", command)
	}
}

func openStore(ctx context.Context, cfg config.Config) (*store.Store, error) {
	return store.Open(ctx, cfg.DBPath)
}

func runETL(ctx context.Context, cfg config.Config) error {
	db, err := openStore(ctx, cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	svc := &etl.Service{Store: db, RawDir: cfg.RawDir}
	summary, err := svc.Run(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("imported places=%d rules=%d events=%d campaigns=%d issues=%d sources=%d elapsed=%s\n",
		summary.Places,
		summary.Rules,
		summary.Events,
		summary.Campaigns,
		summary.Issues,
		summary.Sources,
		summary.Elapsed.Round(time.Millisecond),
	)
	for _, warning := range summary.Warnings {
		fmt.Println("warning:", warning)
	}
	return nil
}

func serve(ctx context.Context, cfg config.Config) error {
	db, err := openStore(ctx, cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	met := metrics.New()
	etlSvc := &etl.Service{Store: db, RawDir: cfg.RawDir}
	if err := maybeBootstrap(ctx, cfg, db, etlSvc, met); err != nil {
		return err
	}
	if counts, err := db.Counts(ctx); err == nil {
		met.ObserveCounts(counts)
	}

	if cfg.RefreshInterval > 0 {
		go refreshLoop(ctx, cfg.RefreshInterval, db, etlSvc, met)
	}

	server, err := app.NewServer(db, cfg, met)
	if err != nil {
		return err
	}
	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		slog.Info("serving Gunoi Arad", "addr", cfg.HTTPAddr, "db", cfg.DBPath)
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func maybeBootstrap(ctx context.Context, cfg config.Config, db *store.Store, etlSvc *etl.Service, met *metrics.Metrics) error {
	empty, err := db.IsEmpty(ctx)
	if err != nil {
		return err
	}
	if !cfg.BootstrapETL && !empty {
		return nil
	}
	if !cfg.BootstrapETL && empty {
		slog.Warn("database is empty; start with --bootstrap-etl or run `gunoiarad etl`")
		return nil
	}
	summary, err := etlSvc.Run(ctx)
	if err != nil {
		met.ObserveETL("error", 0, 0)
		return err
	}
	met.ObserveETL("success", summary.Elapsed, summary.Issues)
	if counts, err := db.Counts(ctx); err == nil {
		met.ObserveCounts(counts)
	}
	return nil
}

func refreshLoop(ctx context.Context, interval time.Duration, db *store.Store, etlSvc *etl.Service, met *metrics.Metrics) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			summary, err := etlSvc.Run(ctx)
			if err != nil {
				slog.Error("background ETL failed", "error", err)
				met.ObserveETL("error", summary.Elapsed, summary.Issues)
				continue
			}
			slog.Info("background ETL complete", "places", summary.Places, "events", summary.Events, "issues", summary.Issues)
			met.ObserveETL("success", summary.Elapsed, summary.Issues)
			if counts, err := db.Counts(ctx); err == nil {
				met.ObserveCounts(counts)
			}
		}
	}
}
