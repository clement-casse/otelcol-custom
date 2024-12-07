{
  description = "Nix flake for creating a custom OpenTelemetry Collector";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, utils }:
    utils.lib.eachDefaultSystem (system:
      let
        # Relative path of the config file describing the modules embedded in the custom OpenTelemetry Collector.
        builderManifestFile = "builder-config.yaml";

        # Generate a user-friendly version number for this development environement.
        version = builtins.substring 0 8 self.lastModifiedDate;

        # Specify the version of Go for all derivétion that will use go later on.
        overlays = [
          (final: prev: {
            go = prev.go_1_22;
          })
        ];

        pkgs = import nixpkgs {
          inherit system overlays;
        };

        nativeBuildInputs = with pkgs; [
          # Go development ecosystem
          delve
          go
          gopls
          gotools
          go-tools
          golangci-lint

          docker-client

          # Build Utilities
          yq-go # To inject Nix Attributes in the builder manifest

          # OpenTelemetry Tools for generating a custom collector (Defined in custom derivations later)
          ocb # Generate and build the code of the custom collector
        ];

        # Referencing the source repository of `opentelemetry-collector` and `opentelemetry-collector-contrib`
        # to build custom tools for collector modules development.
        otelcolVersion = "0.115.0";
        otelcolSource = pkgs.fetchFromGitHub
          {
            owner = "open-telemetry";
            repo = "opentelemetry-collector";
            rev = "v${otelcolVersion}";
            sha256 = "sha256-wl0ThYbDTVLuN9FkrsHF1prydrPMIaiK/9s2Ipeqfik=";
          };

        # Define OpenTelemetry Collector Builder Binary: It does not exist in the nixpkgs repo.
        # In addition, Go binaries of OpenTelemetry Collector does not seem to be up to date.
        ocb = pkgs.buildGoModule {
          pname = "ocb"; # The Package is named `ocb` but buildGoModule installs it as `builder`
          version = otelcolVersion;
          src = otelcolSource + "/cmd/builder";
          vendorHash = "sha256-8g/92NOCj/mH1szrKR04R+Yy9GBYNnQFMi9KhqGKelU=";

          # Tune Build Process
          CGO_ENABLED = 0;
          ldflags = let mod = "go.opentelemetry.io/collector/cmd/builder"; in [
            "-s"
            "-w"
            "-X ${mod}/internal.version=${version}"
            "-X ${mod}/internal.date=${self.lastModifiedDate}"
          ];

          doCheck = false; # Disable running the tests on the source code (the src is external, and tests are run on the repo anyway)

          # Check that the builder is installed by asking it to display its version
          doInstallCheck = true;
          installCheckPhase = ''
            $out/bin/builder version
          '';
        };
      in
      with pkgs;
      {
        # formatter: Specify the formatter that will be used by the command `nix fmt`.
        formatter = nixpkgs-fmt;

        # DevShell create a Shell with all the tools loaded to the appropriate version loaded in the $PATH
        devShells.default = mkShell {
          inherit nativeBuildInputs;
        };

        packages.default = stdenv.mkDerivation rec {
          inherit version nativeBuildInputs;
          pname = "otelcol-custom";
          src = ./.;

          outputs = [ "out" "gen" ];

          # The Patch phase modifies the source code to run with Nix:
          # In that case it retrieves the package name version and OpenTelemetry Collector Builder version
          # to inject them in the builder configuration file.
          patchPhase = ''
            runHook prePatch
            ${yq-go}/bin/yq -i '
              .dist.name = "${pname}" |
              .dist.version = "${version}" |
              .dist.output_path = "'$gen'/go/src/${pname}"' ${builderManifestFile}
            echo "===== FILE PATCHED: ${builderManifestFile} ====="
            cat ${builderManifestFile}
            echo "================================================"
            runHook postPatch
          '';

          # The Configure phase sets the build system up for running OCB:
          # The Go environment is setup to match Nix constraints: the code generated by OCB will be send
          # in the $GO_MOD_GEN_DIR directory that is part of the GOPATH.
          configurePhase = ''
            runHook preConfigure
            mkdir -p "$gen/go/src/${pname}"
            export GOPATH=$gen/go:$GOPATH
            export GOCACHE=$TMPDIR/go-cache
            runHook postConfigure
          '';

          # Custom these values to build on specific platforms
          inherit (go) GOOS GOARCH;

          # The OCB binary is then run with the patched definition and creates the binary
          buildPhase = ''
            runHook preBuild
            ${ocb}/bin/builder --config="${builderManifestFile}"
            runHook postBuild
          '';

          # The Binary is moved from $gen to $out
          installPhase = ''
            runHook preInstall
            install -m755 -D "$gen/go/src/${pname}/${pname}" "$out/bin/${pname}"
            runHook postInstall
          '';
        };
      });
}
