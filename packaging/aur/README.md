# Publishing to the AUR

The `PKGBUILD` here is the source-of-truth copy; the AUR itself is a separate
git remote you push it to. One-time prerequisites, then the release flow.

## One-time setup

1. **Push this repo to GitHub** and make sure the `url=` in `PKGBUILD` matches
   the real repository (also update `module` in `go.mod` if it differs).
2. **Create an AUR account** at <https://aur.archlinux.org> and add your SSH
   public key under *My Account*.
3. **Clone the (new, empty) AUR package repo** — cloning a non-existent
   package name creates it on first push:

   ```sh
   git clone ssh://aur@aur.archlinux.org/cheatsheet-tui.git aur-cheatsheet-tui
   ```

## Releasing a version

```sh
# 1. Tag and push — the GitHub release workflow runs automatically
git tag v0.1.0 && git push --tags

# 2. Update PKGBUILD: set pkgver, reset pkgrel=1, fill the real checksum
cd packaging/aur
updpkgsums                      # from pacman-contrib; rewrites sha256sums

# 3. Test the package build locally (chroot-clean if you have devtools)
makepkg --cleanbuild --syncdeps   # or: pkgctl build
namcap *.pkg.tar.zst              # optional lint

# 4. Regenerate .SRCINFO — the AUR rejects pushes without a current one
makepkg --printsrcinfo > .SRCINFO

# 5. Copy PKGBUILD + .SRCINFO into the AUR clone, commit, push
cp PKGBUILD .SRCINFO ../../../aur-cheatsheet-tui/
cd ../../../aur-cheatsheet-tui
git add PKGBUILD .SRCINFO
git commit -m "v0.1.0"
git push
```

Users then install with `yay -S cheatsheet-tui` (or `paru`, etc.).

## Notes

- `sha256sums=('SKIP')` is a placeholder; the AUR expects real checksums —
  always run `updpkgsums` after tagging.
- The build follows the Arch Go package guidelines (PIE, trimpath,
  readonly modules, external linking) and stamps the version via
  `-X main.version`, so `cheatsheet --version` reports the package version.
- `check()` runs the full test suite (BDD specs + unit tests) during the
  package build.
- A `-bin` variant repackaging the GitHub release binaries is possible later;
  start with this source package — it is what AUR reviewers prefer.
