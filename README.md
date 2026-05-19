# Tomu — Manga Downloader

[![CI](https://github.com/marvinschrd/Tomu-manga-downloader/actions/workflows/ci.yml/badge.svg)](https://github.com/marvinschrd/Tomu-manga-downloader/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Go Report Card](https://goreportcard.com/badge/github.com/marvinschrd/Tomu-manga-downloader)](https://goreportcard.com/report/github.com/marvinschrd/Tomu-manga-downloader)
[![Status](https://img.shields.io/badge/status-WIP-orange)](https://github.com/marvinschrd/Tomu-manga-downloader)

> Your manga, your library, your terminal.

Tomu is a small CLI tool for downloading manga as CBZ files — perfect for loading volumes straight onto an e-reader. No browser, no accounts, no fuss.

> ⚠️ **Work in progress.** Rough edges guaranteed, sharp features incoming.

---

## Stack

Go, Cobra, SQLite. That's it.

---

## Installation

**Prerequisites:** [Go 1.21+](https://go.dev/dl/)

```bash
git clone https://github.com/marvinschrd/Tomu-manga-downloader
cd mangatool
make build
```

---

## Usage (Usage examples currently highlighting MangaDex integration only)

### Search

```
tomu search "fullmetal alchemist"
```

```
[dd8a907a] Fullmetal Alchemist
[c0ee660b] Fullmetal Alchemist (Pilot)
```

### Info

Browse chapters and volumes before downloading:

```
tomu info dd8a907a-3850-4f95-ba03-ba201a8399e3 -l en
```

### Download

```
tomu download dd8a907a-3850-4f95-ba03-ba201a8399e3
```

Download a specific range:

```
tomu download dd8a907a-3850-4f95-ba03-ba201a8399e3 -c 1-27
```

| Flag | Default | Description |
|------|---------|-------------|
| `-c`, `--chapters` | all | Chapter range, e.g. `1-27` |
| `-l`, `--lang` | from config | Language code, e.g. `en`, `fr` |
| `--force` | false | Re-download already downloaded chapters |

### Merge chapters into volumes

Group downloaded chapters into one CBZ per volume:

```
tomu merge dd8a907a-3850-4f95-ba03-ba201a8399e3
```

Clean up the individual chapter files afterwards:

```
tomu merge dd8a907a-... --remove-originals
```

Custom range and name:

```
tomu merge dd8a907a-... -c 1-27 -n "Vol.1"
```

### Update

Download new chapters for all tracked manga:

```
tomu update
```

### List your library

```
tomu list
```

### Remove manga

```
tomu remove dd8a907a-3850-4f95-ba03-ba201a8399e3
```

Keep the files, remove from library only:

```
tomu remove dd8a907a-3850-4f95-ba03-ba201a8399e3 --db-only
```

---

## Configuration

```
tomu config
tomu config set output_dir /mnt/nas/manga
tomu config set default_language fr
```

---

## Source support

Tomu currently only supports **[MangaDex](https://mangadex.org/)**.

MangaDex is great for fan translations, but popular licensed titles regularly disappear after publisher takedowns. Chapters hosted on official platforms like Viz or MangaPlus can't be downloaded and will be skipped automatically.

More sources are on the roadmap.
