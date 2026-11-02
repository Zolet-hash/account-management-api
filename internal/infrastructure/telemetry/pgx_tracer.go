package telemetry

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type QueryTracer struct{}

// QueryTracer traces PostgreSQL queries using OpenTelemetry, recording operation details, execution errors, and affected rows.
func (QueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	// extract first word of sql & convert to upper, SELECT, INSERT etc
	operation := "unknown"
	if fields := strings.Fields(data.SQL); len(fields) > 0 {
		operation = strings.ToUpper(fields[0])
	}
	ctx, _ = otel.Tracer("account-management/postgres").Start(
		ctx,
		"postgres.query",
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation.name", operation)))
	return ctx
}
