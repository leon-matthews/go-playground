# Podman Tutorial: Containerising the Integration Test Database

Working notes from a tutoring session on 2026-08-19. Written so a later session can resume from
here without replaying the conversation.

## Our roles

**Leon (student, driver).** Sets the goal, makes every design decision, and types all of the
project code. The point of the exercise is his understanding of containers, not a working diff.

**Claude (tutor).** Explains one concept at a time and asks a question to check it landed before
moving on. Runs read-only investigation on Leon's behalf - inspecting containers, timing things,
reading source, running tests - because gathering evidence is not the same as doing the work.
Does not write the project code, and waits to be asked before editing anything.

If you are resuming this session: keep to those roles, and keep replies short. Laying out the
whole design space at once was the main failure mode worth avoiding.

## The goal

Replace the integration tests' dependency on a manually installed system MariaDB with a MariaDB
running under podman.

## What was covered

### The host/container boundary

The same database is reachable two ways, and the difference explains everything else:

```bash
# Client runs INSIDE the container, over a unix socket in the container's filesystem
podman exec snippetbox-test-db mariadb -u test_web -ppass -D test_snippetbox

# Client runs on the HOST, over TCP to a published port
mariadb -h 127.0.0.1 -P 3307 -u test_web -ppass -D test_snippetbox
```

`--publish 3307:3306` is `host:container`. The unix socket only exists inside the container's
filesystem, so nothing on the host can use it - including `go test`, which runs on the host.
TCP to the published port is the only door available to the tests.

This was the actual bug. The DSN in `internal/models/testutils_test.go` read:

```
test_web:pass@/test_snippetbox?parseTime=true&multiStatements=true
```

An empty `protocol(address)` does not mean "no network". The driver fills in its default,
`tcp(127.0.0.1:3306)`, which used to be the system MariaDB. Fixed by naming the port explicitly:

```
test_web:pass@tcp(127.0.0.1:3307)/test_snippetbox?parseTime=true&multiStatements=true
```

### Three kinds of storage

Podman's storage root shows the split - `overlay/`, `overlay-images/`, `overlay-layers/` for
images and containers, and a completely separate `volumes/`.

| Kind | What it is | Lifetime |
| --- | --- | --- |
| Image layers | Read-only, stacked, shared between containers | Until `podman rmi` |
| Container layer | Thin writable copy-on-write layer on top | Dies with `podman rm` |
| Volume | A plain host directory mounted into the container | Independent of both |

A volume is a hole punched through the layer stack. At the mount point the overlay machinery is
not consulted at all, which is why the data survives `podman rm`, and why databases want one -
overlayfs copy-on-write is bad at the random writes a database does.

The mariadb image's Dockerfile declares `VOLUME /var/lib/mysql`. Without an explicit `-v`, podman
creates an **anonymous** volume with a 64-hex-character name. Critically, `podman rm` does **not**
delete anonymous volumes, so every experiment left its data behind - 14 orphans totalling ~1.4 GB
by the time we looked. `podman volume prune` clears them.

`podman volume ls` has no size field. Use `podman system df -v`, whose `LINKS` column is the
definition of dangling: `LINKS 0` means no container references the volume.

### Rootless user namespaces

Files in the volume appeared owned by UID 100998, which is not any user on this machine. The
mapping explains it:

```
$ podman unshare cat /proc/self/uid_map
         0       1000          1        # container root -> leon (1000)
         1     100000      65536        # container 1..65536 -> 100000..165535
```

Inside the container, mysql is plain UID 999. Through the second row that is
`100000 + (999 - 1)` = **100998**. The first row consumes one slot, which is where the off-by-one
comes from. The range comes from `/etc/subuid` (`leon:100000:65536`).

Consequences:

- Leon cannot read or write those files directly - UID 1000 is not UID 100998.
- `sudo rm -rf` is not needed. `podman unshare rm -rf <path>` enters the same namespace and works
  as the ordinary user. Worth remembering for the equivalent Docker problem at work, where a
  rootful daemon leaves files owned by real host UID 999 instead.
