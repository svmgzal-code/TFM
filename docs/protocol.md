# sR-1200 protocol reference (condensed, app scope only)

Source: IMST sentire sR-1200 Interface Description v1.7 (`sR1200_Command_Interface.pdf`).
Section numbers below refer to that manual. This file covers only what the app implements:
commands 0x0012, 0x0029, 0x0028, 0x0020 and Automatic Operation mode over Ethernet.
Items marked **[VERIFY]** are assumptions or unclear in the manual. See the last section.

## 1. Conventions
- Byte order: **little-endian** (least significant byte first) for every multi-byte value (2.2.3).
- Ethernet transmits single bytes: `u8` is 1 byte, **no padding** (the 2-byte padding rule is for SPI/CAN). **[VERIFY]**
- Types: `u8/u16/u32/u64` unsigned, `i32` signed two's complement.
- IQ25: signed fixed point, `rad = raw / 2^25`, `deg = raw * 180 / (2^25 * pi)` (A.3).
- Time stamps inside blocks are in units of **10 us**.

## 2. Connections
| Purpose            | Host:Port                          | Notes |
|--------------------|------------------------------------|-------|
| Command connection | radar IP : 1024 (default)          | Default 192.168.1.2:1024, gateway 192.168.1.1, mask 255.255.254.0. Defaults are valid only ~10 s after power-up; afterwards the radar uses the user-defined IP/port (2.1.3). |
| Stream connection  | same IP : 47475 (47475-47480 exist) | Only exists when opMode = 1. Same data goes to every connected socket (2.8.1). |

The app must let the user configure host, command port and stream port.
Changing the radar's IP/port is out of scope (commands 0x0030-0x0032).

## 3. Command framing (Ethernet, 2.3.3)
```
Host -> Radar : CommandID (u16)
Radar -> Host : CommandID (u16)        # echo, sent BEFORE the command executes
Host -> Radar : payload (if any)
Radar -> Host : payload (if any)
```
- No trailing handshake. Nothing is sent after a payload-only command, so completion
  of 0x0028 / 0x0020 cannot be observed directly. Confirm with a read-back.
- Payloads have no length prefix; sizes are fixed by the tables below.
- One command in flight at a time.
- In automatic mode the radar serves the command connection once per measurement loop
  (Table 2.54), so replies can be delayed by up to one cycle.

## 4. Commands

### 4.1 0x0012 CMDID_SEND_INFO (2.5.4) - request has no payload; reply 38 bytes
| Off | Field                 | Type | Notes |
|-----|-----------------------|------|-------|
| 0   | FirmwareVersion       | u32  | `0x00040102` = 4.1.2 (major = bits 16-23, minor = 8-15, patch = 0-7) |
| 4   | FirmwareRevision      | u32  | customer-specific release id |
| 8   | SntBoardVersion       | u32  | internal use |
| 12  | BasebandBoardVersion  | u32  | internal use |
| 16  | FrontendBoardVersion  | u32  | internal use |
| 20  | AvailableChannels     | u16  | bit n = receiver channel Rx(n+1) |
| 22  | AvailableAlgorithms   | u16  | bits 0-1 raw data, 4 Human Tracker, 5 Range Finder (informational only) |
| 24  | RadarHardwareID       | u16  | internal use |
| 26  | RadarNumber           | u32  | unique radar identifier |
| 30  | FlashingDate          | u32  | decimal `YYYYMMDD`, e.g. 20150301 |
| 34  | PhaseOffset           | i32  | degrees, for view-angle estimation |

### 4.2 0x0029 CMDID_SEND_PARAMS (2.5.11) - request has no payload; reply 29 bytes
Reply = the 17-byte R/W block of 4.3 followed by the read-only block:
| Off | Field        | Type | Notes |
|-----|--------------|------|-------|
| 17  | RangeBin     | u32  | um, recalculated by the radar after every 0x0028 |
| 21  | DopplerBin   | u32  | um/s |
| 25  | FrequencyBin | u32  | Hz |

Read this back after every 0x0028.

