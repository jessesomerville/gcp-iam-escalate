
<p align="center">
    <a href="https://goreportcard.com/report/rigup/ephemeral-iam" alt="Go Report Card">
        <img src="https://goreportcard.com/badge/rigup/ephemeral-iam" />
    </a>
    <a href="https://github.com/rigup/ephemeral-iam/actions/workflows/release.yml" alt="Release Workflow">
        <img src="https://img.shields.io/github/workflow/status/rigup/ephemeral-iam/goreleaser" />
    </a>
    <a href="https://github.com/rigup/ephemeral-iam/releases" alt="Latest Release">
      <img src="https://img.shields.io/github/v/release/rigup/ephemeral-iam" />
    </a>
    <a href="https://github.com/rigup/ephemeral-iam/commits/main" alt="Commits since last release">
        <img src="https://img.shields.io/github/commits-since/rigup/ephemeral-iam/latest/main" />
    </a>
    <a href="go.mod" alt="Go Version">
        <img src="https://img.shields.io/github/go-mod/go-version/rigup/ephemeral-iam" />
    </a>
    <a href="https://github.com/rigup/ephemeral-iam/blob/master/LICENSE" alt="License">
        <img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" />
    </a>
    
</p>

<p align="center">
    <img src="docs/img/logo.png" height="160">
</p>

> **NOTICE:** There is a bug in ephemeral-iam v0.0.16 that results in a failure
> to update to the next version.  If you are on v0.0.16, you will need to manually
> update to v0.0.17.

# ephemeral-iam

A CLI tool that utilizes service account token generation to enable users to
temporarily authenticate `gcloud` commands as a service account.  The intended
use-case for this tool is to restrict the permissions that users are granted
by default in their GCP organization while still allowing them to complete
management tasks that require escalated permissions.

> **Notice:** ephemeral-iam requires granting users the `Service Account Token Generator`
> role and does not include any controls to prevent users from using these
> privileges in contexts outside of ephemeral-iam in its current state.
> For more information on ephemeral-iam's security considerations, refer to the
> [security considerations document](docs/security/security_considerations.md).

## FAQ
#### Why not just use `--impersonate-service-account`?
There are several reasons why you would use `ephemeral-iam` in favor of `--impersonate-service-account`.  Here are a few
of note:
 - With the `assume-privileges` command you can start a privileged session and run commands as the service account
   without needing to provide the `--impersonate-service-account` flag each time
 - Using `ephemeral-iam` you can enhance audit logging by adding fields to audit logs using the `request_reason` request
   attribute. For example, you could configure an alert to trigger when a service account token is generated and no
   `request_reason` field is provided.
 - `ephemeral-iam` enforces session length restrictions to limit users to only impersonate a service account for 10 min
   at a time before needing to generate a new OAuth token.
 - This tool provides some QoL features such as being able to list the service accounts that you can impersonate and
   being able to query your permissions on GCP resources
 - When you run `gcloud container clusters get-credentials CLUSTER --impersonate-service-account SA_EMAIL`, a new
   `kubeconfig` entry is generated and persisted on your filesystem.  However, `ephemeral-iam` does not persist privileged
   `kubeconfig` entries to the filesystem for added security.


## Conceptual Overview
This section explains the basic process that happens when running the `eiam assume-privileges`
command.

ephemeral-iam uses the `projects.serviceAccounts.generateAccessToken` method
to generate OAuth 2.0 tokens for service accounts which are then used in subsequent
API calls.  When a user runs the `assume-privileges` command, `eiam` makes a call
to generate an OAuth 2.0 token for the specified service account that expires
in 10 minutes. 

If the token was successfully generated, `eiam` then starts an
HTTPS proxy on the user's localhost. To enable the handling of HTTPS traffic,
a self-signed TLS certificate is generated for the proxy and stored for future
use.

Next, the active `gcloud` config is updated to forward all API calls through
the local proxy.

**Example updated configuration fields:**
```
[core]
  custom_ca_certs_file: [/path/to/eiam/config_dir/server.pem]
[proxy]
  address: [127.0.0.1]
  port: [8084]
  type: [http]
```

For the duration of the privileged session (either until the token expires or
when the user manually stops it), all API calls made with `gcloud` will be 
intercepted by the proxy which will replace the `Authorization` header with the
generated OAuth 2.0 token to authorize the request as the service account.

