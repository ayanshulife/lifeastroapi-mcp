# Homebrew formula for lifeastro-mcp.
#
# Two distribution paths from this single file:
#
#  1. Tap from THIS repo (immediate):
#       brew tap ayanshulife/lifeastroapi-mcp https://github.com/ayanshulife/lifeastroapi-mcp
#       brew install lifeastro-mcp
#
#  2. Promote to the official homebrew-core later for `brew install lifeastro-mcp`
#     without a tap. Requires a stable release history first (per Homebrew's
#     audit policy: 30 days + GitHub stars threshold).
#
# Maintainer release flow when cutting a new version (e.g. `v0.12.0`):
#   1. Tag the commit and push the tag.
#   2. Wait for the GitHub Actions release workflow, then publish the draft.
#   3. Update `version` below.
#   4. Update each `sha256` from the RELEASE's lifeastro-mcp-checksums.txt
#      (not the local dist/ copy: CI and local builds differ byte for byte).
#   5. Commit + push. Users `brew upgrade lifeastro-mcp` picks it up.
#
# Why a binary formula (not source build)?
#  - The MCP server is pure Go and the release pipeline already
#    cross-compiles to every supported triple. Distributing pre-built
#    bottles avoids requiring Go on the user's machine and is faster
#    (sub-second `brew install`).
#  - Homebrew binary formulae are first-class — both `cask`-style
#    (apps) and `formula`-style (CLIs) support shipping pre-built
#    artifacts via `url` + `sha256`.

class LifeastroMcp < Formula
  desc "LifeAstroAPI MCP server — 320 Vedic + Western astrology tools for Claude Desktop, Cursor, Continue"
  homepage "https://github.com/ayanshulife/lifeastroapi-mcp"
  version "1.0.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/ayanshulife/lifeastroapi-mcp/releases/download/v#{version}/lifeastro-mcp-darwin-arm64"
      sha256 "e404feeae298e1cf66476f8102ded5a8d53b7170eed27f9f9659d603520a8a28"
    end
    on_intel do
      url "https://github.com/ayanshulife/lifeastroapi-mcp/releases/download/v#{version}/lifeastro-mcp-darwin-amd64"
      sha256 "3907e064a463b5d9ccd8cc57095712aaf4f2fc5ec1a1980f2be7e4c798731149"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/ayanshulife/lifeastroapi-mcp/releases/download/v#{version}/lifeastro-mcp-linux-arm64"
      sha256 "2b050c15afbd67118199edee3045f062384c40ffd61ddc11dcdddfb32b6b8afc"
    end
    on_intel do
      url "https://github.com/ayanshulife/lifeastroapi-mcp/releases/download/v#{version}/lifeastro-mcp-linux-amd64"
      sha256 "a95a9d5fcc05da29794df018c3fd124d47a81912f70171b917209a43f92e3c37"
    end
  end

  def install
    # The downloaded artifact is the bare binary (not a tarball).
    # Rename to the canonical name and install into Homebrew's bin.
    # Release files use Go's arch names (amd64/arm64); Homebrew reports
    # Intel as x86_64, so map it explicitly.
    arch = Hardware::CPU.intel? ? "amd64" : "arm64"
    src = "lifeastro-mcp-#{OS.kernel_name.downcase}-#{arch}"
    bin.install src => "lifeastro-mcp"
  end

  def caveats
    <<~EOS
      To use lifeastro-mcp with Claude Desktop, add it to your config:

        macOS:   ~/Library/Application Support/Claude/claude_desktop_config.json
        Windows: %APPDATA%\\Claude\\claude_desktop_config.json

        {
          "mcpServers": {
            "lifeastro": {
              "command": "#{HOMEBREW_PREFIX}/bin/lifeastro-mcp",
              "env": { "LIFEASTRO_API_KEY": "dv_live_..." }
            }
          }
        }

      Get a free API key at https://lifeastroapi.com/signup/
      Cursor users: edit ~/.cursor/mcp.json with the same shape.
    EOS
  end

  test do
    # Smoke test — without an API key the server warns on stderr and, with
    # stdin closed, exits cleanly (status 0) at EOF. The warning must name
    # the env var.
    output = shell_output("#{bin}/lifeastro-mcp </dev/null 2>&1")
    assert_match "LIFEASTRO_API_KEY", output
  end
end
