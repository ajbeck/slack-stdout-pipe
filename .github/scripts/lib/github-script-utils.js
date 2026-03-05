/**
 * Mock utilities for local testing of GitHub scripts.
 *
 * These utilities allow scripts to run both in GitHub Actions (via actions/github-script)
 * and locally for testing/debugging.
 */

/**
 * Creates a mock core object compatible with @actions/core.
 *
 * @param {Object} [options]
 * @param {Object} [options.outputs] - Object to store outputs (for inspection in tests)
 * @returns {Object} Mock core object
 */
function createMockCore(options = {}) {
  const outputs = options.outputs || {};

  return {
    getInput: (name) => process.env[`INPUT_${name.toUpperCase()}`] || '',
    setOutput: (name, value) => {
      outputs[name] = value;
      console.log(`::set-output name=${name}::${value}`);
    },
    setFailed: (message) => {
      console.error(`::error::${message}`);
      if (typeof process !== 'undefined' && !process.env.GITHUB_ACTIONS) {
        throw new Error(message);
      }
    },
    error: (message) => console.error(`::error::${message}`),
    warning: (message) => console.warn(`::warning::${message}`),
    info: (message) => console.log(message),
    debug: (message) => {
      if (process.env.ACTIONS_STEP_DEBUG === 'true') {
        console.log(`::debug::${message}`);
      }
    },
    _outputs: outputs,
  };
}

/**
 * Creates a mock context object compatible with actions/github-script.
 *
 * @param {Object} [overrides] - Override specific context properties
 * @returns {Object} Mock context object
 */
function createMockContext(overrides = {}) {
  return {
    repo: {
      owner: process.env.GITHUB_REPOSITORY_OWNER || 'ajbeck',
      repo: process.env.GITHUB_REPOSITORY?.split('/')[1] || 'slack-stdout-pipe',
    },
    sha: process.env.GITHUB_SHA || 'HEAD',
    ref: process.env.GITHUB_REF || 'refs/heads/main',
    runId: parseInt(process.env.GITHUB_RUN_ID, 10) || 0,
    workflow: process.env.GITHUB_WORKFLOW || 'local-test',
    actor: process.env.GITHUB_ACTOR || 'local-user',
    eventName: process.env.GITHUB_EVENT_NAME || 'workflow_dispatch',
    payload: {},
    ...overrides,
  };
}

/**
 * Creates a mock github (Octokit) client for testing.
 *
 * @param {Object} [stubs] - Override specific API methods
 * @returns {Object} Mock github client
 */
function createMockGithub(stubs = {}) {
  const paginate = async (method, options) => {
    if (stubs.paginate) {
      return stubs.paginate(method, options);
    }
    return [];
  };

  return {
    rest: {
      git: {
        getRef: stubs.getRef || (async () => { throw { status: 404 }; }),
        createRef: stubs.createRef || (async () => ({ data: {} })),
        updateRef: stubs.updateRef || (async () => ({ data: {} })),
      },
      actions: {
        listArtifactsForRepo: stubs.listArtifactsForRepo || (async () => ({ data: { artifacts: [] } })),
        getWorkflowRun: stubs.getWorkflowRun || (async () => ({ data: {} })),
      },
      repos: {
        createRelease: stubs.createRelease || (async () => ({ data: { id: 1, html_url: 'https://example.com/release' } })),
        uploadReleaseAsset: stubs.uploadReleaseAsset || (async () => ({ data: {} })),
      },
    },
    paginate,
  };
}

/**
 * Detects if running in GitHub Actions environment.
 *
 * @returns {boolean} True if running in GitHub Actions
 */
function isGitHubActions() {
  return process.env.GITHUB_ACTIONS === 'true';
}

module.exports = {
  createMockCore,
  createMockContext,
  createMockGithub,
  isGitHubActions,
};
