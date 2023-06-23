const strava = require('../shared/strava.js');
const withings = require('../shared/withings.js');

module.exports = async function (context, req) {
  context.log('Refresh function processed a request');
  // strava example:
  // https://ohmyposh.dev/api/refresh?segment=strava&token=<refresh_token>

  try {
    const refresh_token = (req.query.token || (req.body && req.body.token));
    const segment = (req.query.segment || (req.body && req.body.segment));
    if (!refresh_token || !segment) {
      context.res = {
        status: 400
      };
      return;
    }

    context.log(`Refreshing the ${segment} token`);
    let body = null;
    switch (segment) {
      case "strava":
        body = await strava.refreshToken(refresh_token);
        break;
      case "withings":
        body = await withings.refreshToken(refresh_token);
        break;
      default:
        context.log(`Unknown segment: ${segment}`);
        context.res = {
          body: "Unknown segment",
          status: 400
        };
        return;
    }

    context.res.json(body);
  } catch (error) {
    context.log(error);
    context.res = {
      body: {
        "message": (error.message) ? error.message : "unable to refresh token"
      },
      status: 500
    };
  }
}
