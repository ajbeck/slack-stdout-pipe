/**
 * Find the commit where VERSION file matches the requested version constraint.
 *
 * This script walks git history to find the appropriate commit for a release.
 * It supports version constraints like:
 * - "1" - Latest 1.x.x version
 * - "1.1" - Latest 1.1.x version
 * - "1.1.6" - Exact version 1.1.6
 * - "" (empty) - Use HEAD
 *
 * Usage in GitHub Actions:
 *   const { main } = require('./.github/scripts/find-release-commit');
 *   return await main({ context, github, core });
 *
 * Usage locally:
 *   VERSION_CONSTRAINT=0.1 node find-release-commit.js
 */

const { execSync } = require('child_process');
const semver = require('semver');

/**
 * Builds a semver range from a version constraint.
 *
 * @param {string} constraint - Version constraint (e.g., "1", "1.1", "1.1.6")
 * @returns {string} Semver range string
 *
 * @example
 * buildVersionRange("1")     // ">=1.0.0 <2.0.0"
 * buildVersionRange("1.1")   // ">=1.1.0 <1.2.0"
 * buildVersionRange("1.1.6") // "1.1.6"
 */
function buildVersionRange(constraint) {
  if (!constraint) {
    return '*';
  }

  const cleaned = constraint.startsWith('v') ? constraint.slice(1) : constraint;
  const parts = cleaned.split('.');

  if (parts.length === 1) {
    const major = parseInt(parts[0], 10);
    if (isNaN(major)) {
      throw new Error(`Invalid version constraint: ${constraint}`);
    }
    return `>=${major}.0.0 <${major + 1}.0.0`;
  }

  if (parts.length === 2) {
    const major = parseInt(parts[0], 10);
    const minor = parseInt(parts[1], 10);
    if (isNaN(major) || isNaN(minor)) {
      throw new Error(`Invalid version constraint: ${constraint}`);
    }
    return `>=${major}.${minor}.0 <${major}.${minor + 1}.0`;
  }

  if (!semver.valid(cleaned)) {
    throw new Error(`Invalid version constraint: ${constraint}`);
  }
  return cleaned;
}

/**
 * Checks if a version matches a constraint.
 *
 * @param {string} version - The version to check (e.g., "1.1.6")
 * @param {string} constraint - The version constraint (e.g., "1.1")
 * @returns {boolean} True if version matches constraint
 */
function versionMatchesConstraint(version, constraint) {
  if (!version) return false;

  const cleanedVersion = version.trim().replace(/^v/, '');

  if (!semver.valid(cleanedVersion)) {
    return false;
  }

  const range = buildVersionRange(constraint);
  return semver.satisfies(cleanedVersion, range);
}

/**
 * Gets the VERSION file content at a specific commit.
 *
 * @param {string} commitSha - The commit SHA to check
 * @returns {string|null} The version string or null if not found
 */
function getVersionAtCommit(commitSha) {
  try {
    const version = execSync(`git show ${commitSha}:VERSION 2>/dev/null`, {
      encoding: 'utf-8',
      stdio: ['pipe', 'pipe', 'pipe'],
    });
    return version.trim();
  } catch {
    return null;
  }
}

/**
 * Gets the list of commits from HEAD going back in history.
 *
 * @param {number} [maxCommits=500] - Maximum number of commits to retrieve
 * @returns {string[]} Array of commit SHAs
 */
function getCommitHistory(maxCommits = 500) {
  try {
    const output = execSync(`git log --format=%H -n ${maxCommits}`, {
      encoding: 'utf-8',
    });
    return output.trim().split('\n').filter(Boolean);
  } catch (error) {
    throw new Error(`Failed to get commit history: ${error.message}`);
  }
}

/**
 * Find the oldest commit that has the latest version matching a constraint.
 *
 * Algorithm (two-phase):
 * 1. Walk commits from newest to oldest (HEAD first)
 * 2. Find the first commit matching the constraint (this is the "latest matching version")
 * 3. Continue walking to find the oldest commit with that exact version
 * 4. Stop when the version changes (assumes contiguous version history)
 *
 * @param {Array<{commit: string, version: string|null}>} commitVersions - Array of commit/version pairs, newest first
 * @param {function(string): boolean} matchesConstraint - Function to check if a version matches the constraint
 * @returns {{commit: string, version: string}|null} The oldest commit with the latest matching version, or null
 */
function findOldestMatchingCommit(commitVersions, matchesConstraint) {
  let latestMatchingVersion = null;
  let oldestMatchingCommit = null;

  for (const { commit, version } of commitVersions) {
    if (version === null || version === undefined) {
      if (latestMatchingVersion !== null) {
        break;
      }
      continue;
    }

    if (latestMatchingVersion === null) {
      if (matchesConstraint(version)) {
        latestMatchingVersion = version;
        oldestMatchingCommit = { commit, version };
      }
    } else {
      if (version === latestMatchingVersion) {
        oldestMatchingCommit = { commit, version };
      } else {
        break;
      }
    }
  }

  return oldestMatchingCommit;
}

/**
 * Finds the commit where VERSION matches the constraint.
 *
 * @param {string} constraint - Version constraint (e.g., "1", "1.1", "1.1.6")
 * @returns {{ commitSha: string, version: string }} The matching commit and version
 */
function findVersionCommit(constraint) {
  if (!constraint) {
    const headSha = execSync('git rev-parse HEAD', { encoding: 'utf-8' }).trim();
    const version = getVersionAtCommit(headSha);
    if (!version) {
      throw new Error('No VERSION file found at HEAD');
    }
    return { commitSha: headSha, version };
  }

  const range = buildVersionRange(constraint);
  const commits = getCommitHistory();

  const commitVersions = commits.map((sha) => ({
    commit: sha,
    version: getVersionAtCommit(sha),
  }));

  const result = findOldestMatchingCommit(commitVersions, (version) =>
    versionMatchesConstraint(version, constraint),
  );

  if (result) {
    return { commitSha: result.commit, version: result.version };
  }

  throw new Error(`No commit found matching version constraint: ${constraint} (range: ${range})`);
}

/**
 * Main entry point for GitHub Actions.
 *
 * @param {Object} params
 * @param {Object} params.context - GitHub Actions context
 * @param {Object} params.github - Octokit client
 * @param {Object} params.core - @actions/core
 */
async function main({ context, github, core }) {
  const constraint = process.env.VERSION_CONSTRAINT || '';

  console.log(`Version constraint: "${constraint || '(empty - using HEAD)'}"`);

  try {
    const { commitSha, version } = findVersionCommit(constraint);

    console.log(`Found matching commit: ${commitSha}`);
    console.log(`Version: ${version}`);

    core.setOutput('commit_sha', commitSha);
    core.setOutput('version', version);

    return { commitSha, version };
  } catch (error) {
    core.setFailed(error.message);
    throw error;
  }
}

// CLI mode: run directly with node
if (require.main === module) {
  const { createMockCore } = require('./lib/github-script-utils');

  const core = createMockCore();
  const context = {};
  const github = {};

  main({ context, github, core })
    .then((result) => {
      console.log('\nResult:', JSON.stringify(result, null, 2));
    })
    .catch((error) => {
      console.error('\nFailed:', error.message);
      process.exit(1);
    });
}

module.exports = {
  buildVersionRange,
  versionMatchesConstraint,
  getVersionAtCommit,
  getCommitHistory,
  findOldestMatchingCommit,
  findVersionCommit,
  main,
};
