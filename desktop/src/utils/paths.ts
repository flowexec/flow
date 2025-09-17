export function shortenPath(
  path: string,
  maxWidth: number,
  minSegments: number = 2
): string {
  if (!path) return "";

  const normalizedPath = path.replace(/\\/g, "/");
  const segments = normalizedPath
    .split("/")
    .filter((segment) => segment.length > 0);

  if (segments.length <= minSegments) {
    return path;
  }

  const charBudget = Math.floor(maxWidth / 10);
  const lastSegment = segments[segments.length - 1];

  if (maxWidth < 200) {
    if (segments.length === 1) return lastSegment;
    const parentSegment = segments[segments.length - 2];
    return ` …/${parentSegment}/${lastSegment}`;
  }

  const ellipsis = " …/";
  const includedSegments = [lastSegment];
  let currentLength = lastSegment.length;

  // Add segments from right to left while we have budget
  for (let i = segments.length - 2; i >= 0; i--) {
    const segment = segments[i];
    const newLength = currentLength + 1 + segment.length; // +1 for the '/' separator

    if (newLength + ellipsis.length > charBudget && i > 0) {
      return ellipsis + includedSegments.join("/");
    }

    includedSegments.unshift(segment);
    currentLength = newLength;
  }

  // If we included all segments, add the leading slash for absolute path
  const result = includedSegments.join("/");
  return normalizedPath.startsWith("/") ? "/" + result : result;
}
