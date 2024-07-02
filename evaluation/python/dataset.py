import torch
import torch.nn as nn
import torch.optim as optim
import torchvision
from torch.utils.data import Dataset, Subset

MAX_SIZE = 10000

class MNISTCustom (Dataset): 

  def __lable_idxs(self, dataset, label):
    return [i for i, (_, target) in enumerate(dataset) if target == label]

  def __init__(self, mnist):
    self.mnist = Subset(mnist, range(MAX_SIZE))

    self.labels = [i for i in range(10)]
    self.label_indices = [self.__lable_idxs(self.mnist, i) for i in self.labels]

    self.set_params({
      'n_clients': 2,
      'client_idx': 0,
      'bias': 0,
      'noise': 0,
      'knolodge_amount': 0.2,
      'data_round_split': 10,
      'data_round_split_idx': 0
    })

  def set_params(self, params):
    self.n_clients = params['n_clients']
    self.client_idx = params['client_idx']
    self.bias = params['bias']
    self.noise = params['noise']
    self.knolodge_amount = params['knolodge_amount']
    self.data_round_split = params['data_round_split']
    self.data_round_split_idx = params['data_round_split_idx']
    self.n_labels = int(len(self.labels) * (1 - self.bias))

    self.round_ranges = self.get_client_round_ranges(self.client_idx, self.data_round_split_idx)

    self.comp_indexes = [None for _ in range(len(self.round_ranges))]
    curr_idx = 0
    for i in range(len(self.round_ranges)):
      label, r = self.round_ranges[i]
      offset = r[1] - r[0]
      self.comp_indexes[i] = (label, curr_idx, curr_idx + offset - 1, r[0], r[1])
      curr_idx += offset

  def get_label_clients_idxs(self):
    return [list(range(self.n_clients)) for i in range(10)]
    
    lables_clients_idxs = [[] for _ in range(10)]


    curr_label = 0
    for _ in range(self.n_labels):
      for client_idx in range(self.n_clients):
        lables_clients_idxs[curr_label].append(client_idx)
        curr_label += 1 
        curr_label = curr_label % len(self.labels)
 
    return lables_clients_idxs
  

  def get_client_lables(self, idx):
    return [(i, x) for i, x in enumerate(self.get_label_clients_idxs()) if  idx in x]

  def get_client_ranges(self, idx):
    client_labels = self.get_client_lables(idx)

    get_range = lambda size, n, idx: (int(size * self.knolodge_amount / n) * idx, int(size * self.knolodge_amount / n) * (idx + 1))

    return [(label, get_range(len(self.label_indices[label]), len(clients), clients.index(idx))) for label, clients in client_labels]

  def get_client_round_ranges(self, idx, round_idx):
    client_ranges = self.get_client_ranges(idx)

    round_idx = round_idx % self.data_round_split

    get_round_range = lambda start, end: (
        int(start + (end - start) / self.data_round_split * round_idx), 
        int(start + (end - start) / self.data_round_split * (round_idx + 1))
    )

    return [(label, get_round_range(start, end)) for label, (start, end) in client_ranges]

  def compute_len(self):
    return sum([end - start for _, (start, end) in self.round_ranges])
  

  def __len__(self):
    return sum ([end - start for _, (start, end) in self.round_ranges])
  
  def __getitem__(self, idx):
    for label, startIdx, endIdx, start, end in self.comp_indexes:
      if idx >= startIdx and idx <= endIdx:
        real_lable_idx = idx - startIdx + start
        return self.mnist[self.label_indices[label][real_lable_idx]]
      
    raise IndexError("Index out of range for idx", idx, "len", self.current_len)
  

if __name__ == '__main__':
  print("Starting...")
  mnist = torchvision.datasets.MNIST('./data', train=True, download=True, transform=torchvision.transforms.Compose([torchvision.transforms.ToTensor()]))
  print("MNIST dataset loaded")
  mnist = MNISTCustom(mnist)
  print(mnist.get_label_clients_idxs())

  print("Client lables")

  print(mnist.get_client_lables(0))
  print(mnist.get_client_ranges(0))

  print("len(mnist):", len(mnist))
  print("Round ranges")
  for round in range(4):
    print(mnist.get_client_round_ranges(0, round))



  # for i in range(len(mnist)):
  #   item = mnist[i]
  #   print("Item", i, item[1])