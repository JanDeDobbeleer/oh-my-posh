// GET https://${MASTODON_DOMAIN}/.well-known/webfinger?resource=acct:${MASTODON_USER}@${MASTODON_DOMAIN}

function trimPrefix(str, prefix) {
  if (str.startsWith(prefix)) {
    return str.slice(prefix.length)
  }
  return str
}

function webfinger(username) {
  return {
    "subject": `acct:${username}@hachyderm.io`,
    "aliases": [
      `https://hachyderm.io/@${username}`,
      `https://hachyderm.io/users/${username}`
    ],
    "links": [
      {
        "rel": "http://webfinger.net/rel/profile-page",
        "type": "text/html",
        "href": `https://hachyderm.io/@${username}`
      },
      {
        "rel": "self",
        "type": "application/activity+json",
        "href": `https://hachyderm.io/users/${username}`
      },
      {
        "rel": "http://ostatus.org/schema/1.0/subscribe",
        "template": "https://hachyderm.io/authorize_interaction?uri={uri}"
      }
    ]
  };
}

module.exports = async function (context, req) {
  context.log('Webfinger function processed a request');

  try {
    let resource = req.query.resource;
    if (!resource) {
      resource = "jan@social.ohmyposh.dev";
    }

    resource = trimPrefix(resource, "acct:");

    context.log(`Creating Mastodon user info for ${resource}`);

    let body;

    switch (resource) {
      case "releasebot@social.ohmyposh.dev":
        body = webfinger("releasebot");
        break;
      case "jan@social.ohmyposh.dev":
        body = webfinger("jandedobbeleer");
        break;
      default:
        context.log(`Unknown resource: ${resource}`);
        context.res = {
          body: "Unknown resource",
          status: 400
        };
        return;
    }

    // return body as application/jrd+json
    context.res = {
      body: body,
      headers: {
        "Content-Type": "application/jrd+json"
      }
    };
  } catch (error) {
    context.log(error);
    context.res = {
      body: {
        "message": (error.message) ? error.message : "unable to create webfinger"
      },
      status: 500
    };
  }
}
