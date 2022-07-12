const strava = require('../shared/strava.js');
const withings = require('../shared/withings.js');

module.exports = async function (context, req) {
  context.log('Auth function processed a request');
  // strava example:
  // https://www.strava.com/oauth/authorize?client_id=76033&response_type=code&redirect_uri=https://ohmyposh.dev/api/auth&approval_prompt=force&scope=read,activity:read&state=strava
  const code = (req.query.code || req.query._code || (req.body && req.body.code));
  const segment = (req.query.state || (req.body && req.body.state));
  let tokens = {
    access_token: '',
    refresh_token: '',
    expires_in: '',
  };
  try {
    if (!code || !segment) {
      context.log(`Issue processing request: missing code (${code}) or segment (${segment})`);
      redirect(context, segment, tokens, 'missing code or segment');
      return;
    }

    context.log(`Processing the ${segment} code`);

    switch (segment) {
      case "strava":
        tokens = await strava.getToken(code);
        break;
      case "withings":
        tokens = await withings.getToken(code);
        break;
      default:
        context.log(`Unknown segment: ${segment}`);
        redirect(context, segment, tokens, `Unknown segment: ${segment}`);
        return;
    }

    redirect(context, segment, tokens, '');
  } catch (error) {
    if (!error.stack) {
      redirect(context, segment, tokens, toBase64(error));
      return;
    }
    context.log(`Error: ${error.stack}`);
    redirect(context, segment, tokens, toBase64(error.stack));
  }
}

function toBase64(str) {
  return Buffer.from(str).toString('base64');
}

function redirect(context, segment, tokens, error) {
  const url = `${process.env['DOCS_LOCATION']}/docs/auth?segment=${segment}&access_token=${tokens.access_token}&refresh_token=${tokens.refresh_token}&expires_in=${tokens.expires_in}&error=${error}`;
  context.res = {
    status: 302,
    headers: {
      Location: url
    },
    body: {}
  }
  context.done();
}
