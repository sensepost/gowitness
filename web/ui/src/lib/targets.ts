// Helpers for turning free-form user input (pasted lists, uploaded files)
// into a clean list of URLs to submit as probe targets.

const allowedSchemes = ['http:', 'https:'];

export interface ParsedTargets {
  // unique, normalised urls ready to submit
  urls: string[];
  // input entries that could not be turned into a valid url
  invalid: string[];
  // number of entries dropped because they were already in the list
  duplicates: number;
}

// expandCandidate turns a single candidate into one or more urls.
//
// Candidates that already have a scheme are kept as is (after validation).
// Candidates without a scheme (e.g. "example.com" or "10.0.0.1:8080") are
// expanded to both http:// and https://, matching the default behaviour of
// `gowitness scan file`.
const expandCandidate = (candidate: string): string[] | null => {
  const hasScheme = candidate.includes('://');
  const withScheme = hasScheme ? [candidate] : [`http://${candidate}`, `https://${candidate}`];

  for (const raw of withScheme) {
    let parsed: URL;
    try {
      parsed = new URL(raw);
    } catch {
      return null;
    }

    if (!allowedSchemes.includes(parsed.protocol) || parsed.hostname === '') {
      return null;
    }
  }

  // return what the user typed (plus a scheme if needed) rather than the
  // URL-normalised form, so we don't surprise them with e.g. added slashes
  return withScheme;
};

// parseTargets parses newline separated targets, one target per line.
// Empty lines and lines starting with '#' are ignored.
export const parseTargets = (input: string): ParsedTargets => {
  const lines = input.split(/\r?\n/);

  const seen = new Set<string>();
  const urls: string[] = [];
  const invalid: string[] = [];
  let duplicates = 0;

  for (const line of lines) {
    const candidate = line.trim();
    if (candidate === '' || candidate.startsWith('#')) continue;

    // urls can't contain unencoded whitespace. flag these lines rather than
    // guessing, so stray text doesn't silently turn into bogus hostnames.
    const expanded = /\s/.test(candidate) ? null : expandCandidate(candidate);
    if (expanded === null) {
      invalid.push(candidate);
      continue;
    }

    for (const url of expanded) {
      if (seen.has(url)) {
        duplicates++;
        continue;
      }
      seen.add(url);
      urls.push(url);
    }
  }

  return { urls, invalid, duplicates };
};
