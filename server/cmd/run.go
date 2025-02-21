package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-logr/stdr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/llmariner/api-usage/pkg/sender"
	"github.com/llmariner/common/pkg/aws"
	"github.com/llmariner/common/pkg/db"
	"github.com/llmariner/rbac-manager/pkg/auth"
	v1 "github.com/llmariner/user-manager/api/v1"
	"github.com/llmariner/user-manager/server/internal/config"
	"github.com/llmariner/user-manager/server/internal/server"
	"github.com/llmariner/user-manager/server/internal/store"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func runCmd() *cobra.Command {
	var path string
	var logLevel int
	cmd := &cobra.Command{
		Use:   "run",
		Short: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := config.Parse(path)
			if err != nil {
				return err
			}
			if err := c.Validate(); err != nil {
				return err
			}
			stdr.SetVerbosity(logLevel)
			if err := run(cmd.Context(), &c); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "config", "", "Path to the config file")
	cmd.Flags().IntVar(&logLevel, "v", 0, "Log level")
	_ = cmd.MarkFlagRequired("config")
	return cmd
}

func run(ctx context.Context, c *config.Config) error {
	logger := stdr.New(log.Default())
	log := logger.WithName("boot")

	var dbInst *gorm.DB
	var err error
	if c.Debug.Standalone {
		dbInst, err = gorm.Open(sqlite.Open(c.Debug.SqlitePath), &gorm.Config{})
	} else {
		dbInst, err = db.OpenDB(c.Database)
	}
	if err != nil {
		return err
	}

	st := store.New(dbInst)
	if err := st.AutoMigrate(); err != nil {
		return err
	}

	addr := fmt.Sprintf("localhost:%d", c.GRPCPort)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return err
	}
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			// Do not use the camel case for JSON fields to follow OpenAI API.
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:     true,
				EmitDefaultValues: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
		runtime.WithIncomingHeaderMatcher(auth.HeaderMatcher),
		runtime.WithHealthzEndpoint(grpc_health_v1.NewHealthClient(conn)),
	)
	if err := v1.RegisterUsersServiceHandlerFromEndpoint(ctx, mux, addr, opts); err != nil {
		return err
	}

	errCh := make(chan error)
	go func() {
		log.Info("Starting HTTP server...", "port", c.HTTPPort)
		errCh <- http.ListenAndServe(fmt.Sprintf(":%d", c.HTTPPort), mux)
	}()

	var usageSetter sender.UsageSetter
	if c.UsageSender.Enable {
		usage, err := sender.New(ctx, c.UsageSender, grpc.WithTransportCredentials(insecure.NewCredentials()), logger)
		if err != nil {
			return err
		}
		go func() { usage.Run(ctx) }()
		usageSetter = usage
	} else {
		usageSetter = sender.NoopUsageSetter{}
	}

	var dataKey []byte
	if c.KMSConfig.Enable {
		opts := aws.NewConfigOptions{
			Region: c.KMSConfig.Region,
		}
		if ar := c.KMSConfig.AssumeRole; ar != nil {
			opts.AssumeRole = &aws.AssumeRole{
				RoleARN:    ar.RoleARN,
				ExternalID: ar.ExternalID,
			}
		}
		kmsClient, err := aws.NewKMSClient(ctx, opts, c.KMSConfig.KeyAlias)
		if err != nil {
			return err
		}
		dk, err := st.GetDataKey(ctx, kmsClient)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			log.Info("Creating a data key")
			dk, err = st.CreateDataKey(ctx, kmsClient)
			if err != nil {
				return err
			}
		}
		dataKey = dk
	}

	s := server.New(st, dataKey, logger)
	go func() {
		errCh <- s.Run(ctx, c.GRPCPort, c.AuthConfig, usageSetter)
	}()

	if err := createDefaultResources(ctx, s, c); err != nil {
		return err
	}

	go func() {
		s := server.NewInternal(st, dataKey, logger)
		errCh <- s.Run(c.InternalGRPCPort)
	}()

	return <-errCh
}

func createDefaultResources(ctx context.Context, s *server.S, c *config.Config) error {
	if c.DefaultOrganization == nil {
		return nil
	}
	org, err := s.CreateDefaultOrganization(ctx, c.DefaultOrganization)
	if err != nil {
		return err
	}

	if c.DefaultProject == nil {
		return nil
	}
	project, err := s.CreateDefaultProject(ctx, c.DefaultProject, org.OrganizationID, c.DefaultOrganization.TenantID)
	if err != nil {
		return err
	}

	for _, k := range c.DefaultAPIKeys {
		if err := s.CreateDefaultAPIKey(ctx, &k, org.OrganizationID, project.ProjectID, c.DefaultOrganization.TenantID); err != nil {
			return err
		}
	}
	return nil
}
