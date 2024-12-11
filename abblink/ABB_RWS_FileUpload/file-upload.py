import os
import base64
import requests
from requests.auth import HTTPBasicAuth
import urllib3
from tkinter import Tk, Button, Entry, Label, messagebox
from tkinter.filedialog import askopenfilename

# 禁用SSL警告
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

class FileUploadApp:
    def __init__(self):
        self.window = Tk()
        self.window.title("File Upload")
        self.window.geometry("600x200")
        
        self.selected_file_path = ""
        
        # 定义默认URL模板
        self.default_url = "https://192.168.125.1/fileservice/$home/"
        
        # 创建URL输入框并设置默认值
        Label(self.window, text="URL:").grid(row=0, column=0, padx=5, pady=5)
        self.url_entry = Entry(self.window, width=50)
        self.url_entry.grid(row=0, column=1, columnspan=2, padx=5, pady=5)
        self.url_entry.insert(0, self.default_url)  # 设置默认URL
        
        # 创建文件路径显示框
        Label(self.window, text="File:").grid(row=1, column=0, padx=5, pady=5)
        self.file_entry = Entry(self.window, width=50)
        self.file_entry.grid(row=1, column=1, padx=5, pady=5)
        
        # 创建选择文件按钮
        self.select_button = Button(self.window, text="Select File", command=self.select_file)
        self.select_button.grid(row=1, column=2, padx=5, pady=5)
        
        # 创建上传按钮
        self.upload_button = Button(self.window, text="Upload", command=self.upload_file)
        self.upload_button.grid(row=2, column=1, padx=5, pady=5)

    def select_file(self):
        """选择文件"""
        filename = askopenfilename(title="Select a file to upload")
        if filename:
            self.selected_file_path = filename
            self.file_entry.delete(0, 'end')
            self.file_entry.insert(0, filename)

    def upload_file(self):
        """上传文件"""
        if not self.selected_file_path:
            messagebox.showwarning("Warning", "Please select a file first.")
            return
        
        try:
            # 获取文件名
            filename = os.path.basename(self.selected_file_path)
            
            # 设置认证信息
            username = "Default User"
            password = "robotics"
            
            # 设置请求头
            headers = {
                'Accept': 'application/hal+json; v=2.0',
                'Content-Type': 'application/octet-stream; v=2.0'
            }
            
            # 获取当前URL并确保以/结尾
            base_url = self.url_entry.get().rstrip('/') + '/'
            
            # 构建完整URL
            url = f"{base_url}{filename}"
            
            # 读取文件内容
            with open(self.selected_file_path, 'rb') as file:
                file_data = file.read()
            
            # 发送PUT请求
            response = requests.put(
                url,
                data=file_data,
                headers=headers,
                auth=HTTPBasicAuth(username, password),
                verify=False  # 禁用SSL验证
            )
            
            if response.ok:
                messagebox.showinfo("Success", f"{filename} 上传成功.")
            else:
                messagebox.showerror("Error", f"Failed to upload file: {response.status_code}")
                
        except Exception as e:
            error_message = f"An error occurred: {str(e)}"
            if hasattr(e, '__context__') and e.__context__:
                error_message += f"\n{str(e.__context__)}"
            messagebox.showerror("Error", error_message)

    def run(self):
        """运行应用"""
        self.window.mainloop()

if __name__ == "__main__":
    app = FileUploadApp()
    app.run()