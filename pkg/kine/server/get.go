package server

import (
	"context"
	"fmt"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.opentelemetry.io/otel/attribute"
)

func (l *LimitedServer) get(ctx context.Context, r *etcdserverpb.RangeRequest) (*RangeResponse, error) {
	var err error
	getCnt.Add(ctx, 1)
	ctx, span := otelTracer.Start(ctx, fmt.Sprintf("%s.get", otelName))
	defer func() {
		span.RecordError(err)
		span.End()
	}()

	span.SetAttributes(
		attribute.String("key", string(r.Key)),
		attribute.Int64("limit", r.Limit),
		attribute.Int64("revision", r.Revision),
	)

	if len(r.RangeEnd) != 0 {
		return nil, fmt.Errorf("unexpected rangeEnd: want empty, got %s", r.RangeEnd)
	}
	if r.Limit != 0 {
		return nil, fmt.Errorf("unexpected limit: want 0, got %d", r.Limit)
	}

	rev, kv, err := l.backend.List(ctx, string(r.Key), "", 1, r.Revision)
	if err != nil {
		return nil, err
	}
	return &RangeResponse{
		Header: txnHeader(rev),
		Kvs:    kv,
	}, nil
}
