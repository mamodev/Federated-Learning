from http.server import BaseHTTPRequestHandler, HTTPServer
from urllib.parse import urlparse, parse_qs
import json
import subprocess 
import os
import fcntl
import time
import logging

class NullLogger(logging.Logger):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def _log(self, *args, **kwargs):
        pass

# Set the http.server loggers to use the dummy logger
logging.getLogger('http.server').setLevel(logging.CRITICAL)
logging.getLogger('http.server').handlers = [logging.NullHandler()]

BASE_PORT = 8081

running_job = None
running_proc = None
running_proc_stderr = None

jobs_queue = []

def __check_job():
  global running_job, running_proc, running_proc_stderr, jobs_queue, BASE_PORT

  if not running_job:
    return

  ret_code = running_proc.poll()

  if ret_code is not None:
    print(f"Job {running_job['name']} finished with code {ret_code}")
    running_job = None
    running_proc = None

def forward_job():
  __check_job()
  global running_job, running_proc, running_proc_stderr, jobs_queue, BASE_PORT
  if not running_job and len(jobs_queue) > 0:
      running_job = jobs_queue.pop(0)
      print(f"Starting job: {running_job['name']}")
      running_proc = subprocess.Popen(["python3", "coordinator.py", str(BASE_PORT)], stdin=subprocess.PIPE)
      
      BASE_PORT += 1
      if BASE_PORT > 8090:
        BASE_PORT = 8081

      config = json.dumps(running_job).encode()
      running_proc.stdin.write(config)
      running_proc.stdin.write(b"\n")
      running_proc.stdin.flush()
      running_proc.stdin.close()

def run_simulation(c):
  global running_job, running_proc, running_proc_stderr, jobs_queue, BASE_PORT
  jobs_queue.append(c)  
  forward_job()


def get_job_status(name):
  global running_job, running_proc, running_proc_stderr, jobs_queue, BASE_PORT

  if running_job and running_job['name'] == name:
    forward_job()
    if not running_job or running_job['name'] != name:
      return {'status': 'finished'}
    
    return {'status': 'running'}

  else:
    for j in jobs_queue:
      if j['name'] == name:
        return {'status': 'queued'}
  
  return {'status': 'not_found'}  


class NoLoggingHTTPServer(HTTPServer):
    def __init__(self, server_address, handler_class):
        super().__init__(server_address, handler_class)
        self.RequestHandlerClass.log_message = lambda *args: None  # Override log_message method to do nothing


class SimpleHTTPRequestHandler(BaseHTTPRequestHandler):
    
    def do_OPTIONS(self):
        self.send_response(200)
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
        self.send_header("Access-Control-Allow-Headers", "Content-type")
        self.end_headers()

    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)

        if self.path == '/run':
          body = json.loads(post_data)
          run_simulation(body)     

          self.send_response(200)
          self.send_header('Content-type', 'application/json')
          self.send_header('Access-Control-Allow-Origin', '*')
          self.send_header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
          self.send_header("Access-Control-Allow-Headers", "Content-type")

          self.end_headers()
          self.wfile.write(json.dumps({'status': 'ok'}).encode())
          return

        if self.path == '/info':
          body = json.loads(post_data)
          status = get_job_status(body['name'])
          self.send_response(200)
          self.send_header('Content-type', 'application/json')
          self.send_header('Access-Control-Allow-Origin', '*')
          self.send_header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
          self.send_header("Access-Control-Allow-Headers", "Content-type")
          self.end_headers()
          self.wfile.write(json.dumps(status).encode())
          return 

def run(server_class=NoLoggingHTTPServer, handler_class=SimpleHTTPRequestHandler, port=8080):
  
    server_address = ('', port)
    httpd = server_class(server_address, handler_class)
    print(f'Starting httpd server on port {port}...')
    httpd.serve_forever()


if __name__ == "__main__":
    logging.root.setLevel(logging.CRITICAL)
    logging.root.handlers = [logging.NullHandler()]
    run()