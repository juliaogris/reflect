// Whether Hermit should manage Git.
manage-git = false
env = {
  GOBIN : "${HERMIT_ENV}/out/bin",
  PATH : "${GOBIN}:${PATH}",
}
