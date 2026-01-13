# kubemerger

Finds and merges kubeconfigs into a single kubeconfig.

## Why?

This is neat because you can have many different kubeconfig files and still use tools like [`kubectx`}(https://github.com/ahmetb/kubectx) to switch between the different contexts and namespaces.

## How?

`kubemergerd` is a daemon that is meant to run in the background. This daemon watches the provided root directory (recursively), and will dynamically watch / unwatch on changes to any sub-directory of this root directory, by default the directory is `~/.kube/` and the output file (`~/.kube/config` by default) is automatically ignored from this watchlist, so we don`t feedback loop.

## Building

In the cloned repository there is a build script that will just simplifies building by setting the version from the git tag.

```bash
go run build.go
```

## Testing

The same build script can be used to run the tests.

```bash
go run build.go test
```

