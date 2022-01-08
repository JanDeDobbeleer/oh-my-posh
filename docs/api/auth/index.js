const strava = require('../shared/strava.js');

module.exports = async function (context, req) {
  context.log('JavaScript HTTP trigger function processed a request.');
  // strava example:
  // https://www.strava.com/oauth/authorize?client_id=76033&response_type=code&redirect_uri=https://ohmyposh.dev/api/auth&approval_prompt=force&scope=read,activity:read&state=strava

  try {
    const code = (req.query._code || (req.body && req.body.code));
    const segment = (req.query.state || (req.body && req.body.state));
    if (!code || !segment) {
      context.res = {
        status: 400
      };
      return;
    }

    let body = null;
    switch (segment) {
      case "strava":
        body = await strava.getStravaToken(code);
        break;
      default:
        context.res = {
          body: "unknown segment",
          status: 400
        };
        return;
    }

    const url = `https://ohmyposh.dev/docs/auth?segment=${segment}&access_token=${body.access_token}&refresh_token=${body.refresh_token}`;

    const res = {
      status: 302,
      headers: {
        'Location': url
      },
      body: 'Redirecting...'
    };
    context.done(null, res);
  } catch (error) {
    context.log(error);
    context.res = {
      body: error,
      status: 500
    };
  }
}
