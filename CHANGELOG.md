# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
aims to follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html) once the first release is
tagged. Until then, source builds report the version `dev`.

## [0.2.0](https://github.com/Gitlawb/zero/compare/v0.1.0...v0.2.0) (2026-07-04)


### Features

* add --auto flag for LLM-generated commit messages ([#423](https://github.com/Gitlawb/zero/issues/423)) ([b0abde7](https://github.com/Gitlawb/zero/commit/b0abde7d0697e808480cd59d69a6f4d0c6320475))
* **providers:** add `zero providers models` to discover a provider's models ([#386](https://github.com/Gitlawb/zero/issues/386)) ([0bc8074](https://github.com/Gitlawb/zero/commit/0bc8074c97b0310e4a9d70c3f967003ee5e8a59f))
* **providers:** add KiloCode and OpenCode provider support ([#388](https://github.com/Gitlawb/zero/issues/388)) ([b1ccb6d](https://github.com/Gitlawb/zero/commit/b1ccb6d9c1875377f5e5ea81a1304edd1e41ab4f))
* **providers:** add Meituan LongCat catalog preset ([#424](https://github.com/Gitlawb/zero/issues/424)) ([b4275e3](https://github.com/Gitlawb/zero/commit/b4275e350472b2490212bf814709819d354c1216))
* **providers:** split minimax zai into intl cn ([#398](https://github.com/Gitlawb/zero/issues/398)) ([aaad4d2](https://github.com/Gitlawb/zero/commit/aaad4d271270f41af837b6f3b60ae80beba0c645))
* require manual approval before npm publish + drop release-as pin ([#369](https://github.com/Gitlawb/zero/issues/369)) ([bd89a1f](https://github.com/Gitlawb/zero/commit/bd89a1f451643c1b65ec803070abc7b116631ebe))
* **sandbox:** unelevated Windows fallback tier instead of prompts-only degrade ([#427](https://github.com/Gitlawb/zero/issues/427)) ([b9ddd6f](https://github.com/Gitlawb/zero/commit/b9ddd6f42138312a1fee8d8bb67c46c8eb1dea2f))
* **tui:** add search/filter to provider picker in setup wizard ([#400](https://github.com/Gitlawb/zero/issues/400)) ([2fcea71](https://github.com/Gitlawb/zero/commit/2fcea71778d23e050c93409c471aef45b68c1621))
* **update:** add zero upgrade command to apply self-updates ([#461](https://github.com/Gitlawb/zero/issues/461)) ([5f36349](https://github.com/Gitlawb/zero/commit/5f36349c1884e81fa9bc66bb5fe813b627e897b7))


### Bug Fixes

* **action:** keep provider key scoped to zero step ([#448](https://github.com/Gitlawb/zero/issues/448)) ([407a927](https://github.com/Gitlawb/zero/commit/407a92739ff508cba32d2c12b3f36f0efcdd54c3))
* add android platform support for Termux npm install ([#455](https://github.com/Gitlawb/zero/issues/455)) ([9bd93c6](https://github.com/Gitlawb/zero/commit/9bd93c62f8d57fb74057284aa66a1b6e1429dcdd)), closes [#449](https://github.com/Gitlawb/zero/issues/449)
* **agent:** reject a malformed additional_permissions payload before prompting ([#453](https://github.com/Gitlawb/zero/issues/453)) ([e4f760e](https://github.com/Gitlawb/zero/commit/e4f760ee8bd57299cd2fcb37e8e23130037c2607))
* allow non-TLS connections to private-network provider endpoints ([#444](https://github.com/Gitlawb/zero/issues/444)) ([1d86384](https://github.com/Gitlawb/zero/commit/1d8638466ca31517eb9db2b9353d3dce1cbeeabc))
* **auth:** route zero auth login chatgpt to the dedicated ChatGPT flow ([#443](https://github.com/Gitlawb/zero/issues/443)) ([305a62c](https://github.com/Gitlawb/zero/commit/305a62c954ca6cec00bc58d5398f933415156aff))
* **config:** fall back to a usable saved provider instead of forcing full re-onboarding ([#410](https://github.com/Gitlawb/zero/issues/410)) ([c60ad87](https://github.com/Gitlawb/zero/commit/c60ad8729f79bb841114d352ee2d2fe29d5d0e41))
* **config:** unbrick first-run setup — default google/anthropic models, enter setup on fixable config errors ([#385](https://github.com/Gitlawb/zero/issues/385)) ([72eed06](https://github.com/Gitlawb/zero/commit/72eed06b4f94c43d75d31fe54a58d2f566de059e))
* **config:** use ~/.config on macOS and enter setup when no provider ([#371](https://github.com/Gitlawb/zero/issues/371)) ([#372](https://github.com/Gitlawb/zero/issues/372)) ([027a8f2](https://github.com/Gitlawb/zero/commit/027a8f2768b17b89f5c8270887f156e2ccda69ea))
* **docs:** rename AGENTS.MD &gt; AGENTS.md ([#438](https://github.com/Gitlawb/zero/issues/438)) ([4266baf](https://github.com/Gitlawb/zero/commit/4266baf222df583ed2078b776687f12d496475b5))
* **gemini:** strip unsupported JSON Schema fields from tool declarations ([#374](https://github.com/Gitlawb/zero/issues/374)) ([39e7100](https://github.com/Gitlawb/zero/commit/39e7100674150144a1152e3110c64c7cf0321d64)), closes [#373](https://github.com/Gitlawb/zero/issues/373)
* **install:** persist install dir to user PATH on Windows ([#407](https://github.com/Gitlawb/zero/issues/407)) ([bdb1b0e](https://github.com/Gitlawb/zero/commit/bdb1b0ecd15859b1712a6037d296dace7f9c3c3f))
* **mcp:** block cross-origin credential redirects ([#396](https://github.com/Gitlawb/zero/issues/396)) ([f915f70](https://github.com/Gitlawb/zero/commit/f915f70e5a3096e2419fa8d961a0f84a626fa4a9))
* **oauth:** treat Windows ERROR_ACCESS_DENIED as lock contention in createSecretFile ([#445](https://github.com/Gitlawb/zero/issues/445)) ([d05e914](https://github.com/Gitlawb/zero/commit/d05e9148a7f79f67d1d3c31fca2775f21fbd331e))
* preserve conversation context in exec prompts ([#460](https://github.com/Gitlawb/zero/issues/460)) ([949ee43](https://github.com/Gitlawb/zero/commit/949ee43f71e5cb7fab4695c5cb7b442fe4ecfbf7))
* **provider-wizard:** allow multiple custom OpenAI-compatible providers ([#403](https://github.com/Gitlawb/zero/issues/403)) ([3fbbd28](https://github.com/Gitlawb/zero/commit/3fbbd28e4c586822cc4312c86232d94befe56e87))
* **sandbox:** gate /tmp test assertions on GOOS, not path existence ([#426](https://github.com/Gitlawb/zero/issues/426)) ([f653dca](https://github.com/Gitlawb/zero/commit/f653dcac363fb69ad7be5b35e6e0fa6d2bce476d))
* **sandbox:** self-heal a corrupt unelevated setup marker ([#437](https://github.com/Gitlawb/zero/issues/437)) ([8d0c5fe](https://github.com/Gitlawb/zero/commit/8d0c5feccb8bdbfb015df0508aa6e3bcbd1fd0e8))
* **tools:** CRLF line ending mismatch in edit_file tool on Windows ([#378](https://github.com/Gitlawb/zero/issues/378)) ([33dc7ae](https://github.com/Gitlawb/zero/commit/33dc7ae2cc82c5389675531e1416856dae7151ce))
* **tools:** flag piped POSIX utilities before running on Windows ([#412](https://github.com/Gitlawb/zero/issues/412)) ([5658a36](https://github.com/Gitlawb/zero/commit/5658a366274fc59a9d5336b06a21019c9c25cbf1))
* **tools:** require permission before web_search requests ([#382](https://github.com/Gitlawb/zero/issues/382)) ([960db96](https://github.com/Gitlawb/zero/commit/960db9660e4e31dc588fe8f7d6f116ff5e225566))
* **tui:** compose help overlay through the viewport overlay pipeline ([#421](https://github.com/Gitlawb/zero/issues/421)) ([5b2b4de](https://github.com/Gitlawb/zero/commit/5b2b4dea1aaf9e0f68baa25e97e83296fb17b1a2))
* **tui:** keep the profile name on /model switch so the stored key resolves ([#441](https://github.com/Gitlawb/zero/issues/441)) ([9134148](https://github.com/Gitlawb/zero/commit/9134148f4df3e4e556fba6c2f8babfdf6fcfeee1)), closes [#440](https://github.com/Gitlawb/zero/issues/440)
* **tui:** resolve every permission request so the agent can't deadlock ([#397](https://github.com/Gitlawb/zero/issues/397)) ([952788f](https://github.com/Gitlawb/zero/commit/952788f72d32957659fe004521fcc8372b9ba9b4))
* **tui:** show an M suffix for million-scale token counts ([#457](https://github.com/Gitlawb/zero/issues/457)) ([0562e3b](https://github.com/Gitlawb/zero/commit/0562e3bef7df2328610a48a1e81632a8da4aec64))
* **tui:** title /model rows by model name, not the catalog description ([#395](https://github.com/Gitlawb/zero/issues/395)) ([cdf9d83](https://github.com/Gitlawb/zero/commit/cdf9d839ae57a729f292f36f7c5b0c67b41b288d))

## 0.1.0 (2026-07-02)


### Features

* publish zero to npm via release-please ([#367](https://github.com/Gitlawb/zero/issues/367)) ([8eccc26](https://github.com/Gitlawb/zero/commit/8eccc2669887bc38d35bc16a315c888e4d9ec43a))
* **tui:** FILES sidebar panel with click-to-select and file drill-in ([#365](https://github.com/Gitlawb/zero/issues/365)) ([142c548](https://github.com/Gitlawb/zero/commit/142c548c89a8652ce300e64ddf1228ee36df7606))


### Bug Fixes

* **auth:** propagate credentials to every provider-build surface and pin children to the live provider ([#366](https://github.com/Gitlawb/zero/issues/366)) ([6e0a665](https://github.com/Gitlawb/zero/commit/6e0a665118fe0e09c4b07d482dd18f86045acd2b))

## [Unreleased]

### Added
- `SECURITY.md` with a private vulnerability-reporting path, `CODE_OF_CONDUCT.md`, this changelog, and
  GitHub issue/PR templates.
- Interactive `/theme` picker: bare `/theme` opens a popup that live-previews each palette as you move
  and applies on select (Esc reverts).
- Ten built-in color themes alongside the `dark`/`light` built-ins — `dracula`, `nord`, `gruvbox`,
  `tokyo-night`, `catppuccin`, `one-dark`, `solarized-dark`, `rose-pine`, `everforest`, and
  `solarized-light` — selectable via `/theme <name>`, `--theme <name>`, or `ZERO_THEME`. Every palette
  is contrast-audited to WCAG AA. The built-in light theme was reworked for legibility.
- `--theme <name>` flag for the TUI, accepting `auto` or any registered theme (previously only the
  `ZERO_THEME` env var existed).
- "Accessibility / Appearance" section in the README documenting `NO_COLOR`, `ZERO_THEME`, `/theme`,
  and `ZERO_NO_FADE`.

### Changed
- Provider connectivity health checks now allow loopback hosts for explicitly user-configured local
  providers (Ollama / LM Studio), so the keyless local-model path verifies instead of failing with
  "localhost hosts are blocked". The SSRF guard for fetched/remote URLs is unchanged.
- Auth (401/403) errors now show a curated, actionable message pointing at `zero auth` / setup; the
  raw upstream body is shown only under a verbose/debug flag.
- No-provider / missing-key errors now point at `zero setup` and `zero auth`, and distinguish a
  missing key from a rejected key.
- `zero doctor` no longer reports "Overall: pass" when no provider credential is configured, and
  formats the missing-language-server list for humans (no raw Go `map[...]`).
- Raised the `faint`/`faintest` theme tokens (and the light-theme accent) to meet WCAG AA contrast for
  the content they carry.
- `NO_COLOR` is now honored for any non-empty value, per the no-color.org spec.

### Removed
- The inert `/input-style` slash command (it had no backend).

### Fixed
- README/`go.mod` Go-version mismatch and other stale public-release docs claims.
