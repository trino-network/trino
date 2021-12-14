---
order: 1
description: Overview of a new Cosmos SDK blockchain project built with Starport.
---

# Scaffold a Chain

The `starport scaffold chain` command scaffolds a new Cosmos SDK blockchain project.

## Build a Blockchain App

To build the planet application:

```bash
starport scaffold chain github.com/hello/planet
```

## Directory Structure

This command creates a directory called `planet` that contains all the files for your project and initializes a local git repository. The `github.com` URL in the argument is a string that is used for the Go module path. The repository name (`planet`, in this case) is used as the project's name.

The project directory structure:

- `app`: files that wire the blockchain together
- `cmd`: blockchain node's binary
- `proto`: protocol buffer files for custom modules
- `x`: directory with custom modules
- `vue`: scaffolded web application (optional)
- `config.yml`: configuration file

### Application-Specific Logic

Most of the logic of your application-specific blockchain is written in custom modules. Each module effectively encapsulates an independent piece of functionality. Custom modules are stored inside the `x` directory. By default, `starport scaffold chain` scaffolds a module with a name that matches the name of the project. In our example, the module name is `x/planet`.

### Proto Files

Every Cosmos SDK module has protocol buffer files defining data structures, messages, queries, RPCs, and so on. The `proto` directory contains a directory with proto files for each custom module in `x`.

### Global Settings

Global changes to your blockchain are defined in files inside the `app` directory. This includes importing third-party modules, defining relationships between modules, and configuring blockchain-wide settings.

### Configuration

The `config.yml` file contains configuration options that Starport uses to build, initialize, and start your blockchain node in development.

## Address Prefix

Account addresses on Cosmos SDK-based blockchains have string prefixes. For example, Cosmos Hub blockchain uses the default `cosmos` prefix, so that addresses look like this: `cosmos12fjzdtqfrrve7zyg9sv8j25azw2ua6tvu07ypf`.

### Change prefix on new blockchains

When creating a new blockchain, pass a prefix as a value to the `--address-prefix` flag:

```bash
starport scaffold chain github.com/hello/planet --address-prefix moonlight
```

Using the `moonlight` prefix, account addresses on your blockchain look like this: `moonlight12fjzdtqfrrve7zyg9sv8j25azw2ua6tvu07ypf`.

### Change Prefix on Existing Blockchains

To change the prefix after the blockchain has been scaffolded, modify the `AccountAddressPrefix` in the `app/prefix.go` file.

1. Change the `AccountAddressPrefix` variable in the `/app/prefix.go` file. Be sure to preserve other variables in the file.
2. To recognize the new prefix, change the `VUE_APP_ADDRESS_PREFIX` variable in `/vue/.env`.

## Cosmos SDK Version

By default, the `starport scaffold chain` command creates a Cosmos SDK blockchain using the latest stable version of the SDK.
