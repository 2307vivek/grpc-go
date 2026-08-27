/*
 *
 * Copyright 2026 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package transport

import (
	"context"
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// BenchmarkServerHeaderFields benchmarks the real writeHeaderLocked and
// writeStatus code paths with varying amounts of metadata, mirroring the
// approach used in BenchmarkCreateHeaderFields for the client side.
func BenchmarkServerHeaderFields(b *testing.B) {
	for _, mdCount := range []int{0, 4, 12, 48} {
		b.Run(fmt.Sprintf("writeHeaderLocked_mdcoount=%d", mdCount), func(b *testing.B) {

			// Build the metadata once for all iterations in this sub-benchmark.
			md := make(metadata.MD, mdCount)
			for i := 0; i < mdCount; i++ {
				md[fmt.Sprintf("header-%d", i)] = []string{
					fmt.Sprintf("value-%d-a", i),
					fmt.Sprintf("value-%d-b", i),
				}
			}

			// Set up a minimal http2Server with just enough to exercise the
			// real writeHeaderLocked and writeStatus methods.
			done := make(chan struct{})
			t := &http2Server{
				controlBuf:            newControlBuffer(done),
				setResetPingStrikes:   func() {},
				maxSendHeaderListSize: nil, // no limit — skip header list size check
			}

			// Create context and stream once, reuse in loop to avoid allocations.
			ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
			defer cancel()

			s := &ServerStream{
				Stream: Stream{
					id:             1,
					fc:             inFlow{limit: 65536},
					ctx:            ctx,
					contentSubtype: "proto",
				},
				st:     t,
				header: md,
			}
			s.cancel = cancel

			b.ReportAllocs()
			for b.Loop() {
				// Reset stream state for each iteration.
				s.state = streamActive
				s.headerSent.Store(false)
				if err := t.writeHeaderLocked(s); err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("writeSatus_mdcoount=%d", mdCount), func(b *testing.B) {

			// Build the metadata once for all iterations in this sub-benchmark.
			md := make(metadata.MD, mdCount)
			for i := 0; i < mdCount; i++ {
				md[fmt.Sprintf("header-%d", i)] = []string{
					fmt.Sprintf("value-%d-a", i),
					fmt.Sprintf("value-%d-b", i),
				}
			}

			// Set up a minimal http2Server with just enough to exercise the
			// real writeHeaderLocked and writeStatus methods.
			done := make(chan struct{})
			t := &http2Server{
				controlBuf:            newControlBuffer(done),
				setResetPingStrikes:   func() {},
				maxSendHeaderListSize: nil, // no limit — skip header list size check
			}

			// Create context and stream once, reuse in loop to avoid allocations.
			ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
			defer cancel()

			s := &ServerStream{
				Stream: Stream{
					id:             1,
					fc:             inFlow{limit: 65536},
					ctx:            ctx,
					contentSubtype: "proto",
				},
				st:     t,
				header: md,
			}
			s.cancel = cancel
			b.ReportAllocs()
			st := status.New(codes.OK, "OK")
			for b.Loop() {
				// Reset stream state for each iteration.
				s.state = streamActive
				s.headerSent.Store(false)
				s.trailer = nil
				if err := t.writeStatus(s, st); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
