{
  description = "espresso-streamers development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            go-ethereum
            just
            golangci-lint
          ];

          shellHook = ''
            echo "espresso-streamers dev shell"
            echo "  just test  — run tests"
            echo "  just check — fmt-check, vet, lint, test"
          '';
        };
      }
    );
}
