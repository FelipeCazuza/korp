package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const serverAddress = ":8080"

// httpRequestsTotal contabiliza as requisições processadas pelo endpoint.
var httpRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "projeto_korp",
		Name:      "http_requests_total",
		Help:      "Total de requisições HTTP processadas pelo serviço",
	},
	[]string{"method", "path", "status"},
)

// O Projeto korp response repesenta o JSON retornado pelo endpoint.
type ProjetoKorpResponse struct {
	Nome    string `json:"nome"`
	Horario string `json:"horario"`
}

// statusRecorder registra o codigo HTTP retornado pelo handler.
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (recorder *statusRecorder) WriteHeader(statusCode int) {
	if recorder.statusCode != 0 {
		return
	}
	recorder.statusCode = statusCode
	recorder.ResponseWriter.WriteHeader(statusCode)
}
func (recorder *statusRecorder) Write(body []byte) (int, error) {
	if recorder.statusCode == 0 {
		recorder.WriteHeader(http.StatusOK)
	}
	return recorder.ResponseWriter.Write(body)
}

// metricsMiddleware é um middleware que registra métricas de requisições HTTP.
func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{
			ResponseWriter: w,
		}

		next.ServeHTTP(recorder, r)

		httpRequestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			strconv.Itoa(recorder.statusCode),
		).Inc()
	})
}

// writeJSON centraliza a escrita de resposta do  JSON.
func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Erro ao escrever resposta JSON: %v", err)
	}
}

// projetoKorpHandler atende as requisições HTTP para o endpoint /projeto-korp.
func projetoKorpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)

		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Método não permitido"})

		return
	}

	response := ProjetoKorpResponse{
		Nome:    "Projeto Korp",
		Horario: time.Now().UTC().Format(time.RFC3339),
	}
	writeJSON(w, http.StatusOK, response)
}

func main() {
	// Registra a métrica personalizada no catálogo padrão do Prometheus.
	prometheus.MustRegister(httpRequestsTotal)

	mux := http.NewServeMux()

	// O middleware mede as requisições do endpoint principal.
	mux.Handle(
		"/projeto-korp",
		metricsMiddleware(http.HandlerFunc(projetoKorpHandler)),
	)

	// Endpoint utilizado pelo Prometheus para coletar as métricas.
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:              serverAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("http-server-projeto-korp iniciado em http://localhost%s", serverAddress)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
