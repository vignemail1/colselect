# colselect

`colselect` is a command-line tool written in Go that filters columns from a
command's output or a file.  
It supports two modes:

- **With header** (`-header`): the regexp pattern is applied against the **column
  names** found in the first line; the matching columns are then printed for
  every subsequent line.
- **Without header**: the regexp pattern is applied against **every field of
  every line**; only matching fields are printed.

---

## Installation

### From source (requires Go ≥ 1.23)

```bash
git clone https://github.com/vignemail1/colselect.git
cd colselect
go build -o colselect .
```

### With [mise](https://mise.jdx.dev)

```bash
cd colselect
mise install          # installs Go 1.23 declared in .mise.toml
go build -o colselect .
```

### Pre-built binaries

See the [Releases](https://github.com/vignemail1/colselect/releases) page.

---

## Usage

```
colselect [OPTIONS]
```

Reads from **stdin**, writes to **stdout**.

| Flag | Default | Description |
|---|---|---|
| `-pattern` | *(required)* | Regular expression (Go `regexp` syntax) |
| `-header` | `false` | Treat first line as header; match column names |
| `-sep` | `;` | Field separator (`space`, `tab`/`\t`, or any single char) |
| `-invert` | `false` | Invert selection: keep columns that do NOT match |

---

## Examples

### 1 — CSV with header, columns starting with a prefix

```bash
cat data.csv | colselect -header -sep ',' -pattern '^cpu'
```

Input (`data.csv`):
```
hostname,cpu_user,cpu_sys,mem_used,mem_free
server01,12.3,5.1,2048,6144
server02,8.7,2.4,3072,5120
```

Output:
```
cpu_user,cpu_sys
12.3,5.1
8.7,2.4
```

---

### 2 — Semicolon-separated CSV with header

```bash
cat metrics.csv | colselect -header -sep ';' -pattern '^(cpu|mem)'
```

---

### 3 — TSV with header (tab-separated)

```bash
cat report.tsv | colselect -header -sep tab -pattern '^net_'
```

---

### 4 — Command output, whitespace-separated, with header

```bash
ps aux | colselect -header -sep space -pattern '^(PID|CPU|COMMAND)'
```

> `ps aux` uses the first line as a header (`USER PID %CPU %MEM …`).

---

### 5 — No header: filter fields matching a pattern on every line

```bash
cat /proc/net/dev | colselect -sep space -pattern '^[0-9]+'
```

Every line is scanned independently; only numeric fields are kept.

---

### 6 — Invert selection (keep columns that do NOT match)

```bash
cat data.csv | colselect -header -sep ',' -invert -pattern '^mem'
```

Keeps every column whose name does **not** start with `mem`.

---

### 7 — Pipe from a database query (psql)

```bash
psql -c "SELECT * FROM metrics;" -A -F ',' | colselect -header -sep ',' -pattern '^(ts|value)'
```

---

### 8 — Pipe from `awk` output

```bash
awk '{print $1,$2,$3,$4}' access.log | colselect -sep space -pattern '^2[0-9][0-9]'
```

Keeps only fields that look like HTTP 2xx status codes.

---

## Build for all platforms

Using mise tasks:

```bash
mise run build-all
```

Individual targets:

| Task | Output |
|---|---|
| `mise run build-linux-amd64` | `dist/colselect-linux-amd64` |
| `mise run build-linux-arm64` | `dist/colselect-linux-arm64` |
| `mise run build-macos-amd64` | `dist/colselect-macos-amd64` |
| `mise run build-macos-arm64` | `dist/colselect-macos-arm64` |
| `mise run build-windows-amd64` | `dist/colselect-windows-amd64.exe` |
| `mise run build-windows-arm64` | `dist/colselect-windows-arm64.exe` |

---

## Requirements

- Go 1.23+ (managed automatically via mise)
- [mise](https://mise.jdx.dev) (optional but recommended)

---

## License

MIT
