# Weather Data Aggregator Service

This is a minimal implementation of the Golang Developer Test Task: a service that periodically fetches weather data from multiple public APIs, aggregates it, and exposes a REST API using Fiber v2.

Quick start

1. Copy `.env.example` to `.env` and set `WEATHER_API_KEY` if you want OpenWeatherMap integration.
2. Run the server:

```powershell
go run .
```

Environment variables (.env.example)

See `.env.example` for defaults.

API Endpoints

- `GET /api/v1/weather/current?city={city}` — returns aggregated current weather for a city.
- `GET /api/v1/weather/forecast?city={city}&days={1-7}` — returns a simple forecast proxy using recent history.
- `GET /api/v1/health` — returns last successful fetch times per city.

Examples

```powershell
curl "http://localhost:3000/api/v1/weather/current?city=Prague"
curl "http://localhost:3000/api/v1/weather/forecast?city=London&days=3"
curl "http://localhost:3000/api/v1/health"
```

Notes

- Uses Open-Meteo (no API key required) and OpenWeatherMap (requires `WEATHER_API_KEY`).
- Scheduling interval is configurable via `FETCH_INTERVAL` (default `15m`).
- Data is stored in-memory; no persistence.


https://home.openweathermap.org/api_keys