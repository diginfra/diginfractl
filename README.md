# 🧰 diginfractl

[![Diginfra Core Repository](https://github.com/diginfra/evolution/blob/main/repos/badges/diginfra-core-blue.svg)](https://github.com/diginfra/evolution/blob/main/REPOSITORIES.md#core-scope) [![Stable](https://img.shields.io/badge/status-stable-brightgreen?style=for-the-badge)](https://github.com/diginfra/evolution/blob/main/REPOSITORIES.md#stable) [![License](https://img.shields.io/github/license/diginfra/diginfractl?style=for-the-badge)](./LICENSE)

The official CLI tool for working with [Diginfra](https://github.com/diginfra/diginfra) and its [ecosystem components](https://diginfra.org/docs/#what-are-the-ecosystem-projects-that-can-interact-with-diginfra).

## Installation
### Install diginfractl manually
You can download and install *diginfractl* manually following the appropriate instructions based on your operating system architecture.
#### Linux
##### AMD64
```bash
LATEST=$(curl -sI https://github.com/diginfra/diginfractl/releases/latest | awk '/location: /{gsub("\r","",$2);split($2,v,"/");print substr(v[8],2)}')
curl --fail -LS "https://github.com/diginfra/diginfractl/releases/download/v${LATEST}/diginfractl_${LATEST}_linux_amd64.tar.gz" | tar -xz
sudo install -o root -g root -m 0755 diginfractl /usr/local/bin/diginfractl
```
##### ARM64
```bash
LATEST=$(curl -sI https://github.com/diginfra/diginfractl/releases/latest | awk '/location: /{gsub("\r","",$2);split($2,v,"/");print substr(v[8],2)}')
curl --fail -LS "https://github.com/diginfra/diginfractl/releases/download/v${LATEST}/diginfractl_${LATEST}_linux_arm64.tar.gz" | tar -xz
sudo install -o root -g root -m 0755 diginfractl /usr/local/bin/diginfractl
```
> NOTE: Make sure */usr/local/bin* is in your PATH environment variable.

#### MacOS
The easiest way to install on MacOS is via `Homebrew`:
```bash
brew install diginfractl
```

Alternatively, you can download directly from the source:

##### Intel
```bash
LATEST=$(curl -sI https://github.com/diginfra/diginfractl/releases/latest | awk '/location: /{gsub("\r","",$2);split($2,v,"/");print substr(v[8],2)}')
curl --fail -LS "https://github.com/diginfra/diginfractl/releases/download/v${LATEST}/diginfractl_${LATEST}_darwin_amd64.tar.gz" | tar -xz
chmod +x diginfractl
sudo mv diginfractl /usr/local/bin/diginfractl
```
##### Apple Silicon
```bash
LATEST=$(curl -sI https://github.com/diginfra/diginfractl/releases/latest | awk '/location: /{gsub("\r","",$2);split($2,v,"/");print substr(v[8],2)}')
curl --fail -LS "https://github.com/diginfra/diginfractl/releases/download/v${LATEST}/diginfractl_${LATEST}_darwin_arm64.tar.gz" | tar -xz
chmod +x diginfractl
sudo mv diginfractl /usr/local/bin/diginfractl
```

Alternatively, you can manually download *diginfractl* from the [diginfractl releases](https://github.com/diginfra/diginfractl/releases) page on GitHub.

### Install diginfractl from source
You can install *diginfractl* from source. First thing clone the *diginfractl* repository, build the *diginfractl* binary, and move it to a file location in your system **PATH**.
```bash
git clone https://github.com/diginfra/diginfractl.git
cd diginfractl
make diginfractl
sudo mv diginfractl /usr/local/bin/diginfractl
```

# Getting Started

## Installing an artifact

This tutorial aims at presenting how to install a Diginfra artifact. The next few steps will present us with the fundamental commands of *diginfractl* and how to use them.

First thing, we need to add a new `index` to *diginfractl*:
```bash
$ diginfractl index add diginfra https://diginfra.github.io/diginfractl/index.yaml
```
We just downloaded the metadata of the **artifacts** hosted and distributed by the **diginfra** organization and made them available to the *diginfractl* tool.
Now let's check that the `index` file is in place by running:
```
$ diginfractl index list
```
We should get an output similar to this one:
```
NAME            URL                                                     ADDED                   UPDATED            
diginfra   https://diginfra.github.io/diginfractl/index.yaml     2022-10-25 15:01:25     2022-10-25 15:01:25
```
Now let's search all the artifacts related to *cloudtrail*:
```
$ diginfractl artifact search cloudtrail
INDEX           ARTIFACT                TYPE            REGISTRY        REPOSITORY                              
diginfra   cloudtrail              plugin          ghcr.io         diginfra/plugins/plugin/cloudtrail 
diginfra   cloudtrail-rules        rulesfile       ghcr.io         diginfra/plugins/ruleset/cloudtrail
```
Lets install the *cloudtrail plugin*:
```
$ diginfractl artifact install cloudtrail --plugins-dir=./
 INFO  Reading all configured index files from "/home/aldo/.config/diginfractl/indexes.yaml"
 INFO  Preparing to pull "ghcr.io/diginfra/plugins/plugin/cloudtrail:latest"
 INFO  Remote registry "ghcr.io" implements docker registry API V2
 INFO  Pulling 44136fa355b3: ############################################# 100% 
 INFO  Pulling 80e0c33f30c0: ############################################# 100% 
 INFO  Pulling b024dd7a2a63: ############################################# 100% 
 INFO  Artifact successfully installed in "./" 
```
Install the *cloudtrail-rules* rulesfile:
```
$ ./diginfractl artifact install cloudtrail-rules --rulesfiles-dir=./
 INFO  Reading all configured index files from "/home/aldo/.config/diginfractl/indexes.yaml"
 INFO  Preparing to pull "ghcr.io/diginfra/plugins/ruleset/cloudtrail:latest"
 INFO  Remote registry "ghcr.io" implements docker registry API V2
 INFO  Pulling 44136fa355b3: ############################################# 100% 
 INFO  Pulling e0dccb7b0f1d: ############################################# 100% 
 INFO  Pulling 575bced78731: ############################################# 100% 
 INFO  Artifact successfully installed in "./"
```

We should have now two new files in the current directory: `aws_cloudtrail_rules.yaml` and `libcloudtrail.so`.

# Diginfractl Configuration Files

## `/etc/diginfractl/diginfractl.yaml`

The `diginfra configuration file` is a yaml file that contains some metadata about the `diginfractl` behaviour.
It contains the list of the indexes where the artifacts are listed, how often and which artifacts needed to be updated periodically.
The default configuration is stored in `/etc/diginfractl/diginfractl.yaml`.
This is an example of a diginfractl configuration file:

``` yaml
artifact:
  follow:
    every: 6h0m0s
    diginfraVersions: http://localhost:8765/versions
    refs:
    - diginfra-rules:0
    - my-rules:1
  install:
    refs:
      - cloudtrail-rules:latest
      - cloudtrail:latest
    rulesfilesdir: /tmp/rules
    pluginsdir: /tmp/plugins
indexes:
- name: diginfra
  url: https://diginfra.github.io/diginfractl/index.yaml
- name: my-index
  url: https://example.com/diginfractl/index.yaml
registry:
  auth:
    basic:
    - password: password
      registry: myregistry.example.com:5000
      user: user
    oauth:
    - registry: myregistry.example.com:5001
      clientsecret: "999999"
      clientid: "000000"
      tokenurl: http://myregistry.example.com:9096/token
    gcp:
    - registry: europe-docker.pkg.dev
```

## `~/.config/diginfractl/`

The `~/.config/diginfractl/` directory contains:
- *cache objects*
- *OAuth2 client credentials*

### `~/.config/diginfractl/indexes.yaml`

This file is used for cache purposes and contains the *index refs* added by the command `diginfractl index add [name] [ref]`. The *index ref* is enriched with two timestamps to track when it was added and the last time is was updated. Once the *index ref* is added, `diginfractl` will download the real index in the `~/.config/diginfractl/indexes/` directory. Moreover, every time the index is fetched, the `updated_timestamp` is updated.

### `~/.config/diginfractl/clientcredentials.json`

The command `diginfractl registry auth oauth` will add the `clientcredentials.json` file to the `~/.config/diginfractl/` directory. That file will contain all the needed information for the OAuth2 authetication.

# Diginfractl Commands

## Diginfractl index

The `index` file is a yaml file that contains some metadata about the Diginfra **artifacts**. Each entry carries information such as the name, type, registry, repository and other info for the given **artifact**. Different *diginfractl* commands rely on the metadata contained in the `index` file for their operation.
This is an example of an index file:
```yaml
- name: okta
  type: plugin
  registry: ghcr.io
  repository: diginfra/plugins/plugin/okta
  description: Okta Log Events
  home: https://github.com/diginfra/plugins/tree/master/plugins/okta
  keywords:
    - audit
    - log-events
    - okta
  license: Apache-2.0
  maintainers:
    - email: cncf-diginfra-dev@lists.cncf.io
      name: The Diginfra Authors
  sources:
    - https://github.com/diginfra/plugins/tree/master/plugins/okta
- name: okta-rules
  type: rulesfile
  registry: ghcr.io
  repository: diginfra/plugins/ruleset/okta
  description: Okta Log Events
  home: https://github.com/diginfra/plugins/tree/master/plugins/okta
  keywords:
    - audit
    - log-events
    - okta
    - okta-rules
  license: Apache-2.0
  maintainers:
    - email: cncf-diginfra-dev@lists.cncf.io
      name: The Diginfra Authors
  sources:
    - https://github.com/diginfra/plugins/tree/master/plugins/okta/rules
```

### Index Storage Backends

Indices for *diginfractl* can be retrieved from various storage backends. The supported index storage backends are listed in the table below. Note if you do not specify a backend type when adding a new index *diginfractl* will try to guess based on the `URI Scheme`:

| Name  | URI Scheme | Description                                                                                   |
| ----- | ---------- | --------------------------------------------------------------------------------------------- |
| http  | http://    | Can be used to retrieve indices via simple HTTP GET requests.                                 |
| https | https://   | Convenience alias for the HTTP backend.                                                       |
| gcs   | gs://      | For indices stored as Google Cloud Storage objects. Supports application default credentials. |
| file  | file://    | For indices stored on the local file system.                                                  |


#### diginfractl index add
New indexes are configured to be used by the *diginfractl* tool by adding them through the `index add` command. There are no limits to the number of indexes that can be added to the *diginfractl* tool. When adding a new index the tool adds a new entry in a file called **indexes.yaml** and downloads the *index* file in `~/.config/diginfractl`. The same folder is used to store the **indexes.yaml** file, too.
The following command adds a new index named *diginfra*:
```bash
$ diginfractl index add diginfra https://diginfra.github.io/diginfractl/index.yaml
```

The following command adds the same index *diginfra*, but explicitly sets the storage backend to `https`:
```bash
$ diginfractl index add diginfra https://diginfra.github.io/diginfractl/index.yaml https
```
#### diginfractl index list
Using the `index list` command you can check the configured `indexes` in your local system:
```bash
$ diginfractl index list
NAME            URL                                                     ADDED                   UPDATED            
$ diginfra   https://diginfra.github.io/diginfractl/index.yaml     2022-10-25 15:01:25     2022-10-25 15:01:25
```
#### diginfractl index update
The `index update` allows to update a previously configured `index` file by syncing the local one with the remote one:
```bash
$ diginfractl index update diginfra
```
#### diginfractl index remove
When we want to remove an `index` file that we configured previously, the `index remove` command is the one we need:
```bash
$ diginfractl index remove diginfra
```
The above command will remove the **diginfra** index from the local system.

## Diginfractl artifact
The *diginfractl* tool provides different commands to interact with Diginfra **artifacts**. It makes easy to *seach*, *install* and get *info* for the **artifacts** provided by a given `index` file. For these commands to properly work we need to configure at least an `index` file in our system as shown in the previus section.
#### Diginfractl artifact search
The `artifact search` command allows to search for **artifacts** provided by the `index` files configured in *diginfractl*. The command supports searches by name or by keywords and displays all the **artifacts** that match the search. Assuming that we have already configured the `index` provided by the `diginfra` organization, the following command shows all the **artifacts** that work with **Kubernetes**:
```bash
$ diginfractl artifact search kubernetes
INDEX           ARTIFACT        TYPE            REGISTRY        REPOSITORY                            
diginfra   k8saudit        plugin          ghcr.io         diginfra/plugins/plugin/k8saudit 
diginfra   k8saudit-rules  rulesfile       ghcr.io         diginfra/plugins/ruleset/k8saudit
```

#### Diginfractl artifact info
As per the name, `artifact info` prints some info for a given **artifact**:
```bash
$ diginfractl artifact info k8saudit
REF                                             TAGS                                          
ghcr.io/diginfra/plugins/plugin/k8saudit   0.1.0 0.2.0 0.2.1 0.3.0 0.4.0-rc1 0.4.0 latest
```
It shows the OCI **reference** and **tags** for the **artifact** of interest. Thot info is usually used with other commands.

#### Diginfractl artifact install
The above commands help us to find all the necessary info for a given **artifact**. The `artifact install` command installs an **artifact**. It pulls the **artifact** from remote repository, and saves it in a given directory. The following command installs the *k8saudit* plugin in the default path:
```bash
$ diginfractl artifact install k8saudit
 INFO  Reading all configured index files from "/home/aldo/.config/diginfractl/indexes.yaml"
 INFO  Preparing to pull "ghcr.io/diginfra/plugins/plugin/k8saudit:latest"
 INFO  Remote registry "ghcr.io" implements docker registry API V2                                                                                                                                              
 INFO  Pulling 44136fa355b3: ############################################# 100% 
 INFO  Pulling ded0b5419f40: ############################################# 100% 
 INFO  Pulling 107d1230f3f0: ############################################# 100% 
 INFO  Artifact successfully installed in "/usr/share/diginfra/plugins"
```

By default, if we give the name of an **artifact** it will search for the **artifact** in the configured `index` files and downlaod the `latest` version. The commands accepts also the OCI **reference** of an **artifact**. In this case, it will ignore the local `index` files.
 The command has two flags:
 * `--plugins-dir`: directory where to install plugins. Defaults to `/usr/share/diginfra/plugins`;
 * `--rulesfiles-dir`: directory where to install rules. Defaults to `/etc/diginfra`.

 > If the repositories of the **artifacts** your are trying to install are not public then you need to authenticate to the remote registry.

#### Diginfractl artifact follow
The above commands allow us to keep up-to-date one or more given **artifacts**. The `artifact follow` command checks for updates on a periodic basis and then downloads and installs the latest version, as specified by the passed tags. 
It pulls the **artifact** from remote repository, and saves it in a given directory. The following command installs the *github-rules* rulesfile in the default path:
```bash
 $ diginfractl artifact follow github-rules
 WARN  diginfra already exists with the same configuration, skipping
 INFO  Reading all configured index files from "/root/.config/diginfractl/indexes.yaml"
INFO: Creating follower for "github-rules", with check every 6h0m0s
 INFO  Starting follower for "ghcr.io/diginfra/plugins/ruleset/github:latest"
 INFO   (ghcr.io/diginfra/plugins/ruleset/github:latest) found new version under tag "latest"
 INFO   (ghcr.io/diginfra/plugins/ruleset/github:latest) artifact with tag "latest" correctly installed

```

By default, if we give the name of an **artifact** it will search for the **artifact** in the configured `index` files and downlaod the `latest` version. The commands accepts also the OCI **reference** of an **artifact**. In this case, it will ignore the local `index` files.
 The command can specify the directory where to install the *rulesfile* artifacts through the `--rulesfiles-dir` flag (defaults to `/etc/diginfra`).

 > If the repositories of the **artifacts** your are trying to install are not public then you need to authenticate to the remote registry.
 
 > Please note that only **rulesfile** artifact can be followed.

 ## Diginfractl registry

 The `registry` commands interact with OCI registries allowing the user to authenticate, pull and push artifacts. We have tested the *diginfractl* tool with the **ghcr.io** registry, but it should work with all the registries that support the OCI artifacts.

### Diginfractl registry auth
The `registry auth` command authenticates a user to a given OCI registry.

#### Diginfractl registry auth basic
The `registry auth basic` command authenticates a user to a given OCI registry using HTTP Basic Authentication. Run the command in advance for any private registries.

#### Diginfractl registry auth oauth
The `registry auth oauth` command retrieves access and refresh tokens for OAuth2.0 client credentials flow authentication. Run the command in advance for any private registries.

#### Diginfractl registry auth gcp
The `registry auth gcp` command retrieves access tokens using [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials). In particular, it supports access token retrieval using Google Compute Engine metadata server and Workload Identity, useful to authenticate your deployed Diginfra workloads. Run the command in advance for Artifact Registry authentication.

Two typical use cases:

1. You are manipulating some rules or plugins and use `diginfractl` to pull or push to an Artifact Registry:
   1. run `gcloud auth application-default login` to generate a JSON credential file that will be used by applications.
   2. run `diginfractl registry auth gcp europe-docker.pkg.dev` for instance to use Application Default Credentials to connect to any repository hosted at `europe-docker.pkg.dev`.
2. You have a Diginfra instance with Diginfractl as a side car, running in a GKE cluster with Workload Identity enabled:
   1. Workload Identity is correctly set up for the Diginfra instance (see the [documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)).
   2. Add an environment variable like `DIGINFRACTL_REGISTRY_AUTH_GCP=europe-docker.pkg.dev` to enable GCP authentication for the `europe-docker.pkg.dev` registry.
   3. The Diginfractl instance will get access tokens from the metadata server and use them to authenticate to the registry and download your rules.

### Diginfractl registry push
It pushes local files and references the artifact uniquely. The following command shows how to push a local file to a remote registry:
```bash
$ diginfractl registry push --type=plugin ghcr.io/diginfra/plugins/plugin/cloudtrail:0.3.0 clouddrail-0.3.0-linux-x86_64.tar.gz --platform linux/amd64
```
The type denotes the **artifact** type in this case *plugins*. The `ghcr.io/diginfra/plugins/plugin/cloudtrail:0.3.0` is the unique reference that points to the **artifact**.
Currently, *diginfractl* supports only two types of artifacts: **plugin** and **rulesfile**. Based on **artifact type** the commands accepts different flags:
* `--add-floating-tags`: add the floating tags for the major and minor versions
* `--annotation-source`: set annotation source for the artifact;
* `--depends-on`: set an artifact dependency (can be specified multiple times). Example: `--depends-on my-plugin:1.2.3`
* `--tag`: additional artifact tag. Can be repeated multiple time 
* `--type`: type of artifact to be pushed. Allowed values: `rulesfile`, `plugin`, `asset`

### Diginfractl registry pull
Pulling **artifacts** involves specifying the reference. The type of **artifact** is not required since the tool will implicitly extract it from the OCI **artifact**:
```
$ diginfractl registry pull ghcr.io/diginfra/plugins/plugin/cloudtrail:0.3.0
```

# Diginfractl Environment Variables

The arguments of `diginfractl` can passed as arguments through:
 - command line options
 - environment variables
 - configuration file

The `diginfractl` arguments can be passed through these different modalities are prioritized in the following order: command line options, environment variables, and finally the configuration file. This means that if an argument is passed through multiple modalities, the value set in the command line options will take precedence over the value set in environment variables, which will in turn take precedence over the value set in the configuration file.

This is the list of the environment variable that `diginfractl` will use:

| Name                                      | Content                                                          |
| ----------------------------------------- | ---------------------------------------------------------------- |
| `DIGINFRACTL_REGISTRY_AUTH_BASIC`            | `registry,username,password;registry1,username1,password1`       |
| `DIGINFRACTL_REGISTRY_AUTH_OAUTH`            | `registry,client-id,client-secret,token-url;registry1`           |
| `DIGINFRACTL_REGISTRY_AUTH_GCP`              | `registry;registry1`                                             |
| `DIGINFRACTL_INDEXES`                        | `index-name,https://diginfra.github.io/diginfractl/index.yaml` |
| `DIGINFRACTL_ARTIFACT_FOLLOW_EVERY`          | `6h0m0s`                                                         |
| `DIGINFRACTL_ARTIFACT_FOLLOW_CRON`           | `cron-formatted-string`                                          |
| `DIGINFRACTL_ARTIFACT_FOLLOW_REFS`           | `ref1;ref2`                                                      |
| `DIGINFRACTL_ARTIFACT_FOLLOW_DIGINFRAVERSIONS`  | `diginfra-version-url`                                              |
| `DIGINFRACTL_ARTIFACT_FOLLOW_RULESFILEDIR`   | `rules-directory-path`                                           |
| `DIGINFRACTL_ARTIFACT_FOLLOW_PLUGINSDIR`     | `plugins-directory-path`                                         |
| `DIGINFRACTL_ARTIFACT_FOLLOW_TMPDIR`         | `tmp-directory-path`                                             |
| `DIGINFRACTL_ARTIFACT_INSTALL_REFS`          | `ref1;ref2`                                                      |
| `DIGINFRACTL_ARTIFACT_INSTALL_RULESFILESDIR` | `rules-directory-path`                                           |
| `DIGINFRACTL_ARTIFACT_INSTALL_PLUGINSDIR`    | `plugins-directory-path`                                         |
| `DIGINFRACTL_ARTIFACT_NOVERIFY`              |                                                                  | 

Please note that when passing multiple arguments via an environment variable, they must be separated by a semicolon. Moreover, multiple fields of the same argument must be separated by a comma.

Here is an example of `diginfractl` usage with environment variables:

```bash
$ export DIGINFRACTL_REGISTRY_AUTH_OAUTH="localhost:6000,000000,999999,http://localhost:9096/token"
$ diginfractl registry oauth 
```

# Container image signature verification

Official container images for Diginfractl, starting from version 0.5.0, are signed with [cosign](https://github.com/sigstore/cosign) v2. To verify the signature run:

```bash
$ DIGINFRACTL_VERSION=x.y.z # e.g. 0.5.0
$ cosign verify docker.io/diginfra/diginfractl:$DIGINFRACTL_VERSION --certificate-oidc-issuer=https://token.actions.githubusercontent.com --certificate-identity-regexp=https://github.com/diginfra/diginfractl/ --certificate-github-workflow-ref=refs/tags/v$DIGINFRACTL_VERSION
```
