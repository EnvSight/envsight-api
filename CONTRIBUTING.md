# Contributing to EnvSight API

This module is the wire contract between the EnvSight agent and the EnvSight
server, and nothing else. That makes it small, and it makes changes to it more
consequential than their size suggests: both sides resolve this module by
version, and a released version can never be altered afterwards.

## Licence of contributions

Apache 2.0, and contributions arrive under the same terms — section 5 of the
licence says so, so there is no paperwork.

There is no contributor licence agreement. One would buy the right to
relicense your code later; that is not planned, and a signature demanded up
front deters exactly the small fix this repository mostly receives. The
consequence, so nobody discovers it later: **the licence cannot be changed
again without the agreement of everyone who has contributed.**

## Sign your work

Every commit must carry a `Signed-off-by` line:

```
git commit -s -m "your message"
```

That certifies the [Developer Certificate of Origin](./DCO) — that you wrote
the patch, or otherwise have the right to submit it under this licence. It is
a statement about where the code came from, not an assignment of anything. Use
a real name and a working address; your GitHub handle may be a pseudonym, the
sign-off may not.

## Changing the contract

`pb/` is generated, and it is committed. Never edit it by hand — edit
`proto/envsight.proto` and regenerate:

```sh
make generate     # regenerates pb/ using the pinned protoc and plugins
make verify       # regenerates, then fails if pb/ differs from the commit
```

`make verify` is what CI runs, and it compares against the committed tree, so
it reports a difference on any uncommitted change — including one you have
just made deliberately. Commit first, then verify.

The toolchain versions are pinned in the `Makefile` on purpose. The protoc
version is stamped into every generated file's header, so building with the
version your distribution ships produces a diff that has nothing to do with
your change. `make install-tools` fetches the pinned plugins.

### Compatibility

Both the agent and the server are deployed independently, and an operator may
run a new server against an old agent for months. So:

- **Never renumber or reuse a field tag.** A reused tag makes old messages
  decode into the wrong field, silently.
- **Never change a field's type**, and never rename an enum value's meaning.
- Adding a field is safe. Removing one is not — mark it `reserved` instead.
- proto3 has no way to distinguish "absent" from "zero", so a new field whose
  zero value means something must carry that meaning safely on both sides.

## Releases

`main` is protected: changes arrive by pull request with CI green.

Tags are cut from `main` and are immutable — a ruleset forbids deleting or
moving them, and a workflow rejects a tag whose commit is not an ancestor of
`main`. Both exist because a version, once fetched, is recorded permanently in
Go's checksum database and can never be repointed. Tag names are
`vMAJOR.MINOR.PATCH`.

## Before opening a pull request

```sh
go build ./...
go vet ./...
make verify        # after committing your regenerated pb/
```
