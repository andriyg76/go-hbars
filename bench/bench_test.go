//go:build benchgen

package bench

import (
	"fmt"
	"io"
	"testing"
)

var benchData = buildBenchData(100, 12, 8)

func buildBenchData(userCount, ordersPerUser, itemsPerOrder int) map[string]any {
	users := make([]any, 0, userCount)
	totalOrders := 0
	totalRevenue := 0.0
	roles := []string{"admin", "editor", "viewer"}

	for u := 0; u < userCount; u++ {
		orders := make([]any, 0, ordersPerUser)
		for o := 0; o < ordersPerUser; o++ {
			items := make([]any, 0, itemsPerOrder)
			subtotal := 0.0
			for i := 0; i < itemsPerOrder; i++ {
				price := float64(10 + i + u + o)
				subtotal += price
				items = append(items, fmt.Sprintf("item-%02d-%02d-%02d", u, o, i))
			}
			tax := subtotal * 0.08
			totalRevenue += subtotal + tax
			order := map[string]any{
				"id":        fmt.Sprintf("U%02d-O%02d", u, o),
				"items":     items,
				"subtotal":  subtotal,
				"tax":       tax,
				"createdAt": fmt.Sprintf("2024-05-%02d", (o%28)+1),
			}
			orders = append(orders, order)
			totalOrders++
		}

		user := map[string]any{
			"name":   fmt.Sprintf("User %02d", u),
			"email":  fmt.Sprintf("user%02d@example.com", u),
			"role":   roles[u%len(roles)],
			"active": u%2 == 0,
			"score":  60.0 + float64(u%41),
			"orders": orders,
		}
		users = append(users, user)
	}

	return map[string]any{
		"site": map[string]any{
			"title": "Benchmark Suite",
			"url":   "https://example.com/docs?utm=bench",
		},
		"rawHtml": "<p>Hello <strong>bench</strong>!</p>",
		"settings": map[string]any{
			"featureFlags": map[string]any{
				"beta":     true,
				"darkMode": false,
				"newUI":    true,
			},
		},
		"user": map[string]any{
			"name":   "Bench User",
			"email":  "bench@example.com",
			"joined": "2024-04-01T12:00:00Z",
			"tags":   []any{"perf", "bench", "go"},
		},
		"cardPartial": "userCard",
		"users": users,
		"stats": map[string]any{
			"count": totalOrders,
			"total": totalRevenue,
		},
	}
}

func BenchmarkRenderMain(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := RenderMain(io.Discard, benchData); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRenderSummary(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := RenderSummary(io.Discard, benchData); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRenderHelperHeavy(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := RenderHelperHeavy(io.Discard, benchData); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRenderMainString(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := RenderMainString(benchData); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRenderMain_RecreateData(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		data := buildBenchData(100, 12, 8)
		if err := RenderMain(io.Discard, data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRenderSummary_RecreateData(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		data := buildBenchData(100, 12, 8)
		if err := RenderSummary(io.Discard, data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRenderHelperHeavy_RecreateData(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		data := buildBenchData(100, 12, 8)
		if err := RenderHelperHeavy(io.Discard, data); err != nil {
			b.Fatal(err)
		}
	}
}

