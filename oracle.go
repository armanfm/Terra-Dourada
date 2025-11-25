package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "time"
)

// ======== MODELO GENÉRICO DE RESPOSTA ========
type Response struct {
    Status    string      `json:"status"`
    Timestamp string      `json:"timestamp"`
    Data      interface{} `json:"data"`
}

// ======== MODELO UNIVERSAL PARA PROVAS ========
type OraclePayload struct {
    Event   string      `json:"event"`
    Payload interface{} `json:"payload"`
    Device  string      `json:"device"`
}

// ========== CONFIGURAÇÃO ==========
var externalURL = os.Getenv("PRIVATE_ORACLE_URL") // seu prover interno
var mqttBroker = os.Getenv("MQTT_BROKER")
var chainRPC = os.Getenv("BLOCKCHAIN_RPC")
var ipfsURL = os.Getenv("IPFS_URL")

// ===============================================================
// 🔥 INICIO DO SERVIDOR
// ===============================================================
func main() {

    if externalURL == "" {
        log.Println("⚠️  PRIVATE_ORACLE_URL não definido. Usando modo local.")
    }

    mux := http.NewServeMux()

    // === ENDPOINT UNIVERSAL (UI → Backend → ZK-Prover) ===
    mux.HandleFunc("/oracle", handleOracle)

    // === IOT ===
    mux.HandleFunc("/iot/wifi", handleIOTWifi)
    mux.HandleFunc("/iot/ble", handleIOTBle)
    mux.HandleFunc("/iot/serial", handleIOTSerial)
    mux.HandleFunc("/iot/mqtt", handleIOTMQTT)
    mux.HandleFunc("/iot/lora", handleIOTLora)
    mux.HandleFunc("/iot/gps", handleGPS)
    mux.HandleFunc("/iot/zigbee", handleZigbee)

    // === DADOS EXTERNOS ===
    mux.HandleFunc("/price", handlePrice)
    mux.HandleFunc("/weather", handleWeather)
    mux.HandleFunc("/chain", handleChainRPC)
    mux.HandleFunc("/ipfs/upload", handleIPFSUpload)
    mux.HandleFunc("/ipfs/cat", handleIPFSCat)

    // === OUTROS ===
    mux.HandleFunc("/manual", handleManualInput)
    mux.HandleFunc("/test", handleTest)
    mux.HandleFunc("/health", handleHealth)

    log.Println("🔥 ORÁCULO TERRA DOURADA — BACKEND ATIVO EM :8080")
    http.ListenAndServe(":8080", allowCORS(mux))
}

// ===============================================================
// 🔥 UNIVERSAL — recebe o pacote leve da UI e envia ao backend privado
// ===============================================================
func handleOracle(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Método inválido", 405)
        return
    }

    body, _ := io.ReadAll(r.Body)

    if externalURL == "" {
        // modo local (não envia para lugar nenhum)
        respond(w, "ok-local", string(body))
        return
    }

    resp, err := http.Post(externalURL+"/mel", "application/json", bytes.NewReader(body))
    if err != nil {
        http.Error(w, "Erro ao enviar ao provedor privado", 502)
        return
    }

    defer resp.Body.Close()
    dados, _ := io.ReadAll(resp.Body)

    w.Header().Set("Content-Type", "application/json")
    w.Write(dados)
}

// ===============================================================
// 🔥 IOT — WiFi HTTP POST (sensor manda direto)
// ===============================================================
func handleIOTWifi(w http.ResponseWriter, r *http.Request) {
    var payload map[string]interface{}
    json.NewDecoder(r.Body).Decode(&payload)

    respond(w, "iot-wifi", payload)
}

// ===============================================================
// 🔥 IOT — BLE (site lê, backend só recebe json)
// ===============================================================
func handleIOTBle(w http.ResponseWriter, r *http.Request) {
    var data map[string]interface{}
    json.NewDecoder(r.Body).Decode(&data)

    respond(w, "iot-ble", data)
}

// ===============================================================
// 🔥 IOT — SERIAL / USB
// ===============================================================
func handleIOTSerial(w http.ResponseWriter, r *http.Request) {
    var data map[string]interface{}
    json.NewDecoder(r.Body).Decode(&data)

    respond(w, "iot-serial", data)
}

