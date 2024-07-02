import os

class FedStorage:
  def store(self, key: str, data: any):
    raise NotImplementedError
  
  def retrieve(self, key: str) -> any:
    raise NotImplementedError

class MemoryStore(FedStorage):
  def __init__(self):
    self.data = {}

  def store(self, key, data):
    self.data[key] = data

  def retrieve(self, key):
    return self.data.get(key, None)
  
class FileStore(FedStorage):
  def __init__(self, folder):
    self.folder = folder

  def store(self, key, data):
    if not os.path.exists(self.folder):
      os.makedirs(self.folder)

    with open(os.path.join(self.folder, key), 'w') as f:
      f.write(data)

  def retrieve(self, key):
    try:
      with open(os.path.join(self.folder, key), 'r') as f:
        return f.read()
    except:
      return None

