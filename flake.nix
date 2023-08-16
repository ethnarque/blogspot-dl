{
  description = "My Jekyll Website";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};
      in
        with pkgs; {
          devShells.default = mkShell {
            buildInputs = [ruby go wget air];
          };

          packages.default = stdenv.mkDerivation {
            name = "website";
            src = ./.;
            buildInputs = [ruby];
            buildPhase = "echo start && echo `ls` && bundle exec jekyll build";
          };
        }
    );
}