### 4.3 0x0028 CMDID_SETUP (2.5.10) - payload 17 bytes, no reply payload
Writes to EEPROM and restarts the radar with the new settings. Field order follows Table 2.20 **[VERIFY]**:
| Off | Field         | Type | Meaning |
|-----|---------------|------|---------|
| 0   | Band          | u8   | Band index 0-13 (section 5). Ignored if MinFreq and ManualBW are valid. |
| 1   | TRamp         | u8   | Ramp time [ms]; valid range not documented **[VERIFY]** |
| 2   | ZeroPad       | u8   | 1, 2, 4 or 8 |
| 3   | FftDataType   | u8   | 0 magnitude; 1 magnitude+phase; 2 real+imag; 3 magnitude+object angle |
| 4   | FrontendEn    | u8   | 0/1 frontend power |
| 5   | PowerSaveEn   | u8   | 0/1, frontend off between pulses |
| 6   | Norm          | u8   | 0/1 signal normalization |
| 7   | ActChannels   | u16  | bit mask of active receiver channels (bits 0-15) |
| 9   | AdvancedEn    | u8   | internal use. **Echo the value read from the device.** |
| 10  | FreqPoints    | u16  | number of FD samples, valid 5-513 |
| 12  | MinFreq       | u16  | start frequency [MHz]. Out of limits (e.g. 0) means use Band. |
| 14  | ManualBW      | u16  | bandwidth [MHz]. Out of limits (e.g. 0) means use Band. |
| 16  | Attenuation   | u8   | Tx attenuation index 0-7, approx. 0, 0.4, 0.8, 1.4, 2.5, 4, 6, 9 dB (depends on temperature and frequency) |

Frequencies are internally limited to 23-26 GHz, and that range is not guaranteed by the hardware.

### 4.4 0x0020 CMDID_OP_SETUP (2.5.7) - payload 6 bytes, no reply payload
Written to EEPROM. It survives power cycles.
| Off | Field         | Type | Meaning |
|-----|---------------|------|---------|
| 0   | opMode        | u16  | 0 = normal (host-triggered), 1 = automatic (opens the stream sockets) |
| 2   | autoMeasType  | u16  | bit mask, section 6.2 |
| 4   | autoDelay     | u16  | minimum cycle time [ms] |

Reading back is command 0x0021 (out of scope; ask before adding).

## 5. Band table (Table 2.21)
| Band | BW [MHz] | Start [GHz] | Stop [GHz] |
|------|----------|-------------|------------|
| 0    | 50       | 24.000      | 24.050     |
| 1    | 75       | 24.000      | 24.075     |
| 2    | 125      | 24.000      | 24.125     |
| 3    | 230      | 24.000      | 24.230     |
| 4    | 250      | 24.000      | 24.250     |
| 5    | 500      | 24.000      | 24.500     |
| 6    | 800      | 24.000      | 24.800     |
| 7    | 1000     | 24.000      | 25.000     |
| 8    | 1200     | 24.000      | 25.200     |
| 9    | 1500     | 24.000      | 25.500     |
| 10   | 1800     | 24.000      | 25.800     |
| 11   | 2000     | 23.500      | 25.500     |
| 12   | 2200     | 23.500      | 25.700     |
| 13   | 2500     | 23.200      | 25.700     |

Regulatory note (manual, "Important Notices"): license-free use in the EU is 24.000-24.250 GHz,
which means **bands 0-4 only**. The UI must warn on anything above 24.250 GHz.

## 6. Automatic Operation mode (2.8)

### 6.1 How it works
- Ethernet only. Enabled by 0x0020 with opMode = 1; disabled with opMode = 0.
- With opMode = 1 the radar opens six extra TCP sockets (47475-47480) on the user-defined IP.
  The normal command port keeps working.
- The radar measures and transmits **only while at least one stream socket is connected**.
  Each connected socket receives every frame.
- Loop (Table 2.54): measure (bits 0-1) -> send frame to each stream socket -> serve a pending
  command on the command socket -> if elapsed < autoDelay, wait the remainder.
- No requests or acknowledgments on the stream; the host only reads.

