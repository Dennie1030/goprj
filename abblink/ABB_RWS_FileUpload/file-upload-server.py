import os
import base64
import requests
from requests.auth import HTTPBasicAuth
import urllib3
import socket
import json
import threading
import tkinter as tk
from tkinter import ttk, messagebox
from datetime import datetime

# 禁用SSL警告
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

class FileUploadServer:
    def __init__(self):
        # 创建主窗口
        self.window = tk.Tk()
        self.window.title("File Upload Server")
        self.window.geometry("800x600")
        
        # TCP服务器设置
        self.server_socket = None
        self.is_running = False
        self.clients = []
        
        # 上传服务器设置
        self.base_url = "https://192.168.125.1/fileservice/$home/"
        self.username = "Default User"
        self.password = "robotics"

        # 创建GUI组件
        self.create_gui()

    def create_gui(self):
        # 服务器控制框架
        server_frame = ttk.LabelFrame(self.window, text="Server Control", padding="5")
        server_frame.pack(fill="x", padx=5, pady=5)

        # TCP服务器设置
        ttk.Label(server_frame, text="TCP Port:").grid(row=0, column=0, padx=5, pady=5)
        self.port_entry = ttk.Entry(server_frame, width=10)
        self.port_entry.insert(0, "12345")
        self.port_entry.grid(row=0, column=1, padx=5, pady=5)

        # 上传服务器URL设置
        ttk.Label(server_frame, text="Base URL:").grid(row=0, column=2, padx=5, pady=5)
        self.url_entry = ttk.Entry(server_frame, width=50)
        self.url_entry.insert(0, self.base_url)
        self.url_entry.grid(row=0, column=3, padx=5, pady=5)

        # 启动/停止按钮
        self.server_button = ttk.Button(server_frame, text="Start Server", command=self.toggle_server)
        self.server_button.grid(row=0, column=4, padx=5, pady=5)

        # 日志显示区域
        log_frame = ttk.LabelFrame(self.window, text="Server Log", padding="5")
        log_frame.pack(fill="both", expand=True, padx=5, pady=5)

        # 创建日志文本框和滚动条
        self.log_text = tk.Text(log_frame, height=20, width=80)
        scrollbar = ttk.Scrollbar(log_frame, orient="vertical", command=self.log_text.yview)
        self.log_text.configure(yscrollcommand=scrollbar.set)
        
        self.log_text.pack(side="left", fill="both", expand=True)
        scrollbar.pack(side="right", fill="y")

    def log_message(self, message):
        """添加日志消息"""
        timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        self.log_text.insert("end", f"[{timestamp}] {message}\n")
        self.log_text.see("end")

    def toggle_server(self):
        """启动/停止服务器"""
        if not self.is_running:
            try:
                port = int(self.port_entry.get())
                self.base_url = self.url_entry.get()
                
                if not self.base_url:
                    self.base_url = "https://192.168.125.1/fileservice/$home/"
                    self.url_entry.delete(0, tk.END)
                    self.url_entry.insert(0, self.base_url)
                
                # 确保URL以/结尾
                if not self.base_url.endswith('/'):
                    self.base_url += '/'
                    self.url_entry.delete(0, tk.END)
                    self.url_entry.insert(0, self.base_url)
                
                self.start_server(port)
                self.server_button.configure(text="Stop Server")
                self.log_message(f"Server started on port {port}")
                self.log_message(f"Using base URL: {self.base_url}")
            except ValueError:
                messagebox.showerror("Error", "Invalid port number")
        else:
            self.stop_server()
            self.server_button.configure(text="Start Server")
            self.log_message("Server stopped")

    def start_server(self, port):
        """启动TCP服务器"""
        self.server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.server_socket.bind(('0.0.0.0', port))
        self.server_socket.listen(5)
        self.is_running = True
        
        # 在新线程中接受连接
        threading.Thread(target=self.accept_connections, daemon=True).start()

    def stop_server(self):
        """停止TCP服务器"""
        self.is_running = False
        if self.server_socket:
            self.server_socket.close()
        
        # 关闭所有客户端连接
        for client in self.clients:
            try:
                client.close()
            except:
                pass
        self.clients.clear()

    def accept_connections(self):
        """接受客户端连接"""
        while self.is_running:
            try:
                client_socket, address = self.server_socket.accept()
                self.clients.append(client_socket)
                self.log_message(f"New connection from {address}")
                
                # 为每个客户端创建一个处理线程
                threading.Thread(target=self.handle_client, 
                               args=(client_socket, address), 
                               daemon=True).start()
            except:
                if self.is_running:
                    self.log_message("Error accepting connection")

    def handle_client(self, client_socket, address):
        """处理客户端连接"""
        try:
            while self.is_running:
                # 首先接收消息长度（4字节）
                length_data = client_socket.recv(4)
                if not length_data:
                    break
                    
                msg_length = int.from_bytes(length_data, byteorder='big')
                
                # 接收完整消息
                data = b''
                while len(data) < msg_length:
                    chunk = client_socket.recv(min(msg_length - len(data), 1024))
                    if not chunk:
                        break
                    data += chunk
                
                if not data:
                    break
                
                try:
                    command = json.loads(data.decode('utf-8'))
                    self.log_message(f"Received command from {address}: {command}")
                    
                    if command['action'] == 'upload':
                        filename = command['filename']
                        file_path = os.path.join(os.path.dirname(__file__), filename)
                        
                        if not os.path.exists(file_path):
                            response = {"status": "error", "message": f"File not found: {filename}"}
                        else:
                            # 执行上传
                            self.log_message(f"Starting upload for {filename}")
                            upload_result = self.upload_file(file_path)
                            response = {
                                "status": "success" if upload_result else "error",
                                "message": "Upload successful" if upload_result else "Upload failed"
                            }
                        
                        # 发送响应
                        response_json = json.dumps(response)
                        response_bytes = response_json.encode('utf-8')
                        # 首先发送响应长度
                        client_socket.send(len(response_bytes).to_bytes(4, byteorder='big'))
                        # 然后发送响应内容
                        client_socket.send(response_bytes)
                        
                        self.log_message(f"Processed upload request for {filename} from {address}")
                    
                except json.JSONDecodeError:
                    self.log_message(f"Invalid command from {address}")
                    response = {"status": "error", "message": "Invalid command"}
                    response_json = json.dumps(response)
                    response_bytes = response_json.encode('utf-8')
                    client_socket.send(len(response_bytes).to_bytes(4, byteorder='big'))
                    client_socket.send(response_bytes)
        
        except Exception as e:
            self.log_message(f"Error handling client {address}: {str(e)}")
        
        finally:
            if client_socket in self.clients:
                self.clients.remove(client_socket)
            client_socket.close()
            self.log_message(f"Connection closed from {address}")

    def upload_file(self, file_path):
        """上传文件到服务器"""
        try:
            filename = os.path.basename(file_path)
            
            headers = {
                'Accept': 'application/hal+json; v=2.0',
                'Content-Type': 'application/octet-stream; v=2.0'
            }
            
            # 构建完整URL
            url = f"{self.base_url}{filename}"
            
            self.log_message(f"Uploading to URL: {url}")
            
            # 读取文件内容
            with open(file_path, 'rb') as file:
                file_data = file.read()
            
            # 发送PUT请求
            response = requests.put(
                url,
                data=file_data,
                headers=headers,
                auth=HTTPBasicAuth(self.username, self.password),
                verify=False
            )
            
            success = response.ok
            self.log_message(f"File upload {'successful' if success else 'failed'}: {filename}")
            if not success:
                self.log_message(f"Upload failed with status code: {response.status_code}")
                self.log_message(f"Response text: {response.text}")
            
            return success
            
        except Exception as e:
            self.log_message(f"Upload error: {str(e)}")
            if hasattr(e, 'response'):
                self.log_message(f"Response status code: {e.response.status_code}")
                self.log_message(f"Response text: {e.response.text}")
            return False

    def run(self):
        """运行应用"""
        self.window.mainloop()

if __name__ == "__main__":
    app = FileUploadServer()
    app.run()
