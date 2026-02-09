package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// --- DATENBANK STRUKTUREN ---
type ScanResult struct {
	ID        string
	Time      string
	Username  string
	SteamID   string // Simuliert
	RiskLevel string // "Clean", "Suspicious", "Detected"
	Detections []string
}

type CustomRule struct {
	Name      string
	Type      string // "String", "Hash", "File"
	Value     string
}

// Einfacher In-Memory Speicher (Speichert Daten solange der Server läuft)
var (
	GlobalStrings   = []string{"vape", "killaura", "autoclicker"}
	GlobalRules     = []CustomRule{}
	GlobalScans     = []ScanResult{}
	DataMutex       sync.Mutex // Damit beim gleichzeitigen Schreiben nichts abstürzt
)

// --- DER SCANNER CODE (Wird in die EXE kompiliert) ---
const scannerCode = `
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"strings"
	"time"
	"net/http"
)

// Konfig
var ServerURL string = "ERSETZE_MICH" // Wird beim Bauen überschrieben
var BuildStrings string 

// Design Konstanten
const (
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorWhite  = "\033[97m"
	ColorGray   = "\033[90m"
	Reset       = "\033[0m"
)

type ReportPayload struct {
	Username   string   ` + "`json:\"username\"`" + `
	Detections []string ` + "`json:\"detections\"`" + `
}

func main() {
	// 1. UI Initialisierung
	fmt.Print("\033[H\033[2J") // Clear Screen
	fmt.Println(ColorRed + ` + "`" + `
██████╗ ███████╗██████╗ ███████╗ ██████╗ ███╗   ██╗███████╗
██╔══██╗██╔════╝██╔══██╗╚══███╔╝██╔═══██╗████╗  ██║██╔════╝
██████╔╝█████╗  ██║  ██║  ███╔╝ ██║   ██║██╔██╗ ██║█████╗  
██╔══██╗██╔════╝██║  ██║ ███╔╝  ██║   ██║██║╚██╗██║██╔════╝
██║  ██║███████╗██████╔╝███████╗╚██████╔╝██║ ╚████║███████╗
╚═╝  ╚═╝╚══════╝╚═════╝ ╚══════╝ ╚═════╝ ╚═╝  ╚═══╝╚══════╝
` + "`" + ` + Reset)
	
	fmt.Println(ColorGray + "\nConnecting to REDZONE Cloud..." + Reset)
	time.Sleep(1 * time.Second)

	currentUser, _ := user.Current()
	username := "Unknown"
	if currentUser != nil {
		username = currentUser.Username
	}

	fmt.Printf("Scanning Target: " + ColorWhite + "%s\n" + Reset, username)
	fmt.Println(ColorGray + "--------------------------------------------------" + Reset)

	// 2. Der Scan (Simuliert anhand der Strings)
	targetStrings := strings.Split(BuildStrings, ",")
	foundThreats := []string{}

	// Echter Dateiscan (Hier vereinfacht auf lokalen Ordner für Demo)
	files, _ := os.ReadDir("./")
	for _, f := range files {
		for _, s := range targetStrings {
			if s != "" && strings.Contains(strings.ToLower(f.Name()), strings.ToLower(s)) {
				fmt.Printf(ColorRed + "[!] DETECTED: " + ColorWhite + "%s " + ColorGray + "(Sig: %s)\n" + Reset, f.Name(), s)
				foundThreats = append(foundThreats, f.Name() + " (" + s + ")")
			}
		}
	}

	// 3. Ergebnis Senden
	fmt.Println(ColorGray + "\nUploading results to dashboard..." + Reset)
	
	payload := ReportPayload{
		Username:   username,
		Detections: foundThreats,
	}
	jsonData, _ := json.Marshal(payload)

	// Sende POST Request an deine Website
	_, err := http.Post(ServerURL + "/api/report", "application/json", bytes.NewBuffer(jsonData))
	
	fmt.Println(ColorGray + "--------------------------------------------------" + Reset)
	if err != nil {
		fmt.Println(ColorRed + "Upload Failed. Server offline?" + Reset)
	} else {
		fmt.Println(ColorGreen + "UPLOAD SUCCESSFUL." + Reset)
		fmt.Println("Check the 'Scans' tab on the website.")
	}
	
	if len(foundThreats) > 0 {
		fmt.Println("\n" + ColorRed + "RESULT: FAILED (Threats Found)" + Reset)
	} else {
		fmt.Println("\n" + ColorGreen + "RESULT: CLEAN" + Reset)
	}

	fmt.Println("\nPress ENTER to exit.")
	fmt.Scanln()
}
`

