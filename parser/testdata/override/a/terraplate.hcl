
exec {
  extra_args = ["a", "b", "c"]

  plan {
    input = true
    lock  = false
    out   = "outoverride"
  }
}
