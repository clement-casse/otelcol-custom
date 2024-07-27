# A Custom OpenTelemetry Collector *(with Nix)*

## Scope of the project

This project aims to demonstrate how to create custom OpenTelemetry Collectors with custom modules by leveraging the tools and methods used by the OpenTelemetry project.
The structure of this project mimics the struture of [the OpenTelemetry Collector Contrib repo](https://github.com/open-telemetry/opentelemetry-collector-contrib) to arrange the module in their own folders.
Still, with this project, I intend to provide an example of an entry level contribution on an OpenTelemetry-collector module (that is not aim to be merged with the core project) for research purpose.

So, in order to lower complexity to a more familiar level for occasionnal Go developper like me, I took some distance with some of the practices used in the OpenTelemetry repositories that I personaly am not so big of a fan of, like:

- I do not use [chloggen](https://go.opentelemetry.io/build-tools/chloggen) no any of the [opentelemetry-go-build-tools](https://github.com/open-telemetry/opentelemetry-go-build-tools) (yet ?): my project is not at a maturity level where these tools would be benefical.
- The project is built with Nix, although it is not a hard dependancy and can be omited (and should by not conoiseurs), it allows to unify the build process and the required tooling. While Nix may have high potential on build reproductibility, I am not familiar enougth with the tool to use it in CI too, but I am looking forward to it, currently it is only used to provide a dev shell with Go and the ocb binary and to run ocb with on the [./builder-config.yaml](./builder-config.yaml) file.
- a Dockerfile defines the image specification of this custom OpenTelemetry container that can run all sub directories in the [`./examples/`](./examples/) directory.
- CI/CD is lighter and more leniant, it is defined as part of the [playground repository](.github/workflows/).


## References

1. [Some blog explaining Nix to write derivations][1]
2. [Internal of Nix to create an environment to build Go Modules][2]
3. [A Collection of Articles about learning Nix][3]

[1]: https://blog.ysndr.de/posts/internals/2021-01-01-flake-ification/
[2]: https://github.com/NixOS/nixpkgs/blob/e3fbbb1d108988069383a78f424463e6be087707/pkgs/development/go-packages/generic/default.nix#L92-L110
[3]: https://ianthehenry.com/posts/how-to-learn-nix/