For `kubectl` commands, a temporary `kubeconfig` is generated, the `KUBECONFIG`
environment variable is set to the path of the temporary `kubeconfig`,
`gcloud container clusters get-credentials` is called to generate a context
with the GCP Auth Provider, then the OAuth 2.0 token is written to the token
cache fields in that context. See [Issue #49](https://github.com/rigup/ephemeral-iam/issues/49)
for more information about why this is done this way.

Once the session is over, `eiam` gracefully shuts down the proxy server and reverts
the users `gcloud` config to its original state and deletes the temporary `kubeconfig`.

## Installation
Instructions on how to install the `eiam` binary can be found in
[INSTALL.md](docs/INSTALL.md).

## Getting Started

### Help Commands
The root `eiam` invocation and each of its sub-commands have their own help
commands. These commands can be used to gather more information about a command
and to explore the accepted arguments and flags.

Top-level `--help`
```
$ eiam --help

╭────────────────────────────────────────────────────────────╮
│                                                            │
│                        ephemeral-iam                       │
│  ──────────────────────────────────────────────────────    │
│  A CLI tool for temporarily escalating GCP IAM privileges  │
│  to perform high privilege tasks.                          │
│                                                            │
│           https://github.com/rigup/ephemeral-iam           │
│                                                            │
╰────────────────────────────────────────────────────────────╯


╭────────────────────── Example usage ───────────────────────╮
│                                                            │
│                   Start privileged session                 │
│  ──────────────────────────────────────────────────────    │
│  $ eiam assume-privileges \                                │
│      -s example-svc@my-project.iam.gserviceaccount.com \   │
│      --reason "Emergency security patch (JIRA-1234)"       │
│                                                            │
│                                                            │
│                                                            │
│                     Run gcloud command                     │
│  ──────────────────────────────────────────────────────    │
│  $ eiam gcloud compute instances list --format=json \      │
│      -s example@my-project.iam.gserviceaccount.com \       │
│      -R "Reason"                                           │
│                                                            │
╰────────────────────────────────────────────────────────────╯

Please report any bugs or feature requests by opening a new
issue at https://github.com/rigup/ephemeral-iam/issues

Usage:
  eiam [command]

Available Commands:
  assume-privileges        Configure gcloud to make API calls as the provided service account [alias: priv]
  cloud_sql_proxy          Run cloud_sql_proxy with the permissions of the specified service account
  config                   Manage configuration values
  default-service-accounts Configure default service accounts to use in other commands [alias: default-sa]
  gcloud                   Run a gcloud command with the permissions of the specified service account
  help                     Help about any command
  kubectl                  Run a kubectl command with the permissions of the specified service account
  list-service-accounts    List service accounts that can be impersonated [alias: list]
  plugins                  Manage ephemeral-iam plugins
  query-permissions        Query current permissions on a GCP resource
  version                  Print the installed ephemeral-iam version

Flags:
  -f, --format string   Set the output of the current command (default "text")
  -h, --help            help for eiam
  -y, --yes             Assume 'yes' to all prompts

Use "eiam [command] --help" for more information about a command.
```

Sub-command `--help`
```
 $ eiam priv --help

The "assume-privileges" command fetches short-lived credentials for the provided service Account
and configures gcloud to proxy its traffic through an auth proxy. This auth proxy sets the
authorization header to the OAuth2 token generated for the provided service account. Once
the credentials have expired, the auth proxy is shut down and the gcloud config is restored.

The reason flag is used to add additional metadata to audit logs.  The provided reason will
be in 'protoPayload.requestMetadata.requestAttributes.reason'.

Usage:
  eiam assume-privileges [flags]

Aliases:
  assume-privileges, priv

Examples:

eiam assume-privileges \
  --service-account-email example@my-project.iam.gserviceaccount.com \
  --reason "Emergency security patch (JIRA-1234)"

Flags:
  -h, --help                           help for assume-privileges
  -p, --project string                 The GCP project. Inherits from the active gcloud config by default (default "my-project")
  -R, --reason string                  A detailed rationale for assuming higher permissions
  -s, --service-account-email string   The email address for the service account. Defaults to the configured default account for the current project

Global Flags:
  -f, --format string   Set the output of the current command (default "text")
  -y, --yes             Assume 'yes' to all prompts
```

### Tutorial
To better familiarize yourself with `ephemeral-iam` and how it works, you can
follow [the tutorial provided in the documentation](docs/tutorial).
