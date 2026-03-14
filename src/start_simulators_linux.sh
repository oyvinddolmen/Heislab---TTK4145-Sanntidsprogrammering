#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SIM_DIR="$SCRIPT_DIR/Simulator-v2-master"
SIM_BIN="$SIM_DIR/SimElevatorServer"

if [[ ! -x "$SIM_BIN" ]]; then
  echo "Fant ikke kjørbar simulator på: $SIM_BIN"
  echo "Bygg eller last ned Linux-versjonen først."
  exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  echo "Fant ikke 'go' i PATH."
  exit 1
fi

open_terminal() {
  local title="$1"
  local command="$2"

  if command -v gnome-terminal >/dev/null 2>&1; then
    gnome-terminal --title="$title" -- bash -lc "$command; exec bash"
    return
  fi

  if command -v mate-terminal >/dev/null 2>&1; then
    mate-terminal --title="$title" -- bash -lc "$command; exec bash"
    return
  fi

  if command -v tilix >/dev/null 2>&1; then
    tilix --title="$title" -e "bash -lc \"$command; exec bash\""
    return
  fi

  if command -v x-terminal-emulator >/dev/null 2>&1; then
    x-terminal-emulator -T "$title" -e bash -lc "$command; exec bash"
    return
  fi

  if command -v konsole >/dev/null 2>&1; then
    konsole --new-tab -p tabtitle="$title" -e bash -lc "$command; exec bash"
    return
  fi

  if command -v xfce4-terminal >/dev/null 2>&1; then
    xfce4-terminal --title="$title" --command="bash -lc '$command; exec bash'"
    return
  fi

  if command -v terminator >/dev/null 2>&1; then
    terminator --title="$title" -x bash -lc "$command; exec bash"
    return
  fi

  if command -v lxterminal >/dev/null 2>&1; then
    lxterminal --title="$title" -e "bash -lc '$command; exec bash'"
    return
  fi

  if command -v alacritty >/dev/null 2>&1; then
    alacritty --title "$title" -e bash -lc "$command; exec bash"
    return
  fi

  if command -v kitty >/dev/null 2>&1; then
    kitty --title "$title" bash -lc "$command; exec bash"
    return
  fi

  if command -v xterm >/dev/null 2>&1; then
    xterm -T "$title" -e bash -lc "$command; exec bash"
    return
  fi

  return 1
}

open_tmux() {
  if ! command -v tmux >/dev/null 2>&1; then
    return 1
  fi

  local session="heislab"
  if tmux has-session -t "$session" 2>/dev/null; then
    tmux kill-session -t "$session"
  fi

  tmux new-session -d -s "$session" -n "heislab" "cd '$SIM_DIR' && ./SimElevatorServer --port 15657"
  tmux split-window -h -t "$session:0" "cd '$SIM_DIR' && ./SimElevatorServer --port 15667"
  tmux split-window -v -t "$session:0.0" "cd '$SCRIPT_DIR' && go run . -simPort 15657 -peersPort 20001 -bcastPort 20002 -id 1"
  tmux split-window -v -t "$session:0.1" "cd '$SCRIPT_DIR' && go run . -simPort 15667 -peersPort 20001 -bcastPort 20002 -id 2"
  tmux select-layout -t "$session:0" tiled
  tmux select-pane -t "$session:0.0"
  tmux attach-session -t "$session"
}

if open_terminal "Sim1" "cd '$SIM_DIR' && ./SimElevatorServer --port 15657"; then
  open_terminal "Sim2" "cd '$SIM_DIR' && ./SimElevatorServer --port 15667"

  sleep 2

  open_terminal "Elev1" "cd '$SCRIPT_DIR' && go run . -simPort 15657 -peersPort 20001 -bcastPort 20002 -id 1"
  open_terminal "Elev2" "cd '$SCRIPT_DIR' && go run . -simPort 15667 -peersPort 20001 -bcastPort 20002 -id 2"
  exit 0
fi

if open_tmux; then
  exit 0
fi

echo "Fant ingen støttet terminalemulator, og 'tmux' er heller ikke installert."
echo "Installer for eksempel 'xterm' eller 'tmux', og prøv igjen."
exit 1
