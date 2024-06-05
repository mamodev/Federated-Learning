#!/bin/bash


# Check if the required number of panels is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <number_of_client_panels>"
  exit 1
fi

# Number of client panels
clients=$1
n=clients
n=$(($n + 2))

# Session name
session="simulation"

tmux kill-session -t $session
tmux new-session -d -s $session

tmux rename-window -t $session:0 'main'

tmux set pane-border-status top
tmux set pane-border-format "[#T]"

tmux select-pane -T "Server"
tmux send-keys "./server $clients" C-m



tmux split-window 
tmux select-pane -T "Oracle"
tmux send-keys "cd python" C-m
tmux send-keys "source .venv/bin/activate" C-m
tmux send-keys "clear" C-m
tmux send-keys "python oracle.py" C-m


for i in $(seq 3 $n)
do
  tmux split-window 
  panel_id=$(($i - 2))
  
  tmux select-pane  -T "Client: $panel_id"

  tmux send-keys "cd python" C-m
  tmux send-keys "source .venv/bin/activate" C-m
  tmux send-keys "clear" C-m
  tmux send-keys "python main.py $panel_id 1 2 3 4 5 6 7 8 9" C-m

  tmux select-layout tile
done


tmux attach -t $session