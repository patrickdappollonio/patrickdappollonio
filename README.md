### Hello! 👋 Welcome to my Github Profile

My name is Patrick D'appollonio, I'm a technical architect and a staff engineer for [Netlify](https://www.netlify.com/). I work mostly with Go and Kubernetes in my day-to-day, and as such, you'll see a few tools below I've built over time to solve personal itches. If you see them valuable or have any feedback, please do not hesitate and create an issue or message me on Twitter: [@marlex](https://twitter.com/marlex)

### Noteworthy projects 🧑‍💻

* [`kubectl-slice`](https://github.com/patrickdappollonio/kubectl-slice) - have you ever seen a YAML with those triple-dash separators and hoped there would be a tool to split them into individual YAML files? `kubectl-slice` is a `krew` plugin (or a standalone app if you prefer) that can split multi-YAML files into individual files.
* [`tabloid`](https://github.com/patrickdappollonio/tabloid) - whenever I'm working in things like a Kubernetes cluster I often run `kubectl api-resources | grep Application` and my brain hurts after seeing that `grep` kept the original spacing between resources. `tabloid` is very tiny utility that can re-render a table output and filter it as required.
* [`http-server`](https://github.com/patrickdappollonio/http-server) - a simple HTTP server with optional directory listing and markdown rendering support.
* [`tgen`](https://github.com/patrickdappollonio/tgen) - a standalone `consul-template` or mini `helm` tool to render Go templates... Without the complexity of either.
* [`dotenv`](https://github.com/patrickdappollonio/dotenv) - with `dotenv` you can quickly run commands with environment variables set from files, like `dotenv -e sample 'echo $HELLO'` where `sample` is a `sample.env` file with `HELLO=world` in it.
* [`wait-for`](https://github.com/patrickdappollonio/wait-for) - my own personal attempt at creating a typical init container application for Kubernetes that waits before running the main container until a service is available on a given endpoint.

### Additional stats 📈

![Github Stats](https://github-readme-stats.vercel.app/api?username=patrickdappollonio&theme=blue-green)

![Most used languages](https://github-readme-stats.vercel.app/api/top-langs/?username=patrickdappollonio&theme=blue-green)
