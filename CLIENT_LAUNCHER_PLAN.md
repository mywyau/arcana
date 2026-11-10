# Arcana Client Launcher Implementation Plan

This plan describes how to add a launcher that keeps private-beta clients up
to date before starting the Godot game. The initial implementation should
download complete versioned client archives rather than binary delta patches.
That is simpler to secure, test, recover and operate while the client is still
changing quickly.

The launcher is a separate application. A tester opens the launcher, the
launcher checks a small online release manifest, installs an update when
needed, and then starts the correct Arcana client executable. The game server
continues to authenticate players and own gameplay; it does not distribute
large client files.

## Current foundation

The repository already provides several useful pieces:

- `client/project.godot` contains the human-readable client version.
- `build-client-macos.sh` creates a versioned macOS ZIP and SHA-256 checksum.
- `build-client-windows.sh` creates a versioned Windows ZIP and checksum.
- The client sends its protocol and build versions during authentication.
- The server rejects incompatible protocol versions.
- `dist/` is the local output directory for release artifacts.

The launcher work should extend these mechanisms instead of introducing a
second version source.

## Target architecture

```text
Player opens launcher
        |
        v
Launcher downloads latest.json over HTTPS
        |
        v
Compare latest version with installed version
        |
        +-- already installed --> start game
        |
        v
Download platform archive to a temporary file
        |
        v
Verify size and SHA-256 checksum
        |
        v
Extract into a temporary version directory
        |
        v
Validate expected executable and data files
        |
        v
Atomically activate the new version
        |
        v
Start game and retain the previous version for rollback
```

The online components are:

1. A small HTTPS release manifest named `latest.json`.
2. Versioned macOS and Windows client archives.
3. A download host such as GitHub Releases, Cloudflare R2 or Amazon S3.

Render should continue hosting the game and account server. Large client
archives should not be served by the game process.

## Release manifest

The first manifest format should be small, explicit and versioned itself:

```json
{
  "manifest_version": 1,
  "latest_version": "0.1.1-beta",
  "minimum_version": "0.1.1-beta",
  "published_at": "2026-11-10T18:00:00Z",
  "release_notes_url": "https://example.com/arcana/releases/0.1.1-beta",
  "platforms": {
    "windows-x86_64": {
      "url": "https://downloads.example.com/Arcana-Windows-x86_64-0.1.1-beta.zip",
      "sha256": "replace-with-64-lowercase-hex-characters",
      "size_bytes": 123456789,
      "entrypoint": "Arcana.exe"
    },
    "macos-universal": {
      "url": "https://downloads.example.com/Arcana-macOS-0.1.1-beta.zip",
      "sha256": "replace-with-64-lowercase-hex-characters",
      "size_bytes": 123456789,
      "entrypoint": "Arcana.app"
    }
  }
}
```

`latest_version` is the preferred build. `minimum_version` allows an old build
to be made mandatory when it can no longer safely connect. The launcher must
reject unsupported manifest versions, unknown platforms, insecure download
URLs and malformed checksums.

## Local installation layout

Install game versions in user-writable application-data directories rather
than beside the launcher or under a protected system directory.

Windows:

```text
%LOCALAPPDATA%/Arcana/
|-- launcher-state.json
|-- downloads/
`-- versions/
    |-- 0.1.0-beta/
    `-- 0.1.1-beta/
```

macOS:

```text
~/Library/Application Support/Arcana/
|-- launcher-state.json
|-- downloads/
`-- versions/
    |-- 0.1.0-beta/
    `-- 0.1.1-beta/
```

The state file should record only launcher state, not login credentials:

```json
{
  "active_version": "0.1.1-beta",
  "previous_version": "0.1.0-beta",
  "last_manifest_check": "2026-11-10T18:05:00Z"
}
```

Account sessions remain in the Godot client data location and must not be
copied into release archives or launcher logs.

## Implementation slices

Each slice should be completed and tested independently.

### Slice 1: Generate the release manifest

- Add one script that reads the version from `client/project.godot`.
- Read the archive sizes and generated SHA-256 checksums from `dist/`.
- Produce `dist/latest.json` using the documented schema.
- Fail if an expected archive is absent, empty or named for another version.
- Validate the generated JSON in an automated test.
- Do not upload or modify remote services in this slice.

Acceptance criteria:

- One command produces a deterministic manifest for available platforms.
- Manifest versions, filenames, sizes and checksums match the artifacts.
- Invalid or incomplete build output causes a clear non-zero failure.

### Slice 2: Create a minimal launcher shell

- Create a separate lightweight launcher project and executable.
- Display the installed version and update status.
- Detect Windows x86-64 and macOS Universal builds.
- Fetch and parse `latest.json` with timeouts and useful error messages.
- Provide Retry, Play and Exit actions.
- Initially, Play may start a manually installed client.

A small Godot launcher is a reasonable first choice because the project already
uses Godot and needs a graphical cross-platform interface. Keep it in its own
project directory so launcher dependencies do not enter the game client.

Acceptance criteria:

- The launcher remains usable when the update service is temporarily offline.
- A malformed or unsupported manifest cannot crash the launcher.
- The launcher can start an already installed client.

