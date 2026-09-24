# AGENTS.md

## Project
Web app to configure, stream and record data from an IMST sentire sR-1200
24 GHz radar over Ethernet, using **Automatic Operation mode only**.

- Backend (Go) speaks the radar's binary TCP protocol. The browser never talks to the radar.
- Frontend (React + TypeScript) talks only to the Go backend (REST + WebSocket).

Read `docs/protocol.md` before touching `backend/internal/protocol` or
`backend/internal/radar`. The original manual is `docs/sR1200_Command_Interface.pdf`,
but its text extraction is lossy (some formulas are garbled). Prefer `docs/protocol.md`
and flag any conflict you notice instead of silently picking one.

The directory layout below is the **target**. The repo starts empty; scaffolding is the first task.

## Scope
Commands implemented, and ONLY these:

| ID     | Name                  | Purpose                                        |
|--------|-----------------------|------------------------------------------------|
| 0x0012 | CMDID_SEND_INFO       | Read hardware/firmware info (also the connect handshake) |
| 0x0029 | CMDID_SEND_PARAMS     | Read system parameters                         |
| 0x0028 | CMDID_SETUP           | Write system parameters (EEPROM, restarts radar config) |
| 0x0020 | CMDID_OP_SETUP        | Write operation parameters (enables automatic mode, EEPROM) |

Features:
- Connect / disconnect
- View and edit system parameters and measurement mode (FMCW up-ramp, FMCW down-ramp, CW)
- Start / stop live streaming (Automatic Operation mode)
- Start / stop recording streamed data to disk

Out of scope. Do not implement, stub, or "prepare for":
SPI/CAN, request/response measurement commands (0x0040-0x004A), Human Tracker,
Range Finder, autoMeasType bits 7-8, the radar's HTTP server/SSE streams,
interface-parameter commands (0x0030-0x0032), self-test, software extension (0x00FF),
firmware update.

Ask before adding (small, likely useful): 0x0021 (read back operation params),
0x0000 (ping/keepalive), 0x0002 (set radar clock).

## Target layout
```
/
├── AGENTS.md
├── docs/{protocol.md, sR1200_Command_Interface.pdf}
├── backend/
│   ├── cmd/server/
│   └── internal/
│       ├── protocol/    # pure codec: constants, payload encode/decode, stream frame decoder, unit conversions
│       ├── radar/       # CommandClient, StreamClient, Session state machine
│       ├── recorder/    # writes recordings to disk
│       ├── simulator/   # fake radar (command port + stream port) for tests and UI dev
│       └── api/         # REST + WebSocket handlers
└── frontend/src/{api,components,hooks,types}
```

## Stack and commands
- Backend: Go 1.22+, stdlib `net` and `net/http`; ask before adding any dependency (incl. WebSocket lib).
- Frontend: React 18, TypeScript (`strict`), Vite, uPlot for live charts, Vitest + React Testing Library.
- Backend: `cd backend && go run ./cmd/server [--simulate]` | `go test ./... -race` | `go vet ./...`
- Frontend: `cd frontend && npm run dev` | `npm test` | `npm run typecheck` | `npm run lint`
(These work once scaffolded.)

## Architecture rules
- Two kinds of TCP connection to the radar:
  1. **Command connection** (default port 1024): the four commands above.
  2. **Stream connection** (fixed port 47475; 47475-47480 exist): receive-only frames.
     The radar only measures while at least one stream socket is connected.
- `Session` (in `internal/radar`) owns the state machine
  `disconnected -> connected -> streaming`. `recording` is a flag that is only valid while streaming.
- `internal/protocol` is pure: bytes in, structs out. No sockets, no clocks, no globals.
- One command in flight on the command connection (mutex or request channel).
  Every command has a timeout and returns a typed error.
- Frame pipeline: stream reader -> decoder -> hub -> {WebSocket clients, recorder}.
  Slow consumers drop frames. They must never block the reader or the recorder.
- HTTP handlers call `radar.Session` methods; they never touch `net.Conn`.
- Convert raw units to SI in the backend. Send raw and converted values where debugging matters.

## Protocol rules (easy to get wrong; details in docs/protocol.md)
- All multi-byte values are **little-endian**. No padding of single bytes on Ethernet.
- Command framing: send Command ID (u16). The radar **echoes** it before executing. Verify the echo,
  then send/read the payload. Nothing is sent back after a payload-only command (0x0028, 0x0020).
- Payloads have no length prefix. Read exact sizes with `io.ReadFull`.
- Stream frames start with `AA AA 55 55` + autoMeasType (u16), then blocks selected by the bits
  of autoMeasType. Parse every frame using **its own** autoMeasType, not cached settings.
- Sizes depend on data: FD block payload = channels x samples x (4 or 8 bytes); the second
  Int32 is omitted when FftDataType = 0.
- Always **read-modify-write** system parameters: 0x0029 -> change fields -> 0x0028 -> 0x0029 read-back.
  Never send `AdvancedEn` other than the value read from the device (internal-use flag).
