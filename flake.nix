{
  description = "https://github.com/calmestend/whatsapp-messaging-service.git";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
      in {
        devShells.default = pkgs.mkShell {
          name = "whatsapp-messaging-service";

          packages = with pkgs; [
						go
          ];
        };
      }
    );
}


