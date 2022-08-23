# Enable the default zsh options (those marked with <Z> in `man zshoptions`)
# but without `aliases`. Aliases in functions are expanded when functions are
# defined, so if we disable aliases here, we'll be sure to have no pesky
# aliases in any of our functions. This way we won't need prefix every
# command with `command` or to quote every word to defend against global
# aliases. Note that `aliases` is not the only option that's important to
# control. There are several others that could wreck havoc if they are set
# to values we don't expect. With the following `emulate` command we
# sidestep this issue entirely.
'emulate' 'zsh' '-o' 'no_aliases'

# This brace is the start of try-always block. The `always` part is like
# `finally` in lesser languages. We use it to *always* restore user options.
{

# Bail out if not interactive shell.
[[ -o interactive ]] || return 0

# ALT-B - Paste the final snipet into the command line
__snipedit() {
  local cmd="./snipedit git rebase --onto %commit%^ %commit%"
  setopt pipefail no_aliases 2> /dev/null
  eval $cmd
  local ret=$?
  echo
  return $ret
}

snipedit-widget() {
  LBUFFER="${LBUFFER}$(__snipedit)"
  local ret=$?
  zle reset-prompt
  return $ret
}

zle     -N             snipedit-widget
bindkey -M emacs '\eb' snipedit-widget
bindkey -M vicmd '\eb' snipedit-widget
bindkey -M viins '\eb' snipedit-widget

} always {
  # Restore the original options.
}
