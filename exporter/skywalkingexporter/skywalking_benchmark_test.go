package skywalkingexporter

import (
	"context"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/testdata"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"google.golang.org/grpc"
	v3 "skywalking.apache.org/repo/goapi/collect/common/v3"
	logpb "skywalking.apache.org/repo/goapi/collect/logging/v3"
)

func TestIrectReject_SlotCheck_4(t *testing.T) {
	test(2, 1, t)
	test(2, 2, t)
	test(2, 3, t)
	test(2, 4, t)
	test(2, 5, t)
	test(2, 10, t)

	test(3, 1, t)
	test(3, 2, t)
	test(3, 3, t)
	test(3, 4, t)
	test(3, 5, t)
	test(3, 10, t)

	test(4, 1, t)
	test(4, 2, t)
	test(4, 3, t)
	test(4, 4, t)
	test(4, 5, t)
	test(4, 10, t)
}

func test(nThread int, nStream int, t *testing.T) {
	exporter, server := doInit(nStream, t)

	l := testdata.GenerateLogsOneLogRecordNoResource()
	l.ResourceLogs().At(0).InstrumentationLibraryLogs().At(0).Logs().At(0).Body().SetIntVal(0)

	for i := 0; i < 100; i++ {
		exporter.pushLogs(context.Background(), l)
	}

	workers := nThread
	w := &sync.WaitGroup{}
	sum := 1000000
	a := time.Now().UnixMilli()
	for i := 0; i < workers; i++ {
		w.Add(1)
		go func() {
			defer w.Done()
			for i := 0; i < sum/workers; i++ {
				exporter.pushLogs(context.Background(), l)
			}
		}()
	}
	w.Wait()
	b := time.Now().UnixMilli()
	print(" nThread:")
	print(nThread)
	print(" nStream:")
	print(nStream)
	print(" time:")
	println(b - a)
	server.Stop()
	exporter.shutdown(context.Background())
	time.Sleep(time.Second * 5)
}

func doInit(numStream int, t *testing.T) (*swExporter, *grpc.Server) {
	server, addr := initializeGRPC(grpc.MaxConcurrentStreams(100))
	tt := &Config{
		NumStreams:       numStream,
		ExporterSettings: config.NewExporterSettings(config.NewComponentID(typeStr)),
		QueueSettings: exporterhelper.QueueSettings{
			Enabled:      true,
			NumConsumers: 1,
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

	return oce, server
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
		if err != nil {
			stream.SendAndClose(&v3.Commands{})
			break
		}
	}
	return nil
}
