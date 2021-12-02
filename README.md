# go-svm README

The Golang client for [`SVM`](https://github.com/spacemeshos/svm). Its primary goal is supplying an ergonomic API for [`go-spacemesh`](https://github.com/spacemeshos/go-spacemesh)

<br>

## Installation

### Grab your GitHub Personal Access Token

You'll need a GitHub Token in order to continue with the installation.
The reason for that is that SVM currently has no official releases (public release can be downloaded without access token).
The access token is required for using the GitHub Actions API for downloading the latest SVM successful build artifacts.

There are two options for getting your personal access token:

- [GitHub Personal access tokens](https://github.com/settings/tokens)
  You'll need click the `Generate new token` button.
  No need to check anything under the `Select scopes` section.
- [Github CLI](https://github.com/cli/cli)
  Once you've logged-in to the CLI, just run the command:
  ```bash
  gh auth status --show-token
  ```
  The last output line will look like:
  ```bash
  ✓ Token: **HERE_IS_YOUR_ACCESS_TOKEN**
  ```

### go-svm Installation

For installing `go-svm` please follow these instructions:

```bash
git clone https://github.com/spacemeshos/go-svm
cd go-svm
GITHUB_TOKEN=YOUR_TOKEN go run mage.go install
```

## Building

Building `go-svm` requires executing:

```go
go run mage.go build
```

## Testing

It's also very easy to run the `go-svm` tests, just type:

```go
go run mage.go test
```

<br>

## Structs

### Transaction

Each binary (over-the-wire) transaction will always have two parts:

- `Envelope` - A Transaction **agnostic** content.
- `Message` - Transaction-**specific** content.

Each `Message` will be preceded by the `Envelope` - together; both make a complete binary `Transaction`.

In addition to the data sent over the wire, there will be implicit fields inferred from it.
One such example is the `Transaction Id`. That field is part of the `Context` described later.
The computation of the `TransactionId` will be done externally to `go-svm`(i.e., `go-spacemesh`).

There are in total three types of transactions under `SVM`:

- `Deploy` - For deploying Templates (see `Deploying a Template` later).
- `Spawn` - For spawning new accounts out of existing Templates (see `Spawning an Account` later).
- `Call` - For calling an existing account (see `Calling an Account` later).

### Envelope

The `Envelope` contains pieces of data that are part of any transaction.
When the Full-Node (e.g., `go-spacemesh`) receives a binary transaction from the network, it needs to decode the `Envelope` part into a Golang struct.

The other part of the Transaction, a.ka the `Message`, should be kept as `[]byte`
It's the job of `SVM` to decode the transaction `Message`. More information about each `Message` type appears later on this document.

**An `Envelope` will contain the following fields:**

- `Type` - The transaction type (`Deploy / Spawn / Call`)
- `Principal` - The `Address` of the `Account` paying for the `Gas`.
- `Amount` - For funding the `target` account (relevant only for `Spawn/Call` transactions).
- `TxNonce` - The Transaction's `nonce`.
- `GasLimit` - Maximum units of Gas to be paid.
- `GasFee` - Fee per Unit of Gas.

**And this is the corresponding Golang struct:**

```go
type Envelope struct {
	Type      TxType    // Alias for `uint8`
	Principal Address   // Alias for `[20]byte`
	Amount    Amount    // Alias for `uint64`
	TxNonce   TxNonce   // A struct holding a pair of `uint64` (Golang has no `uint128` primitive)
	GasLimit  Gas       // Alias for `uint64`
	GasFee    GasFee    // Alias for `int`
}
```

To create a new **Envelope,** use this helper function:

```go
func NewEnvelope(principal Address, amount Amount, txNonce TxNonce, gasLimit Gas, gasFee GasFee) *Envelope
```

### Message

A `Message` is essentially a blob of bytes. Each `go-svm` API expecting a `Message` will ask for it in its binary form (i.e. `[] byte`).
It's the job of `SVM` itself to decode a binary `Message` and figure out what's inside.

There are in total three types of transactions under `SVM` - each with its corresponding `Message`:

- `Deploy Message` - The `Message` of a `Deploy` transaction.
- `Spawn Message` - The `Message` of a `Spawn` transaction.
- `Call Message` - The `Message` of a `Call` transaction.

Each Transaction is detailed later in this document.

### Context

In addition to the `Transaction`, there is the `Execution Context` (or simply `Context`).
The `Context` structure will contain additional data (alongside the `Transaction`) to be used by `SVM` when executing a transaction.

It will contain data relevant to the currently executing `Context` within the Full-Node (i.e., `go-spacemesh`) and properties computed from the `Transaction` itself.
(such as the `Transaction Id`). It's the role of the Full-Node (e.g., `go-spacemesh`) to create a **Context** instance and pass it forward to `go-svm`.

Here is the current declaration of a `Context`

```go
type Context struct {
	Layer Layer  // The `Layer` (alias for `uint64`) we're about to execute the Transaction in.
	TxId  TxId   // The computed `Transaction Id` out of the `Transaction` data (`TxId` is an alias for `[32]byte`).
}
```

For Creating a new **Context,** use this helper:

```go
func NewContext(layer Layer, txId TxId) *Context
```

### Deploy Message

A `Deploy Message` will be generated using the `Template Toolchain`

By saying a `Template Toolchain`, we mean the process of:

- Compiling the Template code and emitting binary Wasm and Metadata files.
  Currently, the only way to generate such Wasm is by writing Rust code using the `[SVM SDK](https://github.com/spacemeshos/svm/tree/master/crates/sdk)` crate.

Here is a link for such an example Template: (execute the [`build.sh`](https://github.com/spacemeshos/svm/blob/master/crates/runtime/tests/wasm/calldata/build.sh) for compiling into Wasm)
[https://github.com/spacemeshos/svm/tree/master/crates/runtime/tests/wasm/calldata](https://github.com/spacemeshos/svm/tree/master/crates/runtime/tests/wasm/calldata)

- Utilizing the [`SVM CLI`](https://github.com/spacemeshos/svm/tree/master/crates/cli) for crafting a binary `Deploy` transaction.

Here is CLI usage for the generation of a binary `Spawn` message:

```bash
svm-cli craft-deploy --smwasm Template.wasm --meta Template-meta.json --output template.svm
```

In the future, there might be other alternatives to achieve the above.
If Spacemesh has its Smart-Contracts programming language in the future, it'll make sense to let that language compiler take care of everything.
In such a case, the output will be a `Deploy Message`. From here, filling in the missing parts (`Envelope` and signing the Transaction) should be the same solution used today for the `SVM SDK` and `SVM CLI`.

### Spawn Message

Each `Spawn Message` contains the following fields:

- `template` (Template Address) - The `Template` we'll spawn an account of.
- `name` (String) - The name of the `Account` (optional).
- `ctor` (String) - The constructor identifier to execute.
- `calldata` (Blob) - The input for the constructor to run.

Generating a binary `Spawn Message` can be achieved in two ways using the `SVM CLI` and `SVM Codec`

### SVM CLI

Here is an example for a `Spawn JSON`

```json
{
  "version": 0,
  "template": "b5eba98957e6a93173ffb50207cceeedfddb1a72",
  "name": "My Account",
  "ctor_name": "initialize",
  "calldata": {
    "abi": ["address", "bool"],
    "data": ["8f20ed1a0e342c2a75b1b3f8014545dd3d886078", true]
  }
}
```

In order to turn it into a binary `Spawn Message` using the CLI execute:

```bash
svm-cli tx --tx-type=spawn --input=tx.json --output=tx.bin
```

Using the CLI is very useful for tests inputs generation.

### SVM Codec

The `SVM` project ships with an artifact called `svm_codec.wasm`. That Wasm package could be used for encoding a transaction `Message`.
There has been implemented an npm package for interfacing against that Wasm package.

**svm-codec-npm:**

[https://github.com/spacemeshos/svm-codec-npm](https://github.com/spacemeshos/svm-codec-npm)

This npm package will be consumed by [`smapp`](https://github.com/spacemeshos/smapp) or the [`Process Explorer`](https://github.com/spacemeshos/explorer-frontend).
Similarly, new clients could be added in the future (for example, a Golang client to be used by [`smrepl`](https://github.com/spacemeshos/smrepl))

### Call Message

Each `Call Message` contains the following fields:

- `target` (Account Address) - The `Address` of the `Account` which we're calling.
- `function` (String) - The function's name to execute.
- `verifydata` (Blob) - The input for the `svm_verify` function.
- `calldata` (Blob) - The input for the function to run.

In a very similar manner to the `Spawn` - we can generate a binary `Call Message` given a JSON.
And the same information about the `svm_codec.wasm` applies here as well.

Here is an example for a `Call Message` given as a JSON:

```json
{
  "version": 0,
  "target": "066818abe361dd44f425da19e17c45babc40e232",
  "func_name": "store_addr",
  "verifydata": {
    "abi": [],
    "data": []
  },
  "calldata": {
    "abi": ["address"],
    "data": ["102030405060708090102030405060708090AABB"]
  }
}
```

To turn it into a binary `Call Message` using the CLI execute:

```bash
svm-cli tx --tx-type=call --input=tx.json --output=tx.bin
```

<br>

## High-level API

### Init

`Init` is the entry point for interacting with SVM in any way. It runs internal
initialization logic; it is fully thread-safe and idempotent.

```go

func Init() (*API, error)
```

### Creating a Runtime

Creates a new `SVM Runtime`. You can think of it as opening a connection to `SVM`. Please make sure to call `Init` (see above) first.

<br>
Here is the `Create Runtime` API:

```go
func (*API) NewRuntime() (*Runtime, error)
```

### Destroying a Runtime

When the usage of a `Runtime` is over, we need to release its resources. You can think of it as closing a connection.

<br>
And here is the `Destroy Runtime` API:

```go
func (rt *Runtime) Destroy()
```

### Verifying a Transaction

Performs the `verify` stage as dictated by the [Account Unification](https://github.com/spacemeshos/SMIPS/issues/49) design.
Since the `verify` flow involves the running Wasm function as done when running a `Call` transaction, the output will also be of type `CallReceipt`.

<br>
This is the relevant API to be used:

```go
func (rt *Runtime) Verify(env *Envelope, msg []byte, ctx *Context) (*CallReceipt, error)
```

### Starting a new Layer

Signaling `SVM` that we are about to start playing a list of transactions under the input `layer` Layer.
The value of the `Layer` is expected to equal the last known `committed/rewinded` Layer plus one.
Any other `layer` given as input will result in an `error`.

<br>
For starting a new `Layer`, use the following:

```go
func (rt *Runtime) Open(layer Layer) error
```

### Committing Layer changes

Commits `SVM` dirty changes. It also signals the termination of the current Layer.
In other words, after finishing executing the layer transactions, we should call a `Commit`.

<br>
The matching API:

```go
func (rt *Runtime) Commit() (Layer, State, error)
```

If the `Commit` went out fine, it would return a tuple consisting of:

- The `Layer` we have just committed.
- The newly computed `Global State Root Hash`
- Setting `nil` under the `error`

If the `Commit` errored, then the output will be:

- The `Layer` we have just tried to commit (but have failed)
- `nil` under the `State` position.
- The `error` that occurred.

### Rewinding State

Rewinds `SVM Global State` to the given L`ayer`. This capability is necessary for self-healing.

<br>
Here is the API for rewind:

```go
func (rt *Runtime) Rewind(layer Layer) (State, error)
```

If the rewind succeeds, it returns the `Global-State Root Hash` at that given point. (the `error` returned will be assigned with `nil`)
Otherwise, a `nil` will be placed under the `State` position, and the 2nd tuple element will contain the `error` that occurred.

### Retrieving an Account

Given an `Account Address` - retrieves its most basic information encapsulated within an `Account` struct.

<br>
Here is the API to be used for retrieving an account:

```go
func (rt *Runtime) GetAccount(addr Address) (Account, error)
```

<br>

And this is the definition of an `Account` at `go-svm`:

```go
type Account struct {
	Addr    Address  // The `Address` of the account
	Balance Amount   // The account's balance (`Amount` is an alias for `uint64`)
	Counter TxNonce // The account's counter. It's a struct holding a pair of `uint64` (Golang has no `uint128` primitive)
}
```

### Increasing an Account's Balance

Increases an account's balance. The motivation for that API was supporting `Rewards`

<br>
The API for increasing an account's balance:

```go
func (rt *Runtime) IncreaseBalance(addr Address, amount Amount)
```

TODO: What should be the behavior of `go-svm` when there is no account with the given `Address`?

### Deploying a Template

Deploying a Template exposes two dedicated APIs: `ValidateDeploy` and `Deploy`.

### Validate Deploy

Syntactically validates the `Deploy Message` given in a binary form and returns whether it's valid or not.

<br>
The API for validation:

```go
func (rt *Runtime) ValidateDeploy(msg []byte) (bool, error)
```

### Deploy

Performs the actual deployment of a `Template` and returns a `DeployReceipt`.

<br>

The `Deploy` API:

```go
func (rt *Runtime) Deploy(env *Envelope, msg []byte, ctx *Context) (*DeployReceipt, error)
```

<br>

That is the `DeployReceipt` definition:

```go
type DeployReceipt struct {
	Success      bool            // Whether the Transaction succeeded or not
	Error        *RuntimeError   // Returns `nil` when `Success` is true and otherwise the runtime error that occurred
	TemplateAddr TemplateAddr    // The `Template Address` for the newly deployed template
	GasUsed      Gas             // The amount of `Gas` used during the transaction execution (in units of Gas)
	Logs         []Log           // Logs created as part of transaction execution
}
```

### Spawning an Account

Performs the spawning of a new `Account` out of the existing `Template`.
Similarly to `Deploy` - spawning a new `Account` exposes two dedicated APIs: `ValidateSpawn` and `Spawn`.

### Validate Spawn

Syntactically validates the `Spawn Message` given in a binary form and returns whether it's valid or not.

<br>
The validation API:

```go
func (rt *Runtime) ValidateSpawn(msg []byte) (bool, error)
```

### Spawn

Performs the spawning of a new `Account` out of the existing `Template` and returns a `SpawnReceipt`.

<br>

The `Spawn` API:

```go
func (rt *Runtime) Spawn(env *Envelope, msg []byte, ctx *Context) (*SpawnReceipt, error)
```

<br>

Here is the `SpawnReceipt` definition:

```go
type SpawnReceipt struct {
	Success         bool            // Whether the Transaction succeeded or not
	Error           *RuntimeError   // Returns `nil` when `Success` is true and otherwise the runtime error that occurred
	AccountAddr     Address         // The `Address` for the newly spawned Account
	InitState       State           // The newly computed `Global-State Root Hash` after spawning the Account []byte          // The data returned by the constructor running during spawning the account
	GasUsed         Gas             // The amount of `Gas` used during the transaction execution (in units of Gas)
	Logs            []Log           // Logs created as part of transaction execution
	TouchedAccounts []Address       // A list of `Account Addresses` engaged in any at least a single coins-transfer during transaction execution
}
```

### Calling an Account

### Validate Call

Syntactically validates the `Call Message` given in a binary form and returns whether it's valid or not.

<br>
The validation API:

```go
func (rt *Runtime) ValidateCall(msg []byte) (bool, error)
```

### Call

Performs the actual calling an `Account` and returns a `CallReceipt`.

<br>

The `Call` API:

```go
func (rt *Runtime) Call(env *Envelope, msg []byte, ctx *Context) (*CallReceipt, error)
```

<br>

Here is the `CallReceipt` definition:

```go
type CallReceipt struct {
	Success         bool            // Whether the Transaction succeeded or not
	Error           *RuntimeError   // Returns `nil` when `Success` is true and otherwise the runtime error that occurred
	NewState        State           // The newly computed `Global-State Root Hash` after calling the Account []byte          // The data returned by calling the account
	GasUsed         Gas             // The amount of `Gas` used during the transaction execution (in units of Gas)
	Logs            []Log           // Logs created as part of transaction execution
	TouchedAccounts []Address       // A list of `Account Addresses` engaged in any at least a single coins-transfer during transaction execution
}
```

## Tests helpers:

### Runtimes Count

Returns the number of living `SVM Runtimes`

`SVM` was designed to execute transactions sequentially. It means that the number of existing Runtime instances should not exceed one.
There are functions of `SVM` that could have been called in parallel (for example, validation) - it's not recommended at this stage to take extra caution and not do that.
This helper function is intended to be used for testing purposes. However, it could be used for telemetry/tracing/debugging as well.

The API:

```go
func (*API) RuntimesCount() int
```

### Receipts Count

Returns the number of living `Receipts` returned by `SVM`
It's the job of the `go-svm` internals to release binary Receipts returned by `SVM`

If there're no bugs, the reported living Receipt count should be zero after each transaction execution.
The helper should be applied for testing purposes. However, the production code can log (with a fatal severity level) if this number somehow stops being zero.

The API:

```go
func (*API) ReceiptsCount() int
```

### Errors Count

Returns the number of internal errors returned by `SVM`.
First, it's important to stress what we mean by saying an `Error`.
When a transaction has failed due to panic or running out-of-gas - SVM needs to return a valid Receipt setting `Success` to `false`
An `error` should be returned in the case that `SVM` itself panicked - this is undefined behavior.

We, of course, hope never to reach such a point since an internal error might occur only on one Operating-System. However, this will break the consensus.
If we, unfortunately, did hit an internal error, we need to make sure the error data returned by `SVM` will be freed.

This helper function should be used for testing. It's up to the `go-svm` client to decide what to do in case an internal error is being returned.
One way is to crash to process. Another alternative is to convert that error to Golang Receipt Struct and hope for the best. Turning the internal error to Receipt can ease debugging since

the `Process Explorer` will display that Receipt as well.
