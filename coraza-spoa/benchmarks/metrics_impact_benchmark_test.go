package benchmarks

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	flowcontroller "github.com/anticoi/EscudoWAF/coraza-spoa/internal/flow-controller"
	"github.com/rs/zerolog"
)

// BenchmarkMetricsOverhead mide el coste de Metrics en el sistema
func BenchmarkMetricsOverhead(b *testing.B) {
	logger := zerolog.Nop()

	b.Run("WithMetrics", func(b *testing.B) {
		config := flowcontroller.DefaultConfig()
		config.MetricsEnabled = true
		config.Capacity = 1000

		recorder := flowcontroller.NewMemoryIPRecorderWithConfig(config, logger)
		defer recorder.Close()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			ip := "192.168.1.1"
			for pb.Next() {
				recorder.IsIPBlocked(ip)
			}
		})
	})

	b.Run("WithoutMetrics", func(b *testing.B) {
		config := flowcontroller.DefaultConfig()
		config.MetricsEnabled = false
		config.Capacity = 1000

		recorder := flowcontroller.NewMemoryIPRecorderWithConfig(config, logger)
		defer recorder.Close()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			ip := "192.168.1.1"
			for pb.Next() {
				recorder.IsIPBlocked(ip)
			}
		})
	})
}

// BenchmarkIPRecorderRealWorld mide el impacto de rendimiento en escenarios reales
func BenchmarkIPRecorderRealWorld(b *testing.B) {
	logger := zerolog.Nop()
	config := flowcontroller.DefaultConfig()
	config.Capacity = 10000

	recorder := flowcontroller.NewMemoryIPRecorderWithConfig(config, logger)
	defer recorder.Close()

	// Añadir previamente algunos registros de IP
	for i := 0; i < 1000; i++ {
		ip := fmt.Sprintf("192.168.1.%d", i%255)
		recorder.RecordBlockedIP(ip, "test", "/", time.Hour)
	}

	b.Run("IsIPBlocked_Miss", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// Probar el escenario de fallo de caché
				recorder.IsIPBlocked("10.0.0.1")
			}
		})
	})

	b.Run("IsIPBlocked_Hit", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// Probar el escenario de acierto de caché
				recorder.IsIPBlocked("192.168.1.1")
			}
		})
	})

	b.Run("RecordBlockedIP", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				ip := fmt.Sprintf("10.1.1.%d", i%255)
				recorder.RecordBlockedIP(ip, "test", "/", time.Hour)
				i++
			}
		})
	})
}

// BenchmarkAtomicOperations mide el rendimiento de operaciones atómicas
func BenchmarkAtomicOperations(b *testing.B) {
	var counter atomic.Uint64

	b.Run("AtomicAdd", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				counter.Add(1)
			}
		})
	})

	b.Run("AtomicStore", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				counter.Store(uint64(time.Now().UnixNano()))
			}
		})
	})

	b.Run("AtomicLoad", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = counter.Load()
			}
		})
	})
}

// BenchmarkMetricsUpdate simula un escenario real de actualización de métricas
func BenchmarkMetricsUpdate(b *testing.B) {
	metrics := &flowcontroller.Metrics{}

	b.Run("SingleMetricUpdate", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				metrics.CacheHits.Add(1)
			}
		})
	})

	b.Run("MultipleMetricsUpdate", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// Simular la actualización de métricas en IsIPBlocked
				metrics.CacheHits.Add(1)
				metrics.CurrentBlocked.Store(100)
			}
		})
	})

	b.Run("AllMetricsUpdate", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// Simular un escenario completo de actualización de métricas
				metrics.TotalBlocked.Add(1)
				metrics.TotalExpired.Add(1)
				metrics.CurrentBlocked.Store(100)
				metrics.WriteQueueSize.Store(50)
				metrics.CacheHits.Add(1)
				metrics.CacheMisses.Add(1)
				metrics.CleanupDuration.Store(time.Millisecond)
			}
		})
	})
}

// BenchmarkMemoryFootprint mide la huella de memoria
func BenchmarkMemoryFootprint(b *testing.B) {
	b.Run("WithMetrics", func(b *testing.B) {
		var metrics []*flowcontroller.Metrics
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			m := &flowcontroller.Metrics{}
			metrics = append(metrics, m)
		}

		b.StopTimer()
		// Evitar la optimización del compilador
		_ = metrics
	})

	b.Run("WithoutMetrics", func(b *testing.B) {
		var counters []int
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			counters = append(counters, i)
		}

		b.StopTimer()
		// Evitar la optimización del compilador
		_ = counters
	})
}

// BenchmarkCacheSimulation simula el coste de cálculo de tasa de aciertos de caché
func BenchmarkCacheSimulation(b *testing.B) {
	metrics := &flowcontroller.Metrics{}

	b.Run("CacheHitRateCalculation", func(b *testing.B) {
		// Preconfigurar algunos datos base
		metrics.CacheHits.Store(1000)
		metrics.CacheMisses.Store(100)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				hits := metrics.CacheHits.Load()
				misses := metrics.CacheMisses.Load()
				total := hits + misses
				if total > 0 {
					hitRate := float64(hits) / float64(total) * 100
					_ = hitRate // Evitar la optimización del compilador
				}
			}
		})
	})
}
