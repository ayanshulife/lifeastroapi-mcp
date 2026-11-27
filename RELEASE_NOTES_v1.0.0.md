# lifeastro-mcp v1.0.0

The first stable release: easier setup, and a way to list only the tools you need. 321 tools.
From 1.0.0 on, a change that needs a change in your configuration only comes with a new major
version. This release needs none: with the same configuration the server lists the same tools.

## New: `LIFEASTRO_MCP_TOOLS`

- By default the server lists all 321 tools. That list is large: some clients cannot take it,
  and others lose context to it. Set the environment variable `LIFEASTRO_MCP_TOOLS`, next to
  `LIFEASTRO_API_KEY`, to a comma-separated list of tool-name prefixes and only the tools whose
  names start with one of them are listed.
- A prefix can be a family (`panchang` keeps every `panchang_*` tool) or a whole tool name
  (`sade_sati`). Example: `"LIFEASTRO_MCP_TOOLS": "panchang,chart,milan"`.
- The server writes how many tools it kept to its log (stderr). If no tool name matches, it
  lists all tools and says so there, so a typo never leaves you with an empty server.
- Unset, nothing changes: every tool is listed.

## Easier key setup

- The API key in `LIFEASTRO_API_KEY` is trimmed of spaces and line breaks around it, so a key
  pasted into a config file with a stray space or newline works.

## README

- Install: use the real path of the binary in `command`. `which lifeastro-mcp` prints it:
  `/opt/homebrew/bin/lifeastro-mcp` with Homebrew on Apple Silicon Macs,
  `/usr/local/bin/lifeastro-mcp` with Homebrew on Intel Macs and with the manual install.
- macOS: a binary downloaded from the Releases page is quarantined. The README gives the
  `xattr` command that lets it run. Homebrew installs do not need it.
- Troubleshooting shows the 401 message the server returns for a wrong or revoked key, and what
  it says when the config still holds a placeholder or a cut-off key.
