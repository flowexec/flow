/**
 * Shortens a file path based on available width while preserving important context
 * @param path - The full file path to shorten
 * @param maxWidth - Maximum width in pixels for the path display
 * @param minSegments - Minimum number of path segments to show (default: 2)
 * @returns Shortened path string
 */
export function shortenPath(
  path: string,
  maxWidth: number,
  minSegments: number = 2
): string {
  if (!path) return "";

  // Normalize path separators and split into segments
  const normalizedPath = path.replace(/\\/g, "/");
  const segments = normalizedPath
    .split("/")
    .filter((segment) => segment.length > 0);

  // If path is very short, return as-is
  if (segments.length <= minSegments) {
    return path;
  }

  // Estimate character budget based on width (conservative approximation: 10px per character)
  const charBudget = Math.floor(maxWidth / 10);

  // Always preserve the last segment (current directory/file)
  const lastSegment = segments[segments.length - 1];

  // Handle very narrow widths - show minimal path
  if (maxWidth < 200) {
    if (segments.length === 1) return lastSegment;
    const parentSegment = segments[segments.length - 2];
    return ` …/${parentSegment}/${lastSegment}`;
  }

  // For wider displays, try to fit more segments
  const ellipsis = " …/";
  let includedSegments = [lastSegment];
  let currentLength = lastSegment.length;

  // Add segments from right to left while we have budget
  for (let i = segments.length - 2; i >= 0; i--) {
    const segment = segments[i];
    const newLength = currentLength + 1 + segment.length; // +1 for the '/' separator

    // If adding this segment would exceed our budget, stop and add ellipsis
    if (newLength + ellipsis.length > charBudget && i > 0) {
      return ellipsis + includedSegments.join("/");
    }

    // Add the segment
    includedSegments.unshift(segment);
    currentLength = newLength;
  }

  // If we included all segments, add the leading slash for absolute path
  const result = includedSegments.join("/");
  return normalizedPath.startsWith("/") ? "/" + result : result;
}

/**
 * Gets the estimated width needed for a path string
 * @param path - The path string
 * @returns Estimated width in pixels
 */
export function getPathWidth(path: string): number {
  // Rough estimation: 8px per character for monospace-ish display
  return path.length * 8;
}
