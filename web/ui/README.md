# ui

This is the gowitness user interface source.

## development notes

Typically, you'd have the API server running using `go run main.go report server`. That starts the API server on port 7171. `npm run dev` will start another web server on another port which means you will struggle to connect to the API. to help, I added an environment variable you can set. In the default case, it would look something like this:

```text
VITE_GOWITNESS_API_BASE_URL=http://localhost:7171 npm run dev
```

