package apparmor

const template = `
$EXECUTABLE flags=(attach_disconnected) {
  #include <abstractions/base>

  $CAPS

  network,
  mount,

  $PATH mrwix,
}
`
