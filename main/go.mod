module RelayServer/main

go 1.26.1

require (
	RelayServer/relayConfig v0.0.0-00010101000000-000000000000
	RelayServer/relayDB v0.0.0-00010101000000-000000000000
	RelayServer/routes v0.0.0-00010101000000-000000000000
	github.com/gorilla/websocket v1.5.3
)

require (
	github.com/caarlos0/env/v11 v11.4.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/exp v0.0.0-20251023183803-a4bb9ffd2546 // indirect
	golang.org/x/sys v0.37.0 // indirect
	modernc.org/libc v1.67.6 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	modernc.org/sqlite v1.46.1 // indirect
)

replace RelayServer/routes => ../routes/

replace RelayServer/relayDB => ../relayDB

replace RelayServer/relayConfig => ../relayConfig/
