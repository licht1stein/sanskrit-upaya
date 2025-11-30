{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  name = "sanskrit-mitra-dev";

  buildInputs = with pkgs; [
    # Go toolchain
    go

    # Fyne dependencies (Linux)
    pkg-config
    libGL
    xorg.libX11
    xorg.libXcursor
    xorg.libXrandr
    xorg.libXinerama
    xorg.libXi
    xorg.libXxf86vm

    # Fonts for Devanagari rendering
    noto-fonts
  ];

  shellHook = ''
    echo "Sanskrit Mitra development environment"
    echo ""
    echo "Commands:"
    echo "  go run ./cmd/desktop  - Run the app"
    echo "  go run ./cmd/indexer  - Build database"
    echo "  go build -o sanskrit-mitra ./cmd/desktop  - Build binary"
    echo ""
    echo "Releases are built via GitHub Actions on tag push."
  '';
}
