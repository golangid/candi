package static

const (
	// PlaygroundAsset template
	PlaygroundAsset = `
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset=utf-8/>
		<meta name="viewport" content="user-scalable=no, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, minimal-ui">
		<link rel="shortcut icon" href="https://graphcool-playground.netlify.com/favicon.png">
		<link rel="stylesheet" href="//cdn.jsdelivr.net/npm/graphql-playground-react@1.7.8/build/static/css/index.css"/>
		<link rel="shortcut icon" href="//cdn.jsdelivr.net/npm/graphql-playground-react@1.7.8/build/favicon.png"/>
		<script src="//cdn.jsdelivr.net/npm/graphql-playground-react@1.7.8/build/static/js/middleware.js"></script>
		<title>Playground</title>
	</head>
	<body>
	<style type="text/css">
		html { font-family: "Open Sans", sans-serif; overflow: hidden; }
		body { margin: 0; background: #172a3a; }
	</style>
	<div id="root"/>
	<script type="text/javascript">
		window.addEventListener('load', function (event) {
			const root = document.getElementById('root');
			root.classList.add('playgroundIn');
			const wsProto = location.protocol == 'https:' ? 'wss:' : 'ws:'
			GraphQLPlayground.init(root, {
				endpoint: location.protocol + '//' + location.host + '/graphql',
				subscriptionsEndpoint: wsProto + '//' + location.host + '/graphql',
				settings: {
					'request.credentials': 'same-origin'
				}
			})
		})
	</script>
	</body>
	</html>`

	// VoyagerAsset template
	VoyagerAsset = `<!DOCTYPE html>
<html>
  <head>
	<style>
	  body {
		height: 100%;
		margin: 0;
		width: 100%;
		overflow: hidden;
	  }
	  #voyager {
		height: 100vh;
	  }
	</style>

	<!--
	  This GraphQL Voyager example depends on Promise and fetch, which are available in
	  modern browsers, but can be "polyfilled" for older browsers.
	  GraphQL Voyager itself depends on React DOM.
	  If you do not want to rely on a CDN, you can host these files locally or
	  include them directly in your favored resource bunder.
	-->
	<script src="https://cdn.jsdelivr.net/es6-promise/4.0.5/es6-promise.auto.min.js"></script>
	<script src="https://cdn.jsdelivr.net/fetch/0.9.0/fetch.min.js"></script>
	<script src="https://cdn.jsdelivr.net/npm/react@16/umd/react.production.min.js"></script>
	<script src="https://cdn.jsdelivr.net/npm/react-dom@16/umd/react-dom.production.min.js"></script>

	<!--
	  These two files are served from jsDelivr CDN, however you may wish to
	  copy them directly into your environment, or perhaps include them in your
	  favored resource bundler.
	 -->
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphql-voyager/dist/voyager.css" />
	<script src="https://cdn.jsdelivr.net/npm/graphql-voyager/dist/voyager.min.js"></script>
  </head>
  <body>
	<div id="voyager">Loading...</div>
	<script>

	  // Defines a GraphQL introspection fetcher using the fetch API. You're not required to
	  // use fetch, and could instead implement introspectionProvider however you like,
	  // as long as it returns a Promise
	  // Voyager passes introspectionQuery as an argument for this function
	  function introspectionProvider(introspectionQuery) {
		// This example expects a GraphQL server at the path /graphql.
		// Change this to point wherever you host your GraphQL server.
		return fetch(location.protocol + '//' + location.host + '/graphql', {
		  method: 'post',
		  headers: {
			'Accept': 'application/json',
			'Content-Type': 'application/json',
		  },
		  body: JSON.stringify({query: introspectionQuery}),
		  credentials: 'include',
		}).then(function (response) {
		  return response.text();
		}).then(function (responseBody) {
		  try {
			return JSON.parse(responseBody);
		  } catch (error) {
			return responseBody;
		  }
		});
	  }

	  // Render <Voyager /> into the body.
	  GraphQLVoyager.init(document.getElementById('voyager'), {
		introspection: introspectionProvider
	  });
	</script>
  </body>
</html>`
)
