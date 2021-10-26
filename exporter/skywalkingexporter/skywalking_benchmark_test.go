package skywalkingexporter

import (
	"context"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/testdata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"google.golang.org/grpc"
	"io"
	"net"
	v3 "skywalking.apache.org/repo/goapi/collect/common/v3"
	logpb "skywalking.apache.org/repo/goapi/collect/logging/v3"
	"sync"
	"testing"
	"time"
)

func TestIrectReject_SlotCheck_4(t *testing.T) {
	exporter, server := doInit(16, 16, t)
	l := testdata.GenerateLogsOneLogRecordNoResource()
	l.ResourceLogs().At(0).InstrumentationLibraryLogs().At(0).Logs().At(0).Body().SetIntVal(0)

	for i := 0; i < 20; i++ {
		exporter.ConsumeLogs(context.Background(), l)
	}

	a := time.Now().UnixMilli()
	w1 := &sync.WaitGroup{}
	for i := 0; i < 1000000; i++ {
			exporter.ConsumeLogs(context.Background(), l)
		}()
	}
	w1.Wait()
	b := time.Now().UnixMilli()
	println(b - a)
	server.Stop()
}

func doInit(numConsumers int, numStream int, t *testing.T) (component.LogsExporter, *grpc.Server) {
	server, addr := initializeGRPC(grpc.MaxConcurrentStreams(100))
	tt := &Config{
		NumStreams:       numStream,
		ExporterSettings: config.NewExporterSettings(config.NewComponentID(typeStr)),
		QueueSettings: exporterhelper.QueueSettings{
			Enabled:      true,
			NumConsumers: numConsumers,
			QueueSize:    1000,
		},
		GRPCClientSettings: configgrpc.GRPCClientSettings{
			Endpoint: addr.String(),
			TLSSetting: &configtls.TLSClientSetting{
				Insecure: true,
			},
		},
	}

	oce := newExporter(context.Background(), tt)
	got, err := exporterhelper.NewLogsExporter(
		tt,
		componenttest.NewNopExporterCreateSettings(),
		oce.pushLogs,
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		exporterhelper.WithRetry(tt.RetrySettings),
		exporterhelper.WithQueue(tt.QueueSettings),
		exporterhelper.WithTimeout(tt.TimeoutSettings),
		exporterhelper.WithStart(oce.start),
		exporterhelper.WithShutdown(oce.shutdown),
	)

	if err != nil {
		t.Errorf("error")
	}

	err = got.Start(context.Background(), componenttest.NewNopHost())

	return got, server
}

func initializeGRPC(opts ...grpc.ServerOption) (*grpc.Server, net.Addr) {
	server := grpc.NewServer(opts...)
	lis, _ := net.Listen("tcp", "localhost:0")
	m := &mockLogHandler2{
		logChan: nil,
	}
	logpb.RegisterLogReportServiceServer(
		server,
		m,
	)
	go func() {
		err := server.Serve(lis)
		if err != nil {
			return
		}
	}()
	return server, lis.Addr()
}

type mockLogHandler2 struct {
	logChan chan *logpb.LogData
	logpb.UnimplementedLogReportServiceServer
}

func (h *mockLogHandler2) Collect(stream logpb.LogReportService_CollectServer) error {
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&v3.Commands{})
		}
		time.Sleep(time.Millisecond * 5)
	}
}
