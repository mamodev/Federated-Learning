import socket
import json

class Channel():
    def __init__ (self, socket):
        self.socket = socket

    @staticmethod 
    def connect (host, port) -> 'Channel':
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.connect((host, port))
        return Channel(sock)
    
    def send (self, data):
        if isinstance(data, dict):
            data = json.dumps(data).encode('utf-8')

        self.socket.sendall(len(data).to_bytes(4, byteorder='big'))
        self.socket.sendall(data)

    def receive (self) -> bytes:
        size = int.from_bytes(self.__recv_all(4), byteorder='big')
        return self.__recv_all(size)

    def receive_json (self) -> dict:
        return json.loads(self.receive().decode('utf-8'))

    def receive_command (self) -> str:
        return self.receive().decode('ascii')

    def send_command (self, command):
        self.send(command.encode('ascii'))

    def send_ack (self):
        self.send('ACK'.encode('ascii'))

    def receive_ack (self):
        # buff = str(self.receive())
        buff = self.receive().decode('ascii')
        if buff != 'ACK':
            raise Exception(f'Expected ACK, but received {buff}')

    def __recv_all (self, size):
        data = b''
        while len(data) < size:
            data += self.socket.recv(size - len(data))
        return data
    
    def close (self):
        try:
          self.send_command('CLOSE')
        except:
          pass            

        self.socket.close()

    def __enter__(self):
      return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()