### Slice 3: Download and verify complete updates

- Download the selected archive as a `.part` file.
- Display bytes downloaded, total size and progress percentage.
- Enforce a sensible maximum download size.
- Verify `size_bytes` and SHA-256 before extraction.
- Delete rejected or interrupted temporary files safely.
- Support retrying without corrupting an installed version.

Acceptance criteria:

- A valid archive downloads and verifies successfully.
- Modified, truncated and incorrectly checksummed archives are rejected.
- Losing the network leaves the existing client playable when it is not below
  `minimum_version`.

### Slice 4: Install atomically and launch

- Extract into a new temporary directory, never over the active version.
- Reject ZIP entries that escape the destination directory.
- Validate the platform entrypoint after extraction.
- Rename the completed directory into `versions/<version>`.
- Update `launcher-state.json` using a temporary file and atomic rename.
- Start the selected executable with the required server configuration.
- Keep one previous version for rollback and remove older versions safely.

Acceptance criteria:

- Interrupting extraction never damages the active client.
- The launcher starts the newly installed version.
- A failed installation leaves the previous version active.
- Paths containing spaces work on both operating systems.

### Slice 5: Connect updates to server compatibility

- Include an update-manifest URL in the client's build configuration.
- Improve the incompatible-client screen with an Open Launcher or Download
  Update action.
- Decide whether the server should advertise a minimum client build in
  addition to the existing numeric protocol version.
- Keep protocol compatibility authoritative on the server.

Acceptance criteria:

- An incompatible client gives the player a clear route to update.
- Compatible older clients are not forced to update unless intentionally
  placed below `minimum_version`.
- Login sessions survive a normal launcher-driven client update.

### Slice 6: Publish releases automatically

- Choose the artifact host.
- Add a release command or CI workflow that builds, tests and uploads archives.
- Upload versioned archives before publishing `latest.json`.
- Publish the manifest last so it never points at missing files.
- Keep older manifests and artifacts available for rollback.
- Require a manual approval before promoting a build to beta testers.

Acceptance criteria:

- A failed upload cannot advertise an incomplete release.
- Re-publishing an existing version is rejected by default.
- Rolling the manifest back selects a known valid previous client.

### Slice 7: Distribution security and launcher updates

This slice is required before broad public distribution but can follow the
private-beta launcher.

- Sign Windows executables with a code-signing certificate.
- Sign and notarise the macOS launcher and game application.
- Sign the manifest with an offline release key and embed only the public key
  in the launcher.
- Add a tiny updater helper if the launcher must replace itself.
- Never place release signing secrets in the repository or client.

SHA-256 detects corruption only when the checksum source is trusted. A signed
manifest protects against an attacker replacing both an archive and its
checksum.

## Launcher state machine

Keep update behaviour explicit rather than spreading it across UI callbacks:

```text
STARTING
  -> CHECKING
  -> READY
  -> UPDATE_AVAILABLE
  -> DOWNLOADING
  -> VERIFYING
  -> INSTALLING
  -> READY

Any network or validation failure
  -> ERROR
  -> CHECKING when Retry is selected
```

Only `READY` may start the active game. An optional update may allow Play while
offline; a mandatory update must not. Installation and verification states
must not expose a Play action for the incomplete version.

## Failure and recovery rules

- Never overwrite the active installation in place.
- Never mark a version active until download, verification and extraction all
  succeed.
- Never execute a file directly from the downloads directory.
- Treat network responses, manifests and ZIP contents as untrusted input.
- Use timeouts for manifest and archive requests.
- Prevent two launcher instances from installing simultaneously with a local
  lock.
- Keep the previous known-good version until the new build starts successfully.
- Avoid deleting broad directories; only remove validated version or temporary
  paths owned by the launcher.
- Logs may contain versions and error details but never passwords or session
  tokens.

## Testing plan

Automated tests should cover:

- Version comparison, including beta labels.
- Manifest parsing and validation.
- Platform selection.
- Checksum success and failure.
- Interrupted downloads.
- ZIP path traversal attempts.
- Atomic state-file replacement.
- Existing-version detection.
- Rollback selection.
- Paths containing spaces and non-ASCII characters.

Manual release checks should cover:

- Fresh installation on Windows.
- Updating an existing Windows installation.
- Fresh installation on macOS.
- Updating an existing macOS installation.
- Launching while the manifest host is unavailable.
- Rejecting an intentionally modified archive.
- Receiving an update-required response from the game server.
- Preserving the player's saved login session through an update.

## Decisions to make before Slice 2

The manifest-generation slice does not depend on these choices. Decide them
before building the downloadable launcher:

1. Whether the first launcher will be a separate Godot project or a native
   application.
2. Where release archives and `latest.json` will be hosted.
3. Whether beta updates are always mandatory or only mandatory when protocol
   compatibility requires them.
4. Whether the launcher closes after starting the game or remains open to show
   game status.
5. How the launcher discovers its manifest URL for beta and future production
   channels.

## Recommended first implementation

Start with Slice 1 only. It is small, builds directly on the current export
scripts and establishes the contract that every later launcher component will
consume. Review the generated manifest before selecting the launcher UI and
download host.