// ===============================================================
// 🔥 IOT — MQTT
// ===============================================================
func handleIOTMQTT(w http.ResponseWriter, r *http.Request) {
    var incoming map[string]interface{}
    json.NewDecoder(r.Body).Decode(&incoming)

    incoming["broker"] = mqttBroker

    respond(w, "iot-mqtt", incoming)
}

// ===============================================================
// 🔥 IOT — LoRa / LoRaWAN
// ===============================================================
func handleIOTLora(w http.ResponseWriter, r *http.Request) {
    var incoming map[string]interface{}
    json.NewDecoder(r.Body).Decode(&incoming)

    incoming["lora_gateway"] = "LOCAL-GATEWAY"

    respond(w, "iot-lora", incoming)
}

// ===============================================================
// 🔥 IOT — GPS
// ===============================================================
func handleGPS(w http.ResponseWriter, r *http.Request) {
    var gps map[string]interface{}
    json.NewDecoder(r.Body).Decode(&gps)

    respond(w, "iot-gps", gps)
}

// ===============================================================
// 🔥 IOT — Zigbee
// ===============================================================
func handleZigbee(w http.ResponseWriter, r *http.Request) {
    var data map[string]interface{}
    json.NewDecoder(r.Body).Decode(&data)

    respond(w, "iot-zigbee", data)
}

// ===============================================================
// 🔥 PRICE FETCHER (cripto, dólar, commodity, etc.)
// ===============================================================
func handlePrice(w http.ResponseWriter, r *http.Request) {
    moeda := r.URL.Query().Get("symbol")
    if moeda == "" {
        moeda = "BTC"
    }

    api := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sUSDT", moeda)
    resp, _ := http.Get(api)
    defer resp.Body.Close()

    corpo, _ := io.ReadAll(resp.Body)

    w.Header().Set("Content-Type", "application/json")
    w.Write(corpo)
}

// ===============================================================
// 🔥 WEATHER (previsão do tempo)
// ===============================================================
func handleWeather(w http.ResponseWriter, r *http.Request) {
    cidade := r.URL.Query().Get("city")
    if cidade == "" {
        cidade = "São Paulo"
    }

    respond(w, "weather", map[string]string{
        "cidade": cidade,
        "status": "sol",
        "temp":   "28",
    })
}

// ===============================================================
// 🔥 BLOCKCHAIN RPC
// ===============================================================
func handleChainRPC(w http.ResponseWriter, r *http.Request) {
    var req interface{}
    json.NewDecoder(r.Body).Decode(&req)

    b, _ := json.Marshal(req)
    resp, _ := http.Post(chainRPC, "application/json", bytes.NewReader(b))
    defer resp.Body.Close()

    dados, _ := io.ReadAll(resp.Body)
    w.Write(dados)
}

// ===============================================================
// 🔥 IPFS UPLOAD
// ===============================================================
func handleIPFSUpload(w http.ResponseWriter, r *http.Request) {
    var dado map[string]interface{}
    json.NewDecoder(r.Body).Decode(&dado)

    respond(w, "ipfs-upload", dado)
}

// ===============================================================
// 🔥 IPFS CAT
// ===============================================================
func handleIPFSCat(w http.ResponseWriter, r *http.Request) {
    cid := r.URL.Query().Get("cid")
    respond(w, "ipfs-cat", map[string]string{"cid": cid})
}

// ===============================================================
// 🔥 MANUAL INPUT
// ===============================================================
func handleManualInput(w http.ResponseWriter, r *http.Request) {
    var dado map[string]interface{}
    json.NewDecoder(r.Body).Decode(&dado)

    respond(w, "manual", dado)
}

// ===============================================================
// 🔥 MISC
// ===============================================================
func handleTest(w http.ResponseWriter, r *http.Request) {
    respond(w, "test", "ok")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
    respond(w, "health", "alive")
}

// ===============================================================
// 🔧 RESPOSTA UNIVERSAL
// ===============================================================
func respond(w http.ResponseWriter, status string, data interface{}) {
    resposta := Response{
        Status:    status,
        Timestamp: time.Now().Format(time.RFC3339),
        Data:      data,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resposta)
}

// ===============================================================
// 🔧 CORS
// ===============================================================
func allowCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        if r.Method == "OPTIONS" {
            return
        }
        next.ServeHTTP(w, r)
    })
}


