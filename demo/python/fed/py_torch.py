import io
import torch
import torch.nn.functional as F

import numpy as np

from fed.model import FedModel

def numpy_to_torch(data):
  file_sream = io.BytesIO(data) 
  npDict = np.load(file_sream)
  torchDict = {}
  for key in npDict:
    torchDict[key] = torch.from_numpy(npDict[key])
  return torchDict

def torch_to_numpy(data):
  npDict = {}
  for key in data:
    npDict[key] = data[key].numpy()
  file_stream = io.BytesIO()
  np.savez(file_stream, **npDict)
  return file_stream.getvalue()

class PyTorchModel(FedModel):
  def __init__(self, net, optimizer, test_loader, train_loader):
    self.test_loader = test_loader
    self.train_loader = train_loader
    self.optimizer = optimizer
    self.net = net

  # Adapt raw model parameters to PyTorch model parameters
  def _adapt(self, raw_model):
    return numpy_to_torch(raw_model)

  def train(self, raw_model, parameters):
    netParams = self._adapt(raw_model)
    self.net.load_state_dict(netParams)
    train_data = self.train_loader()
    
    self.net.train()
    epochs = parameters["epochs"] if "epochs" in parameters else 3
    for e in range(epochs):
      print("  [Train] Epoch", e + 1)
      for batch_idx, (data, target) in enumerate(train_data):
        self.optimizer.zero_grad()
        output = self.net(data)
        loss = F.nll_loss(output, target)
        loss.backward()
        self.optimizer.step()

    return torch_to_numpy(self.net.state_dict())

  def evaluate(self, raw_model, parameters):
    netParams = self._adapt(raw_model)
    self.net.load_state_dict(netParams)
    test_data = self.test_loader()

    self.net.eval()
    correct = 0
    test_loss = 0

    with torch.no_grad():
        for data, target in test_data:
            output = self.net(data)
            test_loss += F.nll_loss(output, target, reduction='sum').item()
            pred = output.data.max(1, keepdim=True)[1]
            correct += pred.eq(target.data.view_as(pred)).sum().item()

    total = len(test_data.dataset)
    test_loss /= total


    res =  {
      "total": total,
      "correct": correct,
      "loss": test_loss,
      "accuracy": correct / total * 100
    } 

    return res