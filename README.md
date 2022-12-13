# connections-tableau
Small utility to collect Tableau information which is only available with Admin permissions.


## How to use

If you have Golang installed, you can build the binary yourself, otherwise download appropriate binary from the [releases screen](https://github.com/getsynq/connections-tableau/releases) (darwin == macOS).

There are two ways to run the binary, using command line arguments and using interactive wizard:

```
❯ ./connections-tableau --help
Small utility to collect Tableau information which is only available with Admin permissions

Usage:
  connections-tableau [flags]

Flags:
  -h, --help                                       help for connections-tableau
      --site synqtest                              Site name (e.g. synqtest from https://prod-uk-a.online.tableau.com/t/synqtest/)
      --token string                               Value of Personal Access Token for Tableau with Admin permissions
      --token_name synq                            Name of the Private Access Token (e.g. synq)
      --url https://prod-uk-a.online.tableau.com   Full URL of Tableau (e.g. https://prod-uk-a.online.tableau.com)
```

```
❯ ~/Downloads/connections-tableau
? Full URL of Tableau (e.g. `https://prod-uk-a.online.tableau.com`) https://prod-uk-a.online.tableau.com
? Site name (e.g. `synqtest` from https://prod-uk-a.online.tableau.com/t/synqtest/) synqtest
? Name of the Private Access Token synq
? Value of Personal Access Token for Tableau with Admin permissions *********************************************************
```
