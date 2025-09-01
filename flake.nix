{
  description = "A kitten-panel based desktop panel for your desktop";

  inputs = {
    self.submodules = true;
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
      in
      {
        packages.default = pkgs.buildGoModule (finalAttrs: {
          pname = "pawbar";
          version = "0-unstable-2025-08-31";
          src = ./.;
          subPackages = [ "cmd/pawbar" ];
          vendorHash = "sha256-5ysy7DGLE99svDPUw1vS05CT7HRcSP1ov27rTqm6a8Y=";
          buildInputs = with pkgs; [
            udev
            librsvg
            cairo
          ];
          nativeBuildInputs = with pkgs; [ pkg-config ];
        });
      }
    );
}
