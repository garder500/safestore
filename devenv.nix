{ pkgs, lib, config, inputs, ... }:

{
  # https://devenv.sh/basics/
  env.GREET = "User";

  # https://devenv.sh/packages/
  packages = [
    pkgs.git
    pkgs.go_1_23
     ];

  # https://devenv.sh/languages/
  # languages.rust.enable = true;
  languages.go.enable = true;
  # https://devenv.sh/processes/
  # processes.cargo-watch.exec = "cargo-watch";

  # https://devenv.sh/services/
  # services.postgres.enable = true;

  # https://devenv.sh/scripts/
  scripts.hello.exec = ''
    echo hello from $GREET
  '';

  enterShell = ''
    hello
    git --version
    go version
  '';

  services.postgres = {
    enable = true;
    port = 5432;
    package = pkgs.postgresql_15;
    listen_addresses = "127.0.0.1";
    initialDatabases = [{
      name = "safestore";
      user = "safeuser";
      pass = "safepassword";
       }];
  };
  # https://devenv.sh/tasks/
  # tasks = {
  #   "myproj:setup".exec = "mytool build";
  #   "devenv:enterShell".after = [ "myproj:setup" ];
  # };

  # https://devenv.sh/tests/
  enterTest = ''
    echo "Running tests"
    git --version | grep --color=auto "${pkgs.git.version}"
  '';

  # https://devenv.sh/pre-commit-hooks/
  # pre-commit.hooks.shellcheck.enable = true;

  # See full reference at https://devenv.sh/reference/options/
}
