import removeMd from "remove-markdown";

/**
 * Remove markdown and normalize whitespace from a string.
 */
export function cleanMarkdown(input: string | null | undefined): string {
  if (!input) return "";
  const stripped = removeMd(input, { useImgAltText: false });
  // Collapse whitespace and trim
  return stripped.replace(/\s+/g, " ").trim();
}

/**
 * Create a short preview from a markdown-capable description.
 * - Strips markdown
 * - Normalizes whitespace
 * - Truncates to `max` characters with ellipsis if needed
 */
export function shortenCleanDescription(input: string | null | undefined, max = 140): string {
  const cleaned = cleanMarkdown(input);
  if (cleaned.length <= max) return cleaned;
  // Try to cut at a word boundary before the limit
  const slice = cleaned.slice(0, max);
  const lastSpace = slice.lastIndexOf(" ");
  const base = lastSpace > max * 0.6 ? slice.slice(0, lastSpace) : slice;
  return base.replace(/[\s.,;:\-]+$/, "") + "…";
}
