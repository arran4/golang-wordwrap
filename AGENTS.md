# Agent guidance

## Layout and compatibility
- This library deliberately exposes both positional word/box layout and image rendering. Do not conflate the positions and indexes returned by the layout pass with the pixels produced by rendering.
- Preserve the documented Box, Line, Boxer and public option contracts. Treat bounds, line breaks, whitespace, font metrics, embedded images, arbitrary rectangle origins and backtracking as candidate regression cases.
- Avoid forced rewrites of public signatures or blanket normalization of text and whitespace. Changes to geometry or line breaking should include meaningful before/after input/output examples.
- Maintain compatibility with known downstream consumers such as golang-rich-text; coordinate any deprecation with issue #49 rather than removing interfaces incidentally.

## Tests and agents
Prefer small named strings, in-memory images and explicit rectangles for unit tests. Reserve reference screenshots/goldens for rendering changes that require whole-output verification. No artificial XDG or terminal environment is needed for pure layout tests.
Report verified gaps and exact commands to issue #64 before opening implementation PRs.
## Engineering reference and follow-up

- Adopt only relevant guidance from https://arran4.github.io/blog/post/2026/055-suggested-coding-standards-for-repositories-and-agents/ and https://arran4.github.io/blog/post/2026/020-optional-dependency-injection-via-type-switched-variadic-args/ .
- This documentation is not evidence of repository-wide compliance. A coding-agent pass must report checked files, observed gaps, tests/CI and existing issue overlap to https://github.com/arran4/golang-wordwrap/issues/64 before any broad implementation.
- Do not merge implementation automatically; request review and follow repository-specific instructions.
