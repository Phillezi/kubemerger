# kubemerger

Finds and merges kubeconfigs into a single kubeconfig.

## Why?

This is neat because you can have many different kubeconfig files and still use tools like [`kubectx`}(https://github.com/ahmetb/kubectx) to switch between the different contexts and namespaces.

## How?

`kubemergerd` is a daemon that is meant to run in the background. This daemon watches the provided root directory (recursively), and will dynamically watch / unwatch on changes to any sub-directory of this root directory, by default the directory is `~/.kube/` and the output file (`~/.kube/config` by default) is automatically ignored from this watchlist, so we don`t feedback loop.

## Installing

Clone the repository and then inside the repo run:

### Arch on the aur

This project is available as [kubemerger-git](https://aur.archlinux.org/packages/kubemerger-git) on the AUR.

#### With yay

```bash
yay -S kubemerger-git
```

#### With paru

```bash
paru -S kubemerger-git
```

#### No aur helper

```bash
git clone  https://aur.archlinux.org/kubemerger-git.git && cd kubemerger-git && makepkg -si && cd ..
```

### Linux + systemd service

> [!NOTE]
> The command below is only for Linux distros with systemd.

```bash
go run build.go && \
  sudo cp ./bin/kubemergerd /usr/local/bin/kubemergerd && \
  mkdir -p ~/.config/systemd/user/ && \
  cp ./dist/kubemergerd.service ~/.config/systemd/user/kubemergerd.service && \
  systemctl --user daemon-reload && \
  systemctl --user enable --now kubemergerd.service
```

### Other

Build and install:

```bash
go run build.go && \
  sudo cp ./bin/kubemergerd /usr/local/bin/kubemergerd

```

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

