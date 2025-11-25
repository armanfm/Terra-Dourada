package main

import (
    "io"
    "log"
    "net/http"
    "os"
    "bytes"
)

// Estruturas opcionais (mantidas para clareza)
type OracleRequest struct {
    Event   string `json:"event"`
    Payload string `json:"payload"`
    Device  string `json:"device"`
}

type OracleResponse struct {
    Hash       string `json:"hash"`
    Timestamp  string `json:"timestamp"`
    LedgerLine string `json:"ledger_line"`
    Proof      string `json:"proof"`
    Status     string `json:"status"`
}

func main() {
    // URL do backend privado (NÃO aparece no GitHub)
    backendURL := os.Getenv("PRIVATE_ORACLE_URL")
    if backendURL == "" {
        log.Fatal("Erro: a variável PRIVATE_ORACLE_URL não está definida")
    }

    mux := http.NewServeMux()

    // Endpoint chamado pela UI → repassa para o backend /mel
    mux.HandleFunc("/oracle", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Método inválido", http.StatusMethodNotAllowed)
            return
        }

        // Ler JSON vindo da UI
        body, err := io.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "Erro ao ler body", http.StatusBadRequest)
            return
        }

        // Encaminhar para o backend PRIVADO → /mel
        resp, err := http.Post(
            backendURL+"/mel",
            "application/json",
            bytes.NewReader(body),
        )
        if err != nil {
            http.Error(w, "Erro ao contactar backend privado", http.StatusBadGateway)
            return
        }
        defer resp.Body.Close()

        // Ler resposta do backend
        privateResp, err := io.ReadAll(resp.Body)
        if err != nil {
            http.Error(w, "Erro ao ler resposta backend privado", http.StatusBadGateway)
            return
        }

        // Retornar para a UI
        w.Header().Set("Content-Type", "application/json")
        w.Write(privateResp)
    })

    // CORS
    handler := corsMiddleware(mux)

    log.Println("🌐 Gateway Go rodando em http://localhost:7070")
    http.ListenAndServe(":7070", handler)
}

// Middleware simples de CORS
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

        if r.Method == http.MethodOptions {
            return
        }

        next.ServeHTTP(w, r)
    })
}

