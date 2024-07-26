package tracing

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"

	"go.opentelemetry.io/contrib/detectors/aws/eks"
)

func InitTracing(log logrus.FieldLogger) {
	ctx := context.Background()

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "0.0.0.0:4317" // setting default endpoint for exporter
	}

	// Create and start new OTLP trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure(), otlptracegrpc.WithEndpoint(endpoint), otlptracegrpc.WithDialOption(grpc.WithBlock()))
	if err != nil {
		log.Fatalf("failed to create new OTLP trace exporter: %v", err)
		return
	}

	idg := xray.NewIDGenerator()

	// Instantiate a new EKS Resource detector
	eksResourceDetector := eks.NewResourceDetector()
	resource, err := eksResourceDetector.Detect(context.Background())
	if err != nil {
		log.Fatalf("failed to create new OTLP trace exporter: %v", err)
		return
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource),
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithIDGenerator(idg),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(xray.Propagator{})
}
