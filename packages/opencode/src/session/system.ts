import { App } from "../app/app"
import { ListTool } from "../tool/ls"
import { Filesystem } from "../util/filesystem"

import PROMPT_ANTHROPIC from "./prompt/anthropic.txt"
import PROMPT_ANTHROPIC_SPOOF from "./prompt/anthropic_spoof.txt"
import PROMPT_SUMMARIZE from "./prompt/summarize.txt"
import PROMPT_TITLE from "./prompt/title.txt"

export namespace SystemPrompt {
  export function provider(providerID: string) {
    const result = []
    switch (providerID) {
      case "anthropic":
        result.push(PROMPT_ANTHROPIC_SPOOF.trim())
        result.push(PROMPT_ANTHROPIC)
        break
      default:
        result.push(PROMPT_ANTHROPIC)
        break
    }
    return result
  }

  export async function environment(sessionID: string) {
    const app = App.info()
    return [
      [
        `Here is some useful information about the environment you are running in:`,
        `<env>`,
        `  Working directory: ${app.path.cwd}`,
        `  Is directory a git repo: ${app.git ? "yes" : "no"}`,
        `  Platform: ${process.platform}`,
        `  Today's date: ${new Date().toDateString()}`,
        `</env>`,
        `<project>`,
        `  ${app.git ? await ListTool.execute({ path: app.path.cwd, ignore: [] }, { sessionID: sessionID, messageID: "", abort: AbortSignal.any([]) }).then((x) => x.output) : ""}`,
        `</project>`,
      ].join("\n"),
    ]
  }

  const CUSTOM_FILES = [
    "AGENTS.md",
    "CLAUDE.md",
    "CONTEXT.md", // deprecated
  ]
  export async function custom() {
    const { cwd, root } = App.info().path
    const found = []
    for (const item of CUSTOM_FILES) {
      const matches = await Filesystem.findUp(item, cwd, root)
      found.push(...matches.map((x) => Bun.file(x).text()))
    }
    return Promise.all(found)
  }

  export function summarize(providerID: string) {
    switch (providerID) {
      case "anthropic":
        return [PROMPT_ANTHROPIC_SPOOF.trim(), PROMPT_SUMMARIZE]
      default:
        return [PROMPT_SUMMARIZE]
    }
  }

  export function title(providerID: string) {
    switch (providerID) {
      case "anthropic":
        return [PROMPT_ANTHROPIC_SPOOF.trim(), PROMPT_TITLE]
      default:
        return [PROMPT_TITLE]
    }
  }
}