### 6.2 autoMeasType bit mask (Table 2.51)
| Bit | Value | Meaning | In scope |
|-----|-------|---------|----------|
| 0   | 1     | 0 = FMCW, 1 = CW | yes |
| 1   | 2     | FMCW only: 0 = UP-ramp, 1 = DOWN-ramp | yes |
| 2   | 4     | system time block | yes (always on) |
| 3   | 8     | temperatures + Tx power block | yes |
| 4   | 16    | header block | yes (always on) |
| 5   | 32    | time-domain data block | yes |
| 6   | 64    | frequency-domain data block | yes |
| 7   | 128   | Range Finder result | **no, always 0** |
| 8   | 256   | Human Tracker result | **no, always 0** |
| 9-15| -     | undefined | always 0 |

Mode mapping used by the app:
| Mode        | bit 0 | bit 1 |
|-------------|-------|-------|
| `fmcw_up`   | 0     | 0     |
| `fmcw_down` | 0     | 1     |
| `cw`        | 1     | 0     |

Examples: FMCW up + time + temps + header + FD = `0x005C` (92). CW + header + FD = `0x0051`.
FMCW down + header + FD = `0x0052`.

### 6.3 Frame layout (2.8.2)
```
AA AA 55 55        SyncWord1 = 0xAAAA, SyncWord2 = 0x5555 (little-endian bytes)
autoMeasType (u16) bit mask of the blocks that follow
[ blocks for bits 2..6, in ascending bit order ]   **[VERIFY order]**
```
Blocks are the same as the command replies, concatenated with no separators.

**Bit 2, system time (8 B)**
| Field  | Type | Notes |
|--------|------|-------|
| Millis | u64  | ms since UNIX epoch only if the clock was synced (0x0002); otherwise ms since power-up |

**Bit 3, temperatures and power (12 B)** (Table 2.23)
| Field    | Type | Notes |
|----------|------|-------|
| NtcTemp  | i32  | frontend temperature [deg C] |
| TxPower  | u32  | transmitted power [dBm/10] |
| ChipTemp | i32  | transceiver chip temperature [deg C] |

**Bit 4, header (15 B)** (Table 2.53)
| Off | Field                | Type | Notes |
|-----|----------------------|------|-------|
| 0   | MinFreq              | u16  | MHz |
| 2   | ManualBW             | u16  | MHz |
| 4   | TRamp                | u16  | ms (note: u8 in system params) |
| 6   | Attenuation          | u8   | index |
| 7   | RangeBin / DopplerBin| u32  | um (FMCW) or um/s (CW) |
| 11  | FreqPoints           | u16  | |
| 13  | ZeroPad              | u8   | |
| 14  | Norm                 | u8   | |

**Bit 5, time-domain block** (Table 2.33), total = 2 + 4096 x N + 8 bytes
| Field | Type | Notes |
|-------|------|-------|
| NumActChan | u16 | N |
| samples | N x 1024 x i32 | channel-major: all 1024 samples of channel 0, then channel 1, ... |
| PreTime | u32 | 10 us units |
| PostTime | u32 | 10 us units |

**Bit 6, frequency-domain block** (Table 2.34), total = 50 + N x S x (4 or 8) bytes
| Field | Type | Notes |
|-------|------|-------|
| FftDatTyp | u8 | 0-3, same meaning as FftDataType |
| PreProcTime | u32 | 10 us |
| PostProcTime | u32 | 10 us |
| DataMin | 4 x i32 | fixed 4 entries: min TD value per channel |
| DataMax | 4 x i32 | fixed 4 entries |
| Overload | u8 | number of first channel with overload (0 = none? **[VERIFY]**) |
| NumActChan | u16 | N |
| FdSamples | u16 | S |
| samples | N x S x { Data1 i32, Data2 i32 } | **Data2 is omitted when FftDatTyp = 0** |
| PostTime | u32 | 10 us |

Data1 / Data2 by FftDatTyp:
| FftDatTyp | Data1 | Data2 |
|-----------|-------|-------|
| 0 | magnitude | not sent |
| 1 | magnitude | phase [rad, IQ25] |
| 2 | real part | imaginary part |
| 3 | magnitude | object (view) angle [rad, IQ25] |

### 6.4 Decoder requirements
- Parse each frame with the frame's own autoMeasType. Everything else needed
  (N, S, FftDatTyp) is inside the blocks, so frames are self-describing.
