{
  description = "Sanskrit Upaya - A fast Sanskrit dictionary desktop application";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        # Version from VERSION file
        version = builtins.replaceStrings ["\n"] [""] (builtins.readFile ./VERSION);
      in
      {
        packages = {
          default = self.packages.${system}.sanskrit-upaya;

          sanskrit-upaya = pkgs.buildGoModule {
            pname = "sanskrit-upaya";
            inherit version;
            src = ./.;

            vendorHash = "sha256-mh1h/MZFsX4rc9v1KfoQsqKE7WTXedqb+3BwR8njfrw=";

            nativeBuildInputs = with pkgs; [
              pkg-config
              copyDesktopItems
            ];

            desktopItems = [
              (pkgs.makeDesktopItem {
                name = "sanskrit-upaya";
                desktopName = "Sanskrit Upāya";
                comment = "Fast Sanskrit dictionary with full-text search across 36 dictionaries";
                exec = "sanskrit-upaya";
                icon = "sanskrit-upaya";
                terminal = false;
                categories = [ "Education" "Dictionary" "Literature" ];
                keywords = [ "Sanskrit" "Dictionary" "IAST" "Devanagari" ];
              })
            ];

            buildInputs = with pkgs; [
              libGL
              xorg.libX11
              xorg.libXcursor
              xorg.libXrandr
              xorg.libXinerama
              xorg.libXi
              xorg.libXxf86vm
            ];

            ldflags = [
              "-s" "-w"
              "-X main.Version=${version}"
            ];

            subPackages = [ "cmd/desktop" ];

            postInstall = ''
              mv $out/bin/desktop $out/bin/sanskrit-upaya
              install -Dm644 $src/Icon.png $out/share/icons/hicolor/256x256/apps/sanskrit-upaya.png
            '';

            meta = with pkgs.lib; {
              description = "A fast Sanskrit dictionary desktop application";
              homepage = "https://github.com/licht1stein/sanskrit-upaya";
              license = licenses.mit;
              maintainers = [];
              platforms = platforms.linux;
            };
          };
        };

        devShells.default = pkgs.mkShell {
          name = "sanskrit-upaya-dev";

          buildInputs = with pkgs; [
            go_1_24
            pkg-config
            libGL
            xorg.libX11
            xorg.libXcursor
            xorg.libXrandr
            xorg.libXinerama
            xorg.libXi
            xorg.libXxf86vm
            noto-fonts
          ];

          shellHook = ''
            echo "Sanskrit Upaya development environment (flake)"
            echo ""
            echo "Commands:"
            echo "  go run ./cmd/desktop  - Run the app"
            echo "  go run ./cmd/indexer  - Build database"
            echo "  nix build             - Build package"
            echo ""
          '';
        };
      }
    );
}
