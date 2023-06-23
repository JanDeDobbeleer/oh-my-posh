const { BskyAgent } = require('@atproto/api');
const {Octokit} = require("@octokit/rest");

(async function main () {
  const github = new Octokit();
  const response = await github.rest.repos.getLatestRelease({
    owner: process.env.OWNER,
    repo: process.env.REPO,
  });
  const release = response.data;

  let notes = release.body;

  // replace all non-supported characters
  notes = notes.replaceAll('### ', '');
  notes = notes.replaceAll('**', '');
  notes = notes.replace(/ \(\[[0-9a-z]+\]\(.*\)/g, '');
  notes = notes.trim();

  const agent = new BskyAgent({ service: 'https://bsky.social' });
  await agent.login({ identifier: process.env.BLUESKY_IDENTIFIER, password: process.env.BLUESKY_PASSWORD });

  const version = release.name;

  const text = `📦 ${version}

${notes}

#ohmyposh #oss #cli #opensource`;

  console.log(`Posting to Bluesky:\n\n${text}`);

  await agent.post({
    text: text,
    embed: {
      $type: 'app.bsky.embed.external',
      external: {
        uri: `https://github.com/JanDeDobbeleer/oh-my-posh/releases/tag/${version}`,
        title: "The best release yet 🚀",
        description: version,
      },
    },
  });
})();