- Ownership here is delegated rather than direct. That delegation is what lets a rootless
  container run a process that believes it is root without being root on the host.

### Why tmpfs, and the bind-mount trap

A test database wants no persistence at all, so the data belongs in RAM. Two ways to get there,
and only one is painless:

- **Bind-mount a directory under `/tmp`.** A directory created by Leon is owned by UID 1000, and
  1000 maps to container UID 0. Inside the container it looks root-owned, so mysql (999) cannot
  write to it and MariaDB fails to initialise. Podman offers `:U` and `--userns=keep-id` as escape
  hatches, but they are friction that is not needed here.
- **`--tmpfs /var/lib/mysql`.** Podman creates the filesystem inside the container's namespace and
  the entrypoint chowns it as container-root. Nothing on the host to own, nothing to clean up, and
  no anonymous volume is created because the tmpfs occupies the declared mount point.

### Readiness cannot be inferred from the port

Measured on a throwaway container, polling separately for TCP accept, authentication, and a real
DDL statement:

```
podman run returned at 0.10s
t=0.11s  tcp=YES  auth=NO   ddl=NO
t=0.52s  tcp=YES  auth=NO   ddl=NO
t=1.36s  tcp=YES  auth=NO   ddl=NO
t=2.19s  tcp=YES  auth=NO   ddl=NO
t=2.61s  tcp=YES  auth=NO   ddl=YES
```

Podman binds the published host port at 0.11s, essentially instantly, but the database is not
usable until 2.61s. **For two and a half seconds the open port is a lie.**

That rules out the tempting alternatives. A hard-coded sleep tests nothing and will be wrong on
CI. Waiting for the port to open is actively misleading. A container healthcheck runs inside the
container, so it cannot confirm the published port works nor that `MARIADB_USER` has been created
yet, which happens partway through initialisation.

A retry loop feels crude because it is a loop, but it is correct for the reason that matters: it
tests the actual postcondition depended on - that a client on the host, using this DSN and these
credentials, can run a query - rather than a proxy for it.

Two practical notes for writing it. `sql.Open` never dials; it only parses the DSN and builds a
pool, so it cannot report an unreachable server. `db.Ping()` is the call that connects. Bound the
loop with a `context.WithTimeout` deadline rather than a fixed attempt count, so a failure reads
"database not ready after 30s" instead of being a mystery.

## The command

```bash
podman run --rm --detach --name snippetbox-test-db --publish 3307:3306 \
  --tmpfs /var/lib/mysql \
  --env MARIADB_ROOT_PASSWORD=root \
  --env MARIADB_DATABASE=test_snippetbox \
  --env MARIADB_USER=test_web \
  --env MARIADB_PASSWORD=pass \
  docker.io/library/mariadb:12
```

`--rm` removes the container when it stops, and `--tmpfs` puts the datadir in RAM. Together they
mean a test run leaves nothing behind on disk.

`MARIADB_USER` is granted full privileges on `MARIADB_DATABASE`, so `test_web` can run the
`CREATE TABLE` and `DROP TABLE` statements in `testdata/setup.sql` and `testdata/teardown.sql`
without any extra grant work.

## Results so far

| Measure | Before | After |
| --- | --- | --- |
| `go test ./internal/models/` | ~180 ms | 5-6 ms |
| Volumes accumulated per run | 1 (~155 MB) | 0 |
| Container start to usable | n/a | ~2.6-4 s |

The 30x test speedup is because `setup.sql` and `teardown.sql` run per test - create tables,
insert, drop - and every one of those DDL statements had to be durably flushed to SSD. On tmpfs
there is no disk to flush to.

## Where we are up to

The DSN fix is done and the integration tests pass against the container. The container is still
started by hand.

### The open question

How should the test helper acquire the container? The shape we converged on, but which Leon had
not committed to:

