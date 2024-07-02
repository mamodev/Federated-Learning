
import sys
from channel import Channel

from nn import get_model

USAGE = "client.py <server_ip> <server_port>"

def main():
  if len(sys.argv) != 3:
    print(USAGE)
    sys.exit(1)

  server_ip = sys.argv[1]
  server_port = int(sys.argv[2])

  with Channel.connect(server_ip, server_port) as channel:
    modelParams = channel.receive_json()
    # modelParams['shuffle'] = True
    # assert modelParams['name'] == 'MNIST', f"Expected model name to be 'MNIST', but got {modelParams['name']}"

    model = get_model(modelParams)
    model.preload_test_set()
    channel.send_ack()

    weights = model.get_def_weights()
    t_weights = weights
    while True:
      cmd = channel.receive_command()
      channel.send_ack()

      if cmd == 'CLOSE':
        return 0
      
      if cmd == 'RETR':
        weights = channel.receive()
        t_weights = model.train_from(weights)

      if cmd == 'COMM':
        channel.send(t_weights)
        weights = t_weights

if __name__ == "__main__":  
    ret = main()  
    exit(ret)