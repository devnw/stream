{
  description = "development flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    _1password-shell-plugins.url = "github:1Password/shell-plugins";
    gomod2nix = {
      url = "github:tweag/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      gomod2nix,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          config = {
            allowUnfree = true;
          };
          system = system;
          overlays = [
            gomod2nix.overlays.default
            (self: super: {
              go = super.go_1_23;
              python = super.python3.withPackages (
                subpkgs: with subpkgs; [
                  openapi-spec-validator
                  detect-secrets
                  requests
                  python-dotenv
                ]
              );
            })
          ];
        };

        pkglist = with pkgs; [
          # system tools
          automake
          curl
          which
          act
          gcc
          ruby
          git
          sqlite-interactive
          _1password

          # lint tools
          gibberish-detector
          addlicense
          shfmt
          pre-commit
          shellcheck

          # python
          python

          # go tools
          go
          gopls
          gotools
          go-tools
          gomod2nix.packages.${system}.default
          delve
          golangci-lint
          goreleaser
          go-licenses

          arangodb
          #go-jsonschema
        ];
      in
      {
        packages.default = pkgs.buildGoApplication {
          pname = "nist";
          version = "0.1";

          pwd = ./.;
          src = ./.;
          modules = ./gomod2nix.toml;
          buildInputs = pkglist;

          buildPhase = ''
            make build
          '';
          installPhase = ''
            make install
          '';
        };

        devShells.default = pkgs.mkShell { buildInputs = pkglist; };
      }
    );
}
