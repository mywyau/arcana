# Arcana private beta client

This is an early private test build. Progress may be reset, features may be
incomplete, and the server may be unavailable between agreed test sessions.
Please use temporary beta credentials rather than a password used elsewhere.

## macOS

1. Unzip `Arcana-macOS-<version>.zip`.
2. Right-click `Arcana.app` and choose **Open**.
3. Confirm **Open** if macOS warns that the developer cannot be verified.
4. Register an account or log in from the opening screen.

The application already contains the private-beta server address. Godot, Go,
the repository and the terminal are not required. The server must be resumed
by the developer before the client can connect.

This first build is not Apple-notarised. Do not bypass a warning if the file
did not come directly from the developer or its checksum is unexpected.

## Windows 10 or 11 (64-bit Intel/AMD)

1. Right-click `Arcana-Windows-x86_64-<version>.zip` and choose **Extract All**.
2. Keep `Arcana.exe` and `Arcana.pck` together in the extracted folder.
3. Double-click `Arcana.exe`.
4. If SmartScreen appears, select **More info**, followed by **Run anyway**.
5. Register an account or log in from the opening screen.

Do not launch `Arcana.exe` from inside the ZIP archive. The game data is stored
in `Arcana.pck`, so moving or sending the executable by itself will prevent the
client from starting correctly. Godot, Go, the repository and command-line
scripts are not required to play the packaged build.

This first Windows build is not code-signed. Windows SmartScreen may display a
warning; choose **More info** and **Run anyway** only if the archive came
directly from the developer and its checksum matches.

## Connection troubleshooting

The developer must resume the private-beta server before either client can
connect. Open the following address in a web browser:

<https://arcana-beta-server.onrender.com/ready>

When the server is available it returns:

```json
{"status":"ready"}
```

If that page is unavailable or does not report `ready`, contact the developer
rather than reinstalling the client. The first request after the server has
been resumed may take a little longer while the hosted service starts.