> Never stop the container from Go. The helper's job becomes "ensure a usable database exists":
> try to connect, and only if that fails run the container and poll `db.Ping()` until ready.
> `sync.Once` provides once-per-package setup without `TestMain`, and there is then no teardown
> needing a home. The first integration run of the day pays ~4 s; every run after it pays nothing.

Leon avoids `TestMain` because it confuses coworkers. The counterpoint raised: `TestMain` is Go's
**only** package-level teardown hook - `t.Cleanup` is per-test by definition - so avoiding it means
arranging to have no teardown at all, which the never-stop design achieves.

The honest cost is a container that outlives the test run, which some people find untidy. A
`make test-db-down` target covers the rare case where it needs to be gone.

**Resume here:** does a lingering container bother Leon enough to want `TestMain` after all?
Everything else follows from that answer.

### Isolation levels, for reference

Where per-test state is reset. The current choice is the second row, and it is already written.

| Level | Isolation | Cost |
| --- | --- | --- |
| Fresh container per test | Total | ~4 s plus shutdown |
| Create/drop tables per test | Very good | ~6 ms (current) |
| Truncate tables between tests | Good | Sub-millisecond |
| Transaction per test, rolled back | Leaky | Fastest |

The transaction trick breaks as soon as the code under test manages its own transactions or issues
DDL, so it is not a good fit here.

### Then

1. Write the readiness helper in `internal/models/testutils_test.go`.
2. Add a README section on running the integration tests - how to start the database, the port,
   why `--tmpfs`, and the `--short` flag for skipping. Worth deferring until the helper design
   settles, since "start-if-absent" would shrink the instructions to almost nothing.

## Environment notes

- `/tmp` is tmpfs, 30.5 GB, on a machine with 60 GB RAM. The MariaDB datadir is ~155 MB, so the
  in-RAM datadir is comfortable. The container's tmpfs came out at 30.5 GB with no size cap.
- The system `mariadb` service is inactive, so there is no conflict on port 3306.
- **AppArmor gotcha, already solved.** The system-installed MariaDB ships an AppArmor profile that
  trapped containers in the `stopping` state. A reboot cleared it and start/stop/rm have worked
  since. If it recurs, suspect AppArmor before suspecting podman.

## Appendix: the readiness probe

Reproduced here because it produced the timing table above and is generally useful when a
container will not come up. Runs on a spare port so it does not disturb a working container.

```bash
#!/bin/bash
# Probe how a MariaDB container becomes ready, from the host's point of view.
podman rm --force snippetbox-probe >/dev/null 2>&1
START=$(date +%s.%N)
podman run --rm --detach --name snippetbox-probe --publish 3308:3306 --tmpfs /var/lib/mysql \
  --env MARIADB_ROOT_PASSWORD=root --env MARIADB_DATABASE=test_snippetbox \
  --env MARIADB_USER=test_web --env MARIADB_PASSWORD=pass \
  docker.io/library/mariadb:12 >/dev/null

for i in $(seq 1 200); do
  now=$(printf "%.2f" "$(echo "$(date +%s.%N) - $START" | bc)")
  tcp=NO; auth=NO; ddl=NO
  timeout 2 bash -c 'echo > /dev/tcp/127.0.0.1/3308' 2>/dev/null && tcp=YES
  if [ "$tcp" = YES ]; then
    timeout 3 mariadb -h 127.0.0.1 -P 3308 -u test_web -ppass -e "SELECT 1" >/dev/null 2>&1 && auth=YES
    timeout 3 mariadb -h 127.0.0.1 -P 3308 -u test_web -ppass -D test_snippetbox \
        -e "CREATE TABLE probe(id INT); DROP TABLE probe;" >/dev/null 2>&1 && ddl=YES
  fi
  echo "t=${now}s  tcp=$tcp  auth=$auth  ddl=$ddl"
  [ "$ddl" = YES ] && break
  sleep 0.1
done
podman rm --force snippetbox-probe >/dev/null 2>&1
```
