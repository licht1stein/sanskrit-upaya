{
  pkgs ? import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/nixos-25.05.tar.gz") {}
}:

pkgs.mkShell {
  name = "sanskrit-upaya-dev";

  buildInputs = with pkgs; [
    # Go toolchain
    go_1_24

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

    # Google Cloud CLI (for OCR setup)
    google-cloud-sdk
  ];

  shellHook = ''
    echo "Sanskrit Upaya development environment"
    echo ""
    echo "Commands:"
    echo "  go run ./cmd/desktop  - Run the app"
    echo "  go run ./cmd/indexer  - Build database"
    echo "  go build -o sanskrit-upaya ./cmd/desktop  - Build binary"
    echo ""
    echo "Releases are built via GitHub Actions on tag push."
  '';
}
