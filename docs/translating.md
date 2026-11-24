# Translating Cassocial

`src/common/i18n/locales/en.json` is the source language. Translate its values; do not change its keys or edit the English file as part of a translation contribution.

## Contributing a translation

Open a pull request with one locale file, `src/common/i18n/locales/{lang}.json`. For a new language, include the language registry and configuration changes required by AI.md PART 31. A maintainer reviews the translation and runs the project test suite before merging. Existing language corrections follow the same process and should touch only that locale file.

## Translation rules

- Keep UTF-8 JSON, the same key structure and order as `en.json`, two-space indentation, and no trailing commas or comments.
- Translate values only. Keep keys, product names, technical names, code, file names, config keys, URLs, and literal values such as `dark` and `auto` unchanged.
- Preserve every `{placeholder}` exactly, including its spelling and braces. Its position may change to fit the translated sentence.
- Include every plural category required for the language. Use complete natural sentences and consistent, neutral wording.
- Follow the language's typography and punctuation conventions. Right-to-left languages must retain their natural direction.
- Review longer translations for clipping and overflow at 320, 768, and 1280 pixels. Check right-to-left layouts in their mirrored direction.
- Keep the shared glossary consistent. Product and technology names such as Cassocial, API, Tor, `.onion`, I2P, and JSON remain unchanged.
- AI-assisted translations are expected. Use natural, idiomatic wording when the language can be translated reliably; ask for human translation when it cannot. Do not hold a complete translation for native-speaker review, and incorporate human corrections in future updates.
- Keep one consistent, neutral register for each language. Use complete sentences and avoid slang, humor, or idioms that do not translate well.
- Follow script-specific font requirements. CJK, Devanagari, Thai, and other scripts may need a font stack that covers their characters. Text without spaces must wrap with `overflow-wrap: anywhere`.
- Observe language-specific typography, including French non-breaking spaces before `: ; ! ?`. Locale-sensitive numbers and dates must use locale formatting, not translated punctuation or hard-coded separators.

## Review checklist

Maintainers check that the locale has the same message keys and placeholders as English, required plural categories are present, no values are empty, glossary terms are consistent, and translated text fits the interface. Native-speaker review is welcome but is not required before merging. When English strings change, update every shipped locale in the same change where a reliable translation is available; request human translation rather than guessing when it is not. A retired locale is removed only at its maintainer's request.

## Supported locales and credits

The repository currently includes eight locale files: Arabic (`ar`), German (`de`), English (`en`), Spanish (`es`), French (`fr`), Italian (`it`), Japanese (`ja`), and Chinese (`zh`). No individual translator credits have been submitted. Contributors may opt in by adding their chosen name or handle to the locale's `meta.translators` field.

| Locale | Language | Direction |
|--------|----------|-----------|
| `ar` | العربية | Right to left |
| `de` | Deutsch | Left to right |
| `en` | English | Left to right |
| `es` | Español | Left to right |
| `fr` | Français | Left to right |
| `it` | Italiano | Left to right |
| `ja` | 日本語 | Left to right |
| `zh` | 中文 | Left to right |

See AI.md PART 31 for the full language, plural, accessibility, and internationalization requirements.
