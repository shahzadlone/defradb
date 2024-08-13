## defradb client acp relationship add

Add new relationship

### Synopsis

Add new relationship

Notes:
  - ACP must be available (i.e. ACP can not be disabled).
  - The target document must be registered with ACP already (policy & resource specified).
  - The requesting identity MUST either be the owner OR the manager (manages the relation) of the resource.
  - If the specified relation was not granted the miminum DPI permissions (read or write) within the policy,
  and a relationship is formed, the subject/actor will still not be able to access (read or write) the resource.
  - Learn more about [ACP & DPI Rules](/acp/README.md)

Consider the following policy:
'
description: A Policy

actor:
  name: actor

resources:
  users:
    permissions:
      read:
        expr: owner + reader + writer
      write:
        expr: owner + writer
      nothing:
        expr: dummy

    relations:
      owner:
        types:
          - actor
      reader:
        types:
          - actor
      writer:
        types:
          - actor
      admin:
        manages:
          - reader
        types:
          - actor
      dummy:
        types:
          - actor
'

defradb client ... --identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac


Example: Let another actor read my private document:
  defradb client acp relationship add --collection User --docID bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab \
	--relation reader --actor did:key:z6MkkHsQbp3tXECqmUJoCJwyuxSKn1BDF1RHzwDGg9tHbXKw \
	--identity 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f

Example: Create a dummy relation that doesn't do anything (from database prespective):
  defradb client acp relationship add -c User --docID bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab -r dummy \
	-a did:key:z6MkkHsQbp3tXECqmUJoCJwyuxSKn1BDF1RHzwDGg9tHbXKw \
	-i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f



```
defradb client acp relationship add [-i --identity] [policy] [flags]
```

### Options

```
  -a, --actor string        Actor to add relationship with
  -c, --collection string   Collection that has the resource and policy for object
      --docID string        Document Identifier (ObjectID) to make relationship for
  -h, --help                help for add
  -r, --relation string     Relation that needs to be set for the relationship
```

### Options inherited from parent commands

```
  -i, --identity string             Hex formatted private key used to authenticate with ACP
      --keyring-backend string      Keyring backend to use. Options are file or system (default "file")
      --keyring-namespace string    Service name to use when using the system backend (default "defradb")
      --keyring-path string         Path to store encrypted keys when using the file backend (default "keys")
      --log-format string           Log format to use. Options are text or json (default "text")
      --log-level string            Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string           Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string        Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                  Include source location in logs
      --log-stacktrace              Include stacktrace in error and fatal logs
      --no-keyring                  Disable the keyring and generate ephemeral keys
      --no-log-color                Disable colored log output
      --rootdir string              Directory for persistent data (default: $HOME/.defradb)
      --source-hub-address string   The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor
      --tx uint                     Transaction ID
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client acp relationship](defradb_client_acp_relationship.md)	 - Interact with the acp relationship features of DefraDB instance

