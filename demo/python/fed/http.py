import requests
import time

from fed.model import Task, Protocols
from fed.client import FedClient

class HttpFedClient(FedClient):
  def __init__(self, url, store, model, params):
    self.url = url
    self.store = store
    self.initialized = False
    self.subscribed = False
    self.model = model
    self.params = params
  
  def _initialize(self):
    if self.initialized:
      return

    group_key = self.store.retrieve("group_token")
    if group_key is None:
      raise Exception("Group key not found")
    
    self.group_key = group_key
    
    client_token = self.store.retrieve("client_token")
    if client_token is None:
      json = {}
      if self.params is not None:
        json = { 
          "params": self.params
        }

      response = requests.post(f"{self.url}/register", headers={"Group": self.group_key}, json=json)  
      if response.status_code != 200:
        raise Exception("Failed to register client")
      

      body = response.json()
      print("response", body)
      if not "token" in body:
        raise Exception("Token not found in response")

      client_token = body["token"]
      self.store.store("client_token", client_token)

    self.client_token = client_token
    self.initialized = True
   

  def subscribe(self):
    self._initialize()
    print("Subscribing...")
   
    body = {
      "tasks": [
          {
              "type": Task.TRAIN, 
              "protocols": [Protocols.HTTP]
          },
          {
              "type": Task.EVALUATE, 
              "protocols": [Protocols.HTTP]
          }
      ]
    }

    response = requests.post(f"{self.url}/subscribe", headers={"Group": self.group_key, "Authorization": self.client_token}, json=body)

    if response.status_code != 200:
      err_str = "Failed to subscribe with error code: " + str(response.status_code) + " and message: " + response.text
      raise Exception(err_str)

    self.subscribed = True
    print("Subscribed")

    # loop until not subscribed
    while self.subscribed:
      print("Checking for tasks...")
      response = requests.get(f"{self.url}/task", headers={"Group": self.group_key, "Authorization": self.client_token})
      if response.status_code != 200:
        time.sleep(1)
        continue  

      task = response.json()
      print(f"Received task: {task}")
      coordinatorUrl = f"http://{task['host']}:{task['port']}/task/{task['token']}"
      
      response = requests.get(coordinatorUrl, headers={"Authorization": self.client_token})
      if response.status_code != 200:
        print("Failed to get task params", response.status_code, response.text)
        print(self.client_token)
        continue


      task_type = task['type']
      payload = response.content

      if task_type == Task.TRAIN:
        self._train(task, payload)
      elif task_type == Task.EVALUATE:
        self._evaluate(task, payload)
      else:
        print("Unknown task type: ", task_type)
    
  
  def _train(self, task, payload):
    coordinatorUrl = f"http://{task['host']}:{task['port']}/task/{task['token']}"

    print(" ==== Training model ====")
    newModel = self.model.train(payload, {})
    print(" ==== Model Trained ====")

    requests.post(coordinatorUrl, headers={"Authorization": self.client_token, "Content-Type": "application/octet-stream"}, data=newModel)

  def _evaluate(self, task, payload):
    coordinatorUrl = f"http://{task['host']}:{task['port']}/task/{task['token']}"

    print(" ==== Evaluating model ====")
    evaldata = self.model.evaluate(payload, {})
    print(" ==== Model Evaluated ====")

    requests.post(coordinatorUrl, headers={"Authorization": self.client_token, "Content-Type": "application/octet-stream"}, json=evaldata)


  def unsubscribe(self):
    self._initialize()
    response = requests.post(f"{self.url}/unsubscribe", headers={"Group": self.group_key, "Authorization": self.client_token}, json={})
    if response.status_code != 200:
      raise Exception("Failed to unsubscribe")
    
    self.subscribed = False