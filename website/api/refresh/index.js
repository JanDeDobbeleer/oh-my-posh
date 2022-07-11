const strava = require('../shared/strava.js');
const withings = require('../shared/withings.js');

module.exports = async function (context, req) {
  context.log('JavaScript HTTP trigger function processed a request.');
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

    let body = null;
    switch (segment) {
      case "strava":
        body = await strava.refreshToken(refresh_token);
        break;
      case "withings":
        body = await withings.refreshToken(refresh_token);
        break;
      default:
        context.res = {
          body: "unknown segment",
          status: 400
        };
        return;
    }

    context.res.json(body);
  } catch (error) {
    context.log(error);
    context.res = {
      body: error,
      status: 500
    };
  }
}