- Stream decoding is sequential; a TCP read may split a frame anywhere.
- Resync after an error: discard buffered bytes and scan for `AA AA 55 55`. Accept a candidate
  only if bits 7-15 of autoMeasType are 0 and the frame parses to completion. Sync bytes can
  also occur inside sample data, so a sync hit is a candidate, not proof.
- Reject implausible counts (N > 16, S > 513 or < 5) before allocating.

## 7. Derived quantities
- **Range axis (FMCW)**: `d[n] = (n-1) * RangeBin / (ZeroPad * 1e6)` metres, n = 1..FreqPoints.
  Theoretical range bin `c0 / (2*B)`: 250 MHz gives about 0.5996 m. The radar's value differs
  slightly because of hardware timing. Use the radar's value.
- **Velocity axis (CW)**: `v[n] = (n-1) * DopplerBin / (ZeroPad * 1e6)` m/s.
  Theoretical `c0 / (2 * f0 * T)` with `T = 0.98304 ms * TRamp`, f0 = MinFreq.
  CW only detects moving targets.
- **TD sample to volts**: `U = 3 V * TD[n] / (2^12 * AVG * tR)`, AVG = 4, tR = TRamp. **[VERIFY: garbled in PDF]**
- **FD magnitude to relative level**: `LVL = 20 * log10(FD[n] / 2^21)` dB(m). **[VERIFY: garbled in PDF]**
- **Angles**: `deg = raw * 180 / (2^25 * pi)`.
- **Tx power**: `dBm = TxPower / 10`.

## 8. Worked byte examples (use as golden fixtures)
Get info: host sends `12 00`, radar echoes `12 00`, radar sends 38 bytes.

Set system params (illustrative values: band 4, TRamp 2, ZeroPad 2, magnitude only,
frontend on, channels 0x000F, FreqPoints 256, MinFreq/ManualBW 0, attenuation 0):
```
host  -> 28 00
radar -> 28 00
host  -> 04 02 02 00 01 00 00 0F 00 00 00 01 00 00 00 00 00      (17 bytes)
```
Set operation params (opMode 1, autoMeasType 0x005C, autoDelay 100 ms):
```
host  -> 20 00
radar -> 20 00
host  -> 01 00 5C 00 64 00
```
Start of a stream frame for that autoMeasType:
```
AA AA 55 55 5C 00 | 8 B system time | 12 B temps/power | 15 B header | FD block
```

## 9. Behavioural notes
- 0x0028 and 0x0020 write EEPROM: never in a loop, never on every UI change.
- Do not change system or operation parameters while a stream socket is connected. The frame
  layout changes mid-stream. Stop the stream, apply, verify, restart.
- Stopping = closing the stream socket. The radar stays in automatic mode (opMode persists).
- If the stream port refuses connections, opMode is probably 0 (or a firewall is in the way).
- Frame period is at least TRamp plus transfer time, and at least autoDelay.
  More connected stream sockets means a longer transfer per cycle.
- The radar clock is not synced by this app (0x0002 is out of scope), so `Millis` may be time
  since power-up. Recordings therefore carry the host receive time for each frame.

## 10. Unverified / ambiguous (check against real hardware or IMST)
1. Unpadded Ethernet payloads: info reply 38 B, set params 17 B, get params 29 B, header block 15 B.
   If reads are misaligned, check padding first.
2. Order of blocks inside a stream frame. The manual only says "bits 2-8"; ascending order is assumed.
3. Whether the radar sends anything after the payload of 0x0028 / 0x0020 (the manual says no),
   and how long the radar takes to restart after 0x0028.
4. CW mode: which frequency is emitted (MinFreq assumed), and how Band/ManualBW apply.
5. TRamp valid range. ActChannels bit semantics and valid combinations (up to 4 IF channels;
   the manual's DataMin/DataMax arrays are fixed at 4). Meaning of `Overload = 0`.
6. TD voltage and FD level formulas (PDF text is garbled).
7. TRamp is u8 in system params but u16 in the automatic header.
8. Version history 1.6 mentions a command to unlock the full frequency band, but the manual body
   does not describe it. The app does not use it.
9. Manual has table-reference slips (for example FftDatTyp referring to Table 2.19); follow this file
   and record any real-hardware discrepancy here.
