import struct
import sys

import torch
import torch.nn as nn
import torch.optim as optim
import torchvision
from torch.utils.data import Dataset, Subset

from network import Net
from fed.py_torch import PyTorchModel

from dataset import MNISTCustom

import socket

import json

# Usage: python client.py <client_name> <controller_address> <controller_port>


# What i need to control: 
# 1. The amount of noise (espressed as % of noise on the dataset 0-1)
# 2. The amount of bias (espressed as % of labels on the dataset 0-1)
# 3. The amount of data (espressed as the size of the dataset)
# 4. The amount of computation (espressed as minimum training error)
# 5. The timing of DOWNLOADING the model & UPLOADING the model
# 6. The amount of different dataset for different iterations

cached_params = None
cached_subset = None

def notSameParams(p1, p2):
  return p1['n_clients'] != p2['n_clients'] or p1['client_idx'] != p2['client_idx']

def get_subset(dataset, params):
  global cached_params
  global cached_subset

  if cached_params is None or notSameParams(cached_params, params):
    cached_params = params
    n_clients = params['n_clients']
    client_idx = params['client_idx']

    start_idx = (len(dataset) // n_clients) * client_idx
    end_idx = (len(dataset) // n_clients) * (client_idx + 1)

    cached_subset = Subset(dataset, range(start_idx, end_idx))
  
  return cached_subset


def mock_process_model(model, params, ds):

  ds = get_subset(ds, params)
  # ds.set_params(params)
  net = Net()
  optimizer = optim.SGD(net.parameters(), lr=0.01, momentum=0.9)

  # print("Training with size: ", len(ds), "split", ds.round_ranges)

  def loader_fact():
    return torch.utils.data.DataLoader(ds, batch_size=params["batch_size"], shuffle=False)

  net = PyTorchModel(net, optimizer, loader_fact, loader_fact)

  model = net.train(model, params)

  return model

def main():
  if len(sys.argv) < 3:
      print("Usage: python client.py <client_name> <controller_address> <controller_port>")
      sys.exit(1)


  client_name = sys.argv[1]
  controller_address = sys.argv[2]
  controller_port = int(sys.argv[3])

  cotroller_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
  cotroller_socket.connect((controller_address, controller_port))

  # TCP PACKET STRUCTURE (BIG ENDIAN): uint32_t length, byte[] data

  sendData(cotroller_socket, client_name.encode('latin-1'))
  ack = reciveData(cotroller_socket)

  if ack.decode('latin-1') != client_name:
    print(f"Error: Client name mismatch. Expected {client_name} but got {ack.decode('latin-1')}")
    sys.exit(1)

  print("Connected to controller")

  dataset = torchvision.datasets.MNIST(
      './data/', 
      train=True, 
      download=True, 
      transform=torchvision.transforms.Compose([
          torchvision.transforms.ToTensor(),
          torchvision.transforms.Normalize((0.1307,), (0.3081,))
      ])
  )
  

  # dataset = MNISTCustom(dataset)
  dataset = Subset(dataset, range(1000))

  # print("Dataset size: ", len(dataset))
  # filtered_datasets = [filter_dataset(dataset, i) for i in range(10)]
  # for i, ds in enumerate(filtered_datasets):
  #   print(f"Dataset {i} size: ", len(ds))
  
    
  # 1. The controller tells parameters
  # 2. Get the model from agg_server
  # 3. Process the model
  # 4. Wait for the controller to tell to send the model
  # 5. Send the model to the agg_server
  # 6. Send to the controller an ok message
  # 7. Repeat from 1

  while True:
    # 1. The controller tells parameters
    params = json.loads(reciveData(cotroller_socket).decode('latin-1'))
    if "exit" in params:
      break

    # print("Received parameters: ", params)

    address = (params['agg_host'], params['agg_port'])
    agg_sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    agg_sock.connect(address)

    # 2. Get the model from agg_server
    request_type = 1
    sendData(agg_sock, struct.pack('>B', request_type))

    # Receive model length (4 bytes)
    model = reciveData(agg_sock)
    agg_sock.close()

    # 3. Process the model
    processed_model = mock_process_model(model, params, dataset)

    # 4. Wait for the controller to tell to send the model
    cmd = reciveData(cotroller_socket)
    if cmd.decode('latin-1') != "SEND":
      raise Exception("Error: Expected SEND message from controller")


    # print("Connecting to agg_server...")
    agg_sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    agg_sock.connect(address)
    # print("Connected to agg_server")

    # 5. Send the model to the agg_server (request_type = 2, data = model)
    request_type = 2
    sendData(agg_sock, struct.pack('>B', request_type) + processed_model)

    # print("Model sent to agg_server")

    ack = reciveData(agg_sock)
    if ack.decode('latin-1') != "OK":
      raise Exception("Error: Expected OK message from agg_server")
    


    agg_sock.close()

    # 6. Send to the controller an ok message
    sendData(cotroller_socket, "OK".encode('latin-1'))


def reciveAll(s, length):
  readed = 0
  data = b''
  while readed < length:
      rec = s.recv(length - readed)
      readed += len(rec)
      data += rec

  return data

def reciveData(s, debug=False):
  # Receive data length
  if debug:
    print("Receiving data length...")

  length_bytes = reciveAll(s, 4)
  length = struct.unpack('>I', length_bytes)[0]

  if debug:
    print(f"Data length: {length}")

  # Receive data
  data = reciveAll(s, length)

  if debug:
    print(f"Data received: {len(data)} bytes")

  return data


def sendData(s, data):
  length = len(data)
  s.sendall(struct.pack('>I', length))
  s.sendall(data)

def filter_dataset(dataset, label):
  idxs =  [i for i, (_, target) in enumerate(dataset) if target == label]
  return Subset(dataset, idxs)


if __name__ == "__main__":
    main()
    print("Client terminated")