// --- HTML TEMPLATE (Frontend) ---
const htmlTemplate = `
<!DOCTYPE html>
<html lang="de" class="h-full bg-zinc-950">
<head>
    <meta charset="UTF-8">
    <title>REDZONE | Echo Clone</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script>
        tailwind.config = {
            theme: { extend: { colors: { zinc: { 900: '#18181b', 950: '#09090b' }, red: { 600: '#dc2626' } } } }
        }
    </script>
    <style>
        .active-tab { background-color: rgba(220, 38, 38, 0.1); color: #dc2626; border-right: 2px solid #dc2626; }
        ::-webkit-scrollbar { width: 8px; }
        ::-webkit-scrollbar-track { background: #09090b; }
        ::-webkit-scrollbar-thumb { background: #333; border-radius: 4px; }
    </style>
</head>
<body class="h-full flex font-sans text-zinc-400 overflow-hidden">

    <div class="w-64 bg-zinc-900 flex flex-col border-r border-zinc-800 flex-shrink-0">
        <div class="h-16 flex items-center px-6 border-b border-zinc-800">
            <h1 class="text-2xl font-black italic tracking-widest text-white">RED<span class="text-red-600">ZONE</span></h1>
        </div>
        
        <nav class="flex-1 py-6 space-y-1">
            <button onclick="showTab('scans')" id="btn-scans" class="w-full flex items-center px-6 py-3 hover:bg-zinc-800 transition text-left active-tab">
                <svg class="w-5 h-5 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/></svg>
                Scans
            </button>
            <button onclick="showTab('strings')" id="btn-strings" class="w-full flex items-center px-6 py-3 hover:bg-zinc-800 transition text-left">
                <svg class="w-5 h-5 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/></svg>
                String Builder
            </button>
            <button onclick="showTab('custom')" id="btn-custom" class="w-full flex items-center px-6 py-3 hover:bg-zinc-800 transition text-left">
                <svg class="w-5 h-5 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.384-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"/></svg>
                Custom Detections
            </button>
        </nav>

        <div class="p-4 border-t border-zinc-800">
            <form action="/download" method="POST">
                <button class="w-full bg-red-600 hover:bg-red-700 text-white font-bold py-2 rounded shadow-lg shadow-red-900/20 transition">
                    DOWNLOAD EXE
                </button>
            </form>
        </div>
    </div>

    <div class="flex-1 overflow-y-auto bg-zinc-950 p-8 relative">
        
        <div id="tab-scans" class="space-y-6">
            <h2 class="text-xl font-bold text-white mb-4">Neueste Scans</h2>
            
            <div class="overflow-hidden rounded-lg border border-zinc-800">
                <table class="min-w-full bg-zinc-900">
                    <thead>
                        <tr class="border-b border-zinc-800">
                            <th class="px-6 py-3 text-left text-xs font-medium text-zinc-500 uppercase tracking-wider">User</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-zinc-500 uppercase tracking-wider">Ergebnis</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-zinc-500 uppercase tracking-wider">Detections</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-zinc-500 uppercase tracking-wider">Zeit</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-zinc-800 text-zinc-300">
                        {{range .Scans}}
                        <tr class="hover:bg-zinc-800/50 transition">
                            <td class="px-6 py-4 whitespace-nowrap flex items-center">
                                <div class="h-8 w-8 rounded bg-zinc-700 flex items-center justify-center text-white font-bold mr-3">{{slice .Username 0 1}}</div>
                                {{.Username}}
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap">
                                {{if eq .RiskLevel "Clean"}}
                                    <span class="px-2 py-1 text-xs rounded bg-green-900/30 text-green-500 border border-green-900">Clean</span>
                                {{else}}
                                    <span class="px-2 py-1 text-xs rounded bg-red-900/30 text-red-500 border border-red-900">DETECTED</span>
                                {{end}}
                            </td>
                            <td class="px-6 py-4">
                                {{if .Detections}}
                                    <span class="text-red-400 text-sm font-mono">{{.Detections}}</span>
                                {{else}}
                                    <span class="text-zinc-600 text-sm">-</span>
                                {{end}}
                            </td>
                            <td class="px-6 py-4 text-sm text-zinc-500">{{.Time}}</td>
                        </tr>
                        {{else}}
                        <tr>
                            <td colspan="4" class="px-6 py-12 text-center text-zinc-600">Noch keine Scans vorhanden. Starte die EXE!</td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
        </div>

        <div id="tab-strings" class="hidden space-y-6">
            <h2 class="text-xl font-bold text-white mb-4">String Manager</h2>
            <div class="bg-zinc-900 p-6 rounded-lg border border-zinc-800">
                <form action="/add_string" method="POST" class="flex gap-4">
                    <input type="text" name="new_string" placeholder="Cheat String eingeben (z.B. vape_v4)" class="flex-1 bg-zinc-950 border border-zinc-700 rounded px-4 py-2 text-white focus:border-red-600 focus:outline-none">
                    <button class="bg-zinc-800 hover:bg-zinc-700 text-white px-6 py-2 rounded border border-zinc-700">+ Hinzufügen</button>
                </form>
                
                <div class="mt-8">
                    <h3 class="text-sm font-semibold text-zinc-500 uppercase mb-4">Aktive Strings</h3>
                    <div class="flex flex-wrap gap-2">
                        {{range .Strings}}
                        <div class="bg-zinc-950 border border-zinc-700 px-3 py-1 rounded text-sm text-zinc-300 flex items-center">
                            {{.}}
                        </div>
                        {{end}}
                    </div>
                </div>
            </div>
        </div>

        <div id="tab-custom" class="hidden space-y-6">
             <div class="flex justify-between items-center">
                <h2 class="text-xl font-bold text-white">Advanced Detections</h2>
                <button class="bg-red-600 text-white px-4 py-2 rounded text-sm hover:bg-red-700">+ Create New</button>
             </div>

             <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                <div class="bg-zinc-900 rounded-xl border border-zinc-800 overflow-hidden group hover:border-red-600/50 transition">
                    <div class="h-32 bg-zinc-950 relative flex items-center justify-center">
                        <div class="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/carbon-fibre.png')] opacity-20"></div>
                        <h3 class="text-xl font-bold text-white relative z-10">SSTB</h3>
                    </div>
                    <div class="p-4">
                        <div class="flex items-center justify-between mb-4">
                            <span class="text-xs font-mono text-zinc-500">HASH SCAN</span>
                            <div class="flex space-x-2">
                                <div class="w-2 h-2 rounded-full bg-green-500"></div>
                                <span class="text-xs text-green-500">Active</span>
                            </div>
                        </div>
                        <div class="bg-zinc-950 p-3 rounded border border-zinc-800 font-mono text-xs text-zinc-400 overflow-hidden whitespace-nowrap">
                            md5: 4d833a1388...
                        </div>
                    </div>
                </div>
                
                <div class="bg-zinc-900 rounded-xl border border-zinc-800 overflow-hidden group hover:border-red-600/50 transition">
                    <div class="h-32 bg-zinc-950 relative flex items-center justify-center">
                         <div class="absolute inset-0 bg-red-900/10"></div>
                        <h3 class="text-xl font-bold text-white relative z-10">XRC BYPASS</h3>
                    </div>
                    <div class="p-4">
                         <div class="flex items-center justify-between mb-4">
                            <span class="text-xs font-mono text-zinc-500">MEMORY SCAN</span>
                             <div class="flex space-x-2">
                                <div class="w-2 h-2 rounded-full bg-green-500"></div>
                                <span class="text-xs text-green-500">Active</span>
                            </div>
                        </div>
                        <div class="bg-zinc-950 p-3 rounded border border-zinc-800 font-mono text-xs text-zinc-400">
                            Strings: xrc_main, hook_v2
                        </div>
                    </div>
                </div>
             </div>
        </div>

    </div>

    <script>
        function showTab(name) {
            ['scans', 'strings', 'custom'].forEach(t => {
                document.getElementById('tab-' + t).classList.add('hidden');
                document.getElementById('btn-' + t).classList.remove('active-tab');
            });
            document.getElementById('tab-' + name).classList.remove('hidden');
            document.getElementById('btn-' + name).classList.add('active-tab');
        }
    </script>
</body>
</html>
`

// --- SERVER LOGIK ---

func main() {
	// Erstelle das Scanner Template initial
	ioutil.WriteFile("scanner_template.go", []byte(scannerCode), 0644)

	// ROUTE: Dashboard (Hauptseite)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		DataMutex.Lock()
		data := struct {
			Strings []string
			Scans   []ScanResult
		}{
			Strings: GlobalStrings,
			Scans:   GlobalScans, // Wir zeigen die Liste umgekehrt (neueste oben) könnte man noch machen
		}
		DataMutex.Unlock()

		tmpl, _ := template.New("index").Parse(htmlTemplate)
		tmpl.Execute(w, data)
	})

	// ROUTE: String hinzufügen
	http.HandleFunc("/add_string", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			newStr := r.FormValue("new_string")
			if newStr != "" {
				DataMutex.Lock()
				GlobalStrings = append(GlobalStrings, newStr)
				DataMutex.Unlock()
			}
		}
		http.Redirect(w, r, "/", 303)
	})

	// ROUTE: API für die EXE (Report empfangen)
	http.HandleFunc("/api/report", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" { return }

		var payload struct {
			Username   string   `json:"username"`
			Detections []string `json:"detections"`
		}

		json.NewDecoder(r.Body).Decode(&payload)

		result := ScanResult{
			ID:        fmt.Sprintf("%d", time.Now().Unix()),
			Time:      time.Now().Format("15:04:05"),
			Username:  payload.Username,
			RiskLevel: "Clean",
			Detections: payload.Detections,
		}

		if len(payload.Detections) > 0 {
			result.RiskLevel = "DETECTED"
		}

		DataMutex.Lock()
		// Füge neuen Scan vorne an (damit er oben in der Liste ist)
		GlobalScans = append([]ScanResult{result}, GlobalScans...)
		DataMutex.Unlock()

		fmt.Printf("[SERVER] Neuer Report von %s erhalten. Detections: %d\n", payload.Username, len(payload.Detections))
	})

	// ROUTE: EXE Downloaden
	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		// Ermittle die URL der aktuellen Website (für den Client)
		// Bei Render.com nutzen wir die Environment Variable, lokal localhost
		host := "https://" + r.Host
		if strings.Contains(r.Host, "localhost") {
			host = "http://" + r.Host
		}

		DataMutex.Lock()
		currentStrings := strings.Join(GlobalStrings, ",")
		DataMutex.Unlock()

		// Compiler Befehl: Wir brennen die URL und die Strings in die EXE ein
		// WICHTIG: -X main.ServerURL=... sorgt dafür, dass die EXE weiß, wo sie hinmelden muss
		cmd := exec.Command("go", "build", 
			"-ldflags", fmt.Sprintf("-X main.ServerURL=%s -X main.BuildStrings=%s", host, currentStrings), 
			"-o", "redzone_client.exe", "scanner_template.go")
		
		cmd.Env = append(os.Environ(), "GOOS=windows", "GOARCH=amd64")
		output, err := cmd.CombinedOutput()

		if err != nil {
			http.Error(w, "Build Error: "+string(output), 500)
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename=redzone_client.exe")
		http.ServeFile(w, r, "redzone_client.exe")
	})

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	fmt.Println("REDZONE C2 Server läuft auf Port " + port)
	http.ListenAndServe(":"+port, nil)
}
