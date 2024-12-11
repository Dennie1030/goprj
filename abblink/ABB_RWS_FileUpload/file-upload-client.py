import socket
import json
import sys
import time

class FileUploadClient:
    def __init__(self, host, port):
        self.host = host
        self.port = port
        self.socket = None
        self.timeout = 60  # 设置超时时间为60秒

    def connect(self):
        """连接到服务器"""
        try:
            self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.socket.settimeout(self.timeout)  # 设置socket超时
            self.socket.connect((self.host, self.port))
            return True
        except Exception as e:
            print(f"Connection error: {str(e)}")
            return False

    def receive_all(self, size):
        """完整接收指定大小的数据"""
        data = b''
        while len(data) < size:
            packet = self.socket.recv(size - len(data))
            if not packet:
                return None
            data += packet
        return data

    def upload_file(self, filename):
        """发送上传命令"""
        try:
            command = {
                "action": "upload",
                "filename": filename
            }
            
            # 发送命令
            command_json = json.dumps(command)
            msg_length = len(command_json)
            # 首先发送消息长度（4字节）
            self.socket.send(msg_length.to_bytes(4, byteorder='big'))
            # 然后发送实际消息
            self.socket.send(command_json.encode('utf-8'))
            
            print("Waiting for server response...")
            
            # 首先接收响应长度（4字节）
            length_data = self.receive_all(4)
            if length_data is None:
                return {"status": "error", "message": "Connection closed by server"}
            
            response_length = int.from_bytes(length_data, byteorder='big')
            
            # 然后接收完整响应
            response_data = self.receive_all(response_length)
            if response_data is None:
                return {"status": "error", "message": "Connection closed by server"}
            
            response = json.loads(response_data.decode('utf-8'))
            return response
            
        except socket.timeout:
            return {"status": "error", "message": "Operation timed out"}
        except Exception as e:
            print(f"Error: {str(e)}")
            return {"status": "error", "message": str(e)}

    def close(self):
        """关闭连接"""
        if self.socket:
            self.socket.close()

def main():
    if len(sys.argv) != 4:
        print("Usage: python client.py <server_ip> <port> <filename>")
        return

    host = sys.argv[1]
    port = int(sys.argv[2])
    filename = sys.argv[3]

    client = FileUploadClient(host, port)
    
    if client.connect():
        print("Connected to server")
        print("Sending upload request...")
        result = client.upload_file(filename)
        print(f"Upload result: {result}")
        client.close()

if __name__ == "__main__":
    main()
