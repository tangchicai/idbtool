name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  build:
    name: Build
    runs-on: ubuntu-latest

    strategy:
      matrix:
        include:
          - goos: windows
            goarch: amd64
            ext: .exe
          - goos: linux
            goarch: amd64
            ext: ''
          - goos: darwin
            goarch: amd64
            ext: ''

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.26.2'

      - name: Download dependencies
        run: go mod download

      - name: Build
        shell: bash
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          go build -ldflags="-s -w" -o idbtool-${{ matrix.goos }}-${{ matrix.goarch }}${{ matrix.ext }}

      - name: Package
        shell: bash
        run: |
          tar -czf idbtool-${{ matrix.goos }}-${{ matrix.goarch }}.tar.gz idbtool-${{ matrix.goos }}-${{ matrix.goarch }}${{ matrix.ext }}

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: idbtool-${{ matrix.goos }}-${{ matrix.goarch }}
          path: idbtool-${{ matrix.goos }}-${{ matrix.goarch }}.tar.gz

  release:
    needs: build
    runs-on: ubuntu-latest

    steps:
      - name: Download all artifacts
        uses: actions/download-artifact@v4
        with:
          path: dist

      - name: Create Release
        uses: softprops/action-gh-release@v2
        with:
          files: dist/**/*.tar.gz
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
