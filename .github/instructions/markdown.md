---
description: 'Documentation and content creation standards'
applyTo: ['**/*.md', '**/*.mdx']
---

# Markdown Content Rules

The following markdown content rules are enforced in the validators:

1. Headings: Use appropriate heading levels (H2, H3, etc.) to structure your content.
  Do not use an H1 heading, as this will be generated based on the title.
2. Lists: Use bullet points or numbered lists for lists. Ensure proper
  indentation and spacing.
3. Code Blocks: Use fenced code blocks for code snippets. Specify the language
  for syntax highlighting.
4. Links: Use proper markdown syntax for links. Ensure that links are valid and
  accessible.
5. Images: Use proper markdown syntax for images. Include alt text for
  accessibility.
6. Tables: Use markdown tables for tabular data. Ensure proper formatting and
  alignment.
7. Line Length: Limit line length to 120 characters for readability.
8. Whitespace: Use appropriate whitespace to separate sections and improve readability.
9. Front Matter: Include YAML front matter at the beginning of the file with required metadata fields.

## Formatting and Structure

Follow these guidelines for formatting and structuring your markdown content:

- Headings: Use `##` for H2 and `###` for H3. Ensure that headings are used in a
  hierarchical manner. Recommend restructuring if content includes H4, and more
  strongly recommend for H5.
- Lists: Use `-` for bullet points and `1.` for numbered lists. Indent nested lists with two spaces.
- Code Blocks: Use fenced code blocks with a language for syntax highlighting, for example:

  ```go
  fmt.Println("hello")
  ```

- Links:
  - Inline: `[Docs](https://example.com/docs)`
  - Reference-style: `[Docs][docs]` and add at the end of the page:
    `[docs]: https://example.com/docs`
- Images: Use `![alt text](image URL)` for images. Include a brief description
  of the image in the alt text.
- Tables: Use `|` to create tables. Ensure that columns are properly aligned
  and headers are included.
- Line Length: Break lines at 120 characters to improve readability. Use soft
  line breaks for long paragraphs.
- Whitespace: Use blank lines to separate sections and improve readability. Avoid excessive whitespace.

## Validation Requirements

Ensure compliance with the following validation requirements:

- Content Rules: Ensure that the content follows the markdown content rules
  specified above.
- Formatting: Ensure that the content is properly formatted and structured
  according to the guidelines.
- Validation: Run the validation tools to check for compliance with the rules
  and guidelines.
