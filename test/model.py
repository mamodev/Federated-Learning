import io
import torch
from torch import nn

import numpy as np


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
    npDict[key] = data[key].cpu().numpy()
  file_stream = io.BytesIO()
  np.savez(file_stream, **npDict)
  return file_stream.getvalue()



class Model():
    def __init__(self):
      pass

    def train_from(self, weights):
        pass
    
    def evaluate(self, weights):
        pass
    
    def get_def_weights(self):
        pass