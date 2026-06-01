import { createHighlighter } from 'shiki'

let highlighterPromise: ReturnType<typeof createHighlighter> | null = null

export async function getHighlighter() {
  if (!highlighterPromise) {
    highlighterPromise = createHighlighter({
      themes: ['github-dark', 'github-light'],
      langs: [
        'typescript',
        'javascript',
        'tsx',
        'jsx',
        'json',
        'python',
        'rust',
        'go',
        'bash',
        'shell',
        'markdown',
        'css',
        'html',
        'sql',
        'yaml',
        'toml',
        'dockerfile',
      ],
    })
  }
  return highlighterPromise
}

export async function highlightCode(code: string, lang: string = 'text'): Promise<string> {
  try {
    const highlighter = await getHighlighter()
    const lines = highlighter.codeToTokensBase(code, {
      lang: (lang || 'text') as any,
      theme: 'github-dark',
    })

    return lines
      .map((line) =>
        line
          .map((token) => token.color ? `${hexToAnsi(token.color)}${token.content}\x1b[39m` : token.content)
          .join(''),
      )
      .join('\n')
  } catch (error) {
    // Fallback to plain text if highlighting fails
    return code
  }
}

function hexToAnsi(hex: string): string {
  const normalized = hex.replace('#', '')
  const value = Number.parseInt(normalized, 16)

  if (Number.isNaN(value)) {
    return ''
  }

  const red = (value >> 16) & 255
  const green = (value >> 8) & 255
  const blue = value & 255

  return `\x1b[38;2;${red};${green};${blue}m`
}
