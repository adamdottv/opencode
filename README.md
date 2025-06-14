[![OpenCode Terminal UI](screenshot.png)](https://github.com/sst/opencode)

AI coding agent, built for the terminal.

⚠️ **Note:** version 0.1.x is a full rewrite and we do not have proper documentation for it yet. Should have this out week of June 17th 2025 📚

### Installation

If you have a previous version of opencode < 0.1.x installed you might have to remove it first.

#### Curl

```
curl -fsSL https://opencode.ai/install | bash
```

#### NPM

```
npm i -g opencode-ai@latest
bun i -g opencode-ai@latest
pnpm i -g opencode-ai@latest
yarn global add opencode-ai@latest
```

#### Brew

```
brew install sst/tap/opencode
```

#### AUR

```
paru -S opencode-bin
```

### Usage

#### Providers

The recommended approach is to sign up for claude pro or max and do `opencode auth login` and select Anthropic. It is the most cost effective way to use this tool.

Additionally opencode is powered by the provider list at [models.dev](https://models.dev) so you can use `opencode auth login` to configure api keys for any provider you'd like to use. This is stored in `~/.local/share/opencode/auth.json`

The models.dev dataset is also used to detect common environment variables like OPENAI_API_KEY to autoload that provider.

If there are additional providers you want to use you can submit a PR to the [models.dev repo](https://github.com/sst/models.dev). If configuring just for yourself check out the Config section below
