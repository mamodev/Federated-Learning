class Protocols:
  HTTP = "http"

class Task:
  TRAIN = "train"
  EVALUATE = "evaluate"
  PREDICT = "predict"

def dict_to_params(dict):
  params = []
  for key in dict:
    t = type(dict[key])
    if t is str:
      params.append({"name": key, "type": "text", "value": dict[key]})
    elif t is int or t is float:
      params.append({"name": key, "type": "numeric", "value": dict[key]})
    else:
      raise ValueError(f"Unsupported type {t}")
  return params

class FedModel: 
  def train(self, raw_model, parameters):
    raise NotImplementedError
  
  def evaluate(self, raw_model, parameters):
    raise NotImplementedError
  
  def predict(self, input):
    raise NotImplementedError
  