- RangeBin/DopplerBin come from the radar (0x0029 or stream header). Do not compute them from theory.
- Mode mapping (autoMeasType bits 0-1): `fmcw_up`=0b00, `fmcw_down`=0b10, `cw`=0b01.
  FD x-axis is **range** in FMCW and **velocity** in CW.
- Angles are IQ25 fixed point: `rad = raw / 2^25`.

## Behaviour rules
- **Connect**: dial command port -> 0x0012 -> 0x0029. Change nothing on the device.
- **Apply settings** (explicit button only, never on keystroke/slider change; >= 1 s between applies):
  refuse if recording; stop stream; 0x0029 read; merge; 0x0028; 0x0029 read-back;
  0x0020 (opMode=1, autoMeasType, autoDelay); restart stream if it was running.
- **Start stream**: dial stream port. If refused, return `AUTO_MODE_NOT_ENABLED`
  and let the UI offer "Apply settings".
- **Stop stream**: close the stream socket (the radar stops measuring). Do NOT send opMode=0.
- **Disconnect**: stop recording (flush + close file) -> stop stream -> close command socket.
  Leave the radar in automatic mode unless the user ticked "disable automatic mode on disconnect".
  (opMode is stored in EEPROM and survives power cycles.)
- Command timeout must exceed one measurement cycle (about autoDelay + ramp time): the radar
  serves commands between cycles while streaming. Start with max(2 s, 2 x (autoDelay + TRamp)).
- The app always sets autoMeasType bit 4 (header) and bit 2 (system time) so streams are self-describing.
  Bits 3, 5, 6 are user-selectable. Bits 7-15 are always 0.
- **Regulatory guard**: EU license-free band is 24.000-24.250 GHz. Bands 5-13, or any
  MinFreq + ManualBW > 24250 MHz, show a warning and need explicit confirmation. Default is band <= 4.

## Recording (proposed format; confirm before changing)
- Directory `recordings/<UTC-timestamp>/` containing:
  - `meta.json`: app version, decoded 0x0012 info, decoded 0x0029 params, op params, start/stop times, frame count
  - `frames.bin`: repeated records of `u64 LE host receive time (unix ns) | u32 LE length | raw frame bytes (from sync word)`
- Store raw frames exactly as received so data can be re-decoded later.
- Parameter editing is locked while recording (meta.json describes the whole session).
- If the stream drops, finalize the file cleanly and surface an error. Never leave a half-written record.

## Conventions
Go:
- Return `error`, wrap with `%w`, no panics in request paths.
- `context.Context` first parameter on anything doing I/O; use `SetReadDeadline` on sockets.
- Table-driven tests.
TypeScript:
- No `any`. Validate API/WebSocket messages at the boundary (e.g. zod).
- Fetching and socket handling live in `src/hooks`; components stay presentational.
- Render FD/TD arrays with canvas charts (uPlot), never as DOM nodes. Cap UI update rate (~20 fps).

## Backend API (proposed; ask before changing)
`POST /api/connect` `POST /api/disconnect` `GET /api/status` `GET /api/info`
`GET|PUT /api/params` (system params + mode + autoDelay + data selection)
`POST /api/stream/start|stop` `POST /api/record/start|stop` `GET /api/recordings`
`WS /ws`: JSON frames decimated for the UI; the recorder always gets every frame.

## Testing
- Golden byte fixtures in `internal/protocol/testdata/` (worked examples are in docs/protocol.md).
- Stream decoder tests: frame split across TCP reads, garbage before sync, truncated frame,
  every autoMeasType combination, FftDataType 0 vs 1/2/3, plus a Go fuzz test.
- `internal/simulator` implements both the command port and the stream port (FMCW and CW data).
  Integration tests run the real client against it. CI never needs hardware.

## Boundaries
- Ask before: adding dependencies, changing the JSON API or recording format, adding a radar command.
- Never: send 0x0028 or 0x0020 from automated tests against real hardware (EEPROM writes),
  commit radar IPs/serial numbers, add features from the out-of-scope list.

## Definition of done
`go test ./... -race`, `go vet`, `npm run typecheck`, and lint are clean;
`docs/protocol.md` and the simulator are updated if protocol behaviour changed.

## Build order (vertical slices; commit after each)
1. Scaffold + `protocol` codec for 0x0012 / 0x0029 / 0x0028 / 0x0020 payloads (+ golden tests)
2. Stream frame decoder (+ fuzz test)
3. Simulator (command port + stream port)
4. `CommandClient` and `StreamClient`
5. `Session` state machine + REST
6. WebSocket fan-out + recorder
7. Frontend: connect -> info -> params -> live chart -> record

## Verify on hardware early
Payload sizes assume unpadded Ethernet fields (info 38 B, set params 17 B, get params 29 B,
stream header block 15 B). Frame block order is assumed to be ascending by bit.
See "Unverified / ambiguous" in docs/protocol.md. When real captures disagree with the docs,
update the docs and the simulator in the same change